package mmodels

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend/shared/metadata/impl"
	"backend/shared/metadata/interfaces"
	"github.com/jmoiron/sqlx"
)

type SubscriptionPlanLimits struct {
	Users          int `json:"users"`
	Locations      int `json:"locations"`
	Products       int `json:"products"`
	Categories     int `json:"categories"`
	Brands         int `json:"brands"`
	Customers      int `json:"customers"`
	Suppliers      int `json:"suppliers"`
	PaymentMethods int `json:"paymentMethods"`
	PriceLists     int `json:"priceLists"`
	Promotions     int `json:"promotions"`
	StorageGB      int `json:"storageGb"`
}

type SubscriptionPlanMModel struct {
	impl.MModel

	ID          int64                  `json:"id" db:"id"`
	Code        string                 `json:"code" db:"code"`
	Name        string                 `json:"name" db:"name"`
	Price       float64                `json:"price" db:"price"`
	Currency    string                 `json:"currency" db:"currency"`
	Description string                 `json:"description" db:"description"`
	LimitsJSON  string                 `json:"-" db:"limits"`
	Limits      SubscriptionPlanLimits `json:"limits" db:"-"`
	Active      bool                   `json:"active" db:"active"`
	SortOrder   int                    `json:"sortOrder" db:"sort_order"`
	CreatedAt   time.Time              `json:"-" db:"created_at"`
	CreatedBy   string                 `json:"-" db:"created_by"`
	UpdatedAt   sql.NullTime           `json:"-" db:"updated_at"`
	UpdatedBy   sql.NullString         `json:"-" db:"updated_by"`
}

func NewSubscriptionPlanMModel(ctx context.Context, db *sqlx.DB) *SubscriptionPlanMModel {
	return &SubscriptionPlanMModel{
		Active: true,
		MModel: impl.MModel{
			Context: ctx, Name: "SubscriptionPlan", TableName: SubscriptionPlanTable, DB: db,
			Fields: []interfaces.IMField{
				impl.MField{Type: "int64", DbType: "BIGINT", JsonTag: "id", DbTag: "id", IsPrimaryKey: true, IsIdentity: true, Required: true, IsNotUpdateField: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "code", DbTag: "code", MaxLength: 50, Required: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "name", DbTag: "name", MaxLength: 100, Required: true},
				impl.MField{Type: "float64", DbType: "NUMERIC(18,2)", JsonTag: "price", DbTag: "price", Default: "0", Required: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "currency", DbTag: "currency", MaxLength: 10, Default: "'IDR'", Required: true},
				impl.MField{Type: "string", DbType: "TEXT", JsonTag: "description", DbTag: "description", Default: "''", Required: true},
				impl.MField{Type: "string", DbType: "JSONB", JsonTag: "-", DbTag: "limits", Default: "'{}'::JSONB", Required: true},
				impl.MField{Type: "bool", DbType: "BOOLEAN", JsonTag: "active", DbTag: "active", Default: "true", Required: true},
				impl.MField{Type: "int", DbType: "INTEGER", JsonTag: "sortOrder", DbTag: "sort_order", Default: "0", Required: true},
				impl.MField{Type: "time.Time", DbType: "TIMESTAMPTZ", JsonTag: "-", DbTag: "created_at", Default: "NOW()", Required: true, IsNotUpdateField: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "-", DbTag: "created_by", MaxLength: 255, Default: "'system'", Required: true, IsNotUpdateField: true},
				impl.MField{Type: "sql.NullTime", DbType: "TIMESTAMPTZ", JsonTag: "-", DbTag: "updated_at", IsNullable: true, IsNotCreateField: true},
				impl.MField{Type: "sql.NullString", DbType: "VARCHAR", JsonTag: "-", DbTag: "updated_by", MaxLength: 255, IsNullable: true, IsNotCreateField: true},
			},
		},
	}
}

