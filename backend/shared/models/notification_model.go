package models

import (
	"context"
	"time"

	"backend/shared/mmodels"
	"github.com/jmoiron/sqlx"
)

// Notification is retained as the API and service DTO. NotificationMModel owns persistence.
type Notification struct {
	BaseModel

	ID              int64      `json:"id" db:"id"`
	AccountID       int64      `json:"accountId" db:"account_id"`
	RecipientUserID *int64     `json:"recipientUserId,omitempty" db:"recipient_user_id"`
	RecipientEmail  string     `json:"recipientEmail" db:"recipient_email"`
	Title           string     `json:"title" db:"title"`
	Message         string     `json:"message" db:"message"`
	EntityType      string     `json:"entityType" db:"entity_type"`
	EntityID        int64      `json:"entityId" db:"entity_id"`
	EntityNumber    string     `json:"entityNumber" db:"entity_number"`
	Action          string     `json:"action" db:"action"`
	IsRead          bool       `json:"isRead" db:"is_read"`
	ReadAt          *time.Time `json:"readAt,omitempty" db:"read_at"`
}

func NewNotification(db *sqlx.DB, auth Auth) Notification {
	item := Notification{BaseModel: BaseModel{DB: db, Auth: auth}, AccountID: auth.AccountID}
	item.CreatedBy = auth.Email
	return item
}

func (m Notification) mmodel() *mmodels.NotificationMModel {
	ctx := m.Context
	if ctx == nil {
		ctx = context.Background()
	}
	item := mmodels.NewNotificationMModel(ctx, m.DB)
	item.ID = m.ID
	item.AccountID = m.AccountID
	item.RecipientUserID = m.RecipientUserID
	item.RecipientEmail = m.RecipientEmail
	item.Title = m.Title
	item.Message = m.Message
	item.EntityType = m.EntityType
	item.EntityID = m.EntityID
	item.EntityNumber = m.EntityNumber
	item.Action = m.Action
	item.IsRead = m.IsRead
	item.ReadAt = m.ReadAt
	item.CreatedAt = m.CreatedAt
	item.CreatedBy = m.CreatedBy
	if m.UpdatedAt != nil {
		item.UpdatedAt.Valid = true
		item.UpdatedAt.Time = *m.UpdatedAt
	}
	if m.UpdatedBy != nil {
		item.UpdatedBy.Valid = true
		item.UpdatedBy.String = *m.UpdatedBy
	}
	item.AuthAccountID = m.Auth.AccountID
	item.AuthEmail = m.Auth.Email
	item.AuthUserName = m.Auth.UserName
	return item
}

func notificationFromMModel(item *mmodels.NotificationMModel, base BaseModel) *Notification {
	if item == nil {
		return nil
	}
	result := &Notification{
		BaseModel: base,
		ID:        item.ID, AccountID: item.AccountID,
		RecipientUserID: item.RecipientUserID, RecipientEmail: item.RecipientEmail,
		Title: item.Title, Message: item.Message, EntityType: item.EntityType,
		EntityID: item.EntityID, EntityNumber: item.EntityNumber, Action: item.Action,
		IsRead: item.IsRead, ReadAt: item.ReadAt,
	}
	result.CreatedAt, result.CreatedBy = item.CreatedAt, item.CreatedBy
	if item.UpdatedAt.Valid {
		updatedAt := item.UpdatedAt.Time
		result.UpdatedAt = &updatedAt
	}
	if item.UpdatedBy.Valid {
		updatedBy := item.UpdatedBy.String
		result.UpdatedBy = &updatedBy
	}
	return result
}

func (m Notification) Migrate() {
	if err := m.MigrateError(); err != nil {
		panic(err)
	}
}

func (m Notification) MigrateError() error { return m.mmodel().Migrate() }

func (m Notification) Create() (int64, error) { return m.mmodel().Create() }

func (m Notification) FindList(recipientUserID int64, recipientEmail, search string, limit int) ([]Notification, error) {
	items, err := m.mmodel().FindList(recipientUserID, recipientEmail, search, limit)
	if err != nil {
		return nil, err
	}
	result := make([]Notification, 0, len(items))
	for i := range items {
		result = append(result, *notificationFromMModel(&items[i], m.BaseModel))
	}
	return result, nil
}

func (m Notification) CountUnread(recipientUserID int64, recipientEmail string) (int64, error) {
	return m.mmodel().CountUnread(recipientUserID, recipientEmail)
}

func (m Notification) MarkRead(id int64) error { return m.mmodel().MarkRead(id) }

func (m Notification) Normalize() {
	item := m.mmodel()
	item.Normalize()
}
