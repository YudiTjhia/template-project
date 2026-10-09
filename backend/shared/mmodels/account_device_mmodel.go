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

// AccountDeviceMModel stores the last known device and location for an account.
type AccountDeviceMModel struct {
	impl.MModel

	ID                     int64          `json:"id" db:"id"`
	AccountID              int64          `json:"account_id" db:"account_id"`
	DeviceID               string         `json:"device_id" db:"device_id"`
	LocationID             *int64         `json:"location_id,omitempty" db:"location_id"`
	LastLoginAt            *time.Time     `json:"last_login_at,omitempty" db:"last_login_at"`
	LastLocationSelectedAt *time.Time     `json:"last_location_selected_at,omitempty" db:"last_location_selected_at"`
	IsActive               bool           `json:"is_active" db:"is_active"`
	CreatedAt              time.Time      `json:"-" db:"created_at"`
	CreatedBy              string         `json:"created_by" db:"created_by"`
	UpdatedAt              sql.NullTime   `json:"-" db:"updated_at"`
	UpdatedBy              sql.NullString `json:"-" db:"updated_by"`
	AuthAccountID          int64          `json:"-" db:"-"`
	AuthUserName           string         `json:"-" db:"-"`
	AuthEmail              string         `json:"-" db:"-"`
}

func NewAccountDeviceMModel(ctx context.Context, db *sqlx.DB) *AccountDeviceMModel {
	return &AccountDeviceMModel{
		IsActive: true,
		MModel: impl.MModel{
			Context: ctx, Name: "AccountDevice", TableName: "account_devices", DB: db,
			Fields: []interfaces.IMField{
				impl.MField{Type: "int64", DbType: "INTEGER", JsonTag: "id", DbTag: "id", IsPrimaryKey: true, IsIdentity: true, Required: true, IsNotUpdateField: true},
				impl.MField{Type: "int64", DbType: "BIGINT", JsonTag: "account_id", DbTag: "account_id", Required: true, IsNotUpdateField: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "device_id", DbTag: "device_id", MaxLength: 255, Required: true, IsNotUpdateField: true},
				impl.MField{Type: "int64", DbType: "BIGINT", JsonTag: "location_id", DbTag: "location_id", IsNullable: true},
				impl.MField{Type: "time.Time", DbType: "TIMESTAMPTZ", JsonTag: "last_login_at", DbTag: "last_login_at", IsNullable: true},
				impl.MField{Type: "time.Time", DbType: "TIMESTAMPTZ", JsonTag: "last_location_selected_at", DbTag: "last_location_selected_at", IsNullable: true},
				impl.MField{Type: "bool", DbType: "BOOLEAN", JsonTag: "is_active", DbTag: "is_active", Default: "true", Required: true},
				impl.MField{Type: "time.Time", DbType: "TIMESTAMPTZ", JsonTag: "-", DbTag: "created_at", Default: "NOW()", Required: true, IsNotUpdateField: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "created_by", DbTag: "created_by", MaxLength: 255, Default: "''", Required: true, IsNotUpdateField: true},
				impl.MField{Type: "sql.NullTime", DbType: "TIMESTAMPTZ", JsonTag: "-", DbTag: "updated_at", IsNullable: true, IsNotCreateField: true},
				impl.MField{Type: "sql.NullString", DbType: "VARCHAR", JsonTag: "-", DbTag: "updated_by", MaxLength: 255, IsNullable: true, IsNotCreateField: true},
			},
		},
	}
}