func (m *SubscriptionPlanMModel) Migrate() error {
	if m == nil {
		return errors.New("subscription plan metadata model is nil")
	}
	if m.DB == nil {
		return errors.New("subscription plan metadata database is nil")
	}
	ctx := m.ContextOrBackground()
	tx, err := m.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin subscription plans migration: %w", err)
	}
	defer tx.Rollback()

	migrator := impl.NewMMigrator(&m.MModel, tx)
	migrator.SetContext(ctx)
	if err := migrator.Migrate(); err != nil {
		return err
	}
	statements := []string{
		`ALTER TABLE ` + SubscriptionPlanTable + ` ADD COLUMN IF NOT EXISTS code VARCHAR(50) NOT NULL DEFAULT ''`,
		`ALTER TABLE ` + SubscriptionPlanTable + ` ADD COLUMN IF NOT EXISTS name VARCHAR(100) NOT NULL DEFAULT ''`,
		`ALTER TABLE ` + SubscriptionPlanTable + ` ADD COLUMN IF NOT EXISTS price NUMERIC(18,2) NOT NULL DEFAULT 0`,
		`ALTER TABLE ` + SubscriptionPlanTable + ` ADD COLUMN IF NOT EXISTS currency VARCHAR(10) NOT NULL DEFAULT 'IDR'`,
		`ALTER TABLE ` + SubscriptionPlanTable + ` ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE ` + SubscriptionPlanTable + ` ADD COLUMN IF NOT EXISTS limits JSONB NOT NULL DEFAULT '{}'::JSONB`,
		`ALTER TABLE ` + SubscriptionPlanTable + ` ADD COLUMN IF NOT EXISTS active BOOLEAN NOT NULL DEFAULT true`,
		`ALTER TABLE ` + SubscriptionPlanTable + ` ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE ` + SubscriptionPlanTable + ` ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`ALTER TABLE ` + SubscriptionPlanTable + ` ADD COLUMN IF NOT EXISTS created_by VARCHAR(255) NOT NULL DEFAULT 'system'`,
		`ALTER TABLE ` + SubscriptionPlanTable + ` ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ`,
		`ALTER TABLE ` + SubscriptionPlanTable + ` ADD COLUMN IF NOT EXISTS updated_by VARCHAR(255)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_subscription_plans_code ON ` + SubscriptionPlanTable + ` (UPPER(code))`,
		`CREATE INDEX IF NOT EXISTS idx_subscription_plans_active_order ON ` + SubscriptionPlanTable + ` (active, sort_order, code)`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("run subscription plans migration statement %q: %w", statement, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit subscription plans migration: %w", err)
	}
	return nil
}

func (m *SubscriptionPlanMModel) Normalize() error {
	if m == nil {
		return errors.New("subscription plan metadata model is nil")
	}
	m.Code = strings.ToUpper(strings.TrimSpace(m.Code))
	m.Code = strings.ReplaceAll(strings.ReplaceAll(m.Code, "-", "_"), " ", "_")
	m.Name = strings.TrimSpace(m.Name)
	m.Currency = strings.ToUpper(strings.TrimSpace(m.Currency))
	if m.Currency == "" {
		m.Currency = "IDR"
	}
	m.Description = strings.TrimSpace(m.Description)
	if m.Price < 0 {
		return errors.New("harga subscription plan tidak valid")
	}
	if m.Code == "" || m.Name == "" {
		return errors.New("kode dan nama subscription plan wajib diisi")
	}
	if m.LimitsJSON == "" {
		raw, err := json.Marshal(m.Limits)
		if err != nil {
			return err
		}
		m.LimitsJSON = string(raw)
	}
	if err := json.Unmarshal([]byte(m.LimitsJSON), &m.Limits); err != nil {
		return fmt.Errorf("limits subscription plan tidak valid: %w", err)
	}
	return nil
}

func (m *SubscriptionPlanMModel) ValidateID() error {
	if m == nil {
		return errors.New("id subscription plan tidak valid")
	}
	return m.MModel.ValidateID_(m.ID, "subscription plan")
}

func (m *SubscriptionPlanMModel) Create() (int64, error) {
	if err := m.Normalize(); err != nil {
		return 0, err
	}
	if err := m.MModel.ValidateCreate_(m); err != nil {
		return 0, err
	}
	return m.MModel.Create_(m)
}

func (m *SubscriptionPlanMModel) Update() error {
	if err := m.ValidateID(); err != nil {
		return err
	}
	if err := m.Normalize(); err != nil {
		return err
	}
	if err := m.MModel.ValidateUpdate_(m); err != nil {
		return err
	}
	return m.MModel.Update_(m, "id")
}

func (m *SubscriptionPlanMModel) Delete() error {
	if err := m.ValidateID(); err != nil {
		return err
	}
	return m.MModel.Delete_(m, "id")
}

func (m *SubscriptionPlanMModel) FindByCode(code string) (*SubscriptionPlanMModel, error) {
	if m == nil {
		return nil, errors.New("subscription plan metadata model is nil")
	}
	code = strings.ToUpper(strings.TrimSpace(code))
	code = strings.ReplaceAll(strings.ReplaceAll(code, "-", "_"), " ", "_")
	if code == "" {
		code = "STARTER"
	}
	found := &SubscriptionPlanMModel{MModel: m.MModel, Code: code}
	if err := m.MModel.FindByFields(found, "code"); err != nil {
		return nil, err
	}
	if err := found.loadLimits(); err != nil {
		return nil, err
	}
	return found, nil
}

func (m *SubscriptionPlanMModel) FindByID() (*SubscriptionPlanMModel, error) {
	if err := m.ValidateID(); err != nil {
		return nil, err
	}
	found := &SubscriptionPlanMModel{MModel: m.MModel, ID: m.ID}
	if err := m.MModel.FindByID_(found, "id"); err != nil {
		return nil, err
	}
	if err := found.loadLimits(); err != nil {
		return nil, err
	}
	return found, nil
}

func (m *SubscriptionPlanMModel) FindList(activeOnly bool) ([]SubscriptionPlanMModel, error) {
	if m == nil {
		return nil, errors.New("subscription plan metadata model is nil")
	}
	plans := make([]SubscriptionPlanMModel, 0)
	filters := map[string]any(nil)
	if activeOnly {
		filters = map[string]any{"active": true}
	}
	if err := m.MModel.FindListOrdered(&plans, filters, "", nil, "sort_order", "ASC", 0); err != nil {
		return nil, err
	}
	for i := range plans {
		if err := plans[i].loadLimits(); err != nil {
			return nil, err
		}
	}
	return plans, nil
}

func (m *SubscriptionPlanMModel) loadLimits() error {
	if strings.TrimSpace(m.LimitsJSON) == "" {
		m.Limits = SubscriptionPlanLimits{}
		return nil
	}
	return json.Unmarshal([]byte(m.LimitsJSON), &m.Limits)
}
