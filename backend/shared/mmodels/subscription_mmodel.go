package mmodels

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend/shared/metadata/impl"
	"backend/shared/metadata/interfaces"
	"github.com/jmoiron/sqlx"
)

const (
	SubscriptionStatusActive    = "ACTIVE"
	SubscriptionStatusExpired   = "EXPIRED"
	SubscriptionStatusSuspended = "SUSPENDED"

	SubscriptionAppTypePOSKasir      = "POS Kasir"
	SubscriptionAppTypeHumanResource = "Human Resources"
)

type SubscriptionMModel struct {
	impl.MModel

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
	CreatedAt            time.Time              `json:"-" db:"created_at"`
	CreatedBy            string                 `json:"-" db:"created_by"`
	UpdatedAt            sql.NullTime           `json:"-" db:"updated_at"`
	UpdatedBy            sql.NullString         `json:"-" db:"updated_by"`
	AuthAccountID        int64                  `json:"-" db:"-"`
	AuthUserName         string                 `json:"-" db:"-"`
}

func NewSubscriptionMModel(ctx context.Context, db *sqlx.DB) *SubscriptionMModel {
	return &SubscriptionMModel{
		Status:  SubscriptionStatusActive,
		AppType: SubscriptionAppTypePOSKasir,
		MModel: impl.MModel{
			Context: ctx, Name: "Subscription", TableName: SubscriptionTable, DB: db,
			Fields: []interfaces.IMField{
				impl.MField{Type: "int64", DbType: "BIGINT", JsonTag: "id", DbTag: "id", IsPrimaryKey: true, IsIdentity: true, Required: true, IsNotUpdateField: true},
				impl.MField{Type: "int64", DbType: "BIGINT", JsonTag: "account_id", DbTag: "account_id", Required: true, IsNotUpdateField: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "appType", DbTag: "app_type", MaxLength: 100, Default: "'POS Kasir'", Required: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "planCode", DbTag: "plan_code", MaxLength: 50, Default: "'STARTER'", Required: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "planName", DbTag: "plan_name", MaxLength: 100, Default: "'Starter'", Required: true},
				impl.MField{Type: "string", DbType: "JSONB", JsonTag: "-", DbTag: "plan_limits", Default: "'{}'::JSONB", Required: true},
				impl.MField{Type: "int", DbType: "INTEGER", JsonTag: "numberOfUsers", DbTag: "number_of_users", Default: "1", Required: true},
				impl.MField{Type: "float64", DbType: "NUMERIC(18,2)", JsonTag: "pricePerUser", DbTag: "price_per_user", Default: "0", Required: true},
				impl.MField{Type: "float64", DbType: "NUMERIC(18,2)", JsonTag: "amount", DbTag: "amount", Default: "0", Required: true},
				impl.MField{Type: "time.Time", DbType: "DATE", JsonTag: "currentStartDate", DbTag: "current_start_date", Required: true},
				impl.MField{Type: "time.Time", DbType: "DATE", JsonTag: "currentEndDate", DbTag: "current_end_date", Required: true},
				impl.MField{Type: "*int", DbType: "INTEGER", JsonTag: "pendingNumberOfUsers", DbTag: "pending_number_of_users", IsNullable: true},
				impl.MField{Type: "*time.Time", DbType: "DATE", JsonTag: "pendingEffectiveDate", DbTag: "pending_effective_date", IsNullable: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "status", DbTag: "status", MaxLength: 30, Default: "'ACTIVE'", Required: true},
				impl.MField{Type: "time.Time", DbType: "TIMESTAMPTZ", JsonTag: "-", DbTag: "created_at", Default: "NOW()", Required: true, IsNotUpdateField: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "-", DbTag: "created_by", MaxLength: 255, Required: true, IsNotUpdateField: true},
				impl.MField{Type: "sql.NullTime", DbType: "TIMESTAMPTZ", JsonTag: "-", DbTag: "updated_at", IsNullable: true, IsNotCreateField: true},
				impl.MField{Type: "sql.NullString", DbType: "VARCHAR", JsonTag: "-", DbTag: "updated_by", MaxLength: 255, IsNullable: true, IsNotCreateField: true},
			},
		},
	}
}

