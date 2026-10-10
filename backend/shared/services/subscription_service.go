package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"backend/shared/mmodels"
	"backend/shared/models"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type SubscriptionService struct {
	DB            *sqlx.DB
	TransactionDB *sqlx.DB
	Auth          models.Auth
}

type SubscriptionPaymentUserOption struct {
	ID        int64  `json:"id" db:"id"`
	Username  string `json:"username" db:"username"`
	UserEmail string `json:"userEmail" db:"user_email"`
	IsActive  bool   `json:"isActive" db:"is_active"`
}

type SubscriptionPaymentInfo struct {
	Bank         models.SubscriptionBankInfo     `json:"bank"`
	PricePerUser float64                         `json:"pricePerUser"`
	Currency     string                          `json:"currency"`
	Plans        []models.SubscriptionPlan       `json:"plans"`
	Users        []SubscriptionPaymentUserOption `json:"users"`
	Current      *models.Subscription            `json:"current,omitempty"`
}

func NewSubscriptionService(db *sqlx.DB, auth models.Auth) *SubscriptionService {
	return NewSubscriptionServiceWithDatabases(db, db, auth)
}

func NewSubscriptionServiceWithDatabases(db, transactionDB *sqlx.DB, auth models.Auth) *SubscriptionService {
	if transactionDB == nil {
		transactionDB = db
	}
	return &SubscriptionService{DB: db, TransactionDB: transactionDB, Auth: auth}
}

func (s *SubscriptionService) NewSubscription() models.Subscription {
	return models.NewSubscription(s.DB, s.Auth)
}

func (s *SubscriptionService) NewSubscriptionHistory() models.SubscriptionHistory {
	return models.NewSubscriptionHistory(s.TransactionDB, s.Auth)
}

func (s *SubscriptionService) NewSubscriptionPayment() models.SubscriptionPayment {
	return models.NewSubscriptionPayment(s.TransactionDB, s.Auth)
}

func (s *SubscriptionService) Migrate() error {
	if s == nil || s.DB == nil {
		return errors.New("subscription service database is nil")
	}
	planModel := mmodels.NewSubscriptionPlanMModel(context.Background(), s.DB)
	if err := planModel.Migrate(); err != nil {
		return err
	}
	if err := s.seedDefaultPlans(); err != nil {
		return err
	}
	if err := s.NewSubscription().Migrate(); err != nil {
		return err
	}
	s.NewSubscriptionHistory().Migrate()
	s.NewSubscriptionPayment().Migrate()
	return nil
}

func (s *SubscriptionService) seedDefaultPlans() error {
	for _, plan := range models.SubscriptionPlansFromEnv() {
		lookup := mmodels.NewSubscriptionPlanMModel(context.Background(), s.DB)
		if _, err := lookup.FindByCode(plan.Code); err == nil {
			continue
		} else if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		item := mmodels.NewSubscriptionPlanMModel(context.Background(), s.DB)
		item.Code = plan.Code
		item.Name = plan.Name
		item.Price = plan.Price
		item.Currency = plan.Currency
		item.Description = plan.Description
		item.Limits = mmodels.SubscriptionPlanLimits(plan.Limits)
		item.Active = plan.Active
		item.SortOrder = plan.SortOrder
		item.CreatedBy = "system"
		if _, err := item.Create(); err != nil {
			return fmt.Errorf("seed subscription plan %s: %w", plan.Code, err)
		}
	}
	return nil
}

func (s *SubscriptionService) findPlanByCode(code string) (models.SubscriptionPlan, error) {
	return models.SubscriptionPlanByCodeFromDB(s.DB, code)
}

func (s *SubscriptionService) serviceForAccount(accountID int64) *SubscriptionService {
	if accountID <= 0 {
		accountID = s.Auth.AccountID
	}

	return NewSubscriptionServiceWithDatabases(s.DB, s.TransactionDB, models.Auth{
		AccountID: accountID,
		UserName:  s.Auth.UserName,
		Email:     s.Auth.Email,
	})
}

