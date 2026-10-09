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
	"github.com/lib/pq"
)

// RoleMModel describes and operates on the roles table using the shared MModel.
type RoleMModel struct {
	impl.MModel

	ID            int64          `json:"id" db:"id"`
	AccountID     int64          `json:"account_id" db:"account_id"`
	RoleName      string         `json:"role_name" db:"role_name"`
	Actions       pq.StringArray `json:"actions" db:"actions"`
	CreatedAt     time.Time      `json:"-" db:"created_at"`
	CreatedBy     string         `json:"created_by" db:"created_by"`
	UpdatedAt     sql.NullTime   `json:"-" db:"updated_at"`
	UpdatedBy     sql.NullString `json:"-" db:"updated_by"`
	AuthAccountID int64          `json:"-" db:"-"`
	AuthEmail     string         `json:"-" db:"-"`
}

func NewRoleMModel(ctx context.Context, db *sqlx.DB) *RoleMModel {
	return &RoleMModel{
		MModel: impl.MModel{
			Context:   ctx,
			Name:      "Role",
			TableName: "roles",
			DB:        db,
			Fields: []interfaces.IMField{
				impl.MField{Type: "int64", DbType: "INTEGER", JsonTag: "id", DbTag: "id", IsPrimaryKey: true, IsIdentity: true, Required: true, IsNotUpdateField: true},
				impl.MField{Type: "int64", DbType: "BIGINT", JsonTag: "account_id", DbTag: "account_id", Default: "0", Required: true, IsNotUpdateField: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "role_name", DbTag: "role_name", MaxLength: 255, Required: true},
				impl.MField{Type: "[]string", DbType: "TEXT[]", JsonTag: "actions", DbTag: "actions", Default: "'{}'::TEXT[]", Required: true},
				impl.MField{Type: "time.Time", DbType: "TIMESTAMPTZ", JsonTag: "-", DbTag: "created_at", Default: "NOW()", Required: true, IsNotUpdateField: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "created_by", DbTag: "created_by", MaxLength: 255, Required: true, IsNotUpdateField: true},
				impl.MField{Type: "sql.NullTime", DbType: "TIMESTAMPTZ", JsonTag: "-", DbTag: "updated_at", IsNullable: true, IsNotCreateField: true},
				impl.MField{Type: "sql.NullString", DbType: "VARCHAR", JsonTag: "-", DbTag: "updated_by", MaxLength: 255, IsNullable: true, IsNotCreateField: true},
			},
		},
	}
}

