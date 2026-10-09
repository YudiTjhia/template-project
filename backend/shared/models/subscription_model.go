package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	SubscriptionStatusActive    = "ACTIVE"
	SubscriptionStatusExpired   = "EXPIRED"
	SubscriptionStatusSuspended = "SUSPENDED"

	SubscriptionAppTypePOSKasir      = "POS Kasir"
	SubscriptionAppTypeSales         = "Sales"
	SubscriptionAppTypeProcurement   = "Procurement"
	SubscriptionAppTypeHumanResource = "Human Resources"
)

type Subscription struct {
	BaseModel

	ID                   int64                  `json:"id" db:"id"`
	AccountID            int64                  `json:"account_id" db:"account_id"`
	AccountName          string                 `json:"accountName,omitempty" db:"account_name"`
	CompanyName          string                 `json:"companyName,omitempty" db:"company_name"`
	AppType              string                 `json:"appType" db:"app_type"`
	PlanCode             string                 `json:"planCode" db:"plan_code"`
	PlanName             string                 `json:"planName" db:"plan_name"`
	PlanLimitsJSON       string                 `json:"-" db:"plan_limits"`
	PlanLimits           SubscriptionPlanLimits `json:"planLimits" db:"-"`
	NumberOfUsers        int                    `json:"numberOfUsers" db:"number_of_users"`
	PricePerUser         float64                `json:"pricePerUser" db:"price_per_user"`
	Amount               float64                `json:"amount" db:"amount"`
	CurrentStartDate     time.Time              `json:"currentStartDate" db:"current_start_date"`
	CurrentEndDate       time.Time              `json:"currentEndDate" db:"current_end_date"`
	PendingNumberOfUsers *int                   `json:"pendingNumberOfUsers,omitempty" db:"pending_number_of_users"`
	PendingEffectiveDate *time.Time             `json:"pendingEffectiveDate,omitempty" db:"pending_effective_date"`
	Status               string                 `json:"status" db:"status"`
}

type SubscriptionStatus struct {
	IsPaid          bool          `json:"isPaid"`
	HasSubscription bool          `json:"hasSubscription"`
	Message         string        `json:"message"`
	Subscription    *Subscription `json:"subscription,omitempty"`
}

const subscriptionSelectColumns = `s.id, s.account_id, a.account_name, a.company_name, s.app_type, s.plan_code, s.plan_name, s.plan_limits, s.number_of_users, s.price_per_user, s.amount,
s.current_start_date, s.current_end_date, s.pending_number_of_users, s.pending_effective_date, s.status,
s.created_at, s.created_by, s.updated_at, s.updated_by`

func NewSubscription(db *sqlx.DB, auth Auth) Subscription {
	subscription := Subscription{
		BaseModel: BaseModel{DB: db, Auth: auth},
	}
	subscription.AccountID = auth.AccountID
	subscription.CreatedBy = auth.UserName
	subscription.Status = SubscriptionStatusActive
	subscription.AppType = SubscriptionAppTypePOSKasir
	plan := SubscriptionPlanByCode("")
	subscription.PlanCode = plan.Code
	subscription.PlanName = plan.Name
	subscription.PlanLimits = plan.Limits
	return subscription
}

