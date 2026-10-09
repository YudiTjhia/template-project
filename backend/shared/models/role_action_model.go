package models

import (
	"errors"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
)

type RoleAction struct {
	BaseModel

	ID        int64  `json:"id" db:"id"`
	Code      string `json:"value" db:"code"`
	Label     string `json:"label" db:"label"`
	SortOrder int    `json:"sort_order" db:"sort_order"`
	Active    bool   `json:"active" db:"active"`
}

const roleActionSelectColumns = `ra.id, ra.code, ra.label, ra.sort_order, ra.active`

var defaultRoleActions = []RoleAction{
	{Code: "view_pos_dashboard", Label: "View POS Dashboard", SortOrder: 10, Active: true},
	{Code: "view_sales_order", Label: "View Sales Order", SortOrder: 100, Active: true},
	{Code: "create_sales_order", Label: "Create Sales Order", SortOrder: 110, Active: true},
	{Code: "update_sales_order", Label: "Update Sales Order", SortOrder: 120, Active: true},
	{Code: "delete_sales_order", Label: "Delete Sales Order", SortOrder: 130, Active: true},
	{Code: "checkout_sales_order", Label: "Checkout Sales Order", SortOrder: 140, Active: true},
	{Code: "cancel_sales_order", Label: "Cancel Sales Order", SortOrder: 150, Active: true},
	{Code: "view_sales_return", Label: "View Sales Return", SortOrder: 200, Active: true},
	{Code: "create_sales_return", Label: "Create Sales Return", SortOrder: 210, Active: true},
	{Code: "refund_sales_return", Label: "Refund Sales Return", SortOrder: 220, Active: true},
	{Code: "restock_sales_return", Label: "Restock Sales Return", SortOrder: 230, Active: true},
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
	{Code: "view_sales_return_reason", Label: "View Sales Return Reason", SortOrder: 1100, Active: true},
	{Code: "create_sales_return_reason", Label: "Create Sales Return Reason", SortOrder: 1110, Active: true},
	{Code: "update_sales_return_reason", Label: "Update Sales Return Reason", SortOrder: 1120, Active: true},
	{Code: "delete_sales_return_reason", Label: "Delete Sales Return Reason", SortOrder: 1130, Active: true},
	{Code: "view_stock", Label: "View Stock", SortOrder: 1200, Active: true},
	{Code: "view_stock_movement", Label: "View Stock Movement", SortOrder: 1210, Active: true},
	{Code: "create_stock_in", Label: "Create Stock In", SortOrder: 1220, Active: true},
	{Code: "create_stock_out", Label: "Create Stock Out", SortOrder: 1230, Active: true},
	{Code: "view_stock_adjustment", Label: "View Stock Adjustment", SortOrder: 1240, Active: true},
	{Code: "create_stock_adjustment", Label: "Create Stock Adjustment", SortOrder: 1250, Active: true},
	{Code: "approve_stock_adjustment", Label: "Approve Stock Adjustment", SortOrder: 1260, Active: true},
	{Code: "reject_stock_adjustment", Label: "Reject Stock Adjustment", SortOrder: 1270, Active: true},
	{Code: "view_stock_transaction", Label: "View Stock Transaction", SortOrder: 1280, Active: true},
	{Code: "view_inventory_cost", Label: "View Inventory Cost", SortOrder: 1290, Active: true},
	{Code: "view_cashier_shift", Label: "View Cashier Shift", SortOrder: 1300, Active: true},
	{Code: "open_cashier_shift", Label: "Open Cashier Shift", SortOrder: 1310, Active: true},
	{Code: "close_cashier_shift", Label: "Close Cashier Shift", SortOrder: 1320, Active: true},
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

func NewRoleAction(db *sqlx.DB, auth Auth) RoleAction {
	action := RoleAction{
		BaseModel: BaseModel{DB: db, Auth: auth},
	}
	action.Active = true
	return action
}

func (m RoleAction) Migrate() {
	sql := `CREATE TABLE IF NOT EXISTS role_actions (
	id SERIAL PRIMARY KEY,
	code VARCHAR(100) NOT NULL,
	label VARCHAR(255) NOT NULL,
	sort_order INT NOT NULL DEFAULT 0,
	active BOOLEAN NOT NULL DEFAULT true
);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `CREATE UNIQUE INDEX IF NOT EXISTS idx_role_actions_code ON role_actions (code);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `CREATE INDEX IF NOT EXISTS idx_role_actions_active ON role_actions (active, sort_order, label);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE role_actions ADD COLUMN IF NOT EXISTS code VARCHAR(100) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE role_actions ADD COLUMN IF NOT EXISTS label VARCHAR(255) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE role_actions ADD COLUMN IF NOT EXISTS sort_order INT NOT NULL DEFAULT 0;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	sql = `ALTER TABLE role_actions ADD COLUMN IF NOT EXISTS active BOOLEAN NOT NULL DEFAULT true;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	m.seedDefaults()
}

func (m RoleAction) TableName() string {
	return "role_actions"
}

func (m RoleAction) ValidateCreate() error {
	if strings.TrimSpace(m.Code) == "" {
		return errors.New("kode role action wajib diisi")
	}
	if strings.TrimSpace(m.Label) == "" {
		return errors.New("label role action wajib diisi")
	}
	return nil
}

func (m RoleAction) ValidateID() error {
	if m.ID <= 0 {
		return errors.New("id role action tidak valid")
	}
	return nil
}

func (m RoleAction) Create() (int64, error) {
	m.Code = strings.ToLower(strings.TrimSpace(m.Code))
	m.Label = strings.TrimSpace(m.Label)
	if err := m.ValidateCreate(); err != nil {
		return 0, err
	}

	sql := `INSERT INTO role_actions (code, label, sort_order, active)
VALUES ($1, $2, $3, $4)
ON CONFLICT (code) DO UPDATE SET
	label = EXCLUDED.label,
	sort_order = EXCLUDED.sort_order,
	active = EXCLUDED.active
RETURNING id`

	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{
		"code":       m.Code,
		"label":      m.Label,
		"sort_order": m.SortOrder,
		"active":     m.Active,
	})

	var id int64
	err := m.DB.QueryRow(sql, m.Code, m.Label, m.SortOrder, m.Active).Scan(&id)
	return id, err
}