func (s *SubscriptionService) Create(subscription models.Subscription) (int64, error) {
	if current, err := s.FindCurrent(); err == nil && strings.EqualFold(strings.TrimSpace(current.Status), models.SubscriptionStatusActive) {
		return 0, errors.New("subscription aktif sudah ada, nonaktifkan atau tunggu masa berlaku habis")
	}

	model := s.NewSubscription()
	model.AccountID = subscription.AccountID
	model.AppType = models.NormalizeSubscriptionAppType(subscription.AppType)
	model.PlanCode = subscription.PlanCode
	model.NumberOfUsers = subscription.NumberOfUsers
	plan, err := s.findPlanByCode(model.PlanCode)
	if err != nil {
		return 0, err
	}
	model.PlanCode = plan.Code
	model.NumberOfUsers = plan.Limits.Users
	model.PricePerUser = plan.Price
	model.Amount = models.RoundMoney(float64(model.NumberOfUsers) * model.PricePerUser)
	start, end := models.DefaultSubscriptionWindow()
	model.CurrentStartDate = start
	model.CurrentEndDate = end
	model.PendingNumberOfUsers = subscription.PendingNumberOfUsers
	model.PendingEffectiveDate = normalizeTimePtr(subscription.PendingEffectiveDate)
	model.Status = models.SubscriptionStatusActive
	model.CreatedBy = subscription.CreatedBy

	if model.AccountID <= 0 {
		model.AccountID = s.Auth.AccountID
	}
	if model.NumberOfUsers <= 0 {
		return 0, errors.New("number of users subscription wajib lebih dari 0")
	}

	id, err := model.Create()
	if err != nil {
		return 0, err
	}

	if _, err := s.createHistory(models.SubscriptionHistory{
		AccountID:             model.AccountID,
		SubscriptionID:        id,
		AppType:               model.AppType,
		HistoryType:           models.SubscriptionHistoryTypeCreate,
		PreviousNumberOfUsers: 0,
		NewNumberOfUsers:      model.NumberOfUsers,
		DeltaUsers:            model.NumberOfUsers,
		PricePerUser:          model.PricePerUser,
		AmountBefore:          0,
		AmountChange:          model.Amount,
		AmountAfter:           model.Amount,
		CurrentStartDate:      model.CurrentStartDate,
		CurrentEndDate:        model.CurrentEndDate,
		EffectiveDate:         model.CurrentStartDate,
		Notes:                 "subscription created",
	}); err != nil {
		return id, err
	}

	return id, nil
}

func (s *SubscriptionService) Update(subscription models.Subscription) error {
	current, err := s.FindByID(subscription.ID)
	if err != nil {
		return err
	}

	model := s.NewSubscription()
	model.ID = current.ID
	model.AccountID = current.AccountID
	model.AppType = current.AppType
	if rawAppType := strings.TrimSpace(subscription.AppType); rawAppType != "" {
		model.AppType = models.NormalizeSubscriptionAppType(rawAppType)
	}
	model.NumberOfUsers = subscription.NumberOfUsers
	model.PlanCode = subscription.PlanCode
	if strings.TrimSpace(model.PlanCode) == "" {
		model.PlanCode = current.PlanCode
	}
	plan, err := s.findPlanByCode(model.PlanCode)
	if err != nil {
		return err
	}
	model.NumberOfUsers = plan.Limits.Users
	if model.NumberOfUsers <= 0 {
		model.NumberOfUsers = current.NumberOfUsers
	}
	model.PricePerUser = plan.Price
	model.Amount = models.RoundMoney(float64(model.NumberOfUsers) * model.PricePerUser)
	model.CurrentStartDate = current.CurrentStartDate
	model.CurrentEndDate = current.CurrentEndDate
	model.PendingNumberOfUsers = subscription.PendingNumberOfUsers
	model.PendingEffectiveDate = normalizeTimePtr(subscription.PendingEffectiveDate)
	model.Status = current.Status
	model.UpdatedBy = subscription.UpdatedBy

	if err := model.Update(); err != nil {
		return err
	}

	if _, err := s.createHistory(models.SubscriptionHistory{
		AccountID:             model.AccountID,
		SubscriptionID:        model.ID,
		AppType:               model.AppType,
		HistoryType:           models.SubscriptionHistoryTypeUpdate,
		PreviousNumberOfUsers: current.NumberOfUsers,
		NewNumberOfUsers:      model.NumberOfUsers,
		DeltaUsers:            model.NumberOfUsers - current.NumberOfUsers,
		PricePerUser:          model.PricePerUser,
		AmountBefore:          current.Amount,
		AmountChange:          model.Amount - current.Amount,
		AmountAfter:           model.Amount,
		CurrentStartDate:      model.CurrentStartDate,
		CurrentEndDate:        model.CurrentEndDate,
		EffectiveDate:         models.StartOfDay(time.Now().UTC()),
		Notes:                 "subscription updated",
	}); err != nil {
		return err
	}

	return nil
}

