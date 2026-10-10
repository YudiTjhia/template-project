package mmodels

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"backend/shared/helpers"
	"backend/shared/metadata/impl"
	"backend/shared/metadata/interfaces"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// UserMModel combines the users table metadata with values that can be
// validated before creating a user record.
type UserMModel struct {
	impl.MModel

	ID            int64          `json:"id" db:"id"`
	AccountID     int64          `json:"account_id" db:"account_id"`
	CreatedAt     time.Time      `json:"-" db:"created_at"`
	UserEmail     string         `json:"user_email" db:"user_email"`
	MobilePhone   string         `json:"mobilephone" db:"mobilephone"`
	UserPassword  string         `json:"-" db:"user_password"`
	PIN           string         `json:"-" db:"pin"`
	ConfirmPIN    string         `json:"-" db:"-"`
	UserType      string         `json:"user_type" db:"user_type"`
	LoginType     string         `json:"login_type" db:"login_type"`
	UserRoles     pq.StringArray `json:"user_roles" db:"user_roles"`
	IsActive      bool           `json:"is_active" db:"is_active"`
	Username      string         `json:"username" db:"username"`
	CreatedBy     string         `json:"created_by" db:"created_by"`
	UpdatedAt     sql.NullTime   `json:"-" db:"updated_at"`
	UpdatedBy     sql.NullString `json:"-" db:"updated_by"`
	AuthAccountID int64          `json:"-" db:"-"`
	AuthUserName  string         `json:"-" db:"-"`
}

// NewUserMModel builds metadata matching the base users table schema.
// Indexes and data migrations remain in the user model's migration logic.
func NewUserMModel(ctx context.Context, db *sqlx.DB) *UserMModel {
	return &UserMModel{
		MModel: impl.MModel{
			Context:   ctx,
			Name:      "User",
			TableName: UserTable,
			DB:        db,
			Fields: []interfaces.IMField{
				impl.MField{Type: "int64", DbType: "INTEGER", JsonTag: "id", DbTag: "id", IsPrimaryKey: true, IsIdentity: true, Required: true, IsNotUpdateField: true},
				impl.MField{Type: "int64", DbType: "BIGINT", JsonTag: "account_id", DbTag: "account_id", Default: "0", Required: true, IsNotUpdateField: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "user_email", DbTag: "user_email", MaxLength: 255, MinLength: 3, Pattern: `^[^@\s]+@[^@\s]+\.[^@\s]+$`, Required: true, IsEmail: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "mobilephone", DbTag: "mobilephone", MaxLength: 100, Default: "''", Required: true, IsPhone: true},
				impl.MField{Type: "string", DbType: "TEXT", JsonTag: "-", DbTag: "user_password", Required: true, IsPassword: true},
				impl.MField{Type: "string", DbType: "TEXT", JsonTag: "-", DbTag: "pin", Default: "''", Required: true, IsPassword: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "user_type", DbTag: "user_type", MaxLength: 20, Default: "'SUPER_ADMIN'", Required: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "login_type", DbTag: "login_type", MaxLength: 30, Default: "'cashier'", Required: true},
				impl.MField{Type: "[]string", DbType: "TEXT[]", JsonTag: "user_roles", DbTag: "user_roles", Default: "'{}'::TEXT[]", Required: true},
				impl.MField{Type: "bool", DbType: "BOOLEAN", JsonTag: "is_active", DbTag: "is_active", Default: "true", Required: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "username", DbTag: "username", MaxLength: 255, Required: true},
				impl.MField{Type: "time.Time", DbType: "TIMESTAMPTZ", JsonTag: "created_at", DbTag: "created_at", Default: "NOW()", Required: true, IsNotUpdateField: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "created_by", DbTag: "created_by", MaxLength: 255, Required: true, IsNotUpdateField: true},
				impl.MField{Type: "sql.NullTime", DbType: "TIMESTAMPTZ", JsonTag: "updated_at", DbTag: "updated_at", IsNullable: true, IsNotCreateField: true},
				impl.MField{Type: "sql.NullString", DbType: "VARCHAR", JsonTag: "updated_by", DbTag: "updated_by", MaxLength: 255, IsNullable: true, IsNotCreateField: true},
			},
		},
	}
}

