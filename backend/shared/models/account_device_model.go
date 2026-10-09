package models

import (
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type AccountDevice struct {
	BaseModel

	ID                     int64      `json:"id" db:"id"`
	AccountID              int64      `json:"account_id" db:"account_id"`
	DeviceID               string     `json:"device_id" db:"device_id"`
	LocationID             *int64     `json:"location_id,omitempty" db:"location_id"`
	LastLoginAt            *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	LastLocationSelectedAt *time.Time `json:"last_location_selected_at,omitempty" db:"last_location_selected_at"`
	IsActive               bool       `json:"is_active" db:"is_active"`
}

const accountDeviceSelectColumns = `ad.id, ad.account_id, ad.device_id, ad.location_id, ad.last_login_at, ad.last_location_selected_at, ad.is_active,
ad.created_at, ad.created_by, ad.updated_at, ad.updated_by`

func NewAccountDevice(db *sqlx.DB, auth Auth) AccountDevice {
	return AccountDevice{
		BaseModel: BaseModel{DB: db, Auth: auth},
		IsActive:  true,
	}
}

func (m AccountDevice) Migrate() {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS account_devices (
	id SERIAL PRIMARY KEY,
	account_id BIGINT NOT NULL DEFAULT 0,
	device_id VARCHAR(255) NOT NULL DEFAULT '',
	location_id BIGINT,
	last_login_at TIMESTAMPTZ,
	last_location_selected_at TIMESTAMPTZ,
	is_active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_by VARCHAR(255) NOT NULL DEFAULT '',
	updated_at TIMESTAMPTZ,
	updated_by VARCHAR(255)
);`,
		`ALTER TABLE account_devices ADD COLUMN IF NOT EXISTS account_id BIGINT NOT NULL DEFAULT 0;`,
		`ALTER TABLE account_devices ADD COLUMN IF NOT EXISTS device_id VARCHAR(255) NOT NULL DEFAULT '';`,
		`ALTER TABLE account_devices ADD COLUMN IF NOT EXISTS location_id BIGINT;`,
		`ALTER TABLE account_devices ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMPTZ;`,
		`ALTER TABLE account_devices ADD COLUMN IF NOT EXISTS last_location_selected_at TIMESTAMPTZ;`,
		`ALTER TABLE account_devices ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true;`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_account_devices_account_device ON account_devices (account_id, device_id);`,
		`CREATE INDEX IF NOT EXISTS idx_account_devices_account_id ON account_devices (account_id);`,
		`CREATE INDEX IF NOT EXISTS idx_account_devices_location_id ON account_devices (location_id);`,
	}

	for _, statement := range statements {
		log.Println("[SQL]=", statement)
		m.DB.MustExec(statement)
	}
}

func (m AccountDevice) ValidateID() error {
	if m.ID <= 0 {
		return errors.New("id account device tidak valid")
	}
	return nil
}

func (m *AccountDevice) normalize() error {
	if m == nil {
		return errors.New("account device tidak valid")
	}
	m.DeviceID = strings.ToLower(strings.TrimSpace(m.DeviceID))
	if m.DeviceID == "" {
		return errors.New("device id wajib diisi")
	}
	if m.AccountID <= 0 {
		if accountID, err := m.resolveAccountID(); err == nil {
			m.AccountID = accountID
		} else {
			return err
		}
	}
	if m.CreatedBy == "" {
		m.CreatedBy = firstNonEmpty(m.Auth.UserName, m.Auth.Email, "system")
	}
	if m.CreatedBy == "" {
		m.CreatedBy = "system"
	}
	m.IsActive = true
	return nil
}

func (m AccountDevice) resolveAccountID() (int64, error) {
	if m.AccountID > 0 {
		return m.AccountID, nil
	}
	if m.Auth.AccountID > 0 {
		return m.Auth.AccountID, nil
	}
	return 0, errors.New("id account wajib diisi")
}

func (m AccountDevice) FindByDevice() (*AccountDevice, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}
	deviceID := strings.TrimSpace(m.DeviceID)
	if deviceID == "" {
		return nil, errors.New("device id wajib diisi")
	}

	var item AccountDevice
	err = m.DB.Get(&item, `SELECT `+accountDeviceSelectColumns+` FROM account_devices ad WHERE ad.account_id = $1 AND ad.device_id = $2 LIMIT 1`, accountID, strings.ToLower(deviceID))
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (m AccountDevice) UpsertLogin(lastLoginAt time.Time) (*AccountDevice, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}
	if err := m.normalize(); err != nil {
		return nil, err
	}
	if lastLoginAt.IsZero() {
		lastLoginAt = time.Now().UTC()
	}
	createdBy := firstNonEmpty(m.CreatedBy, m.Auth.UserName, m.Auth.Email, "system")
	updatedBy := createdBy

	var locationID interface{}
	if m.LocationID != nil {
		locationID = *m.LocationID
	}

	var id int64
	err = m.DB.QueryRow(`
		INSERT INTO account_devices
		(account_id, device_id, location_id, last_login_at, last_location_selected_at, is_active, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, TRUE, $6, $7)
		ON CONFLICT (account_id, device_id) DO UPDATE SET
			location_id = COALESCE(EXCLUDED.location_id, account_devices.location_id),
			last_login_at = EXCLUDED.last_login_at,
			is_active = TRUE,
			updated_at = NOW(),
			updated_by = EXCLUDED.updated_by
		RETURNING id`,
		accountID,
		m.DeviceID,
		locationID,
		lastLoginAt,
		lastLoginAt,
		createdBy,
		updatedBy,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	item, err := m.findByID(id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (m AccountDevice) UpdateLocation(locationID *int64, actor string) (*AccountDevice, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}
	if err := m.normalize(); err != nil {
		return nil, err
	}
	if locationID != nil && *locationID <= 0 {
		locationID = nil
	}
	actor = firstNonEmpty(actor, m.Auth.UserName, m.Auth.Email, "system")

	var id int64
	var dbLocationID interface{}
	if locationID != nil {
		dbLocationID = *locationID
	}
	err = m.DB.QueryRow(`
		UPDATE account_devices
		SET location_id = $1,
			last_location_selected_at = NOW(),
			is_active = TRUE,
			updated_at = NOW(),
			updated_by = $2
		WHERE account_id = $3 AND device_id = $4
		RETURNING id`,
		dbLocationID,
		actor,
		accountID,
		m.DeviceID,
	).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return m.UpsertLogin(time.Now().UTC())
		}
		return nil, err
	}

	item, err := m.findByID(id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (m AccountDevice) findByID(id int64) (*AccountDevice, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}
	var item AccountDevice
	err = m.DB.Get(&item, `SELECT `+accountDeviceSelectColumns+` FROM account_devices ad WHERE ad.id = $1 AND ad.account_id = $2`, id, accountID)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