func (s *SubscriptionService) Delete(id int64) error {
	model, err := s.FindByID(id)
	if err != nil {
		return err
	}
	return s.NewSubscription().DeleteByID(model.ID)
}

func (s *SubscriptionService) EnsureDefaultSubscriptionForAccount(accountID int64) (*models.Subscription, error) {
	target := s.serviceForAccount(accountID)

	current, err := target.FindCurrent()
	if err == nil {
		return current, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	start, end := models.DefaultSubscriptionWindow()
	model := target.NewSubscription()
	model.AccountID = accountID
	plan, err := target.findPlanByCode("")
	if err != nil {
		return nil, err
	}
	model.PlanCode = plan.Code
	model.NumberOfUsers = plan.Limits.Users
	model.PricePerUser = plan.Price
	model.Amount = models.RoundMoney(float64(model.NumberOfUsers) * model.PricePerUser)
	model.CurrentStartDate = start
	model.CurrentEndDate = end
	model.Status = models.SubscriptionStatusActive
	model.CreatedBy = s.Auth.UserName
	if model.CreatedBy == "" {
		model.CreatedBy = "system"
	}

	id, err := model.Create()
	if err != nil {
		return nil, err
	}

	created, err := target.FindByID(id)
	if err != nil {
		return nil, err
	}

	if _, err := target.createHistory(models.SubscriptionHistory{
		AccountID:             created.AccountID,
		SubscriptionID:        created.ID,
		AppType:               created.AppType,
		HistoryType:           models.SubscriptionHistoryTypeCreate,
		PreviousNumberOfUsers: 0,
		NewNumberOfUsers:      created.NumberOfUsers,
		DeltaUsers:            created.NumberOfUsers,
		PricePerUser:          created.PricePerUser,
		AmountBefore:          0,
		AmountChange:          created.Amount,
		AmountAfter:           created.Amount,
		CurrentStartDate:      created.CurrentStartDate,
		CurrentEndDate:        created.CurrentEndDate,
		EffectiveDate:         created.CurrentStartDate,
		Notes:                 "default subscription created",
	}); err != nil {
		return nil, err
	}

	return created, nil
}

func configuredSubscriptionPricePerUser() float64 {
	for _, key := range []string{"SUBSCRIPTION_PRICE_PER_USER", "NEXT_PUBLIC_SUBSCRIPTION_PRICE_PER_USER"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			if parsed, err := strconv.ParseFloat(value, 64); err == nil && parsed >= 0 {
				return models.RoundMoney(parsed)
			}
		}
	}
	return 0
}

func (s *SubscriptionService) PaymentInfo() (*SubscriptionPaymentInfo, error) {
	if s.Auth.AccountID <= 0 {
		return nil, errors.New("id akun wajib diisi")
	}

	var users []SubscriptionPaymentUserOption
	query := `SELECT id, username, user_email, is_active
	FROM ` + mmodels.UserTable + `
	WHERE account_id = $1
	ORDER BY is_active DESC, username ASC, id ASC`
	if err := s.DB.Select(&users, query, s.Auth.AccountID); err != nil {
		return nil, err
	}

	current, err := s.FindCurrent()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		current = nil
	}

	plans, err := models.SubscriptionPlansFromDB(s.DB, true)
	if err != nil {
		return nil, err
	}
	bank := models.SubscriptionPaymentBankInfo()
	return &SubscriptionPaymentInfo{
		Bank:         bank,
		PricePerUser: configuredSubscriptionPricePerUser(),
		Currency:     bank.Currency,
		Plans:        plans,
		Users:        users,
		Current:      current,
	}, nil
}

