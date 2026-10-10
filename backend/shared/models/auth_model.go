package models

import (
	"context"
	"time"

	"backend/shared/mmodels"
	"github.com/jmoiron/sqlx"
)

// Auth is retained as a public service DTO. AuthMModel owns its persistence.
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

func NewAuth(db *sqlx.DB) Auth { return Auth{DB: db} }

func (m Auth) Migrate() {
	if err := m.mmodel().Migrate(); err != nil {
		panic(err)
	}
}

func (m Auth) TableName() string { return mmodels.AuthTable }

func (m Auth) ValidateCreate() error { return m.mmodel().ValidateCreate() }

func (m Auth) ValidateID() error { return m.mmodel().ValidateID() }

func (m Auth) Create() (int64, error) { return m.mmodel().Create() }

func (m Auth) Update() error { return m.mmodel().Update() }

func (m Auth) Delete() error { return m.mmodel().Delete() }

func (m Auth) FindByID() (*Auth, error) {
	found, err := m.mmodel().FindByID()
	if err != nil {
		return nil, err
	}
	return authFromMModel(found, m.DB), nil
}

func (m Auth) FindByEmail() (*Auth, error) {
	found, err := m.mmodel().FindByEmail()
	if err != nil {
		return nil, err
	}
	return authFromMModel(found, m.DB), nil
}

func (m Auth) FindByToken() (*Auth, error) {
	found, err := m.mmodel().FindByToken()
	if err != nil {
		return nil, err
	}
	return authFromMModel(found, m.DB), nil
}

func (m Auth) ExpireActiveByEmail() error { return m.mmodel().ExpireActiveByEmail() }

func (m Auth) ExpireByToken() error { return m.mmodel().ExpireByToken() }

func (m Auth) ExtendByToken(expiredAt time.Time) error {
	return m.mmodel().ExtendByToken(expiredAt)
}

func (m Auth) UpsertOTP() (int64, error) { return m.mmodel().UpsertOTP() }

func (m Auth) mmodel() *mmodels.AuthMModel {
	model := mmodels.NewAuthMModel(context.Background(), m.DB)
	model.ID = m.ID
	model.AccountID = m.AccountID
	model.UserName = m.UserName
	model.Email = m.Email
	model.Name = m.Name
	model.Token = m.Token
	model.TokenExpired = m.TokenExpired
	model.IsExpired = m.IsExpired
	model.OTP = m.OTP
	model.OTPExpired = m.OTPExpired
	model.CreatedAt = m.CreatedAt
	if m.UpdatedAt != nil {
		model.UpdatedAt.Time = *m.UpdatedAt
		model.UpdatedAt.Valid = true
	}
	return model
}

func authFromMModel(value *mmodels.AuthMModel, db *sqlx.DB) *Auth {
	if value == nil {
		return nil
	}
	auth := &Auth{
		DB: db, ID: value.ID, AccountID: value.AccountID, UserName: value.UserName,
		Email: value.Email, Name: value.Name, Token: value.Token, TokenExpired: value.TokenExpired,
		IsExpired: value.IsExpired, OTP: value.OTP, OTPExpired: value.OTPExpired, CreatedAt: value.CreatedAt,
	}
	if value.UpdatedAt.Valid {
		updatedAt := value.UpdatedAt.Time
		auth.UpdatedAt = &updatedAt
	}
	return auth
}
