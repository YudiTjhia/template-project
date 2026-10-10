package models

import (
	"database/sql"

	"backend/shared/mmodels"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const (
	UserTypeSuperAdmin = mmodels.UserTypeSuperAdmin
	UserTypeUser       = mmodels.UserTypeUser

	LoginTypeCashier       = mmodels.LoginTypeCashier
	LoginTypeCustomerTable = mmodels.LoginTypeCustomerTable
)

type User struct {
	BaseModel

	ID           int64       `json:"id" db:"id"`
	AccountID    int64       `json:"account_id" db:"account_id"`
	UserEmail    string      `json:"user_email" db:"user_email"`
	MobilePhone  string      `json:"mobilephone" db:"mobilephone"`
	UserPassword string      `json:"-" db:"user_password"`
	PIN          string      `json:"-" db:"pin"`
	ConfirmPIN   string      `json:"-" db:"-"`
	UserType     string      `json:"user_type" db:"user_type"`
	LoginType    string      `json:"login_type" db:"login_type"`
	UserRoles    StringArray `json:"user_roles" db:"user_roles"`
	IsActive     bool        `json:"is_active" db:"is_active"`
	Username     string      `json:"username" db:"username"`
}

func NewUser(db *sqlx.DB, auth Auth) User {
	return User{
		BaseModel: BaseModel{DB: db, Auth: auth, CreatedBy: auth.UserName},
		AccountID: auth.AccountID,
		IsActive:  true,
		UserType:  UserTypeSuperAdmin,
		LoginType: LoginTypeCashier,
	}
}

func (m User) Migrate() {
	if err := mmodels.NewUserMModel(m.Context, m.DB).Migrate(); err != nil {
		panic(err)
	}
}

func (User) TableName() string { return mmodels.UserTable }

func (m User) ValidateCreate() error {
	return m.toUserMModel(m.AccountID).ValidateCreate()
}

func (m User) ValidateID() error {
	return m.toUserMModel(m.AccountID).ValidateID()
}

func (m User) Create() (int64, error) {
	return m.toUserMModel(m.AccountID).Create()
}

func (m User) Update() error {
	return m.toUserMModel(m.AccountID).Update()
}

func (m User) Delete() error {
	return m.toUserMModel(m.AccountID).Delete()
}

func (m User) FindByID() (*User, error) {
	found, err := m.toUserMModel(m.AccountID).FindByID()
	if err != nil {
		return nil, err
	}
	user := m.userFromMModel(*found)
	return &user, nil
}

func (m User) FindByEmail() (*User, error) {
	found, err := m.toUserMModel(m.AccountID).FindByEmail(m.UserEmail)
	if err != nil {
		return nil, err
	}
	user := m.userFromMModel(*found)
	return &user, nil
}

func (m User) FindByEmailGlobal() (*User, error) {
	found, err := m.toUserMModel(m.AccountID).FindByEmailGlobal(m.UserEmail)
	if err != nil {
		return nil, err
	}
	user := m.userFromMModel(*found)
	return &user, nil
}

func (m User) FindByLoginIdentifierGlobal(identifier string) (*User, error) {
	found, err := m.toUserMModel(m.AccountID).FindByLoginIdentifierGlobal(identifier)
	if err != nil {
		return nil, err
	}
	user := m.userFromMModel(*found)
	return &user, nil
}

func (m User) FindActiveUserByAccountIDAndPIN(accountID int64, pin string) (*User, error) {
	found, err := m.toUserMModel(accountID).FindActiveUserByAccountIDAndPIN(accountID, pin)
	if err != nil {
		return nil, err
	}
	user := m.userFromMModel(*found)
	return &user, nil
}

func IsNumericIdentifier(identifier string) (int64, bool) {
	return mmodels.IsNumericIdentifier(identifier)
}

func (m User) FindByUsername() (*User, error) {
	found, err := m.toUserMModel(m.AccountID).FindByUsername(m.Username)
	if err != nil {
		return nil, err
	}
	user := m.userFromMModel(*found)
	return &user, nil
}

func (m User) FindList(search string) ([]User, error) {
	rows, err := m.toUserMModel(m.AccountID).FindList(search)
	if err != nil {
		return nil, err
	}
	return m.usersFromMModel(rows), nil
}

func (m User) FindAll() ([]User, error) {
	rows, err := m.toUserMModel(m.AccountID).FindAll()
	if err != nil {
		return nil, err
	}
	return m.usersFromMModel(rows), nil
}

func (m User) FindRecentList(limit int, search string) ([]User, error) {
	rows, err := m.toUserMModel(m.AccountID).FindRecentList(limit, search)
	if err != nil {
		return nil, err
	}
	return m.usersFromMModel(rows), nil
}

func (m User) toUserMModel(accountID int64) *mmodels.UserMModel {
	model := mmodels.NewUserMModel(m.Context, m.DB)
	model.ID = m.ID
	model.AccountID = accountID
	model.AuthAccountID = m.Auth.AccountID
	model.AuthUserName = m.Auth.UserName
	model.CreatedAt = m.CreatedAt
	model.UserEmail = m.UserEmail
	model.MobilePhone = m.MobilePhone
	model.UserPassword = m.UserPassword
	model.PIN = m.PIN
	model.ConfirmPIN = m.ConfirmPIN
	model.UserType = m.UserType
	model.LoginType = m.LoginType
	model.UserRoles = pq.StringArray(m.UserRoles)
	model.IsActive = m.IsActive
	model.Username = m.Username
	model.CreatedBy = m.CreatedBy
	if m.UpdatedAt != nil {
		model.UpdatedAt = sql.NullTime{Time: *m.UpdatedAt, Valid: true}
	}
	if m.UpdatedBy != nil {
		model.UpdatedBy = sql.NullString{String: *m.UpdatedBy, Valid: true}
	}
	return model
}

func (m User) userFromMModel(value mmodels.UserMModel) User {
	user := User{
		BaseModel:    BaseModel{DB: m.DB, Auth: m.Auth, CreatedAt: value.CreatedAt, CreatedBy: value.CreatedBy},
		ID:           value.ID,
		AccountID:    value.AccountID,
		UserEmail:    value.UserEmail,
		MobilePhone:  value.MobilePhone,
		UserPassword: value.UserPassword,
		PIN:          value.PIN,
		UserType:     value.UserType,
		LoginType:    value.LoginType,
		UserRoles:    StringArray(value.UserRoles),
		IsActive:     value.IsActive,
		Username:     value.Username,
	}
	if value.UpdatedAt.Valid {
		updatedAt := value.UpdatedAt.Time
		user.UpdatedAt = &updatedAt
	}
	if value.UpdatedBy.Valid {
		updatedBy := value.UpdatedBy.String
		user.UpdatedBy = &updatedBy
	}
	return user
}

func (m User) usersFromMModel(rows []mmodels.UserMModel) []User {
	users := make([]User, 0, len(rows))
	for _, row := range rows {
		users = append(users, m.userFromMModel(row))
	}
	return users
}