func (s *SubscriptionService) GeneratePaymentToken(request models.SubscriptionPaymentTokenRequest) (*models.SubscriptionPaymentTokenResponse, error) {
	if s.Auth.AccountID <= 0 {
		return nil, errors.New("id akun wajib diisi")
	}

	months := request.Months
	if months <= 0 {
		return nil, errors.New("jumlah bulan pembayaran wajib lebih dari 0")
	}

	userIDs := uniquePositiveInt64(request.UserIDs)
	if len(userIDs) == 0 {
		return nil, errors.New("pilih minimal 1 user yang dibayar")
	}

	if err := s.ensurePaymentUsersBelongToAccount(userIDs); err != nil {
		return nil, err
	}

	proof := request.Proof
	proof.FileName = strings.TrimSpace(proof.FileName)
	proof.ContentType = strings.TrimSpace(proof.ContentType)
	if proof.FileName == "" {
		return nil, errors.New("bukti pembayaran wajib diupload")
	}
	if proof.FileSize <= 0 {
		return nil, errors.New("ukuran bukti pembayaran tidak valid")
	}

	appType := models.NormalizeSubscriptionAppType(request.AppType)
	if appType == "" {
		appType = models.SubscriptionAppTypePOSKasir
	}
	if !models.IsAllowedSubscriptionAppType(appType) {
		return nil, errors.New("app type subscription tidak valid")
	}

	bank := models.SubscriptionPaymentBankInfo()
	plan, err := s.findPlanByCode(request.PlanCode)
	if err != nil {
		return nil, err
	}
	pricePerUser := plan.Price
	payload := models.SubscriptionPaymentTokenPayload{
		Version:       1,
		AccountID:     s.Auth.AccountID,
		AppType:       appType,
		PlanCode:      plan.Code,
		PlanName:      plan.Name,
		UserIDs:       userIDs,
		NumberOfUsers: plan.Limits.Users,
		Months:        months,
		PricePerUser:  pricePerUser,
		TotalAmount:   models.RoundMoney(pricePerUser * float64(months)),
		Currency:      bank.Currency,
		Bank:          bank,
		Proof:         proof,
		IssuedAt:      time.Now().UTC(),
	}

	token, err := models.GenerateSubscriptionPaymentToken(payload)
	if err != nil {
		return nil, err
	}

	return &models.SubscriptionPaymentTokenResponse{
		Token:   token,
		Payload: payload,
	}, nil
}

func (s *SubscriptionService) Pay(request models.SubscriptionPaymentTokenRequest) (*models.Subscription, error) {
	result, err := s.GeneratePaymentToken(request)
	if err != nil {
		return nil, err
	}
	return s.ApplyPaymentToken(result.Token)
}