// Migrate creates or upgrades the users table, then applies indexes and
// user-specific data migrations in one transaction.
func (m *UserMModel) Migrate() error {
	if m == nil {
		return errors.New("user metadata model is nil")
	}
	if m.DB == nil {
		return errors.New("user metadata database is nil")
	}
	tx, err := m.DB.BeginTxx(m.ContextOrBackground(), nil)
	if err != nil {
		return fmt.Errorf("begin users migration: %w", err)
	}
	defer tx.Rollback()

	migrator := impl.NewMMigrator(&m.MModel, tx)
	migrator.SetContext(m.ContextOrBackground())
	if err := migrator.Migrate(); err != nil {
		return err
	}
	for _, statement := range userSchemaUpgradeStatements {
		if _, err := tx.ExecContext(m.ContextOrBackground(), statement); err != nil {
			return fmt.Errorf("run users migration statement %q: %w", statement, err)
		}
	}
	if err := m.migrateLegacyCredentials(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit users migration: %w", err)
	}
	return nil
}

var userSchemaUpgradeStatements = []string{
	`ALTER TABLE ` + UserTable + ` ADD COLUMN IF NOT EXISTS account_id BIGINT NOT NULL DEFAULT 0`,
	`ALTER TABLE ` + UserTable + ` ADD COLUMN IF NOT EXISTS user_email VARCHAR(255) NOT NULL DEFAULT ''`,
	`ALTER TABLE ` + UserTable + ` ADD COLUMN IF NOT EXISTS mobilephone VARCHAR(100) NOT NULL DEFAULT ''`,
	`ALTER TABLE ` + UserTable + ` ADD COLUMN IF NOT EXISTS user_password TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE ` + UserTable + ` ADD COLUMN IF NOT EXISTS pin TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE ` + UserTable + ` ALTER COLUMN pin TYPE TEXT`,
	`ALTER TABLE ` + UserTable + ` ADD COLUMN IF NOT EXISTS user_type VARCHAR(20) NOT NULL DEFAULT 'SUPER_ADMIN'`,
	`ALTER TABLE ` + UserTable + ` ADD COLUMN IF NOT EXISTS login_type VARCHAR(30) NOT NULL DEFAULT 'cashier'`,
	`ALTER TABLE ` + UserTable + ` ADD COLUMN IF NOT EXISTS user_roles TEXT[] NOT NULL DEFAULT '{}'::TEXT[]`,
	`ALTER TABLE ` + UserTable + ` ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true`,
	`ALTER TABLE ` + UserTable + ` ADD COLUMN IF NOT EXISTS username VARCHAR(255) NOT NULL DEFAULT ''`,
	`ALTER TABLE ` + UserTable + ` ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
	`ALTER TABLE ` + UserTable + ` ADD COLUMN IF NOT EXISTS created_by VARCHAR(255) NOT NULL DEFAULT ''`,
	`ALTER TABLE ` + UserTable + ` ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ`,
	`ALTER TABLE ` + UserTable + ` ADD COLUMN IF NOT EXISTS updated_by VARCHAR(255)`,
	`UPDATE ` + UserTable + ` SET login_type = 'cashier' WHERE login_type IS NULL OR TRIM(login_type) = ''`,
	`DROP INDEX IF EXISTS idx_users_account_email`,
	`DROP INDEX IF EXISTS idx_users_account_mobilephone`,
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON ` + UserTable + ` (LOWER(user_email))`,
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_mobilephone ON ` + UserTable + ` (LOWER(mobilephone)) WHERE mobilephone <> ''`,
	`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_account_username ON ` + UserTable + ` (account_id, LOWER(username))`,
	`CREATE INDEX IF NOT EXISTS idx_users_account_id ON ` + UserTable + ` (account_id)`,
	`CREATE INDEX IF NOT EXISTS idx_users_active ON ` + UserTable + ` (is_active)`,
}

func (m *UserMModel) migrateLegacyCredentials(tx *sqlx.Tx) error {
	type credentialRow struct {
		ID       int64  `db:"id"`
		Password string `db:"user_password"`
		PIN      string `db:"pin"`
	}
	var rows []credentialRow
	if err := tx.SelectContext(m.ContextOrBackground(), &rows, `
		SELECT id, user_password, pin
		FROM `+UserTable+`
		WHERE (user_password <> '' AND user_password NOT LIKE '$2%')
		   OR (pin <> '' AND pin NOT LIKE '$2%')`); err != nil {
		return fmt.Errorf("read legacy user credentials: %w", err)
	}
	for _, row := range rows {
		password, pin := row.Password, row.PIN
		var err error
		if password != "" && !helpers.IsCredentialHash(password) {
			password, err = helpers.HashCredential(password)
			if err != nil {
				return fmt.Errorf("hash legacy password for user %d: %w", row.ID, err)
			}
		}
		if pin != "" && !helpers.IsCredentialHash(pin) {
			pin, err = helpers.HashCredential(pin)
			if err != nil {
				return fmt.Errorf("hash legacy PIN for user %d: %w", row.ID, err)
			}
		}
		if _, err := tx.ExecContext(m.ContextOrBackground(), `UPDATE `+UserTable+` SET user_password = $1, pin = $2 WHERE id = $3`, password, pin, row.ID); err != nil {
			return fmt.Errorf("update legacy credentials for user %d: %w", row.ID, err)
		}
	}
	return nil
}

func (m *UserMModel) FindByUsername_(username string, accountID int64) error {
	if m == nil {
		return errors.New("user metadata model is nil")
	}
	m.Username = strings.TrimSpace(username)
	m.AccountID = accountID
	return m.MModel.FindByFields(m, "account_id", "username")
}

func (m *UserMModel) FindAllByAccount_(accountID int64) ([]UserMModel, error) {
	if m == nil {
		return nil, errors.New("user metadata model is nil")
	}
	users := make([]UserMModel, 0)
	err := m.MModel.FindListOrdered(&users, map[string]any{"account_id": accountID}, "", nil, "username", "ASC", 0)
	return users, err
}

func (m *UserMModel) FindList_(search string, limit int) ([]UserMModel, error) {
	if m == nil {
		return nil, errors.New("user metadata model is nil")
	}
	if limit <= 0 {
		limit = 50
	}
	return m.FindListOrdered_(search, "username", "ASC", limit)
}

func (m *UserMModel) FindRecentList_(search string, limit int) ([]UserMModel, error) {
	if m == nil {
		return nil, errors.New("user metadata model is nil")
	}
	if m.AccountID <= 0 {
		return nil, errors.New("id akun wajib diisi")
	}
	if limit <= 0 {
		limit = 5
	}
	users := make([]UserMModel, 0)
	err := m.MModel.FindListOrdered(&users, map[string]any{"account_id": m.AccountID}, search,
		[]string{"user_email", "username"}, "created_at,id", "DESC", limit)
	return users, err
}

// FindListOrdered_ lists users for this account; limit 0 returns all rows.
func (m *UserMModel) FindListOrdered_(search, orderBy, direction string, limit int) ([]UserMModel, error) {
	if m == nil {
		return nil, errors.New("user metadata model is nil")
	}
	if m.AccountID <= 0 {
		return nil, errors.New("id akun wajib diisi")
	}
	users := make([]UserMModel, 0)
	err := m.MModel.FindListOrdered(&users, map[string]any{"account_id": m.AccountID}, search,
		[]string{"user_email", "mobilephone", "username"}, orderBy, direction, limit)
	if err != nil {
		return nil, err
	}
	for i := range users {
		users[i].MModel = m.MModel
	}
	return users, nil
}

func (m *UserMModel) FindByEmail_(email string, accountID *int64) error {
	if m == nil {
		return errors.New("user metadata model is nil")
	}
	m.UserEmail = strings.ToLower(strings.TrimSpace(email))
	if m.UserEmail == "" {
		return errors.New("email user wajib diisi")
	}
	fields := []string{"user_email"}
	if accountID != nil {
		if *accountID <= 0 {
			return errors.New("id akun wajib diisi")
		}
		m.AccountID = *accountID
		fields = append(fields, "account_id")
	}
	return m.MModel.FindByFields(m, fields...)
}

func (m *UserMModel) FindByEmailGlobal_(email string) error {
	if m == nil {
		return errors.New("user metadata model is nil")
	}
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return errors.New("email user wajib diisi")
	}
	return m.MModel.FindByAnyField(m, "user_email", email)
}

func (m *UserMModel) ExistsByEmail_(email string, excludeID int64) (bool, error) {
	if m == nil {
		return false, errors.New("user metadata model is nil")
	}
	return m.MModel.ExistsByField("user_email", strings.ToLower(strings.TrimSpace(email)), "id", excludeID)
}

func (m *UserMModel) ExistsByMobilePhone_(mobilePhone string, excludeID int64) (bool, error) {
	if m == nil {
		return false, errors.New("user metadata model is nil")
	}
	return m.MModel.ExistsByField("mobilephone", strings.ToLower(strings.TrimSpace(mobilePhone)), "id", excludeID)
}

func (m *UserMModel) FindByLoginIdentifierGlobal_(identifier string) error {
	if m == nil {
		return errors.New("user metadata model is nil")
	}
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return errors.New("email atau mobilephone user wajib diisi")
	}
	return m.MModel.FindByAnyField(m, "user_email", strings.ToLower(identifier), "mobilephone", strings.ToLower(identifier))
}

func (m *UserMModel) FindActiveByAccount_(accountID int64) ([]UserMModel, error) {
	if m == nil {
		return nil, errors.New("user metadata model is nil")
	}
	if accountID <= 0 {
		return nil, errors.New("account id user wajib diisi")
	}
	users := make([]UserMModel, 0)
	if err := m.MModel.FindListOrdered(&users, map[string]any{"account_id": accountID, "is_active": true}, "", nil, "id", "ASC", 0); err != nil {
		return nil, err
	}
	return users, nil
}

const (
	UserTypeSuperAdmin = "SUPER_ADMIN"
	UserTypeUser       = "USER"

	LoginTypeCashier       = "cashier"
	LoginTypeCustomerTable = "customer_table"
)

// ValidateCreate applies User-specific rules before metadata validation.
func (m *UserMModel) ValidateCreate() error {
	if m == nil {
		return errors.New("user metadata model is nil")
	}
	if _, err := m.MModel.ValidateAccountID_(m.AuthAccountID, m.AccountID); err != nil {
		return err
	}
	m.normalizeValues()
	if m.UserEmail == "" {
		return errors.New("email user wajib diisi")
	}
	if m.MobilePhone == "" {
		return errors.New("mobilephone user wajib diisi")
	}
	if m.UserPassword == "" {
		return errors.New("password user wajib diisi")
	}
	if err := validateUserPINPair(m.PIN, m.ConfirmPIN); err != nil {
		return err
	}
	if !isAllowedUserMModelType(m.UserType) {
		return errors.New("user_type user tidak valid")
	}
	if !isAllowedUserMModelLoginType(m.LoginType) {
		return errors.New("login_type user tidak valid")
	}
	if m.UserType != UserTypeSuperAdmin && len(m.UserRoles) == 0 {
		return errors.New("user_roles wajib minimal 1 untuk user_type USER")
	}
	if m.Username == "" {
		return errors.New("username user wajib diisi")
	}
	return nil
}

func (m *UserMModel) ValidateID() error {
	if m == nil {
		return errors.New("id user tidak valid")
	}
	return m.MModel.ValidateID_(m.ID, "user")
}

// Create validates and normalizes input, hashes credentials, checks global
// uniqueness, and inserts the user through MModel.
func (m *UserMModel) Create() (int64, error) {
	if m == nil {
		return 0, errors.New("user metadata model is nil")
	}
	accountID, err := m.resolveAccountID()
	if err != nil {
		return 0, err
	}
	m.AccountID = accountID
	m.normalizeValues()
	if err := m.ValidateCreate(); err != nil {
		return 0, err
	}
	if !helpers.IsCredentialHash(m.UserPassword) {
		m.UserPassword, err = helpers.HashCredential(m.UserPassword)
		if err != nil {
			return 0, err
		}
	}
	if m.PIN != "" && !helpers.IsCredentialHash(m.PIN) {
		m.PIN, err = helpers.HashCredential(m.PIN)
		if err != nil {
			return 0, err
		}
	}
	if err := m.ensureEmailAvailable(0); err != nil {
		return 0, err
	}
	if err := m.ensureMobilePhoneAvailable(0); err != nil {
		return 0, err
	}
	m.CreatedBy, err = m.MModel.ResolveCreatedBy_(m.CreatedBy, m.AuthUserName)
	if err != nil {
		return 0, err
	}
	created, err := impl.NewMService().Create_(&m.MModel, m)
	if err != nil {
		return 0, err
	}
	id, ok := created.(int64)
	if !ok {
		return 0, fmt.Errorf("created user ID has type %T, expected int64", created)
	}
	m.ID = id
	return id, nil
}

// Update applies User-specific validation and credential handling, then
// scopes the metadata update to both user ID and account ID.
func (m *UserMModel) Update() error {
	if m == nil {
		return errors.New("user metadata model is nil")
	}
	accountID, err := m.resolveAccountID()
	if err != nil {
		return err
	}
	m.AccountID = accountID
	m.normalizeValues()
	if err := m.ValidateID(); err != nil {
		return err
	}
	if m.UserEmail == "" {
		return errors.New("email user wajib diisi")
	}
	if m.MobilePhone == "" {
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
	if m.PIN != "" && !helpers.IsCredentialHash(m.PIN) {
		m.PIN, err = helpers.HashCredential(m.PIN)
		if err != nil {
			return err
		}
	}
	if !isAllowedUserMModelType(m.UserType) {
		return errors.New("user_type user tidak valid")
	}
	if !isAllowedUserMModelLoginType(m.LoginType) {
		return errors.New("login_type user tidak valid")
	}
	if m.UserType != UserTypeSuperAdmin && len(m.UserRoles) == 0 {
		return errors.New("user_roles wajib minimal 1 untuk user_type USER")
	}
	if m.Username == "" {
		return errors.New("username user wajib diisi")
	}
	if err := m.ensureEmailAvailable(m.ID); err != nil {
		return err
	}
	if err := m.ensureMobilePhoneAvailable(m.ID); err != nil {
		return err
	}
	if !m.UpdatedAt.Valid {
		m.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}
	m.UpdatedBy, err = m.MModel.ResolveUpdatedBy_(m.UpdatedBy, m.AuthUserName)
	if err != nil {
		return err
	}
	return impl.NewMService().Update_(&m.MModel, m, "id", "account_id")
}

func (m *UserMModel) Delete() error {
	if m == nil {
		return errors.New("user metadata model is nil")
	}
	accountID, err := m.resolveAccountID()
	if err != nil {
		return err
	}
	m.AccountID = accountID
	if err := m.ValidateID(); err != nil {
		return err
	}
	found := &UserMModel{ID: m.ID, AccountID: accountID, MModel: m.MModel}
	if err := m.MModel.FindByID_(found, "id", "account_id"); err != nil {
		return err
	}
	if strings.EqualFold(strings.TrimSpace(found.UserType), UserTypeSuperAdmin) {
		return errors.New("user super admin tidak bisa dihapus")
	}
	return impl.NewMService().Delete_(&m.MModel, m, "id", "account_id")
}

func (m *UserMModel) FindByID() (*UserMModel, error) {
	if m == nil {
		return nil, errors.New("user metadata model is nil")
	}
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}
	if err := m.ValidateID(); err != nil {
		return nil, err
	}
	m.AccountID = accountID
	if err := m.MModel.FindByID_(m, "id", "account_id"); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *UserMModel) FindByEmail(email string) (*UserMModel, error) {
	if m == nil {
		return nil, errors.New("user metadata model is nil")
	}
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}
	m.AccountID = accountID
	if err := m.FindByEmail_(email, &accountID); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *UserMModel) FindByEmailGlobal(email string) (*UserMModel, error) {
	if m == nil {
		return nil, errors.New("user metadata model is nil")
	}
	if err := m.FindByEmailGlobal_(email); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *UserMModel) FindByLoginIdentifierGlobal(identifier string) (*UserMModel, error) {
	if m == nil {
		return nil, errors.New("user metadata model is nil")
	}
	if err := m.FindByLoginIdentifierGlobal_(identifier); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *UserMModel) FindByUsername(username string) (*UserMModel, error) {
	if m == nil {
		return nil, errors.New("user metadata model is nil")
	}
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}
	if err := m.FindByUsername_(username, accountID); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *UserMModel) FindActiveUserByAccountIDAndPIN(accountID int64, pin string) (*UserMModel, error) {
	if m == nil {
		return nil, errors.New("user metadata model is nil")
	}
	if accountID <= 0 {
		return nil, errors.New("account id user wajib diisi")
	}
	pin = strings.TrimSpace(pin)
	if pin == "" {
		return nil, errors.New("pin user wajib diisi")
	}
	candidates, err := m.FindActiveByAccount_(accountID)
	if err != nil {
		return nil, err
	}
	for i := range candidates {
		if helpers.CompareCredential(candidates[i].PIN, pin) {
			return &candidates[i], nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *UserMModel) FindList(search string) ([]UserMModel, error) {
	return m.FindList_(search, 50)
}

func (m *UserMModel) FindAll() ([]UserMModel, error) {
	if m == nil {
		return nil, errors.New("user metadata model is nil")
	}
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}
	return m.FindAllByAccount_(accountID)
}

func (m *UserMModel) FindRecentList(limit int, search string) ([]UserMModel, error) {
	return m.FindRecentList_(search, limit)
}

func (m *UserMModel) ensureEmailAvailable(excludeID int64) error {
	if m == nil {
		return errors.New("user metadata model is nil")
	}
	email := strings.ToLower(strings.TrimSpace(m.UserEmail))
	if email == "" {
		return errors.New("email user wajib diisi")
	}
	found, err := m.MModel.ExistsByField("user_email", email, "id", excludeID)
	if err != nil {
		return err
	}
	if found {
		return errors.New("email user sudah digunakan")
	}
	return nil
}

func (m *UserMModel) ensureMobilePhoneAvailable(excludeID int64) error {
	if m == nil {
		return errors.New("user metadata model is nil")
	}
	phone := strings.ToLower(strings.TrimSpace(m.MobilePhone))
	if phone == "" {
		return errors.New("mobilephone user wajib diisi")
	}
	found, err := m.MModel.ExistsByField("mobilephone", phone, "id", excludeID)
	if err != nil {
		return err
	}
	if found {
		return errors.New("mobilephone user sudah digunakan")
	}
	return nil
}

func (m *UserMModel) resolveAccountID() (int64, error) {
	if m == nil {
		return 0, errors.New("user metadata model is nil")
	}
	return m.MModel.ResolveAccountID(m.AuthAccountID, m.AccountID)
}

func (m *UserMModel) normalizeValues() {
	m.UserEmail = strings.ToLower(strings.TrimSpace(m.UserEmail))
	m.MobilePhone = strings.TrimSpace(m.MobilePhone)
	m.UserPassword = strings.TrimSpace(m.UserPassword)
	m.PIN = strings.TrimSpace(m.PIN)
	m.UserType = normalizeUserMModelType(m.UserType)
	m.LoginType = normalizeUserMModelLoginType(m.LoginType)
	m.Username = strings.TrimSpace(m.Username)
	roles := make(pq.StringArray, 0, len(m.UserRoles))
	for _, role := range m.UserRoles {
		if role = strings.TrimSpace(role); role != "" {
			roles = append(roles, role)
		}
	}
	m.UserRoles = roles
}

func normalizeUserMModelType(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return UserTypeSuperAdmin
	}
	return value
}

func isAllowedUserMModelType(value string) bool {
	return value == UserTypeSuperAdmin || value == UserTypeUser
}

func normalizeUserMModelLoginType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return LoginTypeCashier
	}
	return value
}

func isAllowedUserMModelLoginType(value string) bool {
	value = normalizeUserMModelLoginType(value)
	return value == LoginTypeCashier || value == LoginTypeCustomerTable
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
	if len(pin) != 6 {
		return errors.New("pin harus 6 digit angka")
	}
	for _, char := range pin {
		if char < '0' || char > '9' {
			return errors.New("pin harus 6 digit angka")
		}
	}
	if confirmPIN == "" {
		return errors.New("confirm pin wajib diisi")
	}
	if pin != confirmPIN {
		return errors.New("confirm pin tidak cocok")
	}
	return nil
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
