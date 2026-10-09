package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"backend/shared/helpers"

	"github.com/jmoiron/sqlx"
)

const (
	UserTypeSuperAdmin = "SUPER_ADMIN"
	UserTypeUser       = "USER"

	LoginTypeCashier       = "cashier"
	LoginTypeCustomerTable = "customer_table"
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
	UserType     string      `json:"user_type" db:"user_type"` //
	LoginType    string      `json:"login_type" db:"login_type"`
	UserRoles    StringArray `json:"user_roles" db:"user_roles"`
	IsActive     bool        `json:"is_active" db:"is_active"`
	Username     string      `json:"username" db:"username"`
}

const userSelectColumns = `u.id, u.account_id, u.user_email, u.mobilephone, u.user_password, u.pin, u.user_type, u.login_type, u.user_roles, u.is_active, u.username,
u.created_at, u.created_by, u.updated_at, u.updated_by`

func NewUser(db *sqlx.DB, auth Auth) User {
	user := User{
		BaseModel: BaseModel{DB: db, Auth: auth},
	}
	user.AccountID = auth.AccountID
	user.CreatedBy = auth.UserName
	user.IsActive = true
	user.UserType = UserTypeSuperAdmin
	user.LoginType = LoginTypeCashier
	return user
}

func (m User) Migrate() {
	sql := `CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY,
	account_id BIGINT NOT NULL DEFAULT 0,
	user_email VARCHAR(255) NOT NULL,
	mobilephone VARCHAR(100) NOT NULL DEFAULT '',
	user_password TEXT NOT NULL,
	pin TEXT NOT NULL DEFAULT '',
	user_type VARCHAR(20) NOT NULL DEFAULT 'SUPER_ADMIN',
	login_type VARCHAR(30) NOT NULL DEFAULT 'cashier',
	user_roles TEXT[] NOT NULL DEFAULT '{}'::TEXT[],
	is_active BOOLEAN NOT NULL DEFAULT true,
	username VARCHAR(255) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_by VARCHAR(255) NOT NULL,
	updated_at TIMESTAMPTZ,
	updated_by VARCHAR(255)
);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `DROP INDEX IF EXISTS idx_users_account_email;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `DROP INDEX IF EXISTS idx_users_account_mobilephone;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (LOWER(user_email));`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `CREATE UNIQUE INDEX IF NOT EXISTS idx_users_mobilephone ON users (LOWER(mobilephone)) WHERE mobilephone <> '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `CREATE UNIQUE INDEX IF NOT EXISTS idx_users_account_username ON users (account_id, LOWER(username));`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `CREATE INDEX IF NOT EXISTS idx_users_account_id ON users (account_id);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `CREATE INDEX IF NOT EXISTS idx_users_active ON users (is_active);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE users ADD COLUMN IF NOT EXISTS account_id BIGINT NOT NULL DEFAULT 0;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE users ADD COLUMN IF NOT EXISTS user_email VARCHAR(255) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE users ADD COLUMN IF NOT EXISTS mobilephone VARCHAR(100) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE users ADD COLUMN IF NOT EXISTS user_password TEXT NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE users ADD COLUMN IF NOT EXISTS pin VARCHAR(6) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE users ALTER COLUMN pin TYPE TEXT;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE users ADD COLUMN IF NOT EXISTS user_type VARCHAR(20) NOT NULL DEFAULT 'SUPER_ADMIN';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE users ADD COLUMN IF NOT EXISTS login_type VARCHAR(30) NOT NULL DEFAULT 'cashier';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `UPDATE users SET login_type = 'cashier' WHERE login_type IS NULL OR TRIM(login_type) = '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE users ADD COLUMN IF NOT EXISTS user_roles TEXT[] NOT NULL DEFAULT '{}'::TEXT[];`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(255) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	m.migrateLegacyCredentials()
}