func (s *SubscriptionService) ApplyPaymentToken(token string) (*models.Subscription, error) {
	payload, err := models.DecodeSubscriptionPaymentToken(token)
	if err != nil {
		return nil, err
	}
	if payload.AccountID != s.Auth.AccountID {
		return nil, errors.New("token subscription tidak sesuai dengan akun ini")
	}
	if payload.Months <= 0 {
		return nil, errors.New("jumlah bulan token tidak valid")
	}
	if payload.NumberOfUsers <= 0 || len(payload.UserIDs) == 0 {
		return nil, errors.New("jumlah user token tidak valid")
	}
	if err := s.ensurePaymentUsersBelongToAccount(payload.UserIDs); err != nil {
		return nil, err
	}

	current, err := s.FindCurrent()
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		current, err = s.EnsureDefaultSubscriptionForAccount(s.Auth.AccountID)
		if err != nil {
			return nil, err
		}
	}

	appType := models.NormalizeSubscriptionAppType(payload.AppType)
	if appType == "" {
		appType = current.AppType
	}
	if !models.IsAllowedSubscriptionAppType(appType) {
		return nil, errors.New("app type token tidak valid")
	}

	today := models.StartOfDay(time.Now().UTC())
	startDate := today
	endBase := today
	if strings.EqualFold(strings.TrimSpace(current.Status), models.SubscriptionStatusActive) && !models.StartOfDay(current.CurrentEndDate).Before(today) {
		startDate = models.StartOfDay(current.CurrentStartDate)
		endBase = models.StartOfDay(current.CurrentEndDate)
	}
	endDate := endBase.AddDate(0, payload.Months, 0)

	pricePerUser := models.RoundMoney(payload.PricePerUser)
	if pricePerUser < 0 {
		return nil, errors.New("harga token tidak valid")
	}
	amount := models.RoundMoney(payload.TotalAmount)
	if amount < 0 {
		return nil, errors.New("amount token tidak valid")
	}

	tx, err := s.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil && tx != nil {
			_ = tx.Rollback()
		}
	}()

	updatedBy := s.Auth.UserName
	if updatedBy == "" {
		updatedBy = "system"
	}

	plan, err := s.findPlanByCode(payload.PlanCode)
	if payload.PlanCode == "" {
		plan, err = s.findPlanByCode(current.PlanCode)
	}
	if err != nil {
		return nil, err
	}
	updateSQL := `UPDATE ` + mmodels.SubscriptionTable + `
	SET app_type = $1,
		plan_code = $2,
		plan_name = $3,
		plan_limits = $4::JSONB,
		number_of_users = $5,
		price_per_user = $6,
		amount = $7,
		current_start_date = $8,
		current_end_date = $9,
		pending_number_of_users = NULL,
		pending_effective_date = NULL,
		status = $10,
		updated_at = NOW(),
		updated_by = $11
	WHERE id = $12 AND account_id = $13`
	planLimitsJSON := "{}"
	if raw, marshalErr := json.Marshal(plan.Limits); marshalErr == nil {
		planLimitsJSON = string(raw)
	}
	if _, err = tx.Exec(updateSQL, appType, plan.Code, plan.Name, planLimitsJSON, plan.Limits.Users, pricePerUser, amount, startDate, endDate, models.SubscriptionStatusActive, updatedBy, current.ID, current.AccountID); err != nil {
		return nil, err
	}

	if _, err = s.createHistoryTx(tx, models.SubscriptionHistory{
		AccountID:             current.AccountID,
		SubscriptionID:        current.ID,
		AppType:               appType,
		HistoryType:           models.SubscriptionHistoryTypePayment,
		PreviousNumberOfUsers: current.NumberOfUsers,
		NewNumberOfUsers:      payload.NumberOfUsers,
		DeltaUsers:            payload.NumberOfUsers - current.NumberOfUsers,
		PricePerUser:          pricePerUser,
		AmountBefore:          current.Amount,
		AmountChange:          amount - current.Amount,
		AmountAfter:           amount,
		CurrentStartDate:      startDate,
		CurrentEndDate:        endDate,
		EffectiveDate:         today,
		Notes:                 "subscription payment token applied",
	}); err != nil {
		return nil, err
	}

	insertPaymentSQL := `INSERT INTO ` + mmodels.SubscriptionPaymentTable + `
	(account_id, subscription_id, app_type, user_ids, number_of_users, months, price_per_user, total_amount, currency,
		payment_token, proof_file_name, proof_file_size, proof_content_type, current_start_date, current_end_date, paid_at, created_by)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,NOW(),$16)`
	paymentArgs := []interface{}{current.AccountID, current.ID, appType, pq.Int64Array(payload.UserIDs), payload.NumberOfUsers, payload.Months, pricePerUser, amount, payload.Currency, strings.TrimSpace(token), strings.TrimSpace(payload.Proof.FileName), payload.Proof.FileSize, strings.TrimSpace(payload.Proof.ContentType), startDate, endDate, updatedBy}
	if s.TransactionDB != s.DB {
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		tx = nil
		_, err = s.TransactionDB.Exec(insertPaymentSQL, paymentArgs...)
	} else {
		_, err = tx.Exec(insertPaymentSQL, paymentArgs...)
		if err == nil {
			err = tx.Commit()
			tx = nil
		}
	}
	if err != nil {
		return nil, err
	}
	return s.FindByID(current.ID)
}

