package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"backend/shared/mmodels"
	"backend/shared/models"

	"github.com/jmoiron/sqlx"
)

type SalesPartnerService struct {
	DB       *sqlx.DB
	MasterDB *sqlx.DB
	Auth     models.Auth
}

func NewSalesPartnerService(db *sqlx.DB, auth models.Auth) *SalesPartnerService {
	return NewSalesPartnerServiceWithDatabases(db, db, auth)
}

func NewSalesPartnerServiceWithDatabases(db, masterDB *sqlx.DB, auth models.Auth) *SalesPartnerService {
	if masterDB == nil {
		masterDB = db
	}
	return &SalesPartnerService{DB: db, MasterDB: masterDB, Auth: auth}
}

func (s *SalesPartnerService) NewSalesPartner() models.SalesPartner {
	return models.NewSalesPartner(s.DB, s.Auth)
}

func (s *SalesPartnerService) salesPartnerMModel() *mmodels.SalesPartnerMModel {
	model := mmodels.NewSalesPartnerMModel(context.Background(), s.DB)
	model.AccountID = s.Auth.AccountID
	model.AuthAccountID = s.Auth.AccountID
	model.AuthEmail = s.Auth.Email
	return model
}

func (s *SalesPartnerService) Migrate() error {
	return s.salesPartnerMModel().Migrate()
}

func (s *SalesPartnerService) Create(partner models.SalesPartner) (int64, error) {
	if partner.UserID <= 0 {
		return 0, fmt.Errorf("user wajib dipilih")
	}
	var user struct {
		ID          int64  `db:"id"`
		AccountID   int64  `db:"account_id"`
		UserEmail   string `db:"user_email"`
		MobilePhone string `db:"mobilephone"`
		Username    string `db:"username"`
	}
	if err := s.MasterDB.Get(&user, `SELECT id,account_id,user_email,mobilephone,username FROM users WHERE id=$1 AND account_id=$2 AND is_active=true`, partner.UserID, s.Auth.AccountID); err != nil {
		return 0, fmt.Errorf("user tidak ditemukan atau tidak aktif")
	}
	item := s.salesPartnerMModel()
	item.UserID = user.ID
	item.AccountID, item.Name, item.Email, item.Phone = user.AccountID, user.Username, user.UserEmail, user.MobilePhone
	item.ReferralCode, item.Status, item.CommissionPercent, item.CreatedBy = "", partner.Status, partner.CommissionPercent, partner.CreatedBy
	return item.Create()
}