func (m User) migrateLegacyCredentials() {
	type credentialRow struct {
		ID       int64  `db:"id"`
		Password string `db:"user_password"`
		PIN      string `db:"pin"`
	}

	var rows []credentialRow
	err := m.DB.Select(&rows, `
		SELECT id, user_password, pin
		FROM users
		WHERE (user_password <> '' AND user_password NOT LIKE '$2%')
		   OR (pin <> '' AND pin NOT LIKE '$2%')`)
	if err != nil {
		panic(err)
	}

	for _, row := range rows {
		password := row.Password
		pin := row.PIN
		if password != "" && !helpers.IsCredentialHash(password) {
			password, err = helpers.HashCredential(password)
			if err != nil {
				panic(err)
			}
		}
		if pin != "" && !helpers.IsCredentialHash(pin) {
			pin, err = helpers.HashCredential(pin)
			if err != nil {
				panic(err)
			}
		}
		m.DB.MustExec(`UPDATE users SET user_password = $1, pin = $2 WHERE id = $3`, password, pin, row.ID)
	}
}

func (m User) TableName() string {
	return "users"
}

func (m User) ValidateCreate() error {
	if _, err := m.resolveAccountID(); err != nil {
		return err
	}

	if strings.TrimSpace(m.UserEmail) == "" {
		return errors.New("email user wajib diisi")
	}

	if strings.TrimSpace(m.MobilePhone) == "" {
		return errors.New("mobilephone user wajib diisi")
	}

	if strings.TrimSpace(m.UserPassword) == "" {
		return errors.New("password user wajib diisi")
	}

	if err := validateUserPINPair(m.PIN, m.ConfirmPIN); err != nil {
		return err
	}
	m.UserType = normalizeUserType(m.UserType)
	m.LoginType = normalizeLoginType(m.LoginType)
	m.UserRoles = normalizeUserRoles(m.UserRoles)
	if !isAllowedUserType(m.UserType) {
		return errors.New("user_type user tidak valid")
	}
	if !isAllowedLoginType(m.LoginType) {
		return errors.New("login_type user tidak valid")
	}
	if m.UserType != UserTypeSuperAdmin && len(m.UserRoles) == 0 {
		return errors.New("user_roles wajib minimal 1 untuk user_type USER")
	}

	if strings.TrimSpace(m.Username) == "" {
		return errors.New("username user wajib diisi")
	}

	return nil
}

func (m User) ValidateID() error {
	if m.ID <= 0 {
		return errors.New("id user tidak valid")
	}
	return nil
}

func (m User) Create() (int64, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return 0, err
	}

	m.UserEmail = strings.ToLower(strings.TrimSpace(m.UserEmail))
	m.MobilePhone = strings.TrimSpace(m.MobilePhone)
	m.UserPassword = strings.TrimSpace(m.UserPassword)
	m.PIN = strings.TrimSpace(m.PIN)
	m.UserType = normalizeUserType(m.UserType)
	m.LoginType = normalizeLoginType(m.LoginType)
	m.UserRoles = normalizeUserRoles(m.UserRoles)
	m.Username = strings.TrimSpace(m.Username)

	if err := m.ValidateCreate(); err != nil {
		return 0, err
	}
	if !helpers.IsCredentialHash(m.UserPassword) {
		m.UserPassword, err = helpers.HashCredential(m.UserPassword)
		if err != nil {
			return 0, err
		}
	}
	if strings.TrimSpace(m.PIN) != "" && !helpers.IsCredentialHash(m.PIN) {
		m.PIN, err = helpers.HashCredential(m.PIN)
		if err != nil {
			return 0, err
		}
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

	sql := `INSERT INTO users
	(account_id, user_email, mobilephone, user_password, pin, user_type, login_type, user_roles, is_active, username, created_by)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	RETURNING id`

	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{
		"account_id":    accountID,
		"user_email":    m.UserEmail,
		"mobilephone":   m.MobilePhone,
		"user_password": m.UserPassword,
		"pin":           m.PIN,
		"user_type":     m.UserType,
		"login_type":    m.LoginType,
		"user_roles":    []string(m.UserRoles),
		"is_active":     m.IsActive,
		"username":      m.Username,
		"created_by":    createdBy,
	})

	var id int64
	err = m.DB.QueryRow(sql, accountID, m.UserEmail, m.MobilePhone, m.UserPassword, m.PIN, m.UserType, m.LoginType, m.UserRoles, m.IsActive, m.Username, createdBy).Scan(&id)
	return id, err
}

