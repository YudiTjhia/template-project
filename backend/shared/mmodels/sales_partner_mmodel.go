package mmodels

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"backend/shared/metadata/impl"
	"backend/shared/metadata/interfaces"

	"github.com/jmoiron/sqlx"
)

const salesPartnerStatusActive = "ACTIVE"

// SalesPartnerMModel describes and operates on the sales_partners table.
// Aggregate totals and account links are read through the related query methods.
type SalesPartnerMModel struct {
	impl.MModel

	ID                int64          `json:"id" db:"id"`
	UserID            int64          `json:"userId" db:"user_id"`
	AccountID         int64          `json:"account_id" db:"account_id"`
	Name              string         `json:"name" db:"name"`
	Email             string         `json:"email" db:"email"`
	Phone             string         `json:"phone" db:"phone"`
	ReferralCode      string         `json:"referralCode" db:"referral_code"`
	Status            string         `json:"status" db:"status"`
	CommissionPercent float64        `json:"commissionPercent" db:"commission_percent"`
	CreatedAt         time.Time      `json:"-" db:"created_at"`
	CreatedBy         string         `json:"created_by" db:"created_by"`
	UpdatedAt         sql.NullTime   `json:"-" db:"updated_at"`
	UpdatedBy         sql.NullString `json:"-" db:"updated_by"`
	AuthAccountID     int64          `json:"-" db:"-"`
	AuthEmail         string         `json:"-" db:"-"`

	LinkedAccountsCount int     `json:"linkedAccountsCount" db:"linked_accounts_count"`
	TotalSales          float64 `json:"totalSales" db:"total_sales"`
	TotalCommission     float64 `json:"totalCommission" db:"total_commission"`
}

func NewSalesPartnerMModel(ctx context.Context, db *sqlx.DB) *SalesPartnerMModel {
	return &SalesPartnerMModel{
		Status:            salesPartnerStatusActive,
		CommissionPercent: salesPartnerCommissionPercent(),
		MModel: impl.MModel{
			Context: ctx, Name: "SalesPartner", TableName: "sales_partners", DB: db,
			Fields: []interfaces.IMField{
				impl.MField{Type: "int64", DbType: "INTEGER", JsonTag: "id", DbTag: "id", IsPrimaryKey: true, IsIdentity: true, Required: true, IsNotUpdateField: true},
				impl.MField{Type: "int64", DbType: "BIGINT", JsonTag: "userId", DbTag: "user_id", Default: "0", Required: true, IsNotUpdateField: true},
				impl.MField{Type: "int64", DbType: "BIGINT", JsonTag: "account_id", DbTag: "account_id", Default: "0", Required: true, IsNotUpdateField: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "name", DbTag: "name", MaxLength: 255, Required: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "email", DbTag: "email", MaxLength: 255, Default: "''", Required: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "phone", DbTag: "phone", MaxLength: 100, Default: "''", Required: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "referralCode", DbTag: "referral_code", MaxLength: 100, Required: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "status", DbTag: "status", MaxLength: 30, Default: "'ACTIVE'", Required: true},
				impl.MField{Type: "float64", DbType: "NUMERIC(8,4)", JsonTag: "commissionPercent", DbTag: "commission_percent", Required: true},
				impl.MField{Type: "time.Time", DbType: "TIMESTAMPTZ", JsonTag: "-", DbTag: "created_at", Default: "NOW()", Required: true, IsNotUpdateField: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "created_by", DbTag: "created_by", MaxLength: 255, Default: "''", Required: true, IsNotUpdateField: true},
				impl.MField{Type: "sql.NullTime", DbType: "TIMESTAMPTZ", JsonTag: "-", DbTag: "updated_at", IsNullable: true, IsNotCreateField: true},
				impl.MField{Type: "sql.NullString", DbType: "VARCHAR", JsonTag: "-", DbTag: "updated_by", MaxLength: 255, IsNullable: true, IsNotCreateField: true},
			},
		},
	}
}