// EnsureForUser creates or refreshes the global Sales Partner record for a user.
// It accepts inactive users so registration can create the record before OTP verification.
func (s *SalesPartnerService) EnsureForUser(userID int64) (int64, error) {
	if userID <= 0 {
		return 0, fmt.Errorf("user tidak valid")
	}
	var user struct {
		ID          int64  `db:"id"`
		AccountID   int64  `db:"account_id"`
		UserEmail   string `db:"user_email"`
		MobilePhone string `db:"mobilephone"`
		Username    string `db:"username"`
	}
	if err := s.MasterDB.Get(&user, `SELECT id,account_id,user_email,mobilephone,username FROM users WHERE id=$1`, userID); err != nil {
		return 0, fmt.Errorf("user tidak ditemukan")
	}

	var existing models.SalesPartner
	err := s.DB.Get(&existing, `SELECT id,user_id,account_id,name,email,phone,referral_code,status,commission_percent FROM sales_partners WHERE user_id=$1 OR LOWER(email)=LOWER($2) ORDER BY CASE WHEN user_id=$1 THEN 0 ELSE 1 END, id LIMIT 1`, user.ID, user.UserEmail)
	if err == nil {
		_, err = s.DB.Exec(`UPDATE sales_partners SET user_id=$1,account_id=$2,name=$3,email=$4,phone=$5,updated_at=NOW(),updated_by=$6 WHERE id=$7`, user.ID, user.AccountID, user.Username, user.UserEmail, user.MobilePhone, user.UserEmail, existing.ID)
		return existing.ID, err
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	item := s.salesPartnerMModel()
	item.AuthAccountID = 0 // registration may ensure partners for accounts other than the current auth scope
	item.UserID = user.ID
	item.AccountID = user.AccountID
	item.Name = user.Username
	item.Email = user.UserEmail
	item.Phone = user.MobilePhone
	item.CreatedBy = user.UserEmail
	return item.Create()
}

func (s *SalesPartnerService) ApplySubscriptionCommission(accountID, paymentID int64, amount float64, paymentToken, source string) error {
	if accountID <= 0 || paymentID <= 0 || amount <= 0 {
		return nil
	}
	source = strings.TrimSpace(source)
	if source == "" {
		source = "subscription"
	}
	var referralCode string
	if err := s.MasterDB.Get(&referralCode, `SELECT referral_code FROM accounts WHERE id=$1`, accountID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	referralCode = strings.ToUpper(strings.TrimSpace(referralCode))
	if referralCode == "" {
		return nil
	}
	var partner struct {
		ID                int64   `db:"id"`
		CommissionPercent float64 `db:"commission_percent"`
	}
	if err := s.DB.Get(&partner, `SELECT id,commission_percent FROM sales_partners WHERE LOWER(referral_code)=LOWER($1) AND status='ACTIVE'`, referralCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	commission := models.RoundMoney(amount * partner.CommissionPercent / 100)
	_, err := s.DB.Exec(`INSERT INTO sales_partner_commissions (partner_id,account_id,payment_id,payment_source,payment_token,base_amount,commission_percent,commission_amount,created_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT (payment_source,payment_id) DO NOTHING`, partner.ID, accountID, paymentID, source, paymentToken, amount, partner.CommissionPercent, commission, source)
	return err
}

func (s *SalesPartnerService) LinkReferredAccount(accountID int64, referralCode string) error {
	if accountID <= 0 || strings.TrimSpace(referralCode) == "" {
		return nil
	}
	var partnerID int64
	if err := s.DB.Get(&partnerID, `SELECT id FROM sales_partners WHERE LOWER(referral_code)=LOWER($1) AND status='ACTIVE'`, strings.TrimSpace(referralCode)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	_, err := s.DB.Exec(`INSERT INTO partner_accounts (partner_id,account_id,created_by) VALUES ($1,$2,$3) ON CONFLICT (partner_id,account_id) DO UPDATE SET status='ACTIVE',updated_at=NOW(),updated_by=$3`, partnerID, accountID, "registration-referral")
	return err
}

func (s *SalesPartnerService) Update(partner models.SalesPartner) error {
	item := s.salesPartnerMModel()
	item.ID, item.AccountID, item.Name, item.Email, item.Phone = partner.ID, partner.AccountID, partner.Name, partner.Email, partner.Phone
	item.ReferralCode, item.Status, item.CommissionPercent = partner.ReferralCode, partner.Status, partner.CommissionPercent
	return item.Update()
}

func (s *SalesPartnerService) Delete(id int64) error {
	item := s.salesPartnerMModel()
	item.ID = id
	return item.Delete()
}
func (s *SalesPartnerService) FindList(search string) ([]models.SalesPartner, error) {
	return s.NewSalesPartner().FindList(search)
}
func (s *SalesPartnerService) FindByID(id int64) (*models.SalesPartner, error) {
	return s.NewSalesPartner().FindByID(id)
}
func (s *SalesPartnerService) LinkAccount(partnerID, accountID int64) error {
	return s.NewSalesPartner().LinkAccount(partnerID, accountID)
}
func (s *SalesPartnerService) UnlinkAccount(partnerID, accountID int64) error {
	return s.NewSalesPartner().UnlinkAccount(partnerID, accountID)
}
func (s *SalesPartnerService) FindPartnerAccounts(partnerID int64) ([]models.PartnerAccount, error) {
	return s.NewSalesPartner().FindPartnerAccounts(partnerID)
}