func (m User) Update() error {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return err
	}

	m.UserEmail = strings.ToLower(strings.TrimSpace(m.UserEmail))
	m.MobilePhone = strings.TrimSpace(m.MobilePhone)
	m.UserPassword = strings.TrimSpace(m.UserPassword)
	m.PIN = strings.TrimSpace(m.PIN)
	m.UserType = normalizeUserType(m.UserType)
	m.LoginType = normalizeLoginType(m.LoginType)
	m.UserRoles = normalizeUserRoles(m.UserRoles)
	m.Username = strings.TrimSpace(m.Username)

	if err := m.ValidateID(); err != nil {
		return err
	}

	if strings.TrimSpace(m.UserEmail) == "" {
		return errors.New("email user wajib diisi")
	}

	if strings.TrimSpace(m.MobilePhone) == "" {
		return errors.New("mobilephone user wajib diisi")
	}

	if !helpers.IsCredentialHash(m.PIN) {
		if err := validateUserPINPair(m.PIN, m.ConfirmPIN); err != nil {
			return err
		}
	}
	if m.UserPassword != "" && !helpers.IsCredentialHash(m.UserPassword) {
		m.UserPassword, err = helpers.HashCredential(m.UserPassword)
		if err != nil {
			return err
		}
	}
	if strings.TrimSpace(m.PIN) != "" && !helpers.IsCredentialHash(m.PIN) {
		m.PIN, err = helpers.HashCredential(m.PIN)
		if err != nil {
			return err
		}
	}

	if !isAllowedUserType(m.UserType) {
		return errors.New("user_type user tidak valid")
	}
	if !isAllowedLoginType(m.LoginType) {
		return errors.New("login_type user tidak valid")
	}
	if m.UserType != UserTypeSuperAdmin && len(m.UserRoles) == 0 {
		return errors.New("user_roles wajib minimal 1 untuk user_type USER")
	}

	if strings.TrimSpace(m.Username) == "" {
		return errors.New("username user wajib diisi")
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

	sql := `UPDATE users
	SET user_email = $1,
		mobilephone = $2,
		user_password = CASE WHEN $3 = '' THEN user_password ELSE $3 END,
		pin = CASE WHEN $4 = '' THEN pin ELSE $4 END,
		user_type = $5,
		login_type = $6,
		user_roles = $7,
		is_active = $8,
		username = $9,
		updated_at = NOW(),
		updated_by = $10
	WHERE id = $11 AND account_id = $12`

	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{
		"user_email":    m.UserEmail,
		"mobilephone":   m.MobilePhone,
		"user_password": m.UserPassword,
		"pin":           m.PIN,
		"user_type":     m.UserType,
		"login_type":    m.LoginType,
		"user_roles":    []string(m.UserRoles),
		"is_active":     m.IsActive,
		"username":      m.Username,
		"updated_by":    updatedBy,
		"id":            m.ID,
		"account_id":    accountID,
	})

	_, err = m.DB.Exec(sql, m.UserEmail, m.MobilePhone, m.UserPassword, m.PIN, m.UserType, m.LoginType, m.UserRoles, m.IsActive, m.Username, updatedBy, m.ID, accountID)
	return err
}

func (m User) Delete() error {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return err
	}

	if err := m.ValidateID(); err != nil {
		return err
	}

	found, err := m.FindByID()
	if err != nil {
		return err
	}
	if strings.EqualFold(strings.TrimSpace(found.UserType), UserTypeSuperAdmin) {
		return errors.New("user super admin tidak bisa dihapus")
	}

	sql := `DELETE FROM users WHERE id = $1 AND account_id = $2`
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"id": m.ID, "account_id": accountID})

	_, err = m.DB.Exec(sql, m.ID, accountID)
	return err
}

func (m User) FindByID() (*User, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}

	if err := m.ValidateID(); err != nil {
		return nil, err
	}

	sql := userSelectQuery(`u.id = $1 AND u.account_id = $2`)
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"id": m.ID, "account_id": accountID})

	var user User
	err = m.DB.Get(&user, sql, m.ID, accountID)
	user.DB = m.DB
	user.Auth = m.Auth
	return &user, err
}

func (m User) FindByEmail() (*User, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}

	m.UserEmail = strings.ToLower(strings.TrimSpace(m.UserEmail))
	if m.UserEmail == "" {
		return nil, errors.New("email user wajib diisi")
	}

	sql := userSelectQuery(`u.account_id = $1 AND u.user_email = $2`)
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"account_id": accountID, "user_email": m.UserEmail})

	var user User
	err = m.DB.Get(&user, sql, accountID, m.UserEmail)
	user.DB = m.DB
	user.Auth = m.Auth
	return &user, err
}