func (m Subscription) Migrate() {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS subscriptions (
	id SERIAL PRIMARY KEY,
	account_id BIGINT NOT NULL DEFAULT 0,
	app_type VARCHAR(100) NOT NULL DEFAULT 'POS Kasir',
	plan_code VARCHAR(50) NOT NULL DEFAULT 'STARTER',
	plan_name VARCHAR(100) NOT NULL DEFAULT 'Starter',
	plan_limits JSONB NOT NULL DEFAULT '{}'::JSONB,
	number_of_users INTEGER NOT NULL DEFAULT 1,
	price_per_user NUMERIC(18,2) NOT NULL DEFAULT 0,
	amount NUMERIC(18,2) NOT NULL DEFAULT 0,
	current_start_date DATE NOT NULL DEFAULT CURRENT_DATE,
	current_end_date DATE NOT NULL DEFAULT CURRENT_DATE,
	pending_number_of_users INTEGER,
	pending_effective_date DATE,
	status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_by VARCHAR(255) NOT NULL DEFAULT '',
	updated_at TIMESTAMPTZ,
	updated_by VARCHAR(255)
);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_account_id ON subscriptions (account_id);`,
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions (account_id, status);`,
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_end_date ON subscriptions (account_id, current_end_date);`,
		`ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS account_id BIGINT NOT NULL DEFAULT 0;`,
		`ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS app_type VARCHAR(100) NOT NULL DEFAULT 'POS Kasir';`,
		`ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS plan_code VARCHAR(50) NOT NULL DEFAULT 'STARTER';`,
		`ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS plan_name VARCHAR(100) NOT NULL DEFAULT 'Starter';`,
		`ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS plan_limits JSONB NOT NULL DEFAULT '{}'::JSONB;`,
		`ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS number_of_users INTEGER NOT NULL DEFAULT 1;`,
		`ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS price_per_user NUMERIC(18,2) NOT NULL DEFAULT 0;`,
		`ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS amount NUMERIC(18,2) NOT NULL DEFAULT 0;`,
		`ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS current_start_date DATE NOT NULL DEFAULT CURRENT_DATE;`,
		`ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS current_end_date DATE NOT NULL DEFAULT CURRENT_DATE;`,
		`ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS pending_number_of_users INTEGER;`,
		`ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS pending_effective_date DATE;`,
		`ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE';`,
	}

	for _, statement := range statements {
		log.Println("[SQL]=", statement)
		m.DB.MustExec(statement)
	}
}

func (m Subscription) TableName() string {
	return "subscriptions"
}

func (m Subscription) ValidateCreate() error {
	if _, err := m.resolveAccountID(); err != nil {
		return err
	}
	m.AppType = NormalizeSubscriptionAppType(m.AppType)
	m.applyPlan()
	if !IsAllowedSubscriptionAppType(m.AppType) {
		return errors.New("app type subscription tidak valid")
	}
	if m.NumberOfUsers <= 0 {
		return errors.New("number of users subscription wajib lebih dari 0")
	}
	if m.PricePerUser < 0 {
		return errors.New("price per user subscription tidak valid")
	}
	if m.Amount < 0 {
		return errors.New("amount subscription tidak valid")
	}
	if m.CurrentStartDate.IsZero() {
		return errors.New("current start date subscription wajib diisi")
	}
	if m.CurrentEndDate.IsZero() {
		return errors.New("current end date subscription wajib diisi")
	}
	if m.CurrentEndDate.Before(m.CurrentStartDate) {
		return errors.New("current end date tidak boleh lebih kecil dari current start date")
	}
	return nil
}

func (m Subscription) ValidateID() error {
	if m.ID <= 0 {
		return errors.New("id subscription tidak valid")
	}
	return nil
}

func (m Subscription) Create() (int64, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return 0, err
	}

	m.Status = strings.ToUpper(strings.TrimSpace(m.Status))
	if m.Status == "" {
		m.Status = SubscriptionStatusActive
	}
	m.AppType = NormalizeSubscriptionAppType(m.AppType)
	m.CreatedBy = strings.TrimSpace(m.CreatedBy)
	if m.CreatedBy == "" {
		m.CreatedBy = m.Auth.UserName
	}
	if m.CreatedBy == "" {
		m.CreatedBy = "system"
	}

	if m.Amount <= 0 && m.NumberOfUsers > 0 && m.PricePerUser >= 0 {
		m.Amount = float64(m.NumberOfUsers) * m.PricePerUser
	}

	if err := m.ValidateCreate(); err != nil {
		return 0, err
	}

	sql := `INSERT INTO subscriptions
	(account_id, app_type, plan_code, plan_name, plan_limits, number_of_users, price_per_user, amount, current_start_date, current_end_date, pending_number_of_users, pending_effective_date, status, created_by)
	VALUES ($1, $2, $3, $4, $5::JSONB, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	RETURNING id`
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{
		"account_id":              accountID,
		"app_type":                m.AppType,
		"plan_code":               m.PlanCode,
		"plan_name":               m.PlanName,
		"number_of_users":         m.NumberOfUsers,
		"price_per_user":          m.PricePerUser,
		"amount":                  m.Amount,
		"current_start_date":      m.CurrentStartDate,
		"current_end_date":        m.CurrentEndDate,
		"pending_number_of_users": m.PendingNumberOfUsers,
		"pending_effective_date":  m.PendingEffectiveDate,
		"status":                  m.Status,
		"created_by":              m.CreatedBy,
	})

	var id int64
	err = m.DB.QueryRow(
		sql,
		accountID,
		m.AppType,
		m.PlanCode,
		m.PlanName,
		m.PlanLimitsJSON,
		m.NumberOfUsers,
		m.PricePerUser,
		m.Amount,
		m.CurrentStartDate,
		m.CurrentEndDate,
		m.PendingNumberOfUsers,
		m.PendingEffectiveDate,
		m.Status,
		m.CreatedBy,
	).Scan(&id)
	return id, err
}