func (m *AccountDeviceMModel) Migrate() error {
	if m == nil {
		return errors.New("account device metadata model is nil")
	}
	if m.DB == nil {
		return errors.New("account device metadata database is nil")
	}
	tx, err := m.DB.BeginTxx(m.ContextOrBackground(), nil)
	if err != nil {
		return fmt.Errorf("begin account devices migration: %w", err)
	}
	defer tx.Rollback()

	migrator := impl.NewMMigrator(&m.MModel, tx)
	migrator.SetContext(m.ContextOrBackground())
	if err := migrator.Migrate(); err != nil {
		return err
	}
	statements := []string{
		`ALTER TABLE account_devices ADD COLUMN IF NOT EXISTS account_id BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE account_devices ADD COLUMN IF NOT EXISTS device_id VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE account_devices ADD COLUMN IF NOT EXISTS location_id BIGINT`,
		`ALTER TABLE account_devices ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMPTZ`,
		`ALTER TABLE account_devices ADD COLUMN IF NOT EXISTS last_location_selected_at TIMESTAMPTZ`,
		`ALTER TABLE account_devices ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true`,
		`ALTER TABLE account_devices ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`ALTER TABLE account_devices ADD COLUMN IF NOT EXISTS created_by VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE account_devices ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ`,
		`ALTER TABLE account_devices ADD COLUMN IF NOT EXISTS updated_by VARCHAR(255)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_account_devices_account_device ON account_devices (account_id, device_id)`,
		`CREATE INDEX IF NOT EXISTS idx_account_devices_account_id ON account_devices (account_id)`,
		`CREATE INDEX IF NOT EXISTS idx_account_devices_location_id ON account_devices (location_id)`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(m.ContextOrBackground(), statement); err != nil {
			return fmt.Errorf("run account devices migration statement %q: %w", statement, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit account devices migration: %w", err)
	}
	return nil
}

func (m *AccountDeviceMModel) ValidateID() error {
	if m == nil {
		return errors.New("id account device tidak valid")
	}
	return m.MModel.ValidateID_(m.ID, "account device")
}

func (m *AccountDeviceMModel) normalize() error {
	if m == nil {
		return errors.New("account device tidak valid")
	}
	if m.DB == nil {
		return errors.New("account device database is nil")
	}
	m.DeviceID = strings.ToLower(strings.TrimSpace(m.DeviceID))
	if m.DeviceID == "" {
		return errors.New("device id wajib diisi")
	}
	if m.AccountID <= 0 {
		if m.AuthAccountID > 0 {
			m.AccountID = m.AuthAccountID
		} else {
			return errors.New("id account wajib diisi")
		}
	}
	m.CreatedBy = firstNonEmptyDevice(m.CreatedBy, m.AuthUserName, m.AuthEmail, "system")
	m.IsActive = true
	return m.MModel.ValidateCreate_(m)
}

func (m *AccountDeviceMModel) FindByDevice() (*AccountDeviceMModel, error) {
	if err := m.normalize(); err != nil {
		return nil, err
	}
	found := &AccountDeviceMModel{AccountID: m.AccountID, DeviceID: m.DeviceID, MModel: m.MModel}
	if err := m.MModel.FindByFields(found, "account_id", "device_id"); err != nil {
		return nil, err
	}
	return found, nil
}

func (m *AccountDeviceMModel) UpsertLogin(lastLoginAt time.Time) (*AccountDeviceMModel, error) {
	if err := m.normalize(); err != nil {
		return nil, err
	}
	if lastLoginAt.IsZero() {
		lastLoginAt = time.Now().UTC()
	}
	var locationID any
	if m.LocationID != nil {
		locationID = *m.LocationID
	}
	var id int64
	err := m.DB.QueryRowxContext(m.ContextOrBackground(), `
		INSERT INTO account_devices
		(account_id, device_id, location_id, last_login_at, last_location_selected_at, is_active, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, TRUE, $6, $7)
		ON CONFLICT (account_id, device_id) DO UPDATE SET
			location_id = COALESCE(EXCLUDED.location_id, account_devices.location_id),
			last_login_at = EXCLUDED.last_login_at,
			is_active = TRUE,
			updated_at = NOW(),
			updated_by = EXCLUDED.updated_by
		RETURNING id`, m.AccountID, m.DeviceID, locationID, lastLoginAt, lastLoginAt, m.CreatedBy, m.CreatedBy).Scan(&id)
	if err != nil {
		return nil, err
	}
	return m.findByID(id)
}

func (m *AccountDeviceMModel) UpdateLocation(locationID *int64, actor string) (*AccountDeviceMModel, error) {
	if err := m.normalize(); err != nil {
		return nil, err
	}
	if locationID != nil && *locationID <= 0 {
		locationID = nil
	}
	actor = firstNonEmptyDevice(actor, m.AuthUserName, m.AuthEmail, "system")
	var dbLocationID any
	if locationID != nil {
		dbLocationID = *locationID
	}
	var id int64
	err := m.DB.QueryRowxContext(m.ContextOrBackground(), `
		UPDATE account_devices
		SET location_id = $1, last_location_selected_at = NOW(), is_active = TRUE,
			updated_at = NOW(), updated_by = $2
		WHERE account_id = $3 AND device_id = $4
		RETURNING id`, dbLocationID, actor, m.AccountID, m.DeviceID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return m.UpsertLogin(time.Now().UTC())
		}
		return nil, err
	}
	return m.findByID(id)
}

func (m *AccountDeviceMModel) findByID(id int64) (*AccountDeviceMModel, error) {
	found := &AccountDeviceMModel{ID: id, AccountID: m.AccountID, MModel: m.MModel}
	if err := m.MModel.FindByID_(found, "id", "account_id"); err != nil {
		return nil, err
	}
	return found, nil
}

func firstNonEmptyDevice(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