// Migrate creates or upgrades sales partner tables and their supporting indexes.
func (m *SalesPartnerMModel) Migrate() error {
	if m == nil {
		return errors.New("sales partner metadata model is nil")
	}
	if m.DB == nil {
		return errors.New("sales partner metadata database is nil")
	}
	tx, err := m.DB.BeginTxx(m.ContextOrBackground(), nil)
	if err != nil {
		return fmt.Errorf("begin sales partners migration: %w", err)
	}
	defer tx.Rollback()

	migrator := impl.NewMMigrator(&m.MModel, tx)
	migrator.SetContext(m.ContextOrBackground())
	if err := migrator.Migrate(); err != nil {
		return err
	}
	statements := []string{
		`ALTER TABLE sales_partners ADD COLUMN IF NOT EXISTS user_id BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE sales_partners ADD COLUMN IF NOT EXISTS account_id BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE sales_partners ADD COLUMN IF NOT EXISTS name VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE sales_partners ADD COLUMN IF NOT EXISTS email VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE sales_partners ADD COLUMN IF NOT EXISTS phone VARCHAR(100) NOT NULL DEFAULT ''`,
		`ALTER TABLE sales_partners ADD COLUMN IF NOT EXISTS referral_code VARCHAR(100) NOT NULL DEFAULT ''`,
		`ALTER TABLE sales_partners ADD COLUMN IF NOT EXISTS status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE'`,
		`ALTER TABLE sales_partners ADD COLUMN IF NOT EXISTS commission_percent NUMERIC(8,4) NOT NULL DEFAULT 10`,
		`ALTER TABLE sales_partners ALTER COLUMN commission_percent SET DEFAULT 10`,
		`ALTER TABLE sales_partners ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`ALTER TABLE sales_partners ADD COLUMN IF NOT EXISTS created_by VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE sales_partners ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ`,
		`ALTER TABLE sales_partners ADD COLUMN IF NOT EXISTS updated_by VARCHAR(255)`,
		`DROP INDEX IF EXISTS idx_sales_partners_account_code`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sales_partners_referral_code ON sales_partners (LOWER(referral_code))`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sales_partners_user_id_global ON sales_partners (user_id) WHERE user_id > 0`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sales_partners_email_global ON sales_partners (LOWER(email)) WHERE email <> ''`,
		`CREATE INDEX IF NOT EXISTS idx_sales_partners_account_id ON sales_partners (account_id)`,
		`CREATE TABLE IF NOT EXISTS partner_accounts (
			id SERIAL PRIMARY KEY,
			partner_id BIGINT NOT NULL REFERENCES sales_partners(id) ON DELETE CASCADE,
			account_id BIGINT NOT NULL,
			status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_by VARCHAR(255) NOT NULL DEFAULT '',
			updated_at TIMESTAMPTZ,
			updated_by VARCHAR(255)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_partner_accounts_partner_account ON partner_accounts (partner_id, account_id)`,
		`CREATE INDEX IF NOT EXISTS idx_partner_accounts_account_id ON partner_accounts (account_id)`,
		`CREATE TABLE IF NOT EXISTS sales_partner_commissions (
			id SERIAL PRIMARY KEY,
			partner_id BIGINT NOT NULL REFERENCES sales_partners(id) ON DELETE CASCADE,
			account_id BIGINT NOT NULL,
			payment_id BIGINT NOT NULL,
			payment_source VARCHAR(50) NOT NULL DEFAULT 'subscription',
			payment_token TEXT NOT NULL DEFAULT '',
			base_amount NUMERIC(18,2) NOT NULL DEFAULT 0,
			commission_percent NUMERIC(8,4) NOT NULL DEFAULT 0,
			commission_amount NUMERIC(18,2) NOT NULL DEFAULT 0,
			status VARCHAR(30) NOT NULL DEFAULT 'EARNED',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_by VARCHAR(255) NOT NULL DEFAULT ''
		)`,
		`ALTER TABLE sales_partner_commissions ADD COLUMN IF NOT EXISTS payment_source VARCHAR(50) NOT NULL DEFAULT 'subscription'`,
		`DROP INDEX IF EXISTS idx_sales_partner_commissions_payment`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sales_partner_commissions_payment ON sales_partner_commissions (payment_source, payment_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sales_partner_commissions_partner ON sales_partner_commissions (partner_id, created_at DESC)`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(m.ContextOrBackground(), statement); err != nil {
			return fmt.Errorf("run sales partners migration statement %q: %w", statement, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit sales partners migration: %w", err)
	}
	return nil
}

func (m *SalesPartnerMModel) ValidateCreate() error {
	if m == nil {
		return errors.New("sales partner metadata model is nil")
	}
	if _, err := m.MModel.ValidateAccountID_(m.AuthAccountID, m.AccountID); err != nil {
		return err
	}
	if strings.TrimSpace(m.Name) == "" {
		return errors.New("nama sales partner wajib diisi")
	}
	if strings.TrimSpace(m.ReferralCode) == "" {
		return errors.New("referral code wajib diisi")
	}
	return m.MModel.ValidateCreate_(m)
}

func (m *SalesPartnerMModel) ValidateID() error {
	if m == nil {
		return errors.New("id sales partner tidak valid")
	}
	return m.MModel.ValidateID_(m.ID, "sales partner")
}

func (m *SalesPartnerMModel) Create() (int64, error) {
	if m == nil {
		return 0, errors.New("sales partner metadata model is nil")
	}
	accountID, err := m.MModel.ValidateAccountID_(m.AuthAccountID, m.AccountID)
	if err != nil {
		return 0, err
	}
	m.AccountID = accountID
	m.Name = strings.TrimSpace(m.Name)
	m.Email = strings.TrimSpace(m.Email)
	m.Phone = strings.TrimSpace(m.Phone)
	m.ReferralCode = strings.ToUpper(strings.TrimSpace(m.ReferralCode))
	if m.ReferralCode == "" {
		m.ReferralCode, err = m.generateReferralCode()
		if err != nil {
			return 0, err
		}
	}
	m.Status = strings.ToUpper(strings.TrimSpace(m.Status))
	if m.Status == "" {
		m.Status = salesPartnerStatusActive
	}
	if m.CommissionPercent <= 0 {
		m.CommissionPercent = salesPartnerCommissionPercent()
	}
	m.CreatedBy = strings.TrimSpace(m.CreatedBy)
	if m.CreatedBy == "" {
		m.CreatedBy = strings.TrimSpace(m.AuthEmail)
	}
	if err := m.ValidateCreate(); err != nil {
		return 0, err
	}
	return m.MModel.Create_(m)
}

func (m *SalesPartnerMModel) Update() error {
	if m == nil {
		return errors.New("sales partner metadata model is nil")
	}
	if err := m.ValidateID(); err != nil {
		return err
	}
	accountID, err := m.MModel.ValidateAccountID_(m.AuthAccountID, m.AccountID)
	if err != nil {
		return err
	}
	m.AccountID = accountID
	m.Name = strings.TrimSpace(m.Name)
	m.Email = strings.TrimSpace(m.Email)
	m.Phone = strings.TrimSpace(m.Phone)
	m.ReferralCode = strings.ToUpper(strings.TrimSpace(m.ReferralCode))
	m.Status = strings.ToUpper(strings.TrimSpace(m.Status))
	if m.Status == "" {
		m.Status = salesPartnerStatusActive
	}
	if m.CommissionPercent <= 0 {
		m.CommissionPercent = salesPartnerCommissionPercent()
	}
	if err := m.MModel.ValidateUpdate_(m); err != nil {
		return err
	}
	if !m.UpdatedBy.Valid {
		m.UpdatedBy = sql.NullString{String: strings.TrimSpace(m.AuthEmail), Valid: true}
	}
	m.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	return m.MModel.Update_(m, "id", "account_id")
}

func (m *SalesPartnerMModel) Delete() error {
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

func (m *SalesPartnerMModel) FindByID() (*SalesPartnerMModel, error) {
	if err := m.ValidateID(); err != nil {
		return nil, err
	}
	accountID, err := m.MModel.ValidateAccountID_(m.AuthAccountID, m.AccountID)
	if err != nil {
		return nil, err
	}
	found := &SalesPartnerMModel{ID: m.ID, AccountID: accountID, AuthAccountID: m.AuthAccountID, MModel: m.MModel}
	if err := m.MModel.FindByID_(found, "id", "account_id"); err != nil {
		return nil, err
	}
	return found, nil
}

func (m *SalesPartnerMModel) FindList_(search string, limit int) ([]SalesPartnerMModel, error) {
	if m == nil {
		return nil, errors.New("sales partner metadata model is nil")
	}
	accountID, err := m.MModel.ValidateAccountID_(m.AuthAccountID, m.AccountID)
	if err != nil {
		return nil, err
	}
	if limit < 0 {
		return nil, errors.New("sales partner list limit cannot be negative")
	}
	items := make([]SalesPartnerMModel, 0)
	err = m.MModel.FindListOrdered(&items, map[string]any{"account_id": accountID}, strings.TrimSpace(search),
		[]string{"name", "email", "referral_code"}, "created_at,id", "DESC", limit)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].MModel = m.MModel
	}
	return items, nil
}

func (m *SalesPartnerMModel) generateReferralCode() (string, error) {
	base := strings.ToUpper(strings.Join(strings.Fields(m.Name), ""))
	runes := []rune(base)
	if len(runes) > 6 {
		base = string(runes[:6])
	}
	if base == "" {
		base = "PARTNER"
	}
	for sequence := 1; ; sequence++ {
		candidate := fmt.Sprintf("%s%d", base, sequence)
		var exists bool
		if err := m.DB.GetContext(m.ContextOrBackground(), &exists,
			`SELECT EXISTS (SELECT 1 FROM sales_partners WHERE LOWER(referral_code)=LOWER($1))`, candidate); err != nil {
			return "", fmt.Errorf("check referral code availability: %w", err)
		}
		if !exists {
			return candidate, nil
		}
	}
}

func salesPartnerCommissionPercent() float64 {
	value, err := strconv.ParseFloat(strings.TrimSpace(os.Getenv("SALES_PARTNER_COMMISSION_PERCENT")), 64)
	if err != nil || value < 0 {
		return 10
	}
	return value
}