func (m *SubscriptionMModel) Migrate() error {
	if m == nil {
		return errors.New("subscription metadata model is nil")
	}
	if m.DB == nil {
		return errors.New("subscription metadata database is nil")
	}
	ctx := m.ContextOrBackground()
	tx, err := m.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin subscriptions migration: %w", err)
	}
	defer tx.Rollback()

	migrator := impl.NewMMigrator(&m.MModel, tx)
	migrator.SetContext(ctx)
	if err := migrator.Migrate(); err != nil {
		return err
	}
	statements := []string{
		`ALTER TABLE ` + SubscriptionTable + ` ADD COLUMN IF NOT EXISTS account_id BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE ` + SubscriptionTable + ` ADD COLUMN IF NOT EXISTS app_type VARCHAR(100) NOT NULL DEFAULT 'POS Kasir'`,
		`ALTER TABLE ` + SubscriptionTable + ` ADD COLUMN IF NOT EXISTS plan_code VARCHAR(50) NOT NULL DEFAULT 'STARTER'`,
		`ALTER TABLE ` + SubscriptionTable + ` ADD COLUMN IF NOT EXISTS plan_name VARCHAR(100) NOT NULL DEFAULT 'Starter'`,
		`ALTER TABLE ` + SubscriptionTable + ` ADD COLUMN IF NOT EXISTS plan_limits JSONB NOT NULL DEFAULT '{}'::JSONB`,
		`ALTER TABLE ` + SubscriptionTable + ` ADD COLUMN IF NOT EXISTS number_of_users INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE ` + SubscriptionTable + ` ADD COLUMN IF NOT EXISTS price_per_user NUMERIC(18,2) NOT NULL DEFAULT 0`,
		`ALTER TABLE ` + SubscriptionTable + ` ADD COLUMN IF NOT EXISTS amount NUMERIC(18,2) NOT NULL DEFAULT 0`,
		`ALTER TABLE ` + SubscriptionTable + ` ADD COLUMN IF NOT EXISTS current_start_date DATE NOT NULL DEFAULT CURRENT_DATE`,
		`ALTER TABLE ` + SubscriptionTable + ` ADD COLUMN IF NOT EXISTS current_end_date DATE NOT NULL DEFAULT CURRENT_DATE`,
		`ALTER TABLE ` + SubscriptionTable + ` ADD COLUMN IF NOT EXISTS pending_number_of_users INTEGER`,
		`ALTER TABLE ` + SubscriptionTable + ` ADD COLUMN IF NOT EXISTS pending_effective_date DATE`,
		`ALTER TABLE ` + SubscriptionTable + ` ADD COLUMN IF NOT EXISTS status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE'`,
		`ALTER TABLE ` + SubscriptionTable + ` ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`ALTER TABLE ` + SubscriptionTable + ` ADD COLUMN IF NOT EXISTS created_by VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE ` + SubscriptionTable + ` ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ`,
		`ALTER TABLE ` + SubscriptionTable + ` ADD COLUMN IF NOT EXISTS updated_by VARCHAR(255)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_account_id ON ` + SubscriptionTable + ` (account_id)`,
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON ` + SubscriptionTable + ` (account_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_end_date ON ` + SubscriptionTable + ` (account_id, current_end_date)`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("run subscriptions migration statement %q: %w", statement, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit subscriptions migration: %w", err)
	}
	return nil
}

