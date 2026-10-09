package models

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	SubscriptionHistoryTypeCreate            = "CREATE"
	SubscriptionHistoryTypeIncreaseUsers     = "INCREASE_USERS"
	SubscriptionHistoryTypeScheduleReduction = "SCHEDULE_REDUCTION"
	SubscriptionHistoryTypeApplyReduction    = "APPLY_REDUCTION"
	SubscriptionHistoryTypeUpdate            = "UPDATE"
	SubscriptionHistoryTypePayment           = "PAYMENT"
)

type SubscriptionHistory struct {
	BaseModel

	ID                    int64     `json:"id" db:"id"`
	AccountID             int64     `json:"account_id" db:"account_id"`
	SubscriptionID        int64     `json:"subscriptionId" db:"subscription_id"`
	AppType               string    `json:"appType" db:"app_type"`
	HistoryType           string    `json:"historyType" db:"history_type"`
	PreviousNumberOfUsers int       `json:"previousNumberOfUsers" db:"previous_number_of_users"`
	NewNumberOfUsers      int       `json:"newNumberOfUsers" db:"new_number_of_users"`
	DeltaUsers            int       `json:"deltaUsers" db:"delta_users"`
	PricePerUser          float64   `json:"pricePerUser" db:"price_per_user"`
	AmountBefore          float64   `json:"amountBefore" db:"amount_before"`
	AmountChange          float64   `json:"amountChange" db:"amount_change"`
	AmountAfter           float64   `json:"amountAfter" db:"amount_after"`
	CurrentStartDate      time.Time `json:"currentStartDate" db:"current_start_date"`
	CurrentEndDate        time.Time `json:"currentEndDate" db:"current_end_date"`
	EffectiveDate         time.Time `json:"effectiveDate" db:"effective_date"`
	Notes                 string    `json:"notes" db:"notes"`
}

const subscriptionHistorySelectColumns = `h.id, h.account_id, h.subscription_id, h.app_type, h.history_type, h.previous_number_of_users,
h.new_number_of_users, h.delta_users, h.price_per_user, h.amount_before, h.amount_change, h.amount_after,
h.current_start_date, h.current_end_date, h.effective_date, h.notes, h.created_at, h.created_by, h.updated_at, h.updated_by`

func NewSubscriptionHistory(db *sqlx.DB, auth Auth) SubscriptionHistory {
	history := SubscriptionHistory{
		BaseModel: BaseModel{DB: db, Auth: auth},
	}
	history.AccountID = auth.AccountID
	history.CreatedBy = auth.UserName
	return history
}

