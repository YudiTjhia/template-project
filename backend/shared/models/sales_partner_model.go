package models

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

const (
	SalesPartnerStatusActive   = "ACTIVE"
	SalesPartnerStatusInactive = "INACTIVE"
)

type SalesPartner struct {
	BaseModel

	ID                  int64   `json:"id" db:"id"`
	UserID              int64   `json:"userId" db:"user_id"`
	AccountID           int64   `json:"account_id" db:"account_id"`
	Name                string  `json:"name" db:"name"`
	Email               string  `json:"email" db:"email"`
	Phone               string  `json:"phone" db:"phone"`
	ReferralCode        string  `json:"referralCode" db:"referral_code"`
	Status              string  `json:"status" db:"status"`
	CommissionPercent   float64 `json:"commissionPercent" db:"commission_percent"`
	LinkedAccountsCount int     `json:"linkedAccountsCount" db:"linked_accounts_count"`
	TotalSales          float64 `json:"totalSales" db:"total_sales"`
	TotalCommission     float64 `json:"totalCommission" db:"total_commission"`
}

type PartnerAccount struct {
	BaseModel

	ID                 int64   `json:"id" db:"id"`
	PartnerID          int64   `json:"partnerId" db:"partner_id"`
	AccountID          int64   `json:"accountId" db:"account_id"`
	AccountName        string  `json:"accountName" db:"account_name"`
	CompanyName        string  `json:"companyName" db:"company_name"`
	SubscriptionID     *int64  `json:"subscriptionId,omitempty" db:"subscription_id"`
	SubscriptionStatus string  `json:"subscriptionStatus,omitempty" db:"subscription_status"`
	SalesAmount        float64 `json:"salesAmount" db:"sales_amount"`
	CommissionAmount   float64 `json:"commissionAmount" db:"commission_amount"`
	Status             string  `json:"status" db:"status"`
}

func SalesPartnerCommissionPercent() float64 {
	value, err := strconv.ParseFloat(strings.TrimSpace(os.Getenv("SALES_PARTNER_COMMISSION_PERCENT")), 64)
	if err != nil || value < 0 {
		return 10
	}
	return value
}

func NewSalesPartner(db *sqlx.DB, auth Auth) SalesPartner {
	return SalesPartner{BaseModel: BaseModel{DB: db, Auth: auth}, AccountID: auth.AccountID, Status: SalesPartnerStatusActive, CommissionPercent: SalesPartnerCommissionPercent()}
}

