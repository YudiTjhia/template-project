package models

import (
	"backend/shared/mmodels"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type SubscriptionPayment struct {
	BaseModel

	ID                 int64         `json:"id" db:"id"`
	AccountID          int64         `json:"account_id" db:"account_id"`
	SubscriptionID     int64         `json:"subscriptionId" db:"subscription_id"`
	AppType            string        `json:"appType" db:"app_type"`
	UserIDs            pq.Int64Array `json:"userIds" db:"user_ids"`
	NumberOfUsers      int           `json:"numberOfUsers" db:"number_of_users"`
	Months             int           `json:"months" db:"months"`
	PricePerUser       float64       `json:"pricePerUser" db:"price_per_user"`
	TotalAmount        float64       `json:"totalAmount" db:"total_amount"`
	Currency           string        `json:"currency" db:"currency"`
	PaymentToken       string        `json:"paymentToken" db:"payment_token"`
	ProofFileName      string        `json:"proofFileName" db:"proof_file_name"`
	ProofFileSize      int64         `json:"proofFileSize" db:"proof_file_size"`
	ProofContentType   string        `json:"proofContentType" db:"proof_content_type"`
	CurrentStartDate   time.Time     `json:"currentStartDate" db:"current_start_date"`
	CurrentEndDate     time.Time     `json:"currentEndDate" db:"current_end_date"`
	PaidAt             time.Time     `json:"paidAt" db:"paid_at"`
	SelectedUserLabels StringArray   `json:"selectedUserLabels,omitempty" db:"selected_user_labels"`
}

const subscriptionPaymentSelectColumns = `p.id, p.account_id, p.subscription_id, p.app_type, p.user_ids, p.number_of_users,
p.months, p.price_per_user, p.total_amount, p.currency, p.payment_token, p.proof_file_name, p.proof_file_size,
p.proof_content_type, p.current_start_date, p.current_end_date, p.paid_at,
'{}'::TEXT[] selected_user_labels,
p.created_at, p.created_by, p.updated_at, p.updated_by`

func NewSubscriptionPayment(db *sqlx.DB, auth Auth) SubscriptionPayment {
	payment := SubscriptionPayment{
		BaseModel: BaseModel{DB: db, Auth: auth},
		AccountID: auth.AccountID,
	}
	payment.CreatedBy = auth.UserName
	return payment
}

func (m SubscriptionPayment) Migrate() {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS ` + mmodels.SubscriptionPaymentTable + ` (
	id SERIAL PRIMARY KEY,
	account_id BIGINT NOT NULL DEFAULT 0,
	subscription_id BIGINT NOT NULL DEFAULT 0,
	app_type VARCHAR(100) NOT NULL DEFAULT 'POS Kasir',
	user_ids BIGINT[] NOT NULL DEFAULT '{}'::BIGINT[],
	number_of_users INTEGER NOT NULL DEFAULT 0,
	months INTEGER NOT NULL DEFAULT 1,
	price_per_user NUMERIC(18,2) NOT NULL DEFAULT 0,
	total_amount NUMERIC(18,2) NOT NULL DEFAULT 0,
	currency VARCHAR(10) NOT NULL DEFAULT 'IDR',
	payment_token TEXT NOT NULL DEFAULT '',
	proof_file_name VARCHAR(255) NOT NULL DEFAULT '',
	proof_file_size BIGINT NOT NULL DEFAULT 0,
	proof_content_type VARCHAR(255) NOT NULL DEFAULT '',
	current_start_date DATE NOT NULL DEFAULT CURRENT_DATE,
	current_end_date DATE NOT NULL DEFAULT CURRENT_DATE,
	paid_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_by VARCHAR(255) NOT NULL DEFAULT '',
	updated_at TIMESTAMPTZ,
	updated_by VARCHAR(255)
);`,
		`ALTER TABLE ` + mmodels.SubscriptionPaymentTable + ` ADD COLUMN IF NOT EXISTS account_id BIGINT NOT NULL DEFAULT 0;`,
		`ALTER TABLE ` + mmodels.SubscriptionPaymentTable + ` ADD COLUMN IF NOT EXISTS subscription_id BIGINT NOT NULL DEFAULT 0;`,
		`ALTER TABLE ` + mmodels.SubscriptionPaymentTable + ` ADD COLUMN IF NOT EXISTS app_type VARCHAR(100) NOT NULL DEFAULT 'POS Kasir';`,
		`ALTER TABLE ` + mmodels.SubscriptionPaymentTable + ` ADD COLUMN IF NOT EXISTS user_ids BIGINT[] NOT NULL DEFAULT '{}'::BIGINT[];`,
		`ALTER TABLE ` + mmodels.SubscriptionPaymentTable + ` ADD COLUMN IF NOT EXISTS number_of_users INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE ` + mmodels.SubscriptionPaymentTable + ` ADD COLUMN IF NOT EXISTS months INTEGER NOT NULL DEFAULT 1;`,
		`ALTER TABLE ` + mmodels.SubscriptionPaymentTable + ` ADD COLUMN IF NOT EXISTS price_per_user NUMERIC(18,2) NOT NULL DEFAULT 0;`,
		`ALTER TABLE ` + mmodels.SubscriptionPaymentTable + ` ADD COLUMN IF NOT EXISTS total_amount NUMERIC(18,2) NOT NULL DEFAULT 0;`,
		`ALTER TABLE ` + mmodels.SubscriptionPaymentTable + ` ADD COLUMN IF NOT EXISTS currency VARCHAR(10) NOT NULL DEFAULT 'IDR';`,
		`ALTER TABLE ` + mmodels.SubscriptionPaymentTable + ` ADD COLUMN IF NOT EXISTS payment_token TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE ` + mmodels.SubscriptionPaymentTable + ` ADD COLUMN IF NOT EXISTS proof_file_name VARCHAR(255) NOT NULL DEFAULT '';`,
		`ALTER TABLE ` + mmodels.SubscriptionPaymentTable + ` ADD COLUMN IF NOT EXISTS proof_file_size BIGINT NOT NULL DEFAULT 0;`,
		`ALTER TABLE ` + mmodels.SubscriptionPaymentTable + ` ADD COLUMN IF NOT EXISTS proof_content_type VARCHAR(255) NOT NULL DEFAULT '';`,
		`ALTER TABLE ` + mmodels.SubscriptionPaymentTable + ` ADD COLUMN IF NOT EXISTS current_start_date DATE NOT NULL DEFAULT CURRENT_DATE;`,
		`ALTER TABLE ` + mmodels.SubscriptionPaymentTable + ` ADD COLUMN IF NOT EXISTS current_end_date DATE NOT NULL DEFAULT CURRENT_DATE;`,
		`ALTER TABLE ` + mmodels.SubscriptionPaymentTable + ` ADD COLUMN IF NOT EXISTS paid_at TIMESTAMPTZ NOT NULL DEFAULT NOW();`,
		`CREATE INDEX IF NOT EXISTS idx_subscription_payments_account_paid_at ON ` + mmodels.SubscriptionPaymentTable + ` (account_id, paid_at DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_subscription_payments_subscription ON ` + mmodels.SubscriptionPaymentTable + ` (account_id, subscription_id);`,
	}

	for _, statement := range statements {
		log.Println("[SQL]=", statement)
		m.DB.MustExec(statement)
	}
}

func (m SubscriptionPayment) ValidateCreate() error {
	if _, err := m.resolveAccountID(); err != nil {
		return err
	}
	if m.SubscriptionID <= 0 {
		return errors.New("subscription_id payment wajib diisi")
	}
	if len(m.UserIDs) == 0 || m.NumberOfUsers <= 0 {
		return errors.New("user payment subscription wajib diisi")
	}
	if m.Months <= 0 {
		return errors.New("jumlah bulan payment subscription wajib lebih dari 0")
	}
	if strings.TrimSpace(m.PaymentToken) == "" {
		return errors.New("token payment subscription wajib diisi")
	}
	return nil
}

func (m SubscriptionPayment) FindList(search string) ([]SubscriptionPayment, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}

	query := strings.ToLower(strings.TrimSpace(search))
	sql := `SELECT ` + subscriptionPaymentSelectColumns + `
	FROM ` + mmodels.SubscriptionPaymentTable + ` p
	WHERE p.account_id = $1`
	args := []interface{}{accountID}
	if query != "" {
		sql += ` AND (
			LOWER(p.app_type) LIKE $2 OR
			LOWER(p.currency) LIKE $2 OR
			LOWER(p.payment_token) LIKE $2 OR
			LOWER(p.proof_file_name) LIKE $2
		)`
		args = append(args, "%"+query+"%")
	}
	sql += `
	ORDER BY p.paid_at DESC, p.id DESC`

	log.Println("[SQL]=", sql)
	var items []SubscriptionPayment
	err = m.DB.Select(&items, sql, args...)
	return items, err
}

func (m SubscriptionPayment) resolveAccountID() (int64, error) {
	if m.Auth.AccountID > 0 {
		return m.Auth.AccountID, nil
	}
	if m.AccountID > 0 {
		return m.AccountID, nil
	}
	return 0, errors.New("id akun wajib diisi")
}
