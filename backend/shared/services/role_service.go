package services

import (
	"backend/shared/models"

	"github.com/jmoiron/sqlx"
)

type RoleService struct {
	DB   *sqlx.DB
	Auth models.Auth
}

func NewRoleService(db *sqlx.DB, auth models.Auth) *RoleService {
	return &RoleService{DB: db, Auth: auth}
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
	item := s.NewRole()
	item.AccountID = role.AccountID
	item.RoleName = role.RoleName
	item.Actions = role.Actions
	item.CreatedBy = role.CreatedBy
	return item.Create()
}

func (s *RoleService) Update(role models.Role) error {
	item := s.NewRole()
	item.ID = role.ID
	item.AccountID = role.AccountID
	item.RoleName = role.RoleName
	item.Actions = role.Actions
	item.UpdatedBy = role.UpdatedBy
	return item.Update()
}

func (s *RoleService) Delete(id int64) error {
	item := s.NewRole()
	item.ID = id
	return item.Delete()
}

func (s *RoleService) FindByID(id int64) (*models.Role, error) {
	item := s.NewRole()
	item.ID = id
	return item.FindByID()
}

func (s *RoleService) FindByName(name string) (*models.Role, error) {
	item := s.NewRole()
	return item.FindByName(name)
}

func (s *RoleService) FindList(search string) ([]models.Role, error) {
	return s.NewRole().FindList(search)
}

func (s *RoleService) FindAll() ([]models.Role, error) {
	return s.NewRole().FindAll()
}

func (s *RoleService) FindRecentList(limit int, search string) ([]models.Role, error) {
	return s.NewRole().FindRecentList(limit, search)
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
