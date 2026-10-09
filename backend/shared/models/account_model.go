package models

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type BaseModel struct {
	Context   context.Context `json:"-" db:"-"`
	DB        *sqlx.DB        `json:"-" db:"-"`
	Auth      Auth            `json:"-" db:"-"`
	CreatedAt time.Time       `json:"-" db:"created_at"`
	CreatedBy string          `json:"-" db:"created_by"`
	UpdatedAt *time.Time      `json:"-" db:"updated_at"`
	UpdatedBy *string         `json:"-" db:"updated_by"`
}

func LogSQLArgs(args map[string]interface{}) {
	if len(args) == 0 {
		return
	}

	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}

	// Keep SQL arg output stable for easier debugging and test snapshots.
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, args[k]))
	}

	log.Println("[SQL ARGS]", strings.Join(parts, ", "))
}

func nextSQLPlaceholder(args []interface{}) string {
	return fmt.Sprintf("$%d", len(args))
}

type Account struct {
	BaseModel

	ID              int64      `json:"id" db:"id"`
	AccountName     string     `json:"account_name" db:"account_name"`
	CompanyName     string     `json:"company_name" db:"company_name"`
	CompanyLogoURL  string     `json:"company_logo_url" db:"company_logo_url"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	AccountEmail    string     `json:"account_email" db:"account_email"`
	ReferralCode    string     `json:"referral_code" db:"referral_code"`
	Address         string     `json:"address" db:"address"`
	Latitude        float64    `json:"latitude" db:"latitude"`
	Longitude       float64    `json:"longitude" db:"longitude"`
	MobilePhone     string     `json:"mobilephone" db:"mobilephone"`
	Phone           string     `json:"phone" db:"phone"`
	Password        string     `json:"password" db:"-"`
	ConfirmPassword string     `json:"confirm_password" db:"-"`
	PIN             string     `json:"pin" db:"-"`
	ConfirmPIN      string     `json:"confirm_pin" db:"-"`
	TermsAccepted   bool       `json:"terms_accepted" db:"terms_accepted"`
	TermsAcceptedAt *time.Time `json:"terms_accepted_at,omitempty" db:"terms_accepted_at"`
	TermsVersion    string     `json:"terms_version" db:"terms_version"`
	OTP             string     `json:"otp" db:"otp"`
	OTPExpired      *time.Time `json:"otp_expired,omitempty" db:"otp_expired"`
}

const accountSelectColumns = `a.id, a.account_name, a.company_name, a.company_logo_url, a.is_active, a.account_email, a.referral_code, a.address, a.latitude, a.longitude, a.mobilephone, a.phone, a.terms_accepted, a.terms_accepted_at, a.terms_version, a.otp, a.otp_expired,
a.created_at, a.created_by, a.updated_at, a.updated_by`

func NewAccount(db *sqlx.DB, auth Auth) Account {
	account := Account{
		BaseModel: BaseModel{DB: db, Auth: auth},
	}
	account.CreatedBy = auth.UserName
	account.IsActive = true
	return account
}

func (m Account) Migrate() {
	sql := `CREATE TABLE IF NOT EXISTS accounts (
	id SERIAL PRIMARY KEY,
	account_name VARCHAR(255) NOT NULL,
	company_name VARCHAR(255) NOT NULL DEFAULT '',
	company_logo_url TEXT NOT NULL DEFAULT '',
	is_active BOOLEAN NOT NULL DEFAULT true,
	account_email VARCHAR(255) NOT NULL DEFAULT '',
	address TEXT NOT NULL DEFAULT '',
	latitude NUMERIC(12,8) NOT NULL DEFAULT 0,
	longitude NUMERIC(12,8) NOT NULL DEFAULT 0,
	mobilephone VARCHAR(100) NOT NULL DEFAULT '',
	phone VARCHAR(100) NOT NULL DEFAULT '',
	terms_accepted BOOLEAN NOT NULL DEFAULT false,
	terms_accepted_at TIMESTAMPTZ,
	terms_version VARCHAR(50) NOT NULL DEFAULT '',
	otp VARCHAR(20) NOT NULL DEFAULT '',
	otp_expired TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_by VARCHAR(255) NOT NULL,
	updated_at TIMESTAMPTZ,
	updated_by VARCHAR(255)
);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `CREATE INDEX IF NOT EXISTS idx_accounts_name ON accounts (LOWER(account_name));`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `DROP INDEX IF EXISTS idx_accounts_account_email;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `DROP INDEX IF EXISTS idx_accounts_account_mobilephone;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_email ON accounts (LOWER(account_email));`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_mobilephone ON accounts (LOWER(mobilephone)) WHERE mobilephone <> '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `CREATE INDEX IF NOT EXISTS idx_accounts_active ON accounts (is_active);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE accounts ADD COLUMN IF NOT EXISTS account_name VARCHAR(255) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE accounts ADD COLUMN IF NOT EXISTS company_name VARCHAR(255) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE accounts ADD COLUMN IF NOT EXISTS company_logo_url TEXT NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE accounts ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE accounts ADD COLUMN IF NOT EXISTS account_email VARCHAR(255) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE accounts ADD COLUMN IF NOT EXISTS referral_code VARCHAR(100) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE accounts ADD COLUMN IF NOT EXISTS address TEXT NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE accounts ADD COLUMN IF NOT EXISTS latitude NUMERIC(12,8) NOT NULL DEFAULT 0;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE accounts ADD COLUMN IF NOT EXISTS longitude NUMERIC(12,8) NOT NULL DEFAULT 0;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE accounts ADD COLUMN IF NOT EXISTS mobilephone VARCHAR(100) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE accounts ADD COLUMN IF NOT EXISTS phone VARCHAR(100) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE accounts ADD COLUMN IF NOT EXISTS terms_accepted BOOLEAN NOT NULL DEFAULT false;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE accounts ADD COLUMN IF NOT EXISTS terms_accepted_at TIMESTAMPTZ;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE accounts ADD COLUMN IF NOT EXISTS terms_version VARCHAR(50) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE accounts ADD COLUMN IF NOT EXISTS otp VARCHAR(20) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE accounts ADD COLUMN IF NOT EXISTS otp_expired TIMESTAMPTZ;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
}