func (m RoleAction) Update() error {
	m.Code = strings.ToLower(strings.TrimSpace(m.Code))
	m.Label = strings.TrimSpace(m.Label)
	if err := m.ValidateID(); err != nil {
		return err
	}
	if err := m.ValidateCreate(); err != nil {
		return err
	}

	sql := `UPDATE role_actions
SET code = $1,
	label = $2,
	sort_order = $3,
	active = $4
WHERE id = $5`
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{
		"code":       m.Code,
		"label":      m.Label,
		"sort_order": m.SortOrder,
		"active":     m.Active,
		"id":         m.ID,
	})

	_, err := m.DB.Exec(sql, m.Code, m.Label, m.SortOrder, m.Active, m.ID)
	return err
}

func (m RoleAction) Delete() error {
	if err := m.ValidateID(); err != nil {
		return err
	}

	sql := `DELETE FROM role_actions WHERE id = $1`
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"id": m.ID})

	_, err := m.DB.Exec(sql, m.ID)
	return err
}

func (m RoleAction) FindByID() (*RoleAction, error) {
	if err := m.ValidateID(); err != nil {
		return nil, err
	}

	sql := `SELECT ` + roleActionSelectColumns + `
FROM role_actions ra
WHERE ra.id = $1`
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"id": m.ID})

	var item RoleAction
	err := m.DB.Get(&item, sql, m.ID)
	return &item, err
}

func (m RoleAction) seedDefaults() {
	for _, action := range defaultRoleActions {
		action.DB = m.DB
		action.Auth = m.Auth
		_, err := action.Create()
		if err != nil {
			log.Println("[WARN] failed to seed role action:", err)
		}
	}
}

func (m RoleAction) FindAll(search string) ([]RoleAction, error) {
	sql := `SELECT ` + roleActionSelectColumns + `
FROM role_actions ra
WHERE ra.active = true`
	args := []interface{}{}
	if search = strings.TrimSpace(search); search != "" {
		args = append(args, search)
		placeholder := nextSQLPlaceholder(args)
		sql += ` AND (ra.code ILIKE '%' || ` + placeholder + ` || '%' OR ra.label ILIKE '%' || ` + placeholder + ` || '%')`
	}
	sql += ` ORDER BY ra.sort_order ASC, ra.label ASC`

	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"search": search})

	var rows []RoleAction
	err := m.DB.Select(&rows, sql, args...)
	return rows, err
}

func (m RoleAction) FindList(search string) ([]RoleAction, error) {
	sql := `SELECT ` + roleActionSelectColumns + `
FROM role_actions ra`
	args := []interface{}{}
	where := []string{}
	if search = strings.TrimSpace(search); search != "" {
		args = append(args, search)
		placeholder := nextSQLPlaceholder(args)
		where = append(where, `(ra.code ILIKE '%' || `+placeholder+` || '%' OR ra.label ILIKE '%' || `+placeholder+` || '%')`)
	}
	if len(where) > 0 {
		sql += " WHERE " + strings.Join(where, " AND ")
	}
	sql += ` ORDER BY ra.sort_order ASC, ra.label ASC`

	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"search": search})

	var rows []RoleAction
	err := m.DB.Select(&rows, sql, args...)
	return rows, err
}