func (m Subscription) Update() error {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return err
	}
	if err := m.ValidateID(); err != nil {
		return err
	}

	m.Status = strings.ToUpper(strings.TrimSpace(m.Status))
	if m.Status == "" {
		m.Status = SubscriptionStatusActive
	}
	m.AppType = NormalizeSubscriptionAppType(m.AppType)
	m.applyPlan()
	if !IsAllowedSubscriptionAppType(m.AppType) {
		return errors.New("app type subscription tidak valid")
	}

	updatedBy := m.UpdatedBy
	if updatedBy == nil || *updatedBy == "" {
		if m.Auth.UserName != "" {
			updatedBy = &m.Auth.UserName
		}
	}

	sql := `UPDATE subscriptions
	SET app_type = $1,
		plan_code = $2,
		plan_name = $3,
		plan_limits = $4::JSONB,
		number_of_users = $5,
		price_per_user = $6,
		amount = $7,
		current_start_date = $8,
		current_end_date = $9,
		pending_number_of_users = $10,
		pending_effective_date = $11,
		status = $12,
		updated_at = NOW(),
		updated_by = $13
	WHERE id = $14 AND account_id = $15`
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{
		"app_type":                m.AppType,
		"plan_code":               m.PlanCode,
		"plan_name":               m.PlanName,
		"number_of_users":         m.NumberOfUsers,
		"price_per_user":          m.PricePerUser,
		"amount":                  m.Amount,
		"current_start_date":      m.CurrentStartDate,
		"current_end_date":        m.CurrentEndDate,
		"pending_number_of_users": m.PendingNumberOfUsers,
		"pending_effective_date":  m.PendingEffectiveDate,
		"status":                  m.Status,
		"updated_by":              updatedBy,
		"id":                      m.ID,
		"account_id":              accountID,
	})

	_, err = m.DB.Exec(sql,
		m.AppType,
		m.PlanCode,
		m.PlanName,
		m.PlanLimitsJSON,
		m.NumberOfUsers,
		m.PricePerUser,
		m.Amount,
		m.CurrentStartDate,
		m.CurrentEndDate,
		m.PendingNumberOfUsers,
		m.PendingEffectiveDate,
		m.Status,
		updatedBy,
		m.ID,
		accountID,
	)
	return err
}

func (m *Subscription) applyPlan() {
	plan := SubscriptionPlanByCode(m.PlanCode)
	m.PlanCode = plan.Code
	m.PlanName = plan.Name
	m.PlanLimits = plan.Limits
	raw, _ := json.Marshal(plan.Limits)
	m.PlanLimitsJSON = string(raw)
	if m.NumberOfUsers <= 0 {
		m.NumberOfUsers = plan.Limits.Users
	}
	if m.PricePerUser <= 0 {
		m.PricePerUser = plan.Price
	}
}