func (m Account) TableName() string {
	return "accounts"
}

func (m Account) ValidateCreate() error {
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

func (m Account) ValidateID() error {
	if m.ID <= 0 {
		return errors.New("id account tidak valid")
	}
	return nil
}

func (m Account) Create() (int64, error) {
	m.AccountName = strings.TrimSpace(m.AccountName)
	m.CompanyName = strings.TrimSpace(m.CompanyName)
	m.CompanyLogoURL = strings.TrimSpace(m.CompanyLogoURL)
	m.AccountEmail = strings.TrimSpace(m.AccountEmail)
	m.ReferralCode = strings.ToUpper(strings.TrimSpace(m.ReferralCode))
	m.Address = strings.TrimSpace(m.Address)
	m.MobilePhone = strings.TrimSpace(m.MobilePhone)
	m.Phone = strings.TrimSpace(m.Phone)
	m.TermsVersion = strings.TrimSpace(m.TermsVersion)
	m.OTP = strings.TrimSpace(m.OTP)

	if err := m.ValidateCreate(); err != nil {
		return 0, err
	}
	if err := m.ensureEmailAvailable(0); err != nil {
		return 0, err
	}
	if err := m.ensureMobilephoneAvailable(0); err != nil {
		return 0, err
	}

	createdBy := m.CreatedBy
	if createdBy == "" {
		createdBy = m.Auth.UserName
	}
	if createdBy == "" {
		return 0, errors.New("created_by wajib diisi")
	}

	sql := `INSERT INTO accounts
	(account_name, company_name, company_logo_url, is_active, account_email, referral_code, address, latitude, longitude, mobilephone, phone, terms_accepted, terms_accepted_at, terms_version, otp, otp_expired, created_by)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	RETURNING id`

	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{
		"account_name":   m.AccountName,
		"company_name":   m.CompanyName,
		"company_logo":   m.CompanyLogoURL,
		"is_active":      m.IsActive,
		"account_email":  m.AccountEmail,
		"address":        m.Address,
		"latitude":       m.Latitude,
		"longitude":      m.Longitude,
		"mobilephone":    m.MobilePhone,
		"phone":          m.Phone,
		"terms_accepted": m.TermsAccepted,
		"terms_version":  m.TermsVersion,
		"otp":            m.OTP,
		"otp_expired":    m.OTPExpired,
		"created_by":     createdBy,
	})

	var id int64
	err := m.DB.QueryRow(sql, m.AccountName, m.CompanyName, m.CompanyLogoURL, m.IsActive, m.AccountEmail, m.ReferralCode, m.Address, m.Latitude, m.Longitude, m.MobilePhone, m.Phone, m.TermsAccepted, m.TermsAcceptedAt, m.TermsVersion, m.OTP, m.OTPExpired, createdBy).Scan(&id)
	return id, err
}