func (m *SubscriptionMModel) Normalize() error {
	if m == nil {
		return errors.New("subscription metadata model is nil")
	}
	m.AppType = normalizeSubscriptionAppType(m.AppType)
	m.PlanCode = strings.ToUpper(strings.TrimSpace(m.PlanCode))
	m.Status = strings.ToUpper(strings.TrimSpace(m.Status))
	if m.Status == "" {
		m.Status = SubscriptionStatusActive
	}
	if !isAllowedSubscriptionAppType(m.AppType) {
		return errors.New("app type subscription tidak valid")
	}
	if m.NumberOfUsers <= 0 {
		return errors.New("number of users subscription wajib lebih dari 0")
	}
	if m.PricePerUser < 0 || m.Amount < 0 {
		return errors.New("nilai subscription tidak valid")
	}
	if m.CurrentStartDate.IsZero() || m.CurrentEndDate.IsZero() {
		return errors.New("tanggal subscription wajib diisi")
	}
	if m.CurrentEndDate.Before(m.CurrentStartDate) {
		return errors.New("current end date tidak boleh lebih kecil dari current start date")
	}
	if m.PlanLimitsJSON == "" {
		raw, err := json.Marshal(m.PlanLimits)
		if err != nil {
			return err
		}
		m.PlanLimitsJSON = string(raw)
	}
	if err := json.Unmarshal([]byte(m.PlanLimitsJSON), &m.PlanLimits); err != nil {
		return fmt.Errorf("plan limits subscription tidak valid: %w", err)
	}
	return nil
}

func (m *SubscriptionMModel) applyPlan() error {
	planModel := NewSubscriptionPlanMModel(m.ContextOrBackground(), m.DB)
	plan, err := planModel.FindByCode(m.PlanCode)
	if err != nil {
		return err
	}
	m.PlanCode = plan.Code
	m.PlanName = plan.Name
	m.PlanLimits = plan.Limits
	raw, err := json.Marshal(plan.Limits)
	if err != nil {
		return err
	}
	m.PlanLimitsJSON = string(raw)
	if m.NumberOfUsers <= 0 {
		m.NumberOfUsers = plan.Limits.Users
	}
	if m.PricePerUser <= 0 {
		m.PricePerUser = plan.Price
	}
	if m.Amount <= 0 {
		m.Amount = roundSubscriptionMoney(float64(m.NumberOfUsers) * m.PricePerUser)
	}
	return nil
}

func (m *SubscriptionMModel) ValidateID() error {
	if m == nil {
		return errors.New("id subscription tidak valid")
	}
	return m.MModel.ValidateID_(m.ID, "subscription")
}

func (m *SubscriptionMModel) ValidateCreate() error {
	if m == nil {
		return errors.New("subscription metadata model is nil")
	}
	accountID, err := m.MModel.ValidateAccountID_(m.AuthAccountID, m.AccountID)
	if err != nil {
		return err
	}
	m.AccountID = accountID
	if err := m.applyPlan(); err != nil {
		return err
	}
	if err := m.Normalize(); err != nil {
		return err
	}
	m.CreatedBy, err = m.MModel.ResolveCreatedBy_(m.CreatedBy, m.AuthUserName)
	if err != nil {
		return err
	}
	return m.MModel.ValidateCreate_(m)
}

func (m *SubscriptionMModel) Create() (int64, error) {
	if m == nil {
		return 0, errors.New("subscription metadata model is nil")
	}
	accountID, err := m.MModel.ValidateAccountID_(m.AuthAccountID, m.AccountID)
	if err != nil {
		return 0, err
	}
	m.AccountID = accountID
	if err := m.ValidateCreate(); err != nil {
		return 0, err
	}
	return m.MModel.Create_(m)
}

func (m *SubscriptionMModel) Update() error {
	if m == nil {
		return errors.New("subscription metadata model is nil")
	}
	if err := m.ValidateID(); err != nil {
		return err
	}
	accountID, err := m.MModel.ValidateAccountID_(m.AuthAccountID, m.AccountID)
	if err != nil {
		return err
	}
	m.AccountID = accountID
	if err := m.applyPlan(); err != nil {
		return err
	}
	if err := m.Normalize(); err != nil {
		return err
	}
	m.UpdatedBy, err = m.MModel.ResolveUpdatedBy_(m.UpdatedBy, m.AuthUserName)
	if err != nil {
		return err
	}
	m.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	if err := m.MModel.ValidateUpdate_(m); err != nil {
		return err
	}
	return m.MModel.Update_(m, "id", "account_id")
}

func (m *SubscriptionMModel) Delete() error {
	if err := m.ValidateID(); err != nil {
		return err
	}
	accountID, err := m.MModel.ValidateAccountID_(m.AuthAccountID, m.AccountID)
	if err != nil {
		return err
	}
	m.AccountID = accountID
	return m.MModel.Delete_(m, "id", "account_id")
}

