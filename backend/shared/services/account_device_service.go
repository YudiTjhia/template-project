package services

import (
	"context"
	"time"

	"backend/shared/mmodels"
	"backend/shared/models"

	"github.com/jmoiron/sqlx"
)

type AccountDeviceService struct {
	Context context.Context
	DB      *sqlx.DB
	Auth    models.Auth
}

func NewAccountDeviceService(db *sqlx.DB, auth models.Auth) *AccountDeviceService {
	return NewAccountDeviceServiceWithContext(context.Background(), db, auth)
}

func NewAccountDeviceServiceWithContext(ctx context.Context, db *sqlx.DB, auth models.Auth) *AccountDeviceService {
	if ctx == nil {
		ctx = context.Background()
	}
	return &AccountDeviceService{Context: ctx, DB: db, Auth: auth}
}

// NewAccountDeviceMModel returns a metadata model preloaded with the service's
// database, request context, and authenticated account details.
func (s *AccountDeviceService) NewAccountDeviceMModel() *mmodels.AccountDeviceMModel {
	ctx := s.Context
	if ctx == nil {
		ctx = context.Background()
	}
	model := mmodels.NewAccountDeviceMModel(ctx, s.DB)
	model.AccountID = s.Auth.AccountID
	model.AuthAccountID = s.Auth.AccountID
	model.AuthUserName = s.Auth.UserName
	model.AuthEmail = s.Auth.Email
	return model
}

// NewAccountDevice preserves the legacy DTO factory for existing callers.
func (s *AccountDeviceService) NewAccountDevice() models.AccountDevice {
	return models.NewAccountDevice(s.DB, s.Auth)
}

func (s *AccountDeviceService) accountDeviceMModel() *mmodels.AccountDeviceMModel {
	return s.NewAccountDeviceMModel()
}

func (s *AccountDeviceService) accountDeviceFromMModel(value mmodels.AccountDeviceMModel) *models.AccountDevice {
	device := &models.AccountDevice{
		BaseModel: models.BaseModel{DB: s.DB, Auth: s.Auth, CreatedAt: value.CreatedAt, CreatedBy: value.CreatedBy},
		ID:        value.ID, AccountID: value.AccountID, DeviceID: value.DeviceID,
		LocationID: value.LocationID, LastLoginAt: value.LastLoginAt,
		LastLocationSelectedAt: value.LastLocationSelectedAt, IsActive: value.IsActive,
	}
	if value.UpdatedAt.Valid {
		updatedAt := value.UpdatedAt.Time
		device.UpdatedAt = &updatedAt
	}
	if value.UpdatedBy.Valid {
		updatedBy := value.UpdatedBy.String
		device.UpdatedBy = &updatedBy
	}
	return device
}

func (s *AccountDeviceService) Migrate() error {
	return s.accountDeviceMModel().Migrate()
}

func (s *AccountDeviceService) UpsertLogin(deviceID string) (*models.AccountDevice, error) {
	model := s.accountDeviceMModel()
	model.DeviceID = deviceID
	item, err := model.UpsertLogin(time.Now().UTC())
	if err != nil {
		return nil, err
	}
	return s.accountDeviceFromMModel(*item), nil
}

func (s *AccountDeviceService) UpdateLocation(deviceID string, locationID int64) (*models.AccountDevice, error) {
	model := s.accountDeviceMModel()
	model.DeviceID = deviceID
	if locationID > 0 {
		model.LocationID = &locationID
	}
	item, err := model.UpdateLocation(model.LocationID, s.Auth.UserName)
	if err != nil {
		return nil, err
	}
	return s.accountDeviceFromMModel(*item), nil
}

func (s *AccountDeviceService) FindByDevice(deviceID string) (*models.AccountDevice, error) {
	model := s.accountDeviceMModel()
	model.DeviceID = deviceID
	item, err := model.FindByDevice()
	if err != nil {
		return nil, err
	}
	return s.accountDeviceFromMModel(*item), nil
}
