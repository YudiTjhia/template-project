package services

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"

	"backend/shared/helpers"
	"backend/shared/mmodels"
	"backend/shared/models"

	"github.com/jmoiron/sqlx"
)

type NotificationService struct {
	DB   *sqlx.DB
	Auth models.Auth
}

type NotificationEvent struct {
	EntityType   string
	EntityID     int64
	EntityNumber string
	Action       string
	Title        string
	Message      string
	Actor        string
}

type notificationRecipient struct {
	UserID *int64
	Email  string
	Name   string
}

func NewNotificationService(db *sqlx.DB, auth models.Auth) *NotificationService {
	return &NotificationService{DB: db, Auth: auth}
}

func (s *NotificationService) Migrate() {
	if err := s.newNotificationMModel().Migrate(); err != nil {
		panic(err)
	}
}

func (s *NotificationService) newNotificationMModel() *mmodels.NotificationMModel {
	model := mmodels.NewNotificationMModel(context.Background(), s.DB)
	model.AccountID = s.Auth.AccountID
	model.AuthAccountID = s.Auth.AccountID
	model.AuthEmail = s.Auth.Email
	model.AuthUserName = s.Auth.UserName
	return model
}

func notificationMModelToDTO(item *mmodels.NotificationMModel, db *sqlx.DB, auth models.Auth) models.Notification {
	dto := models.Notification{
		BaseModel: models.BaseModel{DB: db, Auth: auth, CreatedAt: item.CreatedAt, CreatedBy: item.CreatedBy},
		ID:        item.ID, AccountID: item.AccountID, RecipientUserID: item.RecipientUserID,
		RecipientEmail: item.RecipientEmail, Title: item.Title, Message: item.Message,
		EntityType: item.EntityType, EntityID: item.EntityID, EntityNumber: item.EntityNumber,
		Action: item.Action, IsRead: item.IsRead, ReadAt: item.ReadAt,
	}
	if item.UpdatedAt.Valid {
		updatedAt := item.UpdatedAt.Time
		dto.UpdatedAt = &updatedAt
	}
	if item.UpdatedBy.Valid {
		updatedBy := item.UpdatedBy.String
		dto.UpdatedBy = &updatedBy
	}
	return dto
}

func (s *NotificationService) NotifyAccountUsers(event NotificationEvent) {
	event = normalizeNotificationEvent(event)
	if event.Title == "" || event.Message == "" {
		return
	}

	recipients, err := s.findActiveAccountRecipients()
	if err != nil {
		log.Printf("[WARN] notification recipient lookup failed: %v", err)
		return
	}
	s.notifyRecipients(event, recipients)
}

func (s *NotificationService) FindList(search string, limit int) ([]models.Notification, error) {
	userID, email := s.currentRecipientIdentity()
	items, err := s.newNotificationMModel().FindList(userID, email, search, limit)
	if err != nil {
		return nil, err
	}
	result := make([]models.Notification, 0, len(items))
	for i := range items {
		result = append(result, notificationMModelToDTO(&items[i], s.DB, s.Auth))
	}
	return result, nil
}

func (s *NotificationService) CountUnread() (int64, error) {
	userID, email := s.currentRecipientIdentity()
	return s.newNotificationMModel().CountUnread(userID, email)
}

func (s *NotificationService) MarkRead(id int64) error {
	return s.newNotificationMModel().MarkRead(id)
}

