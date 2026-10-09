package mmodels

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend/shared/metadata/impl"
	"backend/shared/metadata/interfaces"

	"github.com/jmoiron/sqlx"
)

// AccountMModel describes account persistence and validation using MModel.
type AccountMModel struct {
	impl.MModel

	ID              int64          `json:"id" db:"id"`
	AccountName     string         `json:"account_name" db:"account_name"`
	CompanyName     string         `json:"company_name" db:"company_name"`
	CompanyLogoURL  string         `json:"company_logo_url" db:"company_logo_url"`
	IsActive        bool           `json:"is_active" db:"is_active"`
	AccountEmail    string         `json:"account_email" db:"account_email"`
	ReferralCode    string         `json:"referral_code" db:"referral_code"`
	Address         string         `json:"address" db:"address"`
	Latitude        float64        `json:"latitude" db:"latitude"`
	Longitude       float64        `json:"longitude" db:"longitude"`
	MobilePhone     string         `json:"mobilephone" db:"mobilephone"`
	Phone           string         `json:"phone" db:"phone"`
	TermsAccepted   bool           `json:"terms_accepted" db:"terms_accepted"`
	TermsAcceptedAt *time.Time     `json:"terms_accepted_at,omitempty" db:"terms_accepted_at"`
	TermsVersion    string         `json:"terms_version" db:"terms_version"`
	OTP             string         `json:"otp" db:"otp"`
	OTPExpired      *time.Time     `json:"otp_expired,omitempty" db:"otp_expired"`
	CreatedAt       time.Time      `json:"-" db:"created_at"`
	CreatedBy       string         `json:"created_by" db:"created_by"`
	UpdatedAt       sql.NullTime   `json:"-" db:"updated_at"`
	UpdatedBy       sql.NullString `json:"-" db:"updated_by"`
	Password        string         `json:"password" db:"-"`
	ConfirmPassword string         `json:"confirm_password" db:"-"`
	PIN             string         `json:"pin" db:"-"`
	ConfirmPIN      string         `json:"confirm_pin" db:"-"`
	AuthUserName    string         `json:"-" db:"-"`
}

func NewAccountMModel(ctx context.Context, db *sqlx.DB) *AccountMModel {
	return &AccountMModel{
		IsActive: true,
		MModel: impl.MModel{
			Context: ctx, Name: "Account", TableName: "accounts", DB: db,
			Fields: []interfaces.IMField{
				impl.MField{Type: "int64", DbType: "INTEGER", JsonTag: "id", DbTag: "id", IsPrimaryKey: true, IsIdentity: true, Required: true, IsNotUpdateField: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "account_name", DbTag: "account_name", MaxLength: 255, Required: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "company_name", DbTag: "company_name", MaxLength: 255, Default: "''", Required: true},
				impl.MField{Type: "string", DbType: "TEXT", JsonTag: "company_logo_url", DbTag: "company_logo_url", Default: "''", Required: true},
				impl.MField{Type: "bool", DbType: "BOOLEAN", JsonTag: "is_active", DbTag: "is_active", Default: "true", Required: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "account_email", DbTag: "account_email", MaxLength: 255, MinLength: 3, Pattern: `^[^@\s]+@[^@\s]+\.[^@\s]+$`, Required: true, IsEmail: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "referral_code", DbTag: "referral_code", MaxLength: 100, Default: "''", Required: true},
				impl.MField{Type: "string", DbType: "TEXT", JsonTag: "address", DbTag: "address", Default: "''", Required: true},
				impl.MField{Type: "float64", DbType: "NUMERIC(12,8)", JsonTag: "latitude", DbTag: "latitude", Default: "0", Required: true},
				impl.MField{Type: "float64", DbType: "NUMERIC(12,8)", JsonTag: "longitude", DbTag: "longitude", Default: "0", Required: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "mobilephone", DbTag: "mobilephone", MaxLength: 100, Required: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "phone", DbTag: "phone", MaxLength: 100, Default: "''", Required: true},
				impl.MField{Type: "bool", DbType: "BOOLEAN", JsonTag: "terms_accepted", DbTag: "terms_accepted", Default: "false", Required: true},
				impl.MField{Type: "time.Time", DbType: "TIMESTAMPTZ", JsonTag: "terms_accepted_at", DbTag: "terms_accepted_at", IsNullable: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "terms_version", DbTag: "terms_version", MaxLength: 50, Default: "''", Required: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "otp", DbTag: "otp", MaxLength: 20, Default: "''", Required: true},
				impl.MField{Type: "time.Time", DbType: "TIMESTAMPTZ", JsonTag: "otp_expired", DbTag: "otp_expired", IsNullable: true},
				impl.MField{Type: "time.Time", DbType: "TIMESTAMPTZ", JsonTag: "-", DbTag: "created_at", Default: "NOW()", Required: true, IsNotUpdateField: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "created_by", DbTag: "created_by", MaxLength: 255, Required: true, IsNotUpdateField: true},
				impl.MField{Type: "sql.NullTime", DbType: "TIMESTAMPTZ", JsonTag: "-", DbTag: "updated_at", IsNullable: true, IsNotCreateField: true},
				impl.MField{Type: "sql.NullString", DbType: "VARCHAR", JsonTag: "-", DbTag: "updated_by", MaxLength: 255, IsNullable: true, IsNotCreateField: true},
			},
		},
	}
}