func (m SalesPartner) Migrate() {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS sales_partners (
			id SERIAL PRIMARY KEY,
			account_id BIGINT NOT NULL DEFAULT 0,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL DEFAULT '',
			phone VARCHAR(100) NOT NULL DEFAULT '',
			referral_code VARCHAR(100) NOT NULL,
			status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE',
			commission_percent NUMERIC(8,4) NOT NULL DEFAULT 10,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_by VARCHAR(255) NOT NULL DEFAULT '',
			updated_at TIMESTAMPTZ,
			updated_by VARCHAR(255)
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sales_partners_account_code ON sales_partners (account_id, LOWER(referral_code));`,
		`ALTER TABLE sales_partners ADD COLUMN IF NOT EXISTS user_id BIGINT NOT NULL DEFAULT 0;`,
		`DROP INDEX IF EXISTS idx_sales_partners_account_code;`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sales_partners_referral_code ON sales_partners (LOWER(referral_code));`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sales_partners_user_id_global ON sales_partners (user_id) WHERE user_id > 0;`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sales_partners_email_global ON sales_partners (LOWER(email)) WHERE email <> '';`,
		`CREATE INDEX IF NOT EXISTS idx_sales_partners_account_id ON sales_partners (account_id);`,
		`CREATE TABLE IF NOT EXISTS partner_accounts (
			id SERIAL PRIMARY KEY,
			partner_id BIGINT NOT NULL REFERENCES sales_partners(id) ON DELETE CASCADE,
			account_id BIGINT NOT NULL,
			status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_by VARCHAR(255) NOT NULL DEFAULT '',
			updated_at TIMESTAMPTZ,
			updated_by VARCHAR(255)
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_partner_accounts_partner_account ON partner_accounts (partner_id, account_id);`,
		`CREATE INDEX IF NOT EXISTS idx_partner_accounts_account_id ON partner_accounts (account_id);`,
		`CREATE TABLE IF NOT EXISTS sales_partner_commissions (
			id SERIAL PRIMARY KEY,
			partner_id BIGINT NOT NULL REFERENCES sales_partners(id) ON DELETE CASCADE,
			account_id BIGINT NOT NULL,
			payment_id BIGINT NOT NULL,
			payment_source VARCHAR(50) NOT NULL DEFAULT 'subscription',
			payment_token TEXT NOT NULL DEFAULT '',
			base_amount NUMERIC(18,2) NOT NULL DEFAULT 0,
			commission_percent NUMERIC(8,4) NOT NULL DEFAULT 0,
			commission_amount NUMERIC(18,2) NOT NULL DEFAULT 0,
			status VARCHAR(30) NOT NULL DEFAULT 'EARNED',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_by VARCHAR(255) NOT NULL DEFAULT ''
		);`,
		`ALTER TABLE sales_partner_commissions ADD COLUMN IF NOT EXISTS payment_source VARCHAR(50) NOT NULL DEFAULT 'subscription';`,
		`DROP INDEX IF EXISTS idx_sales_partner_commissions_payment;`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sales_partner_commissions_payment ON sales_partner_commissions (payment_source, payment_id);`,
		`CREATE INDEX IF NOT EXISTS idx_sales_partner_commissions_partner ON sales_partner_commissions (partner_id, created_at DESC);`,
		`ALTER TABLE sales_partners ADD COLUMN IF NOT EXISTS commission_percent NUMERIC(8,4) NOT NULL DEFAULT 10;`,
	}
	for _, statement := range statements {
		log.Println("[SQL]=", statement)
		m.DB.MustExec(statement)
	}
}

func (m SalesPartner) validateAccount() error {
	if m.Auth.AccountID <= 0 && m.AccountID <= 0 {
		return errors.New("id account wajib diisi")
	}
	return nil
}

func (m SalesPartner) Create() (int64, error) {
	if err := m.validateAccount(); err != nil {
		return 0, err
	}
	accountID := m.AccountID
	if accountID <= 0 {
		accountID = m.Auth.AccountID
	}
	name := strings.TrimSpace(m.Name)
	code := strings.ToUpper(strings.TrimSpace(m.ReferralCode))
	if code == "" {
		code = generateReferralCode(m.DB, name)
	}
	if name == "" {
		return 0, errors.New("nama sales partner wajib diisi")
	}
	if code == "" {
		return 0, errors.New("referral code wajib diisi")
	}
	percent := m.CommissionPercent
	if percent <= 0 {
		percent = SalesPartnerCommissionPercent()
	}
	status := strings.ToUpper(strings.TrimSpace(m.Status))
	if status == "" {
		status = SalesPartnerStatusActive
	}
	createdBy := m.CreatedBy
	if createdBy == "" {
		createdBy = m.Auth.Email
	}
	var id int64
	err := m.DB.QueryRow(`INSERT INTO sales_partners (user_id, account_id, name, email, phone, referral_code, status, commission_percent, created_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`, m.UserID, accountID, name, strings.TrimSpace(m.Email), strings.TrimSpace(m.Phone), code, status, percent, createdBy).Scan(&id)
	return id, err
}

func generateReferralCode(db *sqlx.DB, name string) string {
	base := strings.ToUpper(strings.Join(strings.Fields(name), ""))
	if len(base) > 6 {
		base = base[:6]
	}
	if base == "" {
		base = "PARTNER"
	}
	for i := 1; ; i++ {
		candidate := base + strconv.Itoa(i)
		var exists bool
		if err := db.Get(&exists, `SELECT EXISTS (SELECT 1 FROM sales_partners WHERE LOWER(referral_code)=LOWER($1))`, candidate); err == nil && !exists {
			return candidate
		}
	}
}

func (m SalesPartner) Update() error {
	if m.ID <= 0 {
		return errors.New("id sales partner tidak valid")
	}
	accountID := m.AccountID
	if accountID <= 0 {
		accountID = m.Auth.AccountID
	}
	status := strings.ToUpper(strings.TrimSpace(m.Status))
	if status == "" {
		status = SalesPartnerStatusActive
	}
	percent := m.CommissionPercent
	if percent <= 0 {
		percent = SalesPartnerCommissionPercent()
	}
	_, err := m.DB.Exec(`UPDATE sales_partners SET name=$1,email=$2,phone=$3,referral_code=$4,status=$5,commission_percent=$6,updated_at=NOW(),updated_by=$7 WHERE id=$8 AND account_id=$9`, strings.TrimSpace(m.Name), strings.TrimSpace(m.Email), strings.TrimSpace(m.Phone), strings.ToUpper(strings.TrimSpace(m.ReferralCode)), status, percent, m.Auth.Email, m.ID, accountID)
	return err
}

func (m SalesPartner) Delete(id int64) error {
	accountID := m.AccountID
	if accountID <= 0 {
		accountID = m.Auth.AccountID
	}
	_, err := m.DB.Exec(`DELETE FROM sales_partners WHERE id=$1 AND account_id=$2`, id, accountID)
	return err
}

func (m SalesPartner) FindList(search string) ([]SalesPartner, error) {
	accountID := m.AccountID
	if accountID <= 0 {
		accountID = m.Auth.AccountID
	}
	query := `SELECT p.id,p.account_id,p.name,p.email,p.phone,p.referral_code,p.status,p.commission_percent,
		COUNT(DISTINCT pa.id) AS linked_accounts_count,
		COALESCE((SELECT SUM(c.base_amount) FROM sales_partner_commissions c WHERE c.partner_id=p.id),0) AS total_sales,
		COALESCE((SELECT SUM(c.commission_amount) FROM sales_partner_commissions c WHERE c.partner_id=p.id),0) AS total_commission
		FROM sales_partners p LEFT JOIN partner_accounts pa ON pa.partner_id=p.id AND pa.status='ACTIVE'
		WHERE p.account_id=$1`
	args := []interface{}{accountID}
	if strings.TrimSpace(search) != "" {
		query += ` AND (p.name ILIKE $2 OR p.email ILIKE $2 OR p.referral_code ILIKE $2)`
		args = append(args, "%"+strings.TrimSpace(search)+"%")
	}
	query += ` GROUP BY p.id ORDER BY p.created_at DESC`
	items := []SalesPartner{}
	err := m.DB.Select(&items, query, args...)
	return items, err
}

func (m SalesPartner) FindByID(id int64) (*SalesPartner, error) {
	items, err := m.FindList("")
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.ID == id {
			return &item, nil
		}
	}
	return nil, errors.New("sales partner tidak ditemukan")
}

func (m SalesPartner) LinkAccount(partnerID, accountID int64) error {
	if partnerID <= 0 || accountID <= 0 {
		return errors.New("partner_id dan account_id wajib diisi")
	}
	ownerID := m.AccountID
	if ownerID <= 0 {
		ownerID = m.Auth.AccountID
	}
	var exists int
	if err := m.DB.Get(&exists, `SELECT COUNT(1) FROM sales_partners WHERE id=$1 AND account_id=$2`, partnerID, ownerID); err != nil {
		return err
	}
	if exists == 0 {
		return errors.New("sales partner tidak ditemukan")
	}
	_, err := m.DB.Exec(`INSERT INTO partner_accounts (partner_id,account_id,created_by) VALUES ($1,$2,$3) ON CONFLICT (partner_id,account_id) DO UPDATE SET status='ACTIVE',updated_at=NOW(),updated_by=$3`, partnerID, accountID, m.Auth.Email)
	return err
}

func (m SalesPartner) UnlinkAccount(partnerID, accountID int64) error {
	ownerID := m.AccountID
	if ownerID <= 0 {
		ownerID = m.Auth.AccountID
	}
	_, err := m.DB.Exec(`DELETE FROM partner_accounts pa USING sales_partners p WHERE pa.partner_id=p.id AND pa.partner_id=$1 AND pa.account_id=$2 AND p.account_id=$3`, partnerID, accountID, ownerID)
	return err
}

func (m SalesPartner) FindPartnerAccounts(partnerID int64) ([]PartnerAccount, error) {
	ownerID := m.AccountID
	if ownerID <= 0 {
		ownerID = m.Auth.AccountID
	}
	items := []PartnerAccount{}
	err := m.DB.Select(&items, `SELECT pa.id,pa.partner_id,pa.account_id,'' AS account_name,'' AS company_name,NULL::BIGINT AS subscription_id,'' AS subscription_status,0::NUMERIC AS sales_amount,0::NUMERIC AS commission_amount,pa.status FROM partner_accounts pa JOIN sales_partners p ON p.id=pa.partner_id WHERE pa.partner_id=$1 AND p.account_id=$2 ORDER BY pa.created_at DESC`, partnerID, ownerID)
	return items, err
}
