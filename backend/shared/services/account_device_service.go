package services

import (
	"time"

	"backend/shared/models"

	"github.com/jmoiron/sqlx"
)

type AccountDeviceService struct {
	DB   *sqlx.DB
	Auth models.Auth
}

func NewAccountDeviceService(db *sqlx.DB, auth models.Auth) *AccountDeviceService {
	return &AccountDeviceService{DB: db, Auth: auth}
}

func (s *AccountDeviceService) NewAccountDevice() models.AccountDevice {
	return models.NewAccountDevice(s.DB, s.Auth)
}

func (s *AccountDeviceService) Migrate() {
	s.NewAccountDevice().Migrate()
}

func (s *AccountDeviceService) UpsertLogin(deviceID string) (*models.AccountDevice, error) {
	model := s.NewAccountDevice()
	model.DeviceID = deviceID
	return model.UpsertLogin(time.Now().UTC())
}

func (s *AccountDeviceService) UpdateLocation(deviceID string, locationID int64) (*models.AccountDevice, error) {
	model := s.NewAccountDevice()
	model.DeviceID = deviceID
	if locationID > 0 {
		model.LocationID = &locationID
	}
	return model.UpdateLocation(model.LocationID, s.Auth.UserName)
}

func (s *AccountDeviceService) FindByDevice(deviceID string) (*models.AccountDevice, error) {
	model := s.NewAccountDevice()
	model.DeviceID = deviceID
	return model.FindByDevice()
}