func (m Subscription) Delete() error {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return err
	}
	if err := m.ValidateID(); err != nil {
		return err
	}

	sql := `DELETE FROM subscriptions WHERE id = $1 AND account_id = $2`
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"id": m.ID, "account_id": accountID})

	_, err = m.DB.Exec(sql, m.ID, accountID)
	return err
}

func (m Subscription) DeleteByID(id int64) error {
	m.ID = id
	return m.Delete()
}

func (m Subscription) FindByID() (*Subscription, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}
	if err := m.ValidateID(); err != nil {
		return nil, err
	}

	sql := subscriptionSelectQuery(`s.id = $1 AND s.account_id = $2`)
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"id": m.ID, "account_id": accountID})

	var item Subscription
	err = m.DB.Get(&item, sql, m.ID, accountID)
	item.loadPlanLimits()
	return &item, err
}

func (m Subscription) FindByAccount() (*Subscription, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}

	sql := subscriptionSelectQuery(`s.account_id = $1`)
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"account_id": accountID})

	var item Subscription
	err = m.DB.Get(&item, sql, accountID)
	item.loadPlanLimits()
	return &item, err
}

func (m *Subscription) loadPlanLimits() {
	if strings.TrimSpace(m.PlanLimitsJSON) != "" {
		_ = json.Unmarshal([]byte(m.PlanLimitsJSON), &m.PlanLimits)
	}
	if strings.TrimSpace(m.PlanCode) == "" {
		plan := SubscriptionPlanByCode("")
		m.PlanCode = plan.Code
		m.PlanName = plan.Name
		m.PlanLimits = plan.Limits
	}
}

func subscriptionSelectQuery(where string) string {
	return `SELECT ` + subscriptionSelectColumns + `
	FROM subscriptions s
	LEFT JOIN accounts a ON a.id = s.account_id
	WHERE ` + where
}

func (m Subscription) resolveAccountID() (int64, error) {
	if m.Auth.AccountID > 0 {
		return m.Auth.AccountID, nil
	}
	if m.AccountID > 0 {
		return m.AccountID, nil
	}
	return 0, errors.New("id akun wajib diisi")
}

func NormalizeSubscriptionAppType(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return SubscriptionAppTypePOSKasir
	}

	switch strings.ToLower(normalized) {
	case strings.ToLower(SubscriptionAppTypePOSKasir):
		return SubscriptionAppTypePOSKasir
	case strings.ToLower(SubscriptionAppTypeSales):
		return SubscriptionAppTypeSales
	case strings.ToLower(SubscriptionAppTypeProcurement):
		return SubscriptionAppTypeProcurement
	case strings.ToLower(SubscriptionAppTypeHumanResource):
		return SubscriptionAppTypeHumanResource
	default:
		return normalized
	}
}

func IsAllowedSubscriptionAppType(value string) bool {
	switch NormalizeSubscriptionAppType(value) {
	case SubscriptionAppTypePOSKasir, SubscriptionAppTypeSales, SubscriptionAppTypeProcurement, SubscriptionAppTypeHumanResource:
		return true
	default:
		return false
	}
}

func RoundMoney(value float64) float64 {
	return float64(int64(value*100+0.5)) / 100
}

func StartOfDay(value time.Time) time.Time {
	return time.Date(value.UTC().Year(), value.UTC().Month(), value.UTC().Day(), 0, 0, 0, 0, time.UTC)
}

func InclusiveDayDiff(start, end time.Time) int {
	start = StartOfDay(start)
	end = StartOfDay(end)
	if end.Before(start) {
		return 0
	}
	return int(end.Sub(start).Hours()/24) + 1
}

func NextDay(value time.Time) time.Time {
	value = StartOfDay(value)
	return value.AddDate(0, 0, 1)
}

func DefaultSubscriptionWindow() (time.Time, time.Time) {
	start := StartOfDay(time.Now().UTC())
	end := start.AddDate(0, 0, 30)
	return start, end
}

func DaysToDate(value time.Time) string {
	return StartOfDay(value).Format("2006-01-02")
}

func FmtMoney(value float64) string {
	return fmt.Sprintf("%.2f", RoundMoney(value))
}
