package models

import (
	"context"
	"time"

	"backend/shared/mmodels"
	"github.com/jmoiron/sqlx"
)

// AccountDevice remains as a compatibility adapter for existing model callers.
// Persistence and device behavior live in AccountDeviceMModel.
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

func NewAccountDevice(db *sqlx.DB, auth Auth) AccountDevice {
	return AccountDevice{BaseModel: BaseModel{DB: db, Auth: auth}, IsActive: true}
}

func (m AccountDevice) Migrate() {
	if err := m.mmodel().Migrate(); err != nil {
		panic(err)
	}
}

func (m AccountDevice) ValidateID() error {
	return m.mmodel().ValidateID()
}

func (m AccountDevice) FindByDevice() (*AccountDevice, error) {
	item, err := m.mmodel().FindByDevice()
	if err != nil {
		return nil, err
	}
	return accountDeviceFromMModel(item, m.DB, m.Auth), nil
}

func (m AccountDevice) UpsertLogin(lastLoginAt time.Time) (*AccountDevice, error) {
	item, err := m.mmodel().UpsertLogin(lastLoginAt)
	if err != nil {
		return nil, err
	}
	return accountDeviceFromMModel(item, m.DB, m.Auth), nil
}

func (m AccountDevice) UpdateLocation(locationID *int64, actor string) (*AccountDevice, error) {
	item, err := m.mmodel().UpdateLocation(locationID, actor)
	if err != nil {
		return nil, err
	}
	return accountDeviceFromMModel(item, m.DB, m.Auth), nil
}

func (m AccountDevice) mmodel() *mmodels.AccountDeviceMModel {
	model := mmodels.NewAccountDeviceMModel(context.Background(), m.DB)
	model.ID = m.ID
	model.AccountID = m.AccountID
	model.DeviceID = m.DeviceID
	model.LocationID = m.LocationID
	model.LastLoginAt = m.LastLoginAt
	model.LastLocationSelectedAt = m.LastLocationSelectedAt
	model.IsActive = m.IsActive
	model.CreatedBy = m.CreatedBy
	model.AuthAccountID = m.Auth.AccountID
	model.AuthUserName = m.Auth.UserName
	model.AuthEmail = m.Auth.Email
	return model
}

func accountDeviceFromMModel(value *mmodels.AccountDeviceMModel, db *sqlx.DB, auth Auth) *AccountDevice {
	if value == nil {
		return nil
	}
	item := &AccountDevice{
		BaseModel: BaseModel{DB: db, Auth: auth, CreatedAt: value.CreatedAt, CreatedBy: value.CreatedBy},
		ID:        value.ID, AccountID: value.AccountID, DeviceID: value.DeviceID,
		LocationID: value.LocationID, LastLoginAt: value.LastLoginAt,
		LastLocationSelectedAt: value.LastLocationSelectedAt, IsActive: value.IsActive,
	}
	if value.UpdatedAt.Valid {
		updatedAt := value.UpdatedAt.Time
		item.UpdatedAt = &updatedAt
	}
	if value.UpdatedBy.Valid {
		updatedBy := value.UpdatedBy.String
		item.UpdatedBy = &updatedBy
	}
	return item
}
