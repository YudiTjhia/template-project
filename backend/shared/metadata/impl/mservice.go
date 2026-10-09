package impl

import (
	"errors"

	"backend/shared/metadata/interfaces"
)

// MService delegates common CRUD operations to an MModel-backed model.
type MService struct{}

func NewMService() *MService { return &MService{} }

func (s *MService) Create_(model interfaces.ICRUDModel, value any) (any, error) {
	if model == nil {
		return nil, errors.New("service model is nil")
	}
	return model.CreateValue_(value)
}

func (s *MService) Update_(model interfaces.ICRUDModel, value any, keyFields ...string) error {
	if model == nil {
		return errors.New("service model is nil")
	}
	return model.Update_(value, keyFields...)
}

func (s *MService) Delete_(model interfaces.ICRUDModel, value any, keyFields ...string) error {
	if model == nil {
		return errors.New("service model is nil")
	}
	return model.Delete_(value, keyFields...)
}

func (s *MService) FindByID_(model interfaces.ICRUDModel, destination any, keyFields ...string) error {
	if model == nil {
		return errors.New("service model is nil")
	}
	return model.FindByID_(destination, keyFields...)
}

func (s *MService) FindByFields_(model interfaces.ICRUDModel, destination any, filterFields ...string) error {
	if model == nil {
		return errors.New("service model is nil")
	}
	return model.FindByFields(destination, filterFields...)
}

func (s *MService) FindList_(model interfaces.ICRUDModel, destination any, filters map[string]any, search string, searchFields []string, orderBy, orderDirection string, limit int) error {
	if model == nil {
		return errors.New("service model is nil")
	}
	return model.FindListOrdered(destination, filters, search, searchFields, orderBy, orderDirection, limit)
}

var _ interfaces.IService = (*MService)(nil)