func (s *SubscriptionService) SyncActiveUsersForAccount(accountID int64, notes string) error {
	target := s.serviceForAccount(accountID)

	current, err := target.FindCurrent()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	activeCount, err := target.CountActiveUsers(accountID)
	if err != nil {
		return err
	}
	if activeCount <= 0 || activeCount == current.NumberOfUsers {
		return nil
	}

	_, err = target.AdjustUsers(current.ID, activeCount, notes)
	return err
}

func (s *SubscriptionService) SyncLifecycleForAccount(accountID int64) error {
	if accountID <= 0 {
		accountID = s.Auth.AccountID
	}
	if accountID <= 0 {
		return nil
	}

	_, err := s.DB.Exec(
		`UPDATE `+mmodels.SubscriptionTable+`
		SET status = $1, updated_at = NOW(), updated_by = COALESCE(NULLIF(updated_by, ''), 'system')
		WHERE account_id = $2 AND status = $3 AND current_end_date < CURRENT_DATE`,
		models.SubscriptionStatusExpired,
		accountID,
		models.SubscriptionStatusActive,
	)
	return err
}

func (s *SubscriptionService) FindByID(id int64) (*models.Subscription, error) {
	model := s.NewSubscription()
	model.ID = id
	return model.FindByID()
}

func (s *SubscriptionService) FindCurrent() (*models.Subscription, error) {
	if err := s.SyncLifecycleForAccount(s.Auth.AccountID); err != nil {
		return nil, err
	}
	model := s.NewSubscription()
	return model.FindByAccount()
}

func (s *SubscriptionService) StatusForCurrentAccount() (*models.SubscriptionStatus, error) {
	current, err := s.FindCurrent()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.SubscriptionStatus{
				IsPaid:          false,
				HasSubscription: false,
				Message:         "Subscription POS belum dibuat. Silakan hubungi admin untuk mengaktifkan subscription.",
			}, nil
		}
		return nil, err
	}

	today := models.StartOfDay(time.Now().UTC())
	endDate := models.StartOfDay(current.CurrentEndDate)
	active := strings.EqualFold(strings.TrimSpace(current.Status), models.SubscriptionStatusActive) && !endDate.Before(today)
	message := ""
	if !active {
		switch {
		case endDate.Before(today):
			message = "Subscription POS sudah expired. Silakan perpanjang subscription untuk melanjutkan akses aplikasi."
		case strings.EqualFold(strings.TrimSpace(current.Status), models.SubscriptionStatusSuspended):
			message = "Subscription POS sedang suspended. Silakan hubungi admin untuk mengaktifkan kembali akses aplikasi."
		default:
			message = "Subscription POS belum aktif. Silakan hubungi admin untuk melanjutkan akses aplikasi."
		}
	}

	return &models.SubscriptionStatus{
		IsPaid:          active,
		HasSubscription: true,
		Message:         message,
		Subscription:    current,
	}, nil
}

func (s *SubscriptionService) FindList() ([]models.Subscription, error) {
	current, err := s.FindCurrent()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []models.Subscription{}, nil
		}
		return nil, err
	}
	return []models.Subscription{*current}, nil
}

func (s *SubscriptionService) FindHistories(subscriptionID int64) ([]models.SubscriptionHistory, error) {
	history := s.NewSubscriptionHistory()
	return history.FindList(subscriptionID)
}

func (s *SubscriptionService) FindPayments(search string) ([]models.SubscriptionPayment, error) {
	payment := s.NewSubscriptionPayment()
	return payment.FindList(search)
}

func (s *SubscriptionService) AdjustUsers(subscriptionID int64, newNumberOfUsers int, notes string) (*models.Subscription, error) {
	if newNumberOfUsers <= 0 {
		return nil, errors.New("number of users subscription wajib lebih dari 0")
	}

	current, err := s.FindByID(subscriptionID)
	if err != nil {
		return nil, err
	}
	if newNumberOfUsers == current.NumberOfUsers {
		return current, nil
	}

	if newNumberOfUsers > current.NumberOfUsers {
		return s.increaseUsers(current, newNumberOfUsers, notes)
	}

	return s.scheduleReduction(current, newNumberOfUsers, notes)
}

