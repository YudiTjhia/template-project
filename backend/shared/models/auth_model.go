package models

import (
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Auth struct {
	DB           *sqlx.DB   `db:"-" json:"-"`
	ID           int64      `json:"id" db:"id"`
	AccountID    int64      `json:"account_id" db:"account_id"`
	UserName     string     `json:"user_name,omitempty" db:"-"`
	Email        string     `json:"email" db:"email"`
	Name         string     `json:"name" db:"name"`
	Token        string     `json:"token,omitempty" db:"token"`
	TokenExpired *time.Time `json:"token_expired,omitempty" db:"token_expired"`
	IsExpired    bool       `json:"is_expired" db:"is_expired"`
	OTP          string     `json:"otp" db:"otp"`
	OTPExpired   *time.Time `json:"otp_expired,omitempty" db:"otp_expired"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

const authSelectColumns = `a.id, a.account_id, a.email, a.name, a.token, a.token_expired, a.is_expired, a.otp, a.otp_expired, a.created_at, a.updated_at`

func NewAuth(db *sqlx.DB) Auth {
	return Auth{DB: db}
}

func (m Auth) Migrate() {
	sql := `CREATE TABLE IF NOT EXISTS auths (
	id SERIAL PRIMARY KEY,
	account_id BIGINT NOT NULL DEFAULT 0,
	email VARCHAR(255) NOT NULL,
	name VARCHAR(255) NOT NULL DEFAULT '',
	token TEXT NOT NULL DEFAULT '',
	token_expired TIMESTAMPTZ,
	is_expired BOOLEAN NOT NULL DEFAULT false,
	otp VARCHAR(20) NOT NULL DEFAULT '',
	otp_expired TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ
);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE auths ADD COLUMN IF NOT EXISTS account_id BIGINT NOT NULL DEFAULT 0;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE auths ADD COLUMN IF NOT EXISTS email VARCHAR(255) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE auths ADD COLUMN IF NOT EXISTS name VARCHAR(255) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE auths ADD COLUMN IF NOT EXISTS token TEXT NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE auths ADD COLUMN IF NOT EXISTS token_expired TIMESTAMPTZ;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE auths ADD COLUMN IF NOT EXISTS is_expired BOOLEAN NOT NULL DEFAULT false;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE auths ADD COLUMN IF NOT EXISTS otp VARCHAR(20) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE auths ADD COLUMN IF NOT EXISTS otp_expired TIMESTAMPTZ;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `DROP INDEX IF EXISTS idx_auths_account_email;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `CREATE INDEX IF NOT EXISTS idx_auths_account_email ON auths (account_id, LOWER(email));`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `UPDATE auths
	SET is_expired = true,
		token_expired = COALESCE(token_expired, NOW()),
		updated_at = NOW()
	WHERE id IN (
		SELECT a.id
		FROM (
			SELECT a.id,
				ROW_NUMBER() OVER (PARTITION BY a.account_id, LOWER(a.email) ORDER BY a.id DESC) AS session_rank
			FROM auths a
			JOIN users u ON u.account_id = a.account_id AND LOWER(u.user_email) = LOWER(a.email)
			WHERE a.is_expired = false AND COALESCE(NULLIF(LOWER(TRIM(u.login_type)), ''), 'cashier') = 'cashier'
		) a
		WHERE a.session_rank > 1
	);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `DROP INDEX IF EXISTS idx_auths_account_email_active;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `CREATE INDEX IF NOT EXISTS idx_auths_account_id ON auths (account_id);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `CREATE INDEX IF NOT EXISTS idx_auths_token ON auths (token);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
}

func (m Auth) TableName() string {
	return "auths"
}

func (m Auth) ValidateCreate() error {
	if m.AccountID <= 0 {
		return errors.New("id akun wajib diisi")
	}

	if strings.TrimSpace(m.Email) == "" {
		return errors.New("email wajib diisi")
	}

	if strings.TrimSpace(m.Name) == "" {
		return errors.New("nama wajib diisi")
	}

	return nil
}

func (m Auth) ValidateID() error {
	if m.ID <= 0 {
		return errors.New("id auth tidak valid")
	}
	return nil
}

func (m Auth) Create() (int64, error) {
	m.Email = strings.ToLower(strings.TrimSpace(m.Email))
	m.Name = strings.TrimSpace(m.Name)
	m.Token = strings.TrimSpace(m.Token)
	m.OTP = strings.TrimSpace(m.OTP)

	if err := m.ValidateCreate(); err != nil {
		return 0, err
	}

	sql := `INSERT INTO auths
	(account_id, email, name, token, token_expired, is_expired, otp, otp_expired)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING id`

	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{
		"account_id":    m.AccountID,
		"email":         m.Email,
		"name":          m.Name,
		"token":         m.Token,
		"token_expired": m.TokenExpired,
		"is_expired":    m.IsExpired,
		"otp":           m.OTP,
		"otp_expired":   m.OTPExpired,
	})

	var id int64
	err := m.DB.QueryRow(sql, m.AccountID, m.Email, m.Name, m.Token, m.TokenExpired, m.IsExpired, m.OTP, m.OTPExpired).Scan(&id)
	return id, err
}

func (m Auth) Update() error {
	if err := m.ValidateID(); err != nil {
		return err
	}

	m.Email = strings.ToLower(strings.TrimSpace(m.Email))
	m.Name = strings.TrimSpace(m.Name)
	m.Token = strings.TrimSpace(m.Token)
	m.OTP = strings.TrimSpace(m.OTP)

	sql := `UPDATE auths
	SET account_id = $1,
		email = $2,
		name = $3,
		token = $4,
		token_expired = $5,
		is_expired = $6,
		otp = $7,
		otp_expired = $8,
		updated_at = NOW()
	WHERE id = $9`

	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{
		"account_id":    m.AccountID,
		"email":         m.Email,
		"name":          m.Name,
		"token":         m.Token,
		"token_expired": m.TokenExpired,
		"is_expired":    m.IsExpired,
		"otp":           m.OTP,
		"otp_expired":   m.OTPExpired,
		"id":            m.ID,
	})

	_, err := m.DB.Exec(sql, m.AccountID, m.Email, m.Name, m.Token, m.TokenExpired, m.IsExpired, m.OTP, m.OTPExpired, m.ID)
	return err
}

func (m Auth) Delete() error {
	if err := m.ValidateID(); err != nil {
		return err
	}

	sql := `DELETE FROM auths WHERE id = $1`
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"id": m.ID})

	_, err := m.DB.Exec(sql, m.ID)
	return err
}

func (m Auth) FindByID() (*Auth, error) {
	if err := m.ValidateID(); err != nil {
		return nil, err
	}

	sql := authSelectQuery(`a.id = $1`)
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"id": m.ID})

	var auth Auth
	err := m.DB.Get(&auth, sql, m.ID)
	auth.DB = m.DB
	return &auth, err
}

func (m Auth) FindByEmail() (*Auth, error) {
	m.Email = strings.ToLower(strings.TrimSpace(m.Email))
	if m.AccountID <= 0 {
		return nil, errors.New("id akun wajib diisi")
	}
	if m.Email == "" {
		return nil, errors.New("email wajib diisi")
	}

	sql := authSelectQuery(`a.account_id = $1 AND LOWER(a.email) = LOWER($2)
		ORDER BY a.is_expired ASC, a.id DESC
		LIMIT 1`)
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"account_id": m.AccountID, "email": m.Email})

	var auth Auth
	err := m.DB.Get(&auth, sql, m.AccountID, m.Email)
	auth.DB = m.DB
	return &auth, err
}

func (m Auth) FindByToken() (*Auth, error) {
	m.Token = strings.TrimSpace(m.Token)
	if m.Token == "" {
		return nil, errors.New("token wajib diisi")
	}

	sql := authSelectQuery(`a.token = $1`)
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"token": m.Token})

	var auth Auth
	err := m.DB.Get(&auth, sql, m.Token)
	auth.DB = m.DB
	return &auth, err
}

func (m Auth) ExpireActiveByEmail() error {
	m.Email = strings.ToLower(strings.TrimSpace(m.Email))
	if m.AccountID <= 0 {
		return errors.New("id akun wajib diisi")
	}
	if m.Email == "" {
		return errors.New("email wajib diisi")
	}

	query := `UPDATE auths
	SET is_expired = true,
		token_expired = NOW(),
		updated_at = NOW()
	WHERE account_id = $1
		AND LOWER(email) = LOWER($2)
		AND is_expired = false`

	log.Println("[SQL]=", query)
	LogSQLArgs(map[string]interface{}{"account_id": m.AccountID, "email": m.Email})

	_, err := m.DB.Exec(query, m.AccountID, m.Email)
	return err
}

func (m Auth) ExpireByToken() error {
	m.Token = strings.TrimSpace(m.Token)
	if m.Token == "" {
		return errors.New("token wajib diisi")
	}

	query := `UPDATE auths
	SET is_expired = true,
		token_expired = NOW(),
		updated_at = NOW()
	WHERE token = $1`

	log.Println("[SQL]=", query)
	LogSQLArgs(map[string]interface{}{"token": m.Token})

	result, err := m.DB.Exec(query, m.Token)
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

func (m Auth) ExtendByToken(expiredAt time.Time) error {
	m.Token = strings.TrimSpace(m.Token)
	if m.Token == "" {
		return errors.New("token wajib diisi")
	}

	query := `UPDATE auths
	SET is_expired = false,
		token_expired = $2,
		updated_at = NOW()
	WHERE token = $1`

	log.Println("[SQL]=", query)
	LogSQLArgs(map[string]interface{}{"token": m.Token, "token_expired": expiredAt})

	result, err := m.DB.Exec(query, m.Token, expiredAt)
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

func (m Auth) UpsertOTP() (int64, error) {
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

func authSelectQuery(where string) string {
	return `SELECT ` + authSelectColumns + `
	FROM auths a
	WHERE ` + where
}
