package models

import (
	"backend/shared/mmodels"
	"database/sql/driver"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type StringArray []string

func (a *StringArray) Scan(src interface{}) error {
	return pq.Array((*[]string)(a)).Scan(src)
}

func (a StringArray) Value() (driver.Value, error) {
	return pq.Array([]string(a)).Value()
}

type Role struct {
	BaseModel

	ID        int64       `json:"id" db:"id"`
	AccountID int64       `json:"account_id" db:"account_id"`
	RoleName  string      `json:"role_name" db:"role_name"`
	Actions   StringArray `json:"actions" db:"actions"`
}

const roleSelectColumns = `r.id, r.account_id, r.role_name, r.actions,
r.created_at, r.created_by, r.updated_at, r.updated_by`

func NewRole(db *sqlx.DB, auth Auth) Role {
	role := Role{
		BaseModel: BaseModel{DB: db, Auth: auth},
	}
	role.AccountID = auth.AccountID
	role.CreatedBy = auth.Email
	return role
}

func (m Role) Migrate() {
	sql := `CREATE TABLE IF NOT EXISTS ` + mmodels.RoleTable + ` (
	id SERIAL PRIMARY KEY,
	account_id BIGINT NOT NULL DEFAULT 0,
	role_name VARCHAR(255) NOT NULL,
	actions TEXT[] NOT NULL DEFAULT '{}'::TEXT[],
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_by VARCHAR(255) NOT NULL,
	updated_at TIMESTAMPTZ,
	updated_by VARCHAR(255)
);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_account_name ON ` + mmodels.RoleTable + ` (account_id, LOWER(role_name));`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `CREATE INDEX IF NOT EXISTS idx_roles_account_id ON ` + mmodels.RoleTable + ` (account_id);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE ` + mmodels.RoleTable + ` ADD COLUMN IF NOT EXISTS account_id BIGINT NOT NULL DEFAULT 0;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE ` + mmodels.RoleTable + ` ADD COLUMN IF NOT EXISTS role_name VARCHAR(255) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE ` + mmodels.RoleTable + ` ADD COLUMN IF NOT EXISTS actions TEXT[] NOT NULL DEFAULT '{}'::TEXT[];`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
}

func (m Role) TableName() string {
	return mmodels.RoleTable
}

func (m Role) ValidateCreate() error {
	if _, err := m.resolveAccountID(); err != nil {
		return err
	}

	if strings.TrimSpace(m.RoleName) == "" {
		return errors.New("nama role wajib diisi")
	}

	return nil
}

func (m Role) ValidateID() error {
	if m.ID <= 0 {
		return errors.New("id role tidak valid")
	}
	return nil
}

func (m Role) Create() (int64, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return 0, err
	}

	m.RoleName = strings.TrimSpace(m.RoleName)
	m.Actions = normalizeActions(m.Actions)

	if err := m.ValidateCreate(); err != nil {
		return 0, err
	}

	createdBy := m.CreatedBy
	if createdBy == "" {
		createdBy = m.Auth.Email
	}
	if createdBy == "" {
		return 0, errors.New("created_by wajib diisi")
	}

	sql := `INSERT INTO ` + mmodels.RoleTable + `
	(account_id, role_name, actions, created_by)
	VALUES ($1, $2, $3, $4)
	RETURNING id`

	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{
		"account_id": accountID,
		"role_name":  m.RoleName,
		"actions":    []string(m.Actions),
		"created_by": createdBy,
	})

	var id int64
	err = m.DB.QueryRow(sql, accountID, m.RoleName, m.Actions, createdBy).Scan(&id)
	return id, err
}

func (m Role) Update() error {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return err
	}

	m.RoleName = strings.TrimSpace(m.RoleName)
	m.Actions = normalizeActions(m.Actions)

	if err := m.ValidateID(); err != nil {
		return err
	}
	if err := m.ValidateCreate(); err != nil {
		return err
	}

	updatedBy := m.UpdatedBy
	if updatedBy == nil || *updatedBy == "" {
		if m.Auth.Email != "" {
			updatedBy = &m.Auth.Email
		}
	}

	sql := `UPDATE ` + mmodels.RoleTable + `
	SET role_name = $1,
		actions = $2,
		updated_at = NOW(),
		updated_by = $3
	WHERE id = $4 AND account_id = $5`

	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{
		"role_name":  m.RoleName,
		"actions":    []string(m.Actions),
		"updated_by": updatedBy,
		"id":         m.ID,
		"account_id": accountID,
	})

	_, err = m.DB.Exec(sql, m.RoleName, m.Actions, updatedBy, m.ID, accountID)
	return err
}

