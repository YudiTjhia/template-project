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

// AuthMModel describes an authentication session and its optional login OTP.
type AuthMModel struct {
	impl.MModel

	ID           int64        `json:"id" db:"id"`
	AccountID    int64        `json:"account_id" db:"account_id"`
	UserName     string       `json:"user_name,omitempty" db:"-"`
	Email        string       `json:"email" db:"email"`
	Name         string       `json:"name" db:"name"`
	Token        string       `json:"token,omitempty" db:"token"`
	TokenExpired *time.Time   `json:"token_expired,omitempty" db:"token_expired"`
	IsExpired    bool         `json:"is_expired" db:"is_expired"`
	OTP          string       `json:"otp" db:"otp"`
	OTPExpired   *time.Time   `json:"otp_expired,omitempty" db:"otp_expired"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    sql.NullTime `json:"-" db:"updated_at"`
}

func NewAuthMModel(ctx context.Context, db *sqlx.DB) *AuthMModel {
	return &AuthMModel{MModel: impl.MModel{
		Context: ctx, Name: "Auth", TableName: "auths", DB: db,
		Fields: []interfaces.IMField{
			impl.MField{Type: "int64", DbType: "INTEGER", JsonTag: "id", DbTag: "id", IsPrimaryKey: true, IsIdentity: true, Required: true, IsNotUpdateField: true},
			impl.MField{Type: "int64", DbType: "BIGINT", JsonTag: "account_id", DbTag: "account_id", Default: "0", Required: true},
			impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "email", DbTag: "email", MaxLength: 255, Required: true, IsEmail: true},
			impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "name", DbTag: "name", MaxLength: 255, Default: "''", Required: true},
			impl.MField{Type: "string", DbType: "TEXT", JsonTag: "token", DbTag: "token", Default: "''", Required: true},
			impl.MField{Type: "time.Time", DbType: "TIMESTAMPTZ", JsonTag: "token_expired", DbTag: "token_expired", IsNullable: true},
			impl.MField{Type: "bool", DbType: "BOOLEAN", JsonTag: "is_expired", DbTag: "is_expired", Default: "false", Required: true},
			impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "otp", DbTag: "otp", MaxLength: 20, Default: "''", Required: true},
			impl.MField{Type: "time.Time", DbType: "TIMESTAMPTZ", JsonTag: "otp_expired", DbTag: "otp_expired", IsNullable: true},
			impl.MField{Type: "time.Time", DbType: "TIMESTAMPTZ", JsonTag: "created_at", DbTag: "created_at", Default: "NOW()", Required: true, IsNotUpdateField: true},
			impl.MField{Type: "sql.NullTime", DbType: "TIMESTAMPTZ", JsonTag: "-", DbTag: "updated_at", IsNullable: true, IsNotCreateField: true},
		},
	}}
}

func (m *AuthMModel) Migrate() error {
	if m == nil {
		return errors.New("auth metadata model is nil")
	}
	if m.DB == nil {
		return errors.New("auth metadata database is nil")
	}
	tx, err := m.DB.BeginTxx(m.ContextOrBackground(), nil)
	if err != nil {
		return fmt.Errorf("begin auths migration: %w", err)
	}
	defer tx.Rollback()

	migrator := impl.NewMMigrator(&m.MModel, tx)
	migrator.SetContext(m.ContextOrBackground())
	if err := migrator.Migrate(); err != nil {
		return err
	}
	statements := []string{
		`ALTER TABLE auths ADD COLUMN IF NOT EXISTS account_id BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE auths ADD COLUMN IF NOT EXISTS email VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE auths ADD COLUMN IF NOT EXISTS name VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE auths ADD COLUMN IF NOT EXISTS token TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE auths ADD COLUMN IF NOT EXISTS token_expired TIMESTAMPTZ`,
		`ALTER TABLE auths ADD COLUMN IF NOT EXISTS is_expired BOOLEAN NOT NULL DEFAULT false`,
		`ALTER TABLE auths ADD COLUMN IF NOT EXISTS otp VARCHAR(20) NOT NULL DEFAULT ''`,
		`ALTER TABLE auths ADD COLUMN IF NOT EXISTS otp_expired TIMESTAMPTZ`,
		`ALTER TABLE auths ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`ALTER TABLE auths ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ`,
		`DROP INDEX IF EXISTS idx_auths_account_email`,
		`CREATE INDEX IF NOT EXISTS idx_auths_account_email ON auths (account_id, LOWER(email))`,
		`UPDATE auths
		 SET is_expired = true,
		     token_expired = COALESCE(token_expired, NOW()),
		     updated_at = NOW()
		 WHERE id IN (
			SELECT a.id FROM (
				SELECT a.id,
					ROW_NUMBER() OVER (PARTITION BY a.account_id, LOWER(a.email) ORDER BY a.id DESC) AS session_rank
				FROM auths a
				JOIN users u ON u.account_id = a.account_id AND LOWER(u.user_email) = LOWER(a.email)
				WHERE a.is_expired = false
				  AND COALESCE(NULLIF(LOWER(TRIM(u.login_type)), ''), 'cashier') = 'cashier'
			) a WHERE a.session_rank > 1
		 )`,
		`DROP INDEX IF EXISTS idx_auths_account_email_active`,
		`CREATE INDEX IF NOT EXISTS idx_auths_account_id ON auths (account_id)`,
		`CREATE INDEX IF NOT EXISTS idx_auths_token ON auths (token)`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(m.ContextOrBackground(), statement); err != nil {
			return fmt.Errorf("run auths migration statement %q: %w", statement, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit auths migration: %w", err)
	}
	return nil
}

func (m *AuthMModel) ValidateCreate() error {
	if m == nil {
		return errors.New("auth metadata model is nil")
	}
	m.normalize()
	if m.AccountID <= 0 {
		return errors.New("id akun wajib diisi")
	}
	if m.Email == "" {
		return errors.New("email wajib diisi")
	}
	if m.Name == "" {
		return errors.New("nama wajib diisi")
	}
	return m.MModel.ValidateCreate_(m)
}

func (m *AuthMModel) ValidateID() error {
	if m == nil {
		return errors.New("id auth tidak valid")
	}
	return m.MModel.ValidateID_(m.ID, "auth")
}

func (m *AuthMModel) Create() (int64, error) {
	if err := m.ValidateCreate(); err != nil {
		return 0, err
	}
	return m.MModel.Create_(m)
}

func (m *AuthMModel) Update() error {
	if err := m.ValidateID(); err != nil {
		return err
	}
	m.normalize()
	if err := m.MModel.ValidateUpdate_(m); err != nil {
		return err
	}
	m.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	return m.MModel.Update_(m, "id")
}

func (m *AuthMModel) Delete() error {
	if err := m.ValidateID(); err != nil {
		return err
	}
	return m.MModel.Delete_(m, "id")
}

func (m *AuthMModel) FindByID() (*AuthMModel, error) {
	if err := m.ValidateID(); err != nil {
		return nil, err
	}
	found := &AuthMModel{ID: m.ID, MModel: m.MModel}
	if err := m.MModel.FindByID_(found, "id"); err != nil {
		return nil, err
	}
	return found, nil
}

func (m *AuthMModel) FindByEmail() (*AuthMModel, error) {
	if m == nil {
		return nil, errors.New("auth metadata model is nil")
	}
	if m.DB == nil {
		return nil, errors.New("auth metadata database is nil")
	}
	m.normalize()
	if m.AccountID <= 0 {
		return nil, errors.New("id akun wajib diisi")
	}
	if m.Email == "" {
		return nil, errors.New("email wajib diisi")
	}
	found := &AuthMModel{MModel: m.MModel}
	query := `SELECT id, account_id, email, name, token, token_expired, is_expired, otp, otp_expired, created_at, updated_at
		FROM auths WHERE account_id = $1 AND LOWER(email) = LOWER($2)
		ORDER BY is_expired ASC, id DESC LIMIT 1`
	if err := m.DB.GetContext(m.ContextOrBackground(), found, query, m.AccountID, m.Email); err != nil {
		return nil, err
	}
	return found, nil
}

func (m *AuthMModel) FindByToken() (*AuthMModel, error) {
	if m == nil {
		return nil, errors.New("auth metadata model is nil")
	}
	m.Token = strings.TrimSpace(m.Token)
	if m.Token == "" {
		return nil, errors.New("token wajib diisi")
	}
	found := &AuthMModel{Token: m.Token, MModel: m.MModel}
	if err := m.MModel.FindByFields(found, "token"); err != nil {
		return nil, err
	}
	return found, nil
}

func (m *AuthMModel) ExpireActiveByEmail() error {
	if m == nil {
		return errors.New("auth metadata model is nil")
	}
	if m.DB == nil {
		return errors.New("auth metadata database is nil")
	}
	m.normalize()
	if m.AccountID <= 0 {
		return errors.New("id akun wajib diisi")
	}
	if m.Email == "" {
		return errors.New("email wajib diisi")
	}
	_, err := m.DB.ExecContext(m.ContextOrBackground(), `
		UPDATE auths SET is_expired = true, token_expired = NOW(), updated_at = NOW()
		WHERE account_id = $1 AND LOWER(email) = LOWER($2) AND is_expired = false`, m.AccountID, m.Email)
	return err
}

func (m *AuthMModel) ExpireByToken() error {
	if err := m.validateToken(); err != nil {
		return err
	}
	result, err := m.DB.ExecContext(m.ContextOrBackground(), `
		UPDATE auths SET is_expired = true, token_expired = NOW(), updated_at = NOW()
		WHERE token = $1`, m.Token)
	return requireAuthRowsAffected(result, err)
}

func (m *AuthMModel) ExtendByToken(expiredAt time.Time) error {
	if err := m.validateToken(); err != nil {
		return err
	}
	result, err := m.DB.ExecContext(m.ContextOrBackground(), `
		UPDATE auths SET is_expired = false, token_expired = $2, updated_at = NOW()
		WHERE token = $1`, m.Token, expiredAt)
	return requireAuthRowsAffected(result, err)
}

func (m *AuthMModel) UpsertOTP() (int64, error) {
	if m == nil {
		return 0, errors.New("auth metadata model is nil")
	}
	m.normalize()
	existing, err := m.FindByEmail()
	if err == nil && existing != nil {
		m.ID = existing.ID
		if m.Token == "" {
			m.Token = existing.Token
		}
		if m.TokenExpired == nil {
			m.TokenExpired = existing.TokenExpired
		}
		m.IsExpired = existing.IsExpired
		if err := m.Update(); err != nil {
			return 0, err
		}
		return m.ID, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	return m.Create()
}

func (m *AuthMModel) normalize() {
	if m == nil {
		return
	}
	m.Email = strings.ToLower(strings.TrimSpace(m.Email))
	m.Name = strings.TrimSpace(m.Name)
	m.Token = strings.TrimSpace(m.Token)
	m.OTP = strings.TrimSpace(m.OTP)
}

func (m *AuthMModel) validateToken() error {
	if m == nil {
		return errors.New("auth metadata model is nil")
	}
	if m.DB == nil {
		return errors.New("auth metadata database is nil")
	}
	m.Token = strings.TrimSpace(m.Token)
	if m.Token == "" {
		return errors.New("token wajib diisi")
	}
	return nil
}

func requireAuthRowsAffected(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