func (s *SubscriptionService) increaseUsers(current *models.Subscription, newNumberOfUsers int, notes string) (*models.Subscription, error) {
	effectiveDate := models.StartOfDay(time.Now().UTC())
	remainingDays := models.InclusiveDayDiff(effectiveDate, current.CurrentEndDate)
	totalDays := models.InclusiveDayDiff(current.CurrentStartDate, current.CurrentEndDate)
	if remainingDays <= 0 {
		return nil, errors.New("subscription period sudah berakhir, silakan renew terlebih dahulu")
	}
	if totalDays <= 0 {
		totalDays = 1
	}

	delta := newNumberOfUsers - current.NumberOfUsers
	amountChange := models.RoundMoney(float64(delta) * current.PricePerUser * float64(remainingDays) / float64(totalDays))
	nextAmount := models.RoundMoney(current.Amount + amountChange)

	tx, err := s.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	updatedBy := s.Auth.UserName
	if updatedBy == "" {
		updatedBy = "system"
	}

	sql := `UPDATE ` + mmodels.SubscriptionTable + `
	SET number_of_users = $1,
		amount = $2,
		pending_number_of_users = NULL,
		pending_effective_date = NULL,
		updated_at = NOW(),
		updated_by = $3
	WHERE id = $4 AND account_id = $5`
	if _, err = tx.Exec(sql, newNumberOfUsers, nextAmount, updatedBy, current.ID, current.AccountID); err != nil {
		return nil, err
	}

	current.NumberOfUsers = newNumberOfUsers
	current.Amount = nextAmount
	current.PendingNumberOfUsers = nil
	current.PendingEffectiveDate = nil

	if _, err = s.createHistoryTx(tx, models.SubscriptionHistory{
		AccountID:             current.AccountID,
		SubscriptionID:        current.ID,
		AppType:               current.AppType,
		HistoryType:           models.SubscriptionHistoryTypeIncreaseUsers,
		PreviousNumberOfUsers: current.NumberOfUsers - delta,
		NewNumberOfUsers:      current.NumberOfUsers,
		DeltaUsers:            delta,
		PricePerUser:          current.PricePerUser,
		AmountBefore:          models.RoundMoney(current.Amount - amountChange),
		AmountChange:          amountChange,
		AmountAfter:           nextAmount,
		CurrentStartDate:      current.CurrentStartDate,
		CurrentEndDate:        current.CurrentEndDate,
		EffectiveDate:         effectiveDate,
		Notes:                 strings.TrimSpace(notes),
	}); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *SubscriptionService) scheduleReduction(current *models.Subscription, newNumberOfUsers int, notes string) (*models.Subscription, error) {
	targetDate := models.StartOfDay(current.CurrentEndDate)
	tx, err := s.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	updatedBy := s.Auth.UserName
	if updatedBy == "" {
		updatedBy = "system"
	}

	sql := `UPDATE ` + mmodels.SubscriptionTable + `
	SET pending_number_of_users = $1,
		pending_effective_date = $2,
		updated_at = NOW(),
		updated_by = $3
	WHERE id = $4 AND account_id = $5`
	if _, err = tx.Exec(sql, newNumberOfUsers, targetDate, updatedBy, current.ID, current.AccountID); err != nil {
		return nil, err
	}

	if _, err = s.createHistoryTx(tx, models.SubscriptionHistory{
		AccountID:             current.AccountID,
		SubscriptionID:        current.ID,
		AppType:               current.AppType,
		HistoryType:           models.SubscriptionHistoryTypeScheduleReduction,
		PreviousNumberOfUsers: current.NumberOfUsers,
		NewNumberOfUsers:      newNumberOfUsers,
		DeltaUsers:            newNumberOfUsers - current.NumberOfUsers,
		PricePerUser:          current.PricePerUser,
		AmountBefore:          current.Amount,
		AmountChange:          0,
		AmountAfter:           current.Amount,
		CurrentStartDate:      current.CurrentStartDate,
		CurrentEndDate:        current.CurrentEndDate,
		EffectiveDate:         targetDate,
		Notes:                 strings.TrimSpace(notes),
	}); err != nil {
		return nil, err
	}

	current.PendingNumberOfUsers = &newNumberOfUsers
	current.PendingEffectiveDate = &targetDate

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return current, nil
}

