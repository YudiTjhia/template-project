package models

import (
	"context"

	"backend/shared/mmodels"
	"github.com/jmoiron/sqlx"
)

// RoleAction remains as the API DTO while RoleActionMModel owns persistence.
type RoleAction struct {
	BaseModel

	ID        int64  `json:"id" db:"id"`
	Code      string `json:"value" db:"code"`
	Label     string `json:"label" db:"label"`
	SortOrder int    `json:"sort_order" db:"sort_order"`
	Active    bool   `json:"active" db:"active"`
}

func NewRoleAction(db *sqlx.DB, auth Auth) RoleAction {
	return RoleAction{BaseModel: BaseModel{DB: db, Auth: auth}, Active: true}
}

func (m RoleAction) mmodel() *mmodels.RoleActionMModel {
	ctx := m.Context
	if ctx == nil {
		ctx = context.Background()
	}
	item := mmodels.NewRoleActionMModel(ctx, m.DB)
	item.ID, item.Code, item.Label = m.ID, m.Code, m.Label
	item.SortOrder, item.Active = m.SortOrder, m.Active
	return item
}

func roleActionFromMModel(item *mmodels.RoleActionMModel, base BaseModel) *RoleAction {
	if item == nil {
		return nil
	}
	return &RoleAction{
		BaseModel: base,
		ID:        item.ID, Code: item.Code, Label: item.Label,
		SortOrder: item.SortOrder, Active: item.Active,
	}
}

func (m RoleAction) Migrate() {
	if err := m.MigrateError(); err != nil {
		panic(err)
	}
}

func (m RoleAction) MigrateError() error { return m.mmodel().Migrate() }

func (m RoleAction) TableName() string { return "role_actions" }

func (m RoleAction) ValidateCreate() error { return m.mmodel().ValidateCreate() }

func (m RoleAction) ValidateID() error { return m.mmodel().ValidateID() }

func (m RoleAction) Create() (int64, error) { return m.mmodel().Create() }

func (m RoleAction) Update() error { return m.mmodel().Update() }

func (m RoleAction) Delete() error { return m.mmodel().Delete() }

func (m RoleAction) FindByID() (*RoleAction, error) {
	found, err := m.mmodel().FindByID()
	if err != nil {
		return nil, err
	}
	return roleActionFromMModel(found, m.BaseModel), nil
}

func (m RoleAction) FindAll(search string) ([]RoleAction, error) {
	rows, err := m.mmodel().FindAll(search)
	if err != nil {
		return nil, err
	}
	return roleActionsFromMModels(rows, m.BaseModel), nil
}

func (m RoleAction) FindList(search string) ([]RoleAction, error) {
	rows, err := m.mmodel().FindList(search)
	if err != nil {
		return nil, err
	}
	return roleActionsFromMModels(rows, m.BaseModel), nil
}

func roleActionsFromMModels(rows []mmodels.RoleActionMModel, base BaseModel) []RoleAction {
	items := make([]RoleAction, 0, len(rows))
	for i := range rows {
		items = append(items, *roleActionFromMModel(&rows[i], base))
	}
	return items
}
