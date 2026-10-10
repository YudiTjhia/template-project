package mmodels

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"backend/shared/metadata/impl"
	"backend/shared/metadata/interfaces"

	"github.com/jmoiron/sqlx"
)

// RoleActionMModel owns persistence for the global role action catalog.
type RoleActionMModel struct {
	impl.MModel

	ID        int64  `json:"id" db:"id"`
	Code      string `json:"value" db:"code"`
	Label     string `json:"label" db:"label"`
	SortOrder int    `json:"sort_order" db:"sort_order"`
	Active    bool   `json:"active" db:"active"`
}

var defaultRoleActions = []RoleActionMModel{
	{Code: "view_pos_dashboard", Label: "View POS Dashboard", SortOrder: 10, Active: true},
	{Code: "view_customer", Label: "View Customer", SortOrder: 300, Active: true},
	{Code: "create_customer", Label: "Create Customer", SortOrder: 310, Active: true},
	{Code: "update_customer", Label: "Update Customer", SortOrder: 320, Active: true},
	{Code: "delete_customer", Label: "Delete Customer", SortOrder: 330, Active: true},
	{Code: "view_event", Label: "View Event", SortOrder: 340, Active: true},
	{Code: "create_event", Label: "Create Event", SortOrder: 350, Active: true},
	{Code: "update_event", Label: "Update Event", SortOrder: 360, Active: true},
	{Code: "delete_event", Label: "Delete Event", SortOrder: 370, Active: true},
	{Code: "view_room", Label: "View Room", SortOrder: 380, Active: true},
	{Code: "create_room", Label: "Create Room", SortOrder: 390, Active: true},
	{Code: "update_room", Label: "Update Room", SortOrder: 400, Active: true},
	{Code: "delete_room", Label: "Delete Room", SortOrder: 410, Active: true},
	{Code: "view_meja", Label: "View Meja", SortOrder: 420, Active: true},
	{Code: "create_meja", Label: "Create Meja", SortOrder: 430, Active: true},
	{Code: "update_meja", Label: "Update Meja", SortOrder: 440, Active: true},
	{Code: "delete_meja", Label: "Delete Meja", SortOrder: 450, Active: true},
	{Code: "view_item", Label: "View Item", SortOrder: 400, Active: true},
	{Code: "create_item", Label: "Create Item", SortOrder: 410, Active: true},
	{Code: "update_item", Label: "Update Item", SortOrder: 420, Active: true},
	{Code: "delete_item", Label: "Delete Item", SortOrder: 430, Active: true},
	{Code: "generate_item_image", Label: "Generate Item Image", SortOrder: 440, Active: true},
	{Code: "view_brand", Label: "View Brand", SortOrder: 500, Active: true},
	{Code: "create_brand", Label: "Create Brand", SortOrder: 510, Active: true},
	{Code: "update_brand", Label: "Update Brand", SortOrder: 520, Active: true},
	{Code: "delete_brand", Label: "Delete Brand", SortOrder: 530, Active: true},
	{Code: "view_item_category", Label: "View Item Category", SortOrder: 600, Active: true},
	{Code: "create_item_category", Label: "Create Item Category", SortOrder: 610, Active: true},
	{Code: "update_item_category", Label: "Update Item Category", SortOrder: 620, Active: true},
	{Code: "delete_item_category", Label: "Delete Item Category", SortOrder: 630, Active: true},
	{Code: "view_item_type", Label: "View Item Type", SortOrder: 700, Active: true},
	{Code: "create_item_type", Label: "Create Item Type", SortOrder: 710, Active: true},
	{Code: "update_item_type", Label: "Update Item Type", SortOrder: 720, Active: true},
	{Code: "delete_item_type", Label: "Delete Item Type", SortOrder: 730, Active: true},
	{Code: "view_uom", Label: "View UOM", SortOrder: 800, Active: true},
	{Code: "create_uom", Label: "Create UOM", SortOrder: 810, Active: true},
	{Code: "update_uom", Label: "Update UOM", SortOrder: 820, Active: true},
	{Code: "delete_uom", Label: "Delete UOM", SortOrder: 830, Active: true},
	{Code: "view_location", Label: "View Location", SortOrder: 900, Active: true},
	{Code: "create_location", Label: "Create Location", SortOrder: 910, Active: true},
	{Code: "update_location", Label: "Update Location", SortOrder: 920, Active: true},
	{Code: "delete_location", Label: "Delete Location", SortOrder: 930, Active: true},
	{Code: "select_location", Label: "Select Active Location", SortOrder: 940, Active: true},
	{Code: "view_payment_method", Label: "View Payment Method", SortOrder: 1000, Active: true},
	{Code: "create_payment_method", Label: "Create Payment Method", SortOrder: 1010, Active: true},
	{Code: "update_payment_method", Label: "Update Payment Method", SortOrder: 1020, Active: true},
	{Code: "delete_payment_method", Label: "Delete Payment Method", SortOrder: 1030, Active: true},
	{Code: "view_pos_reports", Label: "View POS Reports", SortOrder: 1400, Active: true},
	{Code: "export_pos_reports", Label: "Export POS Reports", SortOrder: 1410, Active: true},
	{Code: "view_approval_rule", Label: "View Approval Rule", SortOrder: 1500, Active: true},
	{Code: "create_approval_rule", Label: "Create Approval Rule", SortOrder: 1510, Active: true},
	{Code: "update_approval_rule", Label: "Update Approval Rule", SortOrder: 1520, Active: true},
	{Code: "delete_approval_rule", Label: "Delete Approval Rule", SortOrder: 1530, Active: true},
	{Code: "view_approval_request", Label: "View Approval Request", SortOrder: 1540, Active: true},
	{Code: "approve_approval_request", Label: "Approve Approval Request", SortOrder: 1550, Active: true},
	{Code: "reject_approval_request", Label: "Reject Approval Request", SortOrder: 1560, Active: true},
	{Code: "view_user", Label: "View User", SortOrder: 1600, Active: true},
	{Code: "create_user", Label: "Create User", SortOrder: 1610, Active: true},
	{Code: "update_user", Label: "Update User", SortOrder: 1620, Active: true},
	{Code: "delete_user", Label: "Delete User", SortOrder: 1630, Active: true},
	{Code: "view_role", Label: "View Role", SortOrder: 1640, Active: true},
	{Code: "create_role", Label: "Create Role", SortOrder: 1650, Active: true},
	{Code: "update_role", Label: "Update Role", SortOrder: 1660, Active: true},
	{Code: "delete_role", Label: "Delete Role", SortOrder: 1670, Active: true},
	{Code: "manage_role_action", Label: "Manage Role Action", SortOrder: 1680, Active: true},
	{Code: "view_ai_chat", Label: "View AI Chat", SortOrder: 1700, Active: true},
	{Code: "use_ai_chat", Label: "Use AI Chat", SortOrder: 1710, Active: true},
}