// Migrate creates or upgrades the accounts table and its indexes.
func (m *AccountMModel) Migrate() error {
	if m == nil {
		return errors.New("account metadata model is nil")
	}
	if m.DB == nil {
		return errors.New("account metadata database is nil")
	}
	tx, err := m.DB.BeginTxx(m.ContextOrBackground(), nil)
	if err != nil {
		return fmt.Errorf("begin accounts migration: %w", err)
	}
	defer tx.Rollback()

	migrator := impl.NewMMigrator(&m.MModel, tx)
	migrator.SetContext(m.ContextOrBackground())
	if err := migrator.Migrate(); err != nil {
		return err
	}
	statements := []string{
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS account_name VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS company_name VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS company_logo_url TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS account_email VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS referral_code VARCHAR(100) NOT NULL DEFAULT ''`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS address TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS latitude NUMERIC(12,8) NOT NULL DEFAULT 0`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS longitude NUMERIC(12,8) NOT NULL DEFAULT 0`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS mobilephone VARCHAR(100) NOT NULL DEFAULT ''`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS phone VARCHAR(100) NOT NULL DEFAULT ''`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS terms_accepted BOOLEAN NOT NULL DEFAULT false`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS terms_accepted_at TIMESTAMPTZ`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS terms_version VARCHAR(50) NOT NULL DEFAULT ''`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS otp VARCHAR(20) NOT NULL DEFAULT ''`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS otp_expired TIMESTAMPTZ`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS created_by VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ`,
		`ALTER TABLE accounts ADD COLUMN IF NOT EXISTS updated_by VARCHAR(255)`,
		`CREATE INDEX IF NOT EXISTS idx_accounts_name ON accounts (LOWER(account_name))`,
		`DROP INDEX IF EXISTS idx_accounts_account_email`,
		`DROP INDEX IF EXISTS idx_accounts_account_mobilephone`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_email ON accounts (LOWER(account_email))`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_mobilephone ON accounts (LOWER(mobilephone)) WHERE mobilephone <> ''`,
		`CREATE INDEX IF NOT EXISTS idx_accounts_active ON accounts (is_active)`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(m.ContextOrBackground(), statement); err != nil {
			return fmt.Errorf("run accounts migration statement %q: %w", statement, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit accounts migration: %w", err)
	}
	return nil
}

func (m *AccountMModel) ValidateCreate() error {
	if m == nil {
		return errors.New("account metadata model is nil")
	}
	m.normalize()
	if err := m.validateRequiredFields(); err != nil {
		return err
	}
	return m.MModel.ValidateCreate_(m)
}

func (m *AccountMModel) ValidateID() error {
	if m == nil {
		return errors.New("id account tidak valid")
	}
	return m.MModel.ValidateID_(m.ID, "account")
}

func (m *AccountMModel) Create() (int64, error) {
	if m == nil {
		return 0, errors.New("account metadata model is nil")
	}
	m.normalize()
	if err := m.ValidateCreate(); err != nil {
		return 0, err
	}
	if err := m.ensureUniqueFields(0); err != nil {
		return 0, err
	}
	m.CreatedBy = strings.TrimSpace(m.CreatedBy)
	if m.CreatedBy == "" {
		m.CreatedBy = strings.TrimSpace(m.AuthUserName)
	}
	if m.CreatedBy == "" {
		return 0, errors.New("created_by wajib diisi")
	}
	return m.MModel.Create_(m)
}

func (m *AccountMModel) Update() error {
	if m == nil {
		return errors.New("account metadata model is nil")
	}
	m.normalize()
	if err := m.ValidateID(); err != nil {
		return err
	}
	if err := m.validateRequiredFields(); err != nil {
		return err
	}
	if err := m.MModel.ValidateUpdate_(m); err != nil {
		return err
	}
	if err := m.ensureUniqueFields(m.ID); err != nil {
		return err
	}
	if !m.UpdatedBy.Valid || strings.TrimSpace(m.UpdatedBy.String) == "" {
		m.UpdatedBy = sql.NullString{String: strings.TrimSpace(m.AuthUserName), Valid: strings.TrimSpace(m.AuthUserName) != ""}
	}
	m.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	return m.MModel.Update_(m, "id")
}

func (m *AccountMModel) Delete() error {
	if err := m.ValidateID(); err != nil {
		return err
	}
	return m.MModel.Delete_(m, "id")
}

func (m *AccountMModel) FindByID() (*AccountMModel, error) {
	if err := m.ValidateID(); err != nil {
		return nil, err
	}
	found := &AccountMModel{ID: m.ID, MModel: m.MModel}
	if err := m.MModel.FindByID_(found, "id"); err != nil {
		return nil, err
	}
	return found, nil
}

func (m *AccountMModel) FindByEmail_(email string) (*AccountMModel, error) {
	if m == nil {
		return nil, errors.New("account metadata model is nil")
	}
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, errors.New("email account wajib diisi")
	}
	found := &AccountMModel{AccountEmail: email, MModel: m.MModel}
	if err := m.MModel.FindByAnyField(found, "account_email", email); err != nil {
		return nil, err
	}
	return found, nil
}

func (m *AccountMModel) FindByOTP_(otp string) (*AccountMModel, error) {
	if m == nil {
		return nil, errors.New("account metadata model is nil")
	}
	otp = strings.TrimSpace(otp)
	if otp == "" {
		return nil, errors.New("otp wajib diisi")
	}
	found := &AccountMModel{OTP: otp, MModel: m.MModel}
	if err := m.MModel.FindByFields(found, "otp"); err != nil {
		return nil, err
	}
	return found, nil
}

func (m *AccountMModel) FindList_(search string, limit int) ([]AccountMModel, error) {
	if m == nil {
		return nil, errors.New("account metadata model is nil")
	}
	if limit < 0 {
		return nil, errors.New("account list limit cannot be negative")
	}
	accounts := make([]AccountMModel, 0)
	err := m.MModel.FindListOrdered(&accounts, nil, strings.TrimSpace(search),
		[]string{"account_name", "company_name", "account_email", "address", "mobilephone", "phone"}, "account_name", "ASC", limit)
	if err != nil {
		return nil, err
	}
	for i := range accounts {
		accounts[i].MModel = m.MModel
	}
	return accounts, nil
}

func (m *AccountMModel) FindRecentList_(search string, limit int) ([]AccountMModel, error) {
	if limit <= 0 {
		limit = 5
	}
	if m == nil {
		return nil, errors.New("account metadata model is nil")
	}
	if limit <= 0 {
		limit = 5
	}
	accounts := make([]AccountMModel, 0)
	err := m.MModel.FindListOrdered(&accounts, nil, strings.TrimSpace(search),
		[]string{"account_name", "company_name", "account_email", "address", "mobilephone", "phone"}, "created_at,id", "DESC", limit)
	if err != nil {
		return nil, err
	}
	for i := range accounts {
		accounts[i].MModel = m.MModel
	}
	return accounts, nil
}

func (m *AccountMModel) normalize() {
	m.AccountName = strings.TrimSpace(m.AccountName)
	m.CompanyName = strings.TrimSpace(m.CompanyName)
	m.CompanyLogoURL = strings.TrimSpace(m.CompanyLogoURL)
	m.AccountEmail = strings.ToLower(strings.TrimSpace(m.AccountEmail))
	m.ReferralCode = strings.ToUpper(strings.TrimSpace(m.ReferralCode))
	m.Address = strings.TrimSpace(m.Address)
	m.MobilePhone = strings.TrimSpace(m.MobilePhone)
	m.Phone = strings.TrimSpace(m.Phone)
	m.TermsVersion = strings.TrimSpace(m.TermsVersion)
	m.OTP = strings.TrimSpace(m.OTP)
}

func (m *AccountMModel) validateRequiredFields() error {
	if strings.TrimSpace(m.AccountName) == "" {
		return errors.New("nama account wajib diisi")
	}
	if strings.TrimSpace(m.AccountEmail) == "" {
		return errors.New("email account wajib diisi")
	}
	if strings.TrimSpace(m.MobilePhone) == "" {
		return errors.New("mobilephone account wajib diisi")
	}
	return nil
}

func (m *AccountMModel) ensureUniqueFields(excludeID int64) error {
	exists, err := m.MModel.ExistsByField("account_email", m.AccountEmail, "id", excludeID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("email account sudah digunakan")
	}
	exists, err = m.MModel.ExistsByField("mobilephone", m.MobilePhone, "id", excludeID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("mobilephone account sudah digunakan")
	}
	return nil
}