func (m User) FindByEmailGlobal() (*User, error) {
	m.UserEmail = strings.ToLower(strings.TrimSpace(m.UserEmail))
	if m.UserEmail == "" {
		return nil, errors.New("email user wajib diisi")
	}

	sql := userSelectQuery(`LOWER(u.user_email) = $1`)
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"user_email": m.UserEmail})

	var user User
	err := m.DB.Get(&user, sql, m.UserEmail)
	user.DB = m.DB
	user.Auth = m.Auth
	return &user, err
}

func (m User) FindByLoginIdentifierGlobal(identifier string) (*User, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return nil, errors.New("email atau mobilephone user wajib diisi")
	}

	normalizedEmail := strings.ToLower(identifier)
	sql := userSelectQuery(`LOWER(u.user_email) = $1 OR LOWER(u.mobilephone) = $2`)
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"identifier": identifier})

	var user User
	err := m.DB.Get(&user, sql, normalizedEmail, strings.ToLower(identifier))
	user.DB = m.DB
	user.Auth = m.Auth
	return &user, err
}

func (m User) FindActiveUserByAccountIDAndPIN(accountID int64, pin string) (*User, error) {
	if accountID <= 0 {
		return nil, errors.New("account id user wajib diisi")
	}
	pin = strings.TrimSpace(pin)
	if pin == "" {
		return nil, errors.New("pin user wajib diisi")
	}

	query := userSelectQuery(`u.account_id = $1 AND u.is_active = true`)
	query += ` ORDER BY u.id ASC`
	log.Println("[SQL]=", query)
	LogSQLArgs(map[string]interface{}{"account_id": accountID})

	var users []User
	if err := m.DB.Select(&users, query, accountID); err != nil {
		return nil, err
	}
	for i := range users {
		users[i].DB = m.DB
		users[i].Auth = m.Auth
		if helpers.CompareCredential(users[i].PIN, pin) {
			return &users[i], nil
		}
	}
	return nil, sql.ErrNoRows
}

func IsNumericIdentifier(identifier string) (int64, bool) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return 0, false
	}

	value, err := strconv.ParseInt(identifier, 10, 64)
	if err != nil || value <= 0 {
		return 0, false
	}

	return value, true
}

func (m User) FindByUsername() (*User, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}

	m.Username = strings.TrimSpace(m.Username)
	if m.Username == "" {
		return nil, errors.New("username user wajib diisi")
	}

	sql := userSelectQuery(`u.account_id = $1 AND u.username = $2`)
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"account_id": accountID, "username": m.Username})

	var user User
	err = m.DB.Get(&user, sql, accountID, m.Username)
	user.DB = m.DB
	user.Auth = m.Auth
	return &user, err
}

func (m User) FindList(search string) ([]User, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}

	sql := userSelectQuery(`u.account_id = $1`)
	args := []interface{}{accountID}
	where := []string{}

	if search != "" {
		args = append(args, search)
		placeholder := nextSQLPlaceholder(args)
		where = append(where, `(u.user_email ILIKE '%' || `+placeholder+` || '%' OR u.mobilephone ILIKE '%' || `+placeholder+` || '%' OR u.username ILIKE '%' || `+placeholder+` || '%')`)
	}

	if len(where) > 0 {
		sql += " AND " + strings.Join(where, " AND ")
	}

	sql += ` ORDER BY u.username ASC LIMIT 50`

	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"account_id": accountID, "search": search})

	var users []User
	err = m.DB.Select(&users, sql, args...)
	return users, err
}

func (m User) FindAll() ([]User, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}

	sql := userSelectQuery(`u.account_id = $1`)
	sql += ` ORDER BY u.username ASC`

	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"account_id": accountID})

	var users []User
	err = m.DB.Select(&users, sql, accountID)
	return users, err
}