func (m SubscriptionHistory) Migrate() {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS subscription_histories (
	id SERIAL PRIMARY KEY,
	account_id BIGINT NOT NULL DEFAULT 0,
	subscription_id BIGINT NOT NULL DEFAULT 0,
	app_type VARCHAR(100) NOT NULL DEFAULT 'POS Kasir',
	history_type VARCHAR(50) NOT NULL DEFAULT '',
	previous_number_of_users INTEGER NOT NULL DEFAULT 0,
	new_number_of_users INTEGER NOT NULL DEFAULT 0,
	delta_users INTEGER NOT NULL DEFAULT 0,
	price_per_user NUMERIC(18,2) NOT NULL DEFAULT 0,
	amount_before NUMERIC(18,2) NOT NULL DEFAULT 0,
	amount_change NUMERIC(18,2) NOT NULL DEFAULT 0,
	amount_after NUMERIC(18,2) NOT NULL DEFAULT 0,
	current_start_date DATE NOT NULL DEFAULT CURRENT_DATE,
	current_end_date DATE NOT NULL DEFAULT CURRENT_DATE,
	effective_date DATE NOT NULL DEFAULT CURRENT_DATE,
	notes TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_by VARCHAR(255) NOT NULL DEFAULT '',
	updated_at TIMESTAMPTZ,
	updated_by VARCHAR(255)
);`,
		`CREATE INDEX IF NOT EXISTS idx_subscription_histories_account_id ON subscription_histories (account_id, created_at DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_subscription_histories_subscription_id ON subscription_histories (subscription_id, created_at DESC, id DESC);`,
		`ALTER TABLE subscription_histories ADD COLUMN IF NOT EXISTS account_id BIGINT NOT NULL DEFAULT 0;`,
		`ALTER TABLE subscription_histories ADD COLUMN IF NOT EXISTS subscription_id BIGINT NOT NULL DEFAULT 0;`,
		`ALTER TABLE subscription_histories ADD COLUMN IF NOT EXISTS app_type VARCHAR(100) NOT NULL DEFAULT 'POS Kasir';`,
		`ALTER TABLE subscription_histories ADD COLUMN IF NOT EXISTS history_type VARCHAR(50) NOT NULL DEFAULT '';`,
		`ALTER TABLE subscription_histories ADD COLUMN IF NOT EXISTS previous_number_of_users INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE subscription_histories ADD COLUMN IF NOT EXISTS new_number_of_users INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE subscription_histories ADD COLUMN IF NOT EXISTS delta_users INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE subscription_histories ADD COLUMN IF NOT EXISTS price_per_user NUMERIC(18,2) NOT NULL DEFAULT 0;`,
		`ALTER TABLE subscription_histories ADD COLUMN IF NOT EXISTS amount_before NUMERIC(18,2) NOT NULL DEFAULT 0;`,
		`ALTER TABLE subscription_histories ADD COLUMN IF NOT EXISTS amount_change NUMERIC(18,2) NOT NULL DEFAULT 0;`,
		`ALTER TABLE subscription_histories ADD COLUMN IF NOT EXISTS amount_after NUMERIC(18,2) NOT NULL DEFAULT 0;`,
		`ALTER TABLE subscription_histories ADD COLUMN IF NOT EXISTS current_start_date DATE NOT NULL DEFAULT CURRENT_DATE;`,
		`ALTER TABLE subscription_histories ADD COLUMN IF NOT EXISTS current_end_date DATE NOT NULL DEFAULT CURRENT_DATE;`,
		`ALTER TABLE subscription_histories ADD COLUMN IF NOT EXISTS effective_date DATE NOT NULL DEFAULT CURRENT_DATE;`,
		`ALTER TABLE subscription_histories ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT '';`,
	}

	for _, statement := range statements {
		log.Println("[SQL]=", statement)
		m.DB.MustExec(statement)
	}
}

func (m SubscriptionHistory) TableName() string {
	return "subscription_histories"
}

func (m SubscriptionHistory) ValidateCreate() error {
	if _, err := m.resolveAccountID(); err != nil {
		return err
	}
	m.AppType = NormalizeSubscriptionAppType(m.AppType)
	if !IsAllowedSubscriptionAppType(m.AppType) {
		return errors.New("app type subscription history tidak valid")
	}
	if m.SubscriptionID <= 0 {
		return errors.New("subscription_id history wajib diisi")
	}
	if strings.TrimSpace(m.HistoryType) == "" {
		return errors.New("history type wajib diisi")
	}
	return nil
}

func (m SubscriptionHistory) Create() (int64, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return 0, err
	}

	m.HistoryType = strings.ToUpper(strings.TrimSpace(m.HistoryType))
	m.AppType = NormalizeSubscriptionAppType(m.AppType)
	m.Notes = strings.TrimSpace(m.Notes)
	if m.CreatedBy == "" {
		m.CreatedBy = m.Auth.UserName
	}
	if m.CreatedBy == "" {
		m.CreatedBy = "system"
	}

	if err := m.ValidateCreate(); err != nil {
		return 0, err
	}

	sql := `INSERT INTO subscription_histories
	(account_id, subscription_id, app_type, history_type, previous_number_of_users, new_number_of_users, delta_users, price_per_user,
	amount_before, amount_change, amount_after, current_start_date, current_end_date, effective_date, notes, created_by)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
	RETURNING id`
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{
		"account_id":               accountID,
		"subscription_id":          m.SubscriptionID,
		"app_type":                 m.AppType,
		"history_type":             m.HistoryType,
		"previous_number_of_users": m.PreviousNumberOfUsers,
		"new_number_of_users":      m.NewNumberOfUsers,
		"delta_users":              m.DeltaUsers,
		"price_per_user":           m.PricePerUser,
		"amount_before":            m.AmountBefore,
		"amount_change":            m.AmountChange,
		"amount_after":             m.AmountAfter,
		"current_start_date":       m.CurrentStartDate,
		"current_end_date":         m.CurrentEndDate,
		"effective_date":           m.EffectiveDate,
		"notes":                    m.Notes,
		"created_by":               m.CreatedBy,
	})

	var id int64
	err = m.DB.QueryRow(sql,
		accountID,
		m.SubscriptionID,
		m.AppType,
		m.HistoryType,
		m.PreviousNumberOfUsers,
		m.NewNumberOfUsers,
		m.DeltaUsers,
		m.PricePerUser,
		m.AmountBefore,
		m.AmountChange,
		m.AmountAfter,
		m.CurrentStartDate,
		m.CurrentEndDate,
		m.EffectiveDate,
		m.Notes,
		m.CreatedBy,
	).Scan(&id)
	return id, err
}

func (m SubscriptionHistory) FindList(subscriptionID int64) ([]SubscriptionHistory, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}
	if subscriptionID <= 0 {
		return []SubscriptionHistory{}, nil
	}

	sql := `SELECT ` + subscriptionHistorySelectColumns + `
	FROM subscription_histories h
	WHERE h.account_id = $1 AND h.subscription_id = $2
	ORDER BY h.created_at DESC, h.id DESC`
	log.Println("[SQL]=", sql)
	LogSQLArgs(map[string]interface{}{"account_id": accountID, "subscription_id": subscriptionID})

	var items []SubscriptionHistory
	err = m.DB.Select(&items, sql, accountID, subscriptionID)
	return items, err
}

func (m SubscriptionHistory) resolveAccountID() (int64, error) {
	if m.Auth.AccountID > 0 {
		return m.Auth.AccountID, nil
	}
	if m.AccountID > 0 {
		return m.AccountID, nil
	}
	return 0, errors.New("id akun wajib diisi")
}