func (m Account) Update() error {
	m.AccountName = strings.TrimSpace(m.AccountName)
	m.CompanyName = strings.TrimSpace(m.CompanyName)
	m.CompanyLogoURL = strings.TrimSpace(m.CompanyLogoURL)
	m.AccountEmail = strings.TrimSpace(m.AccountEmail)
	m.ReferralCode = strings.ToUpper(strings.TrimSpace(m.ReferralCode))
	m.Address = strings.TrimSpace(m.Address)
	m.MobilePhone = strings.TrimSpace(m.MobilePhone)
	m.Phone = strings.TrimSpace(m.Phone)
	m.TermsVersion = strings.TrimSpace(m.TermsVersion)
	m.OTP = strings.TrimSpace(m.OTP)

	if err := m.ValidateID(); err != nil {
		return err
	}

	if err := m.ValidateCreate(); err != nil {
		return err
	}
	if err := m.ensureEmailAvailable(m.ID); err != nil {
		return err
	}
	if err := m.ensureMobilephoneAvailable(m.ID); err != nil {
		return err
	}

	updatedBy := m.UpdatedBy
	if updatedBy == nil || *updatedBy == "" {
		if m.Auth.UserName != "" {
			updatedBy = &m.Auth.UserName
		}
	}

	sql := `UPDATE accounts
	SET account_name = $1,
		company_name = $2,
		company_logo_url = $3,
		is_active = $4,
		account_email = $5,
		referral_code = $6,
		address = $7,
		latitude = $8,
		longitude = $9,
		mobilephone = $10,
		phone = $11,
		terms_accepted = $12,
		terms_accepted_at = $13,
		terms_version = $14,
		otp = $15,
		otp_expired = $16,
		updated_at = NOW(),
		updated_by = $17
	WHERE id = $18`

	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{
		"account_name":   m.AccountName,
		"company_name":   m.CompanyName,
		"company_logo":   m.CompanyLogoURL,
		"is_active":      m.IsActive,
		"account_email":  m.AccountEmail,
		"address":        m.Address,
		"latitude":       m.Latitude,
		"longitude":      m.Longitude,
		"mobilephone":    m.MobilePhone,
		"phone":          m.Phone,
		"terms_accepted": m.TermsAccepted,
		"terms_version":  m.TermsVersion,
		"otp":            m.OTP,
		"otp_expired":    m.OTPExpired,
		"updated_by":     updatedBy,
		"id":             m.ID,
	})

	_, err := m.DB.Exec(sql, m.AccountName, m.CompanyName, m.CompanyLogoURL, m.IsActive, m.AccountEmail, m.ReferralCode, m.Address, m.Latitude, m.Longitude, m.MobilePhone, m.Phone, m.TermsAccepted, m.TermsAcceptedAt, m.TermsVersion, m.OTP, m.OTPExpired, updatedBy, m.ID)
	return err
}

func (m Account) ensureEmailAvailable(excludeID int64) error {
	email := strings.ToLower(strings.TrimSpace(m.AccountEmail))
	if email == "" {
		return errors.New("email account wajib diisi")
	}

	sql := `SELECT id FROM accounts WHERE LOWER(account_email) = $1`
	args := []interface{}{email}
	if excludeID > 0 {
		sql += ` AND id <> $2`
		args = append(args, excludeID)
	}

	var existingID int64
	err := m.DB.Get(&existingID, sql, args...)
	if err == nil {
		return errors.New("email account sudah digunakan")
	}
	if strings.Contains(strings.ToLower(err.Error()), "no rows") {
		return nil
	}

	return err
}

func (m Account) ensureMobilephoneAvailable(excludeID int64) error {
	mobilephone := strings.ToLower(strings.TrimSpace(m.MobilePhone))
	if mobilephone == "" {
		return errors.New("mobilephone account wajib diisi")
	}

	sql := `SELECT id FROM accounts WHERE LOWER(mobilephone) = $1`
	args := []interface{}{mobilephone}
	if excludeID > 0 {
		sql += ` AND id <> $2`
		args = append(args, excludeID)
	}

	var existingID int64
	err := m.DB.Get(&existingID, sql, args...)
	if err == nil {
		return errors.New("mobilephone account sudah digunakan")
	}
	if strings.Contains(strings.ToLower(err.Error()), "no rows") {
		return nil
	}

	return err
}