func (m User) FindRecentList(limit int, search string) ([]User, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 5
	}

	sql := userSelectQuery(`u.account_id = $1`)
	args := []interface{}{accountID}
	where := []string{}

	if search != "" {
		args = append(args, search)
		placeholder := nextSQLPlaceholder(args)
		where = append(where, `(u.user_email ILIKE '%' || `+placeholder+` || '%' OR u.username ILIKE '%' || `+placeholder+` || '%')`)
	}

	if len(where) > 0 {
		sql += " AND " + strings.Join(where, " AND ")
	}

	sql += fmt.Sprintf(` ORDER BY u.created_at DESC, u.id DESC LIMIT %d`, limit)

	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"account_id": accountID, "search": search, "limit": limit})

	var users []User
	err = m.DB.Select(&users, sql, args...)
	return users, err
}

func userSelectQuery(where string) string {
	return `SELECT ` + userSelectColumns + `
	FROM users u
	WHERE ` + where
}

func (m User) resolveAccountID() (int64, error) {
	if m.Auth.AccountID > 0 {
		return m.Auth.AccountID, nil
	}

	if m.AccountID > 0 {
		return m.AccountID, nil
	}

	return 0, errors.New("id akun wajib diisi")
}

func (m User) ensureEmailAvailable(excludeID int64) error {
	email := strings.ToLower(strings.TrimSpace(m.UserEmail))
	if email == "" {
		return errors.New("email user wajib diisi")
	}

	sql := `SELECT id FROM users WHERE LOWER(user_email) = $1`
	args := []interface{}{email}
	if excludeID > 0 {
		sql += ` AND id <> $2`
		args = append(args, excludeID)
	}

	var existingID int64
	err := m.DB.Get(&existingID, sql, args...)
	if err == nil {
		return errors.New("email user sudah digunakan")
	}
	if strings.Contains(strings.ToLower(err.Error()), "no rows") {
		return nil
	}

	return err
}

func (m User) ensureMobilephoneAvailable(excludeID int64) error {
	mobilephone := strings.ToLower(strings.TrimSpace(m.MobilePhone))
	if mobilephone == "" {
		return errors.New("mobilephone user wajib diisi")
	}

	sql := `SELECT id FROM users WHERE LOWER(mobilephone) = $1`
	args := []interface{}{mobilephone}
	if excludeID > 0 {
		sql += ` AND id <> $2`
		args = append(args, excludeID)
	}

	var existingID int64
	err := m.DB.Get(&existingID, sql, args...)
	if err == nil {
		return errors.New("mobilephone user sudah digunakan")
	}
	if strings.Contains(strings.ToLower(err.Error()), "no rows") {
		return nil
	}

	return err
}

func normalizeUserType(userType string) string {
	userType = strings.ToUpper(strings.TrimSpace(userType))
	if userType == "" {
		return UserTypeSuperAdmin
	}
	return userType
}

func isAllowedUserType(userType string) bool {
	switch userType {
	case UserTypeSuperAdmin, UserTypeUser:
		return true
	default:
		return false
	}
}

func normalizeLoginType(loginType string) string {
	loginType = strings.ToLower(strings.TrimSpace(loginType))
	if loginType == "" {
		return LoginTypeCashier
	}
	return loginType
}

func isAllowedLoginType(loginType string) bool {
	switch normalizeLoginType(loginType) {
	case LoginTypeCashier, LoginTypeCustomerTable:
		return true
	default:
		return false
	}
}

func normalizeUserRoles(roles StringArray) StringArray {
	normalized := make(StringArray, 0, len(roles))
	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role != "" {
			normalized = append(normalized, role)
		}
	}
	return normalized
}

func validateUserPIN(pin string) error {
	pin = strings.TrimSpace(pin)
	if pin == "" {
		return nil
	}
	if len(pin) != 6 {
		return errors.New("pin harus 6 digit angka")
	}
	for _, char := range pin {
		if char < '0' || char > '9' {
			return errors.New("pin harus 6 digit angka")
		}
	}
	return nil
}

func validateUserPINPair(pin, confirmPIN string) error {
	pin = strings.TrimSpace(pin)
	confirmPIN = strings.TrimSpace(confirmPIN)

	if pin == "" && confirmPIN == "" {
		return nil
	}
	if pin == "" {
		return errors.New("pin wajib diisi")
	}
	if err := validateUserPIN(pin); err != nil {
		return err
	}
	if confirmPIN == "" {
		return errors.New("confirm pin wajib diisi")
	}
	if pin != confirmPIN {
		return errors.New("confirm pin tidak cocok")
	}
	return nil
}