func (m Role) Delete() error {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return err
	}
	if err := m.ValidateID(); err != nil {
		return err
	}

	sql := `DELETE FROM ` + mmodels.RoleTable + ` WHERE id = $1 AND account_id = $2`
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"id": m.ID, "account_id": accountID})

	_, err = m.DB.Exec(sql, m.ID, accountID)
	return err
}

func (m Role) FindByID() (*Role, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}
	if err := m.ValidateID(); err != nil {
		return nil, err
	}

	sql := roleSelectQuery(`r.id = $1 AND r.account_id = $2`)
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"id": m.ID, "account_id": accountID})

	var role Role
	err = m.DB.Get(&role, sql, m.ID, accountID)
	return &role, err
}

func (m Role) FindByName(name string) (*Role, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("nama role wajib diisi")
	}

	sql := roleSelectQuery(`r.account_id = $1 AND LOWER(r.role_name) = LOWER($2)`)
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"account_id": accountID, "role_name": name})

	var role Role
	err = m.DB.Get(&role, sql, accountID, name)
	return &role, err
}

func (m Role) FindList(search string) ([]Role, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}

	sql := roleSelectQuery(`r.account_id = $1`)
	args := []interface{}{accountID}
	where := []string{}
	if search != "" {
		args = append(args, search)
		placeholder := nextSQLPlaceholder(args)
		where = append(where, `(r.role_name ILIKE '%' || `+placeholder+` || '%' OR EXISTS (SELECT 1 FROM unnest(r.actions) a WHERE a ILIKE '%' || `+placeholder+` || '%'))`)
	}
	if len(where) > 0 {
		sql += " AND " + strings.Join(where, " AND ")
	}
	sql += ` ORDER BY r.role_name ASC LIMIT 50`

	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"account_id": accountID, "search": search})

	var roles []Role
	err = m.DB.Select(&roles, sql, args...)
	return roles, err
}

func (m Role) FindAll() ([]Role, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}

	sql := roleSelectQuery(`r.account_id = $1`)
	sql += ` ORDER BY r.role_name ASC`

	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"account_id": accountID})

	var roles []Role
	err = m.DB.Select(&roles, sql, accountID)
	return roles, err
}

func (m Role) FindRecentList(limit int, search string) ([]Role, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 5
	}

	sql := roleSelectQuery(`r.account_id = $1`)
	args := []interface{}{accountID}
	where := []string{}
	if search != "" {
		args = append(args, search)
		placeholder := nextSQLPlaceholder(args)
		where = append(where, `(r.role_name ILIKE '%' || `+placeholder+` || '%' OR EXISTS (SELECT 1 FROM unnest(r.actions) a WHERE a ILIKE '%' || `+placeholder+` || '%'))`)
	}
	if len(where) > 0 {
		sql += " AND " + strings.Join(where, " AND ")
	}
	sql += fmt.Sprintf(` ORDER BY r.created_at DESC, r.id DESC LIMIT %d`, limit)

	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"account_id": accountID, "search": search, "limit": limit})

	var roles []Role
	err = m.DB.Select(&roles, sql, args...)
	return roles, err
}

func roleSelectQuery(where string) string {
	return `SELECT ` + roleSelectColumns + `
	FROM ` + mmodels.RoleTable + ` r
	WHERE ` + where
}

func normalizeActions(actions StringArray) StringArray {
	normalized := make(StringArray, 0, len(actions))
	for _, action := range actions {
		action = strings.TrimSpace(action)
		if action != "" {
			normalized = append(normalized, action)
		}
	}
	return normalized
}

func (m Role) resolveAccountID() (int64, error) {
	if m.Auth.AccountID > 0 {
		return m.Auth.AccountID, nil
	}
	if m.AccountID > 0 {
		return m.AccountID, nil
	}
	return 0, errors.New("id akun wajib diisi")
}