func (m *RoleMModel) Migrate() error {
	if m == nil {
		return errors.New("role metadata model is nil")
	}
	if m.DB == nil {
		return errors.New("role metadata database is nil")
	}
	tx, err := m.DB.BeginTxx(m.ContextOrBackground(), nil)
	if err != nil {
		return fmt.Errorf("begin roles migration: %w", err)
	}
	defer tx.Rollback()

	migrator := impl.NewMMigrator(&m.MModel, tx)
	migrator.SetContext(m.ContextOrBackground())
	if err := migrator.Migrate(); err != nil {
		return err
	}
	statements := []string{
		`ALTER TABLE roles ADD COLUMN IF NOT EXISTS account_id BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE roles ADD COLUMN IF NOT EXISTS role_name VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE roles ADD COLUMN IF NOT EXISTS actions TEXT[] NOT NULL DEFAULT '{}'::TEXT[]`,
		`ALTER TABLE roles ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`ALTER TABLE roles ADD COLUMN IF NOT EXISTS created_by VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE roles ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ`,
		`ALTER TABLE roles ADD COLUMN IF NOT EXISTS updated_by VARCHAR(255)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_account_name ON roles (account_id, LOWER(role_name))`,
		`CREATE INDEX IF NOT EXISTS idx_roles_account_id ON roles (account_id)`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(m.ContextOrBackground(), statement); err != nil {
			return fmt.Errorf("run roles migration statement %q: %w", statement, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit roles migration: %w", err)
	}
	return nil
}

func (m *RoleMModel) ValidateCreate() error {
	if m == nil {
		return errors.New("role metadata model is nil")
	}
	if _, err := m.MModel.ValidateAccountID_(m.AuthAccountID, m.AccountID); err != nil {
		return err
	}
	m.RoleName = strings.TrimSpace(m.RoleName)
	if m.RoleName == "" {
		return errors.New("nama role wajib diisi")
	}
	return m.MModel.ValidateCreate_(m)
}

func (m *RoleMModel) ValidateID() error {
	if m == nil {
		return errors.New("id role tidak valid")
	}
	return m.MModel.ValidateID_(m.ID, "role")
}

func (m *RoleMModel) Create() (int64, error) {
	if m == nil {
		return 0, errors.New("role metadata model is nil")
	}
	accountID, err := m.resolveAccountID()
	if err != nil {
		return 0, err
	}
	m.AccountID = accountID
	m.RoleName = strings.TrimSpace(m.RoleName)
	m.Actions = normalizedRoleActions(m.Actions)
	if err := m.ValidateCreate(); err != nil {
		return 0, err
	}
	m.CreatedBy, err = m.MModel.ResolveCreatedBy_(m.CreatedBy, m.AuthEmail)
	if err != nil {
		return 0, err
	}
	return m.MModel.Create_(m)
}

func (m *RoleMModel) Update() error {
	if m == nil {
		return errors.New("role metadata model is nil")
	}
	accountID, err := m.resolveAccountID()
	if err != nil {
		return err
	}
	if err := m.ValidateID(); err != nil {
		return err
	}
	m.AccountID = accountID
	m.RoleName = strings.TrimSpace(m.RoleName)
	m.Actions = normalizedRoleActions(m.Actions)
	if m.RoleName == "" {
		return errors.New("nama role wajib diisi")
	}
	if err := m.MModel.ValidateUpdate_(m); err != nil {
		return err
	}
	m.UpdatedBy, err = m.MModel.ResolveUpdatedBy_(m.UpdatedBy, m.AuthEmail)
	if err != nil {
		return err
	}
	m.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	return m.MModel.Update_(m, "id", "account_id")
}

func (m *RoleMModel) Delete() error {
	if err := m.ValidateID(); err != nil {
		return err
	}
	accountID, err := m.resolveAccountID()
	if err != nil {
		return err
	}
	m.AccountID = accountID
	return m.MModel.Delete_(m, "id", "account_id")
}

func (m *RoleMModel) FindByID() (*RoleMModel, error) {
	if err := m.ValidateID(); err != nil {
		return nil, err
	}
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}
	found := &RoleMModel{ID: m.ID, AccountID: accountID, AuthAccountID: m.AuthAccountID, MModel: m.MModel}
	if err := m.MModel.FindByID_(found, "id", "account_id"); err != nil {
		return nil, err
	}
	return found, nil
}

func (m *RoleMModel) FindByName(name string) (*RoleMModel, error) {
	if m == nil {
		return nil, errors.New("role metadata model is nil")
	}
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("nama role wajib diisi")
	}
	return m.findByNameInsensitive(accountID, name)
}

func (m *RoleMModel) FindList(search string) ([]RoleMModel, error) {
	return m.findList(50, strings.TrimSpace(search), false)
}

func (m *RoleMModel) FindAll() ([]RoleMModel, error) {
	return m.findList(0, "", false)
}

func (m *RoleMModel) FindRecentList(limit int, search string) ([]RoleMModel, error) {
	if limit <= 0 {
		limit = 5
	}
	return m.findList(limit, strings.TrimSpace(search), true)
}

func (m *RoleMModel) findList(limit int, search string, recent bool) ([]RoleMModel, error) {
	if m == nil {
		return nil, errors.New("role metadata model is nil")
	}
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}
	if m.DB == nil {
		return nil, errors.New("role metadata database is nil")
	}
	query := `SELECT id, account_id, role_name, actions, created_at, created_by, updated_at, updated_by FROM roles WHERE account_id = $1`
	args := []any{accountID}
	if search != "" {
		args = append(args, search)
		query += fmt.Sprintf(` AND (role_name ILIKE '%%' || $%d || '%%' OR EXISTS (SELECT 1 FROM unnest(actions) a WHERE a ILIKE '%%' || $%d || '%%'))`, len(args), len(args))
	}
	if recent {
		query += ` ORDER BY created_at DESC, id DESC`
	} else {
		query += ` ORDER BY role_name ASC`
	}
	if limit > 0 {
		args = append(args, limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
	}
	items := make([]RoleMModel, 0)
	if err := m.DB.SelectContext(m.ContextOrBackground(), &items, query, args...); err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	for i := range items {
		items[i].MModel = m.MModel
	}
	return items, nil
}

func (m *RoleMModel) findByNameInsensitive(accountID int64, name string) (*RoleMModel, error) {
	if m.DB == nil {
		return nil, errors.New("role metadata database is nil")
	}
	query := `SELECT id, account_id, role_name, actions, created_at, created_by, updated_at, updated_by FROM roles WHERE account_id = $1 AND LOWER(role_name) = LOWER($2)`
	found := &RoleMModel{MModel: m.MModel}
	if err := m.DB.GetContext(m.ContextOrBackground(), found, query, accountID, name); err != nil {
		return nil, err
	}
	return found, nil
}

func (m *RoleMModel) resolveAccountID() (int64, error) {
	if m == nil {
		return 0, errors.New("role metadata model is nil")
	}
	return m.MModel.ResolveAccountID(m.AuthAccountID, m.AccountID)
}

func normalizedRoleActions(actions pq.StringArray) pq.StringArray {
	result := make(pq.StringArray, 0, len(actions))
	for _, action := range actions {
		if action = strings.TrimSpace(action); action != "" {
			result = append(result, action)
		}
	}
	return result
}
