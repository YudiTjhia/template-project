package services

import (
	"context"
	"database/sql"

	"backend/shared/metadata/impl"
	"backend/shared/metadata/interfaces"
	"backend/shared/mmodels"
	"backend/shared/models"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type RoleService struct {
	DB   *sqlx.DB
	Auth models.Auth
	CRUD interfaces.IService
}

func NewRoleService(db *sqlx.DB, auth models.Auth) *RoleService {
	return &RoleService{DB: db, Auth: auth, CRUD: impl.NewMService()}
}

func (s *RoleService) roleMModel() *mmodels.RoleMModel {
	model := mmodels.NewRoleMModel(context.Background(), s.DB)
	model.AccountID = s.Auth.AccountID
	model.AuthAccountID = s.Auth.AccountID
	model.AuthEmail = s.Auth.Email
	return model
}

func (s *RoleService) roleFromMModel(value mmodels.RoleMModel) models.Role {
	role := models.Role{
		BaseModel: models.BaseModel{
			DB: s.DB, Auth: s.Auth, CreatedAt: value.CreatedAt, CreatedBy: value.CreatedBy,
		},
		ID: value.ID, AccountID: value.AccountID, RoleName: value.RoleName,
		Actions: models.StringArray(value.Actions),
	}
	if value.UpdatedAt.Valid {
		updatedAt := value.UpdatedAt.Time
		role.UpdatedAt = &updatedAt
	}
	if value.UpdatedBy.Valid {
		updatedBy := value.UpdatedBy.String
		role.UpdatedBy = &updatedBy
	}
	return role
}

func (s *RoleService) NewRole() models.Role {
	return models.NewRole(s.DB, s.Auth)
}

func (s *RoleService) NewRoleAction() models.RoleAction {
	return models.NewRoleAction(s.DB, s.Auth)
}

func (s *RoleService) Migrate() {
	s.NewRole().Migrate()
	s.NewRoleAction().Migrate()
}

func (s *RoleService) Create(role models.Role) (int64, error) {
	item := s.roleMModel()
	item.AccountID = role.AccountID
	item.RoleName = role.RoleName
	item.Actions = pq.StringArray(role.Actions)
	item.CreatedBy = role.CreatedBy
	return item.Create()
}

func (s *RoleService) Update(role models.Role) error {
	item := s.roleMModel()
	item.ID = role.ID
	item.AccountID = role.AccountID
	item.RoleName = role.RoleName
	item.Actions = pq.StringArray(role.Actions)
	if role.UpdatedBy != nil {
		item.UpdatedBy = sql.NullString{String: *role.UpdatedBy, Valid: true}
	}
	return item.Update()
}

func (s *RoleService) Delete(id int64) error {
	item := s.roleMModel()
	item.ID = id
	return s.CRUD.Delete_(item, item, "id", "account_id")
}

func (s *RoleService) FindByID(id int64) (*models.Role, error) {
	item := s.roleMModel()
	item.ID = id
	var result mmodels.RoleMModel
	if err := s.CRUD.FindByID_(item, &result, "id", "account_id"); err != nil {
		return nil, err
	}
	role := s.roleFromMModel(result)
	return &role, nil
}

func (s *RoleService) FindByName(name string) (*models.Role, error) {
	item, err := s.roleMModel().FindByName(name)
	if err != nil {
		return nil, err
	}
	role := s.roleFromMModel(*item)
	return &role, nil
}

func (s *RoleService) FindList(search string) ([]models.Role, error) {
	items, err := s.roleMModel().FindList(search)
	if err != nil {
		return nil, err
	}
	roles := make([]models.Role, 0, len(items))
	for _, item := range items {
		roles = append(roles, s.roleFromMModel(item))
	}
	return roles, nil
}

func (s *RoleService) FindAll() ([]models.Role, error) {
	items, err := s.roleMModel().FindAll()
	if err != nil {
		return nil, err
	}
	roles := make([]models.Role, 0, len(items))
	for _, item := range items {
		roles = append(roles, s.roleFromMModel(item))
	}
	return roles, nil
}

func (s *RoleService) FindRecentList(limit int, search string) ([]models.Role, error) {
	items, err := s.roleMModel().FindRecentList(limit, search)
	if err != nil {
		return nil, err
	}
	roles := make([]models.Role, 0, len(items))
	for _, item := range items {
		roles = append(roles, s.roleFromMModel(item))
	}
	return roles, nil
}

func (s *RoleService) FindActionList(search string) ([]models.RoleAction, error) {
	return s.NewRoleAction().FindAll(search)
}

func (s *RoleService) FindActionAll(search string) ([]models.RoleAction, error) {
	return s.NewRoleAction().FindList(search)
}

func (s *RoleService) FindActionByID(id int64) (*models.RoleAction, error) {
	item := s.NewRoleAction()
	item.ID = id
	return item.FindByID()
}

func (s *RoleService) CreateAction(action models.RoleAction) (int64, error) {
	item := s.NewRoleAction()
	item.Code = action.Code
	item.Label = action.Label
	item.SortOrder = action.SortOrder
	item.Active = action.Active
	return item.Create()
}

func (s *RoleService) UpdateAction(action models.RoleAction) error {
	item := s.NewRoleAction()
	item.ID = action.ID
	item.Code = action.Code
	item.Label = action.Label
	item.SortOrder = action.SortOrder
	item.Active = action.Active
	return item.Update()
}

func (s *RoleService) DeleteAction(id int64) error {
	item := s.NewRoleAction()
	item.ID = id
	return item.Delete()
}