func (m *SubscriptionMModel) FindByID() (*SubscriptionMModel, error) {
	if err := m.ValidateID(); err != nil {
		return nil, err
	}
	accountID, err := m.MModel.ValidateAccountID_(m.AuthAccountID, m.AccountID)
	if err != nil {
		return nil, err
	}
	found := &SubscriptionMModel{MModel: m.MModel, ID: m.ID, AccountID: accountID}
	if err := m.findJoined(found, `s.id = $1 AND s.account_id = $2`, m.ID, accountID); err != nil {
		return nil, err
	}
	if err := found.loadPlanLimits(); err != nil {
		return nil, err
	}
	return found, nil
}

func (m *SubscriptionMModel) FindByAccount() (*SubscriptionMModel, error) {
	if m == nil {
		return nil, errors.New("subscription metadata model is nil")
	}
	accountID, err := m.MModel.ValidateAccountID_(m.AuthAccountID, m.AccountID)
	if err != nil {
		return nil, err
	}
	found := &SubscriptionMModel{MModel: m.MModel, AccountID: accountID}
	if err := m.findJoined(found, `s.account_id = $1`, accountID); err != nil {
		return nil, err
	}
	if err := found.loadPlanLimits(); err != nil {
		return nil, err
	}
	return found, nil
}

func (m *SubscriptionMModel) FindList(search string, limit int) ([]SubscriptionMModel, error) {
	if m == nil {
		return nil, errors.New("subscription metadata model is nil")
	}
	accountID, err := m.MModel.ValidateAccountID_(m.AuthAccountID, m.AccountID)
	if err != nil {
		return nil, err
	}
	if limit < 0 {
		return nil, errors.New("limit subscription tidak valid")
	}
	items := make([]SubscriptionMModel, 0)
	if err := m.MModel.FindListOrdered(&items, map[string]any{"account_id": accountID}, search,
		[]string{"app_type", "plan_code", "plan_name", "status"}, "created_at,id", "DESC", limit); err != nil {
		return nil, err
	}
	for i := range items {
		items[i].MModel = m.MModel
		if err := items[i].loadPlanLimits(); err != nil {
			return nil, err
		}
	}
	return items, nil
}

func (m *SubscriptionMModel) findJoined(destination any, where string, args ...any) error {
	if m == nil || m.DB == nil {
		return errors.New("subscription metadata database is nil")
	}
	query := `SELECT s.id, s.account_id, a.account_name, a.company_name, s.app_type, s.plan_code, s.plan_name, s.plan_limits,
		s.number_of_users, s.price_per_user, s.amount, s.current_start_date, s.current_end_date,
		s.pending_number_of_users, s.pending_effective_date, s.status, s.created_at, s.created_by, s.updated_at, s.updated_by
		FROM ` + SubscriptionTable + ` s LEFT JOIN ` + AccountTable + ` a ON a.id = s.account_id WHERE ` + where
	return m.DB.GetContext(m.ContextOrBackground(), destination, query, args...)
}

func (m *SubscriptionMModel) loadPlanLimits() error {
	if strings.TrimSpace(m.PlanLimitsJSON) == "" {
		m.PlanLimits = SubscriptionPlanLimits{}
		return nil
	}
	return json.Unmarshal([]byte(m.PlanLimitsJSON), &m.PlanLimits)
}

func normalizeSubscriptionAppType(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return SubscriptionAppTypePOSKasir
	}
	switch strings.ToLower(value) {
	case strings.ToLower(SubscriptionAppTypePOSKasir):
		return SubscriptionAppTypePOSKasir
	case strings.ToLower(SubscriptionAppTypeHumanResource):
		return SubscriptionAppTypeHumanResource
	default:
		return value
	}
}

func isAllowedSubscriptionAppType(value string) bool {
	switch normalizeSubscriptionAppType(value) {
	case SubscriptionAppTypePOSKasir, SubscriptionAppTypeHumanResource:
		return true
	default:
		return false
	}
}

func roundSubscriptionMoney(value float64) float64 {
	return float64(int64(value*100+0.5)) / 100
}