func (s *SubscriptionService) CountActiveUsers(accountID int64) (int, error) {
	if accountID <= 0 {
		accountID = s.Auth.AccountID
	}
	if accountID <= 0 {
		return 0, errors.New("id akun wajib diisi")
	}

	var count int
	if err := s.DB.Get(&count, `SELECT COUNT(*) FROM `+mmodels.UserTable+` WHERE account_id = $1 AND is_active = true`, accountID); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *SubscriptionService) ensurePaymentUsersBelongToAccount(userIDs []int64) error {
	userIDs = uniquePositiveInt64(userIDs)
	if len(userIDs) == 0 {
		return errors.New("pilih minimal 1 user yang dibayar")
	}

	var count int
	query, args, err := sqlx.In(
		`SELECT COUNT(*)
		FROM `+mmodels.UserTable+`
		WHERE account_id = ? AND is_active = true AND id IN (?)`,
		s.Auth.AccountID,
		userIDs,
	)
	if err != nil {
		return err
	}
	query = s.DB.Rebind(query)
	if err := s.DB.Get(&count, query, args...); err != nil {
		return err
	}
	if count != len(userIDs) {
		return errors.New("ada user yang tidak aktif atau tidak sesuai akun")
	}
	return nil
}

func uniquePositiveInt64(values []int64) []int64 {
	seen := map[int64]bool{}
	result := []int64{}
	for _, value := range values {
		if value <= 0 || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func (s *SubscriptionService) createHistory(history models.SubscriptionHistory) (int64, error) {
	return s.createHistoryTx(nil, history)
}

func (s *SubscriptionService) createHistoryTx(tx *sqlx.Tx, history models.SubscriptionHistory) (int64, error) {
	model := s.NewSubscriptionHistory()
	model.AccountID = history.AccountID
	model.SubscriptionID = history.SubscriptionID
	model.AppType = models.NormalizeSubscriptionAppType(history.AppType)
	model.HistoryType = history.HistoryType
	model.PreviousNumberOfUsers = history.PreviousNumberOfUsers
	model.NewNumberOfUsers = history.NewNumberOfUsers
	model.DeltaUsers = history.DeltaUsers
	model.PricePerUser = history.PricePerUser
	model.AmountBefore = history.AmountBefore
	model.AmountChange = history.AmountChange
	model.AmountAfter = history.AmountAfter
	model.CurrentStartDate = normalizeDate(history.CurrentStartDate)
	model.CurrentEndDate = normalizeDate(history.CurrentEndDate)
	model.EffectiveDate = normalizeDate(history.EffectiveDate)
	model.Notes = history.Notes
	model.CreatedBy = history.CreatedBy
	if model.CreatedBy == "" {
		model.CreatedBy = s.Auth.UserName
	}
	if model.CreatedBy == "" {
		model.CreatedBy = "system"
	}
	if s.TransactionDB != s.DB {
		return model.Create()
	}

	if tx == nil {
		return model.Create()
	}

	accountID := model.AccountID
	if accountID <= 0 {
		accountID = s.Auth.AccountID
	}
	sql := `INSERT INTO ` + mmodels.SubscriptionHistoryTable + `
		(account_id, subscription_id, app_type, history_type, previous_number_of_users, new_number_of_users, delta_users, price_per_user,
		amount_before, amount_change, amount_after, current_start_date, current_end_date, effective_date, notes, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		RETURNING id`

	var id int64
	err := tx.QueryRow(
		sql,
		accountID,
		model.SubscriptionID,
		model.AppType,
		model.HistoryType,
		model.PreviousNumberOfUsers,
		model.NewNumberOfUsers,
		model.DeltaUsers,
		model.PricePerUser,
		model.AmountBefore,
		model.AmountChange,
		model.AmountAfter,
		model.CurrentStartDate,
		model.CurrentEndDate,
		model.EffectiveDate,
		model.Notes,
		model.CreatedBy,
	).Scan(&id)
	return id, err
}

func normalizeDate(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}
	return models.StartOfDay(value)
}

func normalizeTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	normalized := models.StartOfDay(*value)
	return &normalized
}