func NewRoleActionMModel(ctx context.Context, db *sqlx.DB) *RoleActionMModel {
	return &RoleActionMModel{
		Active: true,
		MModel: impl.MModel{
			Context: ctx, Name: "RoleAction", TableName: RoleActionTable, DB: db,
			Fields: []interfaces.IMField{
				impl.MField{Type: "int64", DbType: "INTEGER", JsonTag: "id", DbTag: "id", IsPrimaryKey: true, IsIdentity: true, Required: true, IsNotUpdateField: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "value", DbTag: "code", MaxLength: 100, Required: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "label", DbTag: "label", MaxLength: 255, Required: true},
				impl.MField{Type: "int", DbType: "INTEGER", JsonTag: "sort_order", DbTag: "sort_order", Default: "0", Required: true},
				impl.MField{Type: "bool", DbType: "BOOLEAN", JsonTag: "active", DbTag: "active", Default: "true", Required: true},
			},
		},
	}
}

func (m *RoleActionMModel) Migrate() error {
	if m == nil {
		return errors.New("role action metadata model is nil")
	}
	if m.DB == nil {
		return errors.New("role action metadata database is nil")
	}
	ctx := m.ContextOrBackground()
	tx, err := m.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin role actions migration: %w", err)
	}
	defer tx.Rollback()

	migrator := impl.NewMMigrator(&m.MModel, tx)
	migrator.SetContext(ctx)
	if err := migrator.Migrate(); err != nil {
		return err
	}
	statements := []string{
		`ALTER TABLE ` + RoleActionTable + ` ADD COLUMN IF NOT EXISTS code VARCHAR(100) NOT NULL DEFAULT ''`,
		`ALTER TABLE ` + RoleActionTable + ` ADD COLUMN IF NOT EXISTS label VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE ` + RoleActionTable + ` ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE ` + RoleActionTable + ` ADD COLUMN IF NOT EXISTS active BOOLEAN NOT NULL DEFAULT true`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_role_actions_code ON ` + RoleActionTable + ` (code)`,
		`CREATE INDEX IF NOT EXISTS idx_role_actions_active ON ` + RoleActionTable + ` (active, sort_order, label)`,
		`UPDATE ` + RoleActionTable + ` SET active = false WHERE code ILIKE '%sales%' OR code ILIKE '%stock%' OR code ILIKE '%purchase%' OR code ILIKE '%cash_drawer%' OR code ILIKE '%cashier_shift%'`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("run role actions migration statement %q: %w", statement, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit role actions migration: %w", err)
	}
	return m.seedDefaults()
}

func (m *RoleActionMModel) ValidateCreate() error {
	if m == nil {
		return errors.New("role action metadata model is nil")
	}
	m.normalize()
	if m.Code == "" {
		return errors.New("kode role action wajib diisi")
	}
	if m.Label == "" {
		return errors.New("label role action wajib diisi")
	}
	return m.MModel.ValidateCreate_(m)
}

func (m *RoleActionMModel) ValidateID() error {
	if m == nil {
		return errors.New("id role action tidak valid")
	}
	return m.MModel.ValidateID_(m.ID, "role action")
}

func (m *RoleActionMModel) Create() (int64, error) {
	if err := m.ValidateCreate(); err != nil {
		return 0, err
	}
	query := `INSERT INTO ` + RoleActionTable + ` (code, label, sort_order, active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (code) DO UPDATE SET label = EXCLUDED.label,
			sort_order = EXCLUDED.sort_order, active = EXCLUDED.active
		RETURNING id`
	var id int64
	if err := m.DB.QueryRowxContext(m.ContextOrBackground(), query, m.Code, m.Label, m.SortOrder, m.Active).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (m *RoleActionMModel) Update() error {
	if err := m.ValidateID(); err != nil {
		return err
	}
	if err := m.ValidateCreate(); err != nil {
		return err
	}
	return m.MModel.Update_(m, "id")
}

func (m *RoleActionMModel) Delete() error {
	if err := m.ValidateID(); err != nil {
		return err
	}
	return m.MModel.Delete_(m, "id")
}

func (m *RoleActionMModel) FindByID() (*RoleActionMModel, error) {
	if err := m.ValidateID(); err != nil {
		return nil, err
	}
	found := &RoleActionMModel{MModel: m.MModel}
	if err := m.MModel.FindByID_(found, "id"); err != nil {
		return nil, err
	}
	return found, nil
}

func (m *RoleActionMModel) FindAll(search string) ([]RoleActionMModel, error) {
	return m.find(search, true)
}

func (m *RoleActionMModel) FindList(search string) ([]RoleActionMModel, error) {
	return m.find(search, false)
}

func (m *RoleActionMModel) find(search string, activeOnly bool) ([]RoleActionMModel, error) {
	if m == nil {
		return nil, errors.New("role action metadata model is nil")
	}
	if m.DB == nil {
		return nil, errors.New("role action metadata database is nil")
	}
	query := `SELECT id, code, label, sort_order, active FROM ` + RoleActionTable + ``
	conditions := make([]string, 0, 2)
	args := make([]any, 0, 2)
	if activeOnly {
		conditions = append(conditions, `active = true`)
	}
	if search = strings.TrimSpace(search); search != "" {
		args = append(args, search)
		conditions = append(conditions, `(code ILIKE '%' || $1 || '%' OR label ILIKE '%' || $1 || '%')`)
	}
	if len(conditions) > 0 {
		query += ` WHERE ` + strings.Join(conditions, ` AND `)
	}
	query += ` ORDER BY sort_order ASC, label ASC`
	var rows []RoleActionMModel
	if err := m.DB.SelectContext(m.ContextOrBackground(), &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *RoleActionMModel) seedDefaults() error {
	for _, action := range defaultRoleActions {
		item := *NewRoleActionMModel(m.ContextOrBackground(), m.DB)
		item.Code, item.Label = action.Code, action.Label
		item.SortOrder, item.Active = action.SortOrder, action.Active
		if _, err := item.Create(); err != nil {
			return fmt.Errorf("seed role action %q: %w", item.Code, err)
		}
	}
	return nil
}

func (m *RoleActionMModel) normalize() {
	m.Code = strings.ToLower(strings.TrimSpace(m.Code))
	m.Label = strings.TrimSpace(m.Label)
}