func (s *NotificationService) notifyRecipients(event NotificationEvent, recipients []notificationRecipient) {
	if len(recipients) == 0 {
		return
	}

	seen := map[string]struct{}{}
	for _, recipient := range recipients {
		recipient.Email = strings.ToLower(strings.TrimSpace(recipient.Email))
		if recipient.UserID == nil && recipient.Email == "" {
			continue
		}
		key := recipient.Email
		if recipient.UserID != nil {
			key = fmt.Sprintf("id:%d", *recipient.UserID)
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		model := s.newNotificationMModel()
		model.RecipientUserID = recipient.UserID
		model.RecipientEmail = recipient.Email
		model.Title = event.Title
		model.Message = event.Message
		model.EntityType = event.EntityType
		model.EntityID = event.EntityID
		model.EntityNumber = event.EntityNumber
		model.Action = event.Action
		if _, err := model.Create(); err != nil {
			log.Printf("[WARN] create notification failed: %v", err)
			continue
		}

		if recipient.Email != "" {
			s.sendNotificationEmailAsync(recipient, event)
		}
	}
}

func (s *NotificationService) findActiveAccountRecipients() ([]notificationRecipient, error) {
	userModel := mmodels.NewUserMModel(context.Background(), s.DB)
	rows, err := userModel.FindActiveByAccount_(s.Auth.AccountID)
	if err != nil {
		log.Printf("[WARN] active user notification lookup failed, fallback to auths: %v", err)
	}

	recipients := make([]notificationRecipient, 0, len(rows))
	for _, row := range rows {
		id := row.ID
		recipients = append(recipients, notificationRecipient{
			UserID: &id,
			Email:  row.UserEmail,
			Name:   row.Username,
		})
	}

	authRecipients, err := s.findAuthRecipients()
	if err != nil {
		log.Printf("[WARN] auth notification recipient lookup failed: %v", err)
	}
	recipients = append(recipients, authRecipients...)

	if len(recipients) == 0 && strings.TrimSpace(s.Auth.Email) != "" {
		recipients = append(recipients, notificationRecipient{Email: s.Auth.Email, Name: s.Auth.UserName})
	}
	return recipients, nil
}

func (s *NotificationService) findAuthRecipients() ([]notificationRecipient, error) {
	var rows []struct {
		Email string `db:"email"`
		Name  string `db:"name"`
	}
	err := s.DB.Select(&rows, `SELECT email, name FROM `+mmodels.AuthTable+` WHERE account_id = $1 AND is_expired = false ORDER BY id ASC`, s.Auth.AccountID)
	if err != nil {
		return nil, err
	}
	recipients := make([]notificationRecipient, 0, len(rows))
	for _, row := range rows {
		recipients = append(recipients, notificationRecipient{Email: row.Email, Name: row.Name})
	}
	return recipients, nil
}

func (s *NotificationService) currentRecipientIdentity() (int64, string) {
	userModel := mmodels.NewUserMModel(context.Background(), s.DB)
	userModel.AccountID = s.Auth.AccountID
	userModel.AuthAccountID = s.Auth.AccountID
	if strings.TrimSpace(s.Auth.Email) != "" {
		if user, err := userModel.FindByEmail(s.Auth.Email); err == nil && user != nil {
			return user.ID, user.UserEmail
		}
	}
	if strings.TrimSpace(s.Auth.UserName) != "" {
		if user, err := userModel.FindByUsername(s.Auth.UserName); err == nil && user != nil {
			return user.ID, user.UserEmail
		}
	}
	return 0, strings.TrimSpace(s.Auth.Email)
}

func (s *NotificationService) sendNotificationEmailAsync(recipient notificationRecipient, event NotificationEvent) {
	to := strings.TrimSpace(recipient.Email)
	if to == "" {
		return
	}

	helpers.SendEmailInBackground("transaction notification email", func() error {
		cfg := helpers.LoadSMTPConfig()
		if !cfg.Enabled() {
			return nil
		}

		htmlBody, err := helpers.RenderTransactionNotificationEmail(s.transactionEmailData(recipient, event))
		if err != nil {
			log.Printf("[WARN] render transaction email template failed: %v", err)
			message := helpers.WriteEmailMessage(cfg.FromHeader(), to, event.Title, event.Message)
			var auth smtp.Auth
			if cfg.Username != "" {
				auth = cfg.Auth()
			}
			return smtp.SendMail(cfg.Addr(), auth, cfg.FromEmail, []string{to}, message)
		}

		message := helpers.WriteHTMLMultipartEmailMessage(cfg.FromHeader(), to, event.Title, event.Message, htmlBody)
		var auth smtp.Auth
		if cfg.Username != "" {
			auth = cfg.Auth()
		}
		return smtp.SendMail(cfg.Addr(), auth, cfg.FromEmail, []string{to}, message)
	})
}

func (s *NotificationService) transactionEmailData(recipient notificationRecipient, event NotificationEvent) helpers.TransactionNotificationEmailData {
	entityLabel := notificationEntityLabel(event.EntityType)
	actionLabel := notificationActionLabel(event.Action)
	actor := strings.TrimSpace(event.Actor)
	if actor == "" {
		actor = strings.TrimSpace(s.Auth.Email)
	}
	if actor == "" {
		actor = strings.TrimSpace(s.Auth.UserName)
	}
	if actor == "" {
		actor = "system"
	}

	fields := []helpers.EmailTemplateField{
		{Label: "Tipe Dokumen", Value: fallbackEmailValue(entityLabel)},
		{Label: "Nomor Dokumen", Value: fallbackEmailValue(event.EntityNumber)},
		{Label: "Action", Value: fallbackEmailValue(actionLabel)},
		{Label: "Actor", Value: fallbackEmailValue(actor)},
		{Label: "Recipient", Value: fallbackEmailValue(recipient.Email)},
	}

	systemURL := notificationSystemURL(event.EntityType)
	documentURL := notificationDocumentURL(event.EntityType, event.EntityID)
	if strings.EqualFold(strings.TrimSpace(event.Action), "DELETE") {
		documentURL = ""
	}

	return helpers.TransactionNotificationEmailData{
		Subject:        event.Title,
		Title:          event.Title,
		Message:        event.Message,
		RecipientName:  recipient.Name,
		RecipientEmail: recipient.Email,
		EntityType:     event.EntityType,
		EntityLabel:    entityLabel,
		EntityNumber:   event.EntityNumber,
		Action:         event.Action,
		ActionLabel:    actionLabel,
		Actor:          actor,
		SystemURL:      systemURL,
		DocumentURL:    documentURL,
		Fields:         fields,
	}
}

func normalizeNotificationEvent(event NotificationEvent) NotificationEvent {
	event.EntityType = strings.ToUpper(strings.TrimSpace(event.EntityType))
	event.EntityNumber = strings.TrimSpace(event.EntityNumber)
	event.Action = strings.ToUpper(strings.TrimSpace(event.Action))
	event.Title = strings.TrimSpace(event.Title)
	event.Message = strings.TrimSpace(event.Message)
	event.Actor = strings.TrimSpace(event.Actor)
	return event
}

func notificationEntityLabel(entityType string) string {
	switch strings.ToUpper(strings.TrimSpace(entityType)) {
	case models.ApprovalEntityBudgetControl:
		return "Budget Control"
	case models.ApprovalEntityBudgetAdjustment:
		return "Budget Adjustment"
	default:
		return strings.Title(strings.ToLower(strings.ReplaceAll(strings.TrimSpace(entityType), "_", " ")))
	}
}

func notificationActionLabel(action string) string {
	switch strings.ToUpper(strings.TrimSpace(action)) {
	case "CREATE":
		return "Created"
	case "EDIT", "UPDATE":
		return "Updated"
	case "DELETE":
		return "Deleted"
	case "SUBMIT_APPROVAL":
		return "Submitted for Approval"
	case "APPROVAL_APPROVED":
		return "Approval Approved"
	case "APPROVAL_REJECTED":
		return "Approval Rejected"
	case "APPROVAL_REVISION":
		return "Approval Revision Requested"
	case "APPROVAL_PENDING":
		return "Approval Pending"
	default:
		value := strings.TrimSpace(strings.ReplaceAll(action, "_", " "))
		if value == "" {
			return "-"
		}
		return strings.Title(strings.ToLower(value))
	}
}

func notificationDocumentURL(entityType string, entityID int64) string {
	if entityID <= 0 {
		return ""
	}

	systemURL := notificationSystemURL(entityType)
	if systemURL == "" {
		return ""
	}
	return strings.TrimRight(systemURL, "/") + "/print/" + fmt.Sprintf("%d", entityID)
}

func notificationSystemURL(entityType string) string {
	baseURL := strings.TrimSpace(os.Getenv("BUYER_FRONTEND_BASE_URL"))
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv("FRONTEND_BASE_URL"))
	}
	if baseURL == "" {
		return ""
	}

	path := notificationDocumentPath(entityType)
	if path == "" {
		return ""
	}
	return strings.TrimRight(baseURL, "/") + "/" + path
}

func notificationDocumentPath(entityType string) string {
	switch strings.ToUpper(strings.TrimSpace(entityType)) {
	case models.ApprovalEntityBudgetControl, models.ApprovalEntityBudgetAdjustment:
		return "budget-controls"
	default:
		return ""
	}
}

func fallbackEmailValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	return value
}