func (m Account) Delete() error {
	if err := m.ValidateID(); err != nil {
		return err
	}

	sql := `DELETE FROM accounts WHERE id = $1`
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"id": m.ID})

	_, err := m.DB.Exec(sql, m.ID)
	return err
}

func (m Account) FindByID() (*Account, error) {
	if err := m.ValidateID(); err != nil {
		return nil, err
	}

	sql := accountSelectQuery(`a.id = $1`)
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"id": m.ID})

	var account Account
	err := m.DB.Get(&account, sql, m.ID)
	account.DB = m.DB
	account.Auth = m.Auth
	return &account, err
}

func (m Account) FindByEmail() (*Account, error) {
	m.AccountEmail = strings.TrimSpace(m.AccountEmail)
	if m.AccountEmail == "" {
		return nil, errors.New("email account wajib diisi")
	}

	sql := accountSelectQuery(`a.account_email = $1`)
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"account_email": m.AccountEmail})

	var account Account
	err := m.DB.Get(&account, sql, m.AccountEmail)
	account.DB = m.DB
	account.Auth = m.Auth
	return &account, err
}

func (m Account) FindByOTP(otp string) (*Account, error) {
	otp = strings.TrimSpace(otp)
	if otp == "" {
		return nil, errors.New("otp wajib diisi")
	}

	sql := accountSelectQuery(`a.otp = $1`)
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"otp": otp})

	var account Account
	err := m.DB.Get(&account, sql, otp)
	account.DB = m.DB
	account.Auth = m.Auth
	return &account, err
}

func (m Account) FindList(search string) ([]Account, error) {
	sql := accountSelectQuery(`true`)
	args := []interface{}{}
	where := []string{}

	if search != "" {
		args = append(args, search)
		placeholder := nextSQLPlaceholder(args)
		where = append(where, `(a.account_name ILIKE '%' || `+placeholder+` || '%' OR a.company_name ILIKE '%' || `+placeholder+` || '%' OR a.account_email ILIKE '%' || `+placeholder+` || '%' OR a.address ILIKE '%' || `+placeholder+` || '%' OR a.mobilephone ILIKE '%' || `+placeholder+` || '%' OR a.phone ILIKE '%' || `+placeholder+` || '%')`)
	}

	if len(where) > 0 {
		sql += " AND " + strings.Join(where, " AND ")
	}

	sql += ` ORDER BY a.account_name ASC LIMIT 50`

	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"search": search})

	var accounts []Account
	err := m.DB.Select(&accounts, sql, args...)
	return accounts, err
}

func (m Account) FindAll() ([]Account, error) {
	sql := accountSelectQuery(`true`)
	sql += ` ORDER BY a.account_name ASC`

	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{})

	var accounts []Account
	err := m.DB.Select(&accounts, sql)
	return accounts, err
}

func (m Account) FindRecentList(limit int, search string) ([]Account, error) {
	if limit <= 0 {
		limit = 5
	}

	sql := accountSelectQuery(`true`)
	args := []interface{}{}
	where := []string{}

	if search != "" {
		args = append(args, search)
		placeholder := nextSQLPlaceholder(args)
		where = append(where, `(a.account_name ILIKE '%' || `+placeholder+` || '%' OR a.company_name ILIKE '%' || `+placeholder+` || '%' OR a.account_email ILIKE '%' || `+placeholder+` || '%' OR a.address ILIKE '%' || `+placeholder+` || '%' OR a.mobilephone ILIKE '%' || `+placeholder+` || '%' OR a.phone ILIKE '%' || `+placeholder+` || '%')`)
	}

	if len(where) > 0 {
		sql += " AND " + strings.Join(where, " AND ")
	}

	sql += fmt.Sprintf(` ORDER BY a.created_at DESC, a.id DESC LIMIT %d`, limit)

	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"search": search, "limit": limit})

	var accounts []Account
	err := m.DB.Select(&accounts, sql, args...)
	return accounts, err
}

func accountSelectQuery(where string) string {
	return `SELECT ` + accountSelectColumns + `
	FROM accounts a
	WHERE ` + where
}
