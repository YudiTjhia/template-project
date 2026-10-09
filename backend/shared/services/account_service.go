package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"backend/shared/helpers"
	"backend/shared/metadata/impl"
	"backend/shared/metadata/interfaces"
	"backend/shared/mmodels"
	"backend/shared/models"

	"github.com/jmoiron/sqlx"
)

const accountTermsVersion = "POS-SaaS-ID-2026-09-12"

type AccountService struct {
	DB            *sqlx.DB
	TransactionDB *sqlx.DB
	Auth          models.Auth
	CRUD          interfaces.IService
}

func NewAccountService(db *sqlx.DB, auth models.Auth) *AccountService {
	return NewAccountServiceWithDatabases(db, db, auth)
}

func NewAccountServiceWithDatabases(db, transactionDB *sqlx.DB, auth models.Auth) *AccountService {
	if transactionDB == nil {
		transactionDB = db
	}
	return &AccountService{DB: db, TransactionDB: transactionDB, Auth: auth, CRUD: impl.NewMService()}
}

func (s *AccountService) NewAccount() models.Account {
	return models.NewAccount(s.DB, s.Auth)
}

func (s *AccountService) accountMModel() *mmodels.AccountMModel {
	model := mmodels.NewAccountMModel(context.Background(), s.DB)
	model.AuthUserName = s.Auth.UserName
	model.CreatedBy = s.Auth.UserName
	return model
}

func (s *AccountService) accountFromMModel(value mmodels.AccountMModel) models.Account {
	account := models.Account{
		BaseModel: models.BaseModel{DB: s.DB, Auth: s.Auth, CreatedAt: value.CreatedAt, CreatedBy: value.CreatedBy},
		ID:        value.ID, AccountName: value.AccountName, CompanyName: value.CompanyName,
		CompanyLogoURL: value.CompanyLogoURL, IsActive: value.IsActive, AccountEmail: value.AccountEmail,
		ReferralCode: value.ReferralCode, Address: value.Address, Latitude: value.Latitude,
		Longitude: value.Longitude, MobilePhone: value.MobilePhone, Phone: value.Phone,
		TermsAccepted: value.TermsAccepted, TermsVersion: value.TermsVersion, OTP: value.OTP,
		TermsAcceptedAt: value.TermsAcceptedAt, OTPExpired: value.OTPExpired,
	}
	if value.UpdatedAt.Valid {
		updatedAt := value.UpdatedAt.Time
		account.UpdatedAt = &updatedAt
	}
	if value.UpdatedBy.Valid {
		updatedBy := value.UpdatedBy.String
		account.UpdatedBy = &updatedBy
	}
	return account
}

func (s *AccountService) Migrate() error {
	return s.accountMModel().Migrate()
}

func (s *AccountService) CreateAccount(account models.Account) (int64, error) {
	acc := s.accountMModel()
	acc.AccountName = account.AccountName
	acc.CompanyName = account.CompanyName
	acc.CompanyLogoURL = account.CompanyLogoURL
	acc.IsActive = account.IsActive
	acc.AccountEmail = account.AccountEmail
	acc.ReferralCode = account.ReferralCode
	acc.Address = account.Address
	acc.Latitude = account.Latitude
	acc.Longitude = account.Longitude
	acc.MobilePhone = account.MobilePhone
	acc.Phone = account.Phone
	acc.TermsAccepted = account.TermsAccepted
	acc.TermsAcceptedAt = account.TermsAcceptedAt
	acc.TermsVersion = account.TermsVersion
	acc.OTP = account.OTP
	acc.OTPExpired = account.OTPExpired
	acc.CreatedBy = account.CreatedBy
	if acc.CreatedBy == "" {
		acc.CreatedBy = account.AccountEmail
	}
	if acc.CreatedBy == "" {
		acc.CreatedBy = account.AccountName
	}
	if acc.CreatedBy == "" {
		acc.CreatedBy = "system"
	}
	id, err := acc.Create()
	if err != nil {
		return 0, err
	}

	subscriptionService := NewSubscriptionServiceWithDatabases(s.DB, s.TransactionDB, models.Auth{
		AccountID: id,
		UserName:  s.Auth.UserName,
		Email:     s.Auth.Email,
	})
	if _, err := subscriptionService.EnsureDefaultSubscriptionForAccount(id); err != nil {
		return 0, err
	}

	return id, nil
}

func (s *AccountService) RegisterAccount(account models.Account) (int64, string, time.Time, error) {
	if err := validateRegisterAccount(account); err != nil {
		return 0, "", time.Time{}, err
	}
	if err := s.validateReferralCode(account.ReferralCode); err != nil {
		return 0, "", time.Time{}, err
	}
	if err := validateRegisterPassword(account.Password, account.ConfirmPassword); err != nil {
		return 0, "", time.Time{}, err
	}
	if err := validateRegisterPIN(account.PIN, account.ConfirmPIN); err != nil {
		return 0, "", time.Time{}, err
	}
	if err := validateRegisterTerms(account.TermsAccepted); err != nil {
		return 0, "", time.Time{}, err
	}

	existingAccount, err := s.FindByEmail(account.AccountEmail)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, "", time.Time{}, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		existingAccount = nil
	}

	if existingAccount != nil {
		if existingAccount.IsActive {
			return 0, "", time.Time{}, errors.New("email sudah digunakan")
		}
		return s.resendRegistrationOTP(*existingAccount, account)
	}

	return s.createRegistrationAccount(account)
}

func (s *AccountService) createRegistrationAccount(account models.Account) (int64, string, time.Time, error) {
	otp, err := s.generateUniqueRegisterOTP(0)
	if err != nil {
		return 0, "", time.Time{}, err
	}

	expired := time.Now().UTC().Add(15 * time.Minute)
	termsAcceptedAt := time.Now().UTC()
	account.IsActive = false
	account.TermsAccepted = true
	account.TermsAcceptedAt = &termsAcceptedAt
	if strings.TrimSpace(account.TermsVersion) == "" {
		account.TermsVersion = accountTermsVersion
	}
	account.OTP = otp
	account.OTPExpired = &expired
	account.CreatedBy = account.AccountEmail
	if strings.TrimSpace(account.CreatedBy) == "" {
		account.CreatedBy = account.AccountName
	}
	if strings.TrimSpace(account.CreatedBy) == "" {
		account.CreatedBy = "system"
	}

	id, err := s.CreateAccount(account)
	if err != nil {
		return 0, "", time.Time{}, err
	}
	account.ID = id

	userID, err := s.CreateUserFromAccount(account, account.Password)
	if err != nil {
		return 0, "", time.Time{}, err
	}
	partnerService := NewSalesPartnerServiceWithDatabases(s.TransactionDB, s.DB, s.Auth)
	if _, err := partnerService.EnsureForUser(userID); err != nil {
		return 0, "", time.Time{}, err
	}
	if err := partnerService.LinkReferredAccount(id, account.ReferralCode); err != nil {
		return 0, "", time.Time{}, err
	}

	s.sendRegisterOTPEmailAsync(account, otp, expired)

	return id, otp, expired, nil
}

func (s *AccountService) resendRegistrationOTP(existingAccount models.Account, account models.Account) (int64, string, time.Time, error) {
	otp, err := s.generateUniqueRegisterOTP(existingAccount.ID)
	if err != nil {
		return 0, "", time.Time{}, err
	}

	expired := time.Now().UTC().Add(15 * time.Minute)
	termsAcceptedAt := time.Now().UTC()
	existingAccount.AccountName = account.AccountName
	existingAccount.CompanyName = account.CompanyName
	existingAccount.CompanyLogoURL = account.CompanyLogoURL
	existingAccount.IsActive = false
	existingAccount.AccountEmail = account.AccountEmail
	if strings.TrimSpace(account.ReferralCode) != "" {
		existingAccount.ReferralCode = account.ReferralCode
	}
	existingAccount.Address = account.Address
	existingAccount.Latitude = account.Latitude
	existingAccount.Longitude = account.Longitude
	existingAccount.MobilePhone = account.MobilePhone
	existingAccount.Phone = account.Phone
	existingAccount.TermsAccepted = true
	existingAccount.TermsAcceptedAt = &termsAcceptedAt
	if strings.TrimSpace(account.TermsVersion) != "" {
		existingAccount.TermsVersion = account.TermsVersion
	} else if strings.TrimSpace(existingAccount.TermsVersion) == "" {
		existingAccount.TermsVersion = accountTermsVersion
	}
	existingAccount.OTP = otp
	existingAccount.OTPExpired = &expired
	existingAccount.CreatedBy = account.AccountEmail
	if strings.TrimSpace(existingAccount.CreatedBy) == "" {
		existingAccount.CreatedBy = account.AccountName
	}
	if strings.TrimSpace(existingAccount.CreatedBy) == "" {
		existingAccount.CreatedBy = "system"
	}

	if err := existingAccount.Update(); err != nil {
		return 0, "", time.Time{}, err
	}

	account.ID = existingAccount.ID
	account.IsActive = false
	account.TermsAccepted = true
	account.TermsAcceptedAt = &termsAcceptedAt
	account.TermsVersion = existingAccount.TermsVersion
	account.OTP = otp
	account.OTPExpired = &expired
	account.CreatedBy = existingAccount.CreatedBy
	user, err := s.upsertRegistrationUser(account)
	if err != nil {
		return 0, "", time.Time{}, err
	}
	partnerService := NewSalesPartnerServiceWithDatabases(s.TransactionDB, s.DB, s.Auth)
	if _, err := partnerService.EnsureForUser(user.ID); err != nil {
		return 0, "", time.Time{}, err
	}
	if err := partnerService.LinkReferredAccount(existingAccount.ID, existingAccount.ReferralCode); err != nil {
		return 0, "", time.Time{}, err
	}

	s.sendRegisterOTPEmailAsync(account, otp, expired)

	return existingAccount.ID, otp, expired, nil
}

func (s *AccountService) upsertRegistrationUser(account models.Account) (*models.User, error) {
	user := models.NewUser(s.DB, models.Auth{
		AccountID: account.ID,
		UserName:  s.Auth.UserName,
		Email:     account.AccountEmail,
	})
	user.AccountID = account.ID
	user.UserEmail = account.AccountEmail
	user.MobilePhone = account.MobilePhone
	user.UserPassword = account.Password
	user.PIN = account.PIN
	user.ConfirmPIN = account.ConfirmPIN
	user.UserType = models.UserTypeSuperAdmin
	user.UserRoles = nil
	user.IsActive = false
	user.Username = account.AccountName
	user.CreatedBy = s.Auth.UserName
	if user.CreatedBy == "" {
		user.CreatedBy = account.AccountEmail
	}
	if user.CreatedBy == "" {
		user.CreatedBy = account.AccountName
	}
	if user.CreatedBy == "" {
		user.CreatedBy = "system"
	}

	existing, err := user.FindByLoginIdentifierGlobal(account.AccountEmail)
	if err == nil && existing != nil && existing.ID > 0 {
		existing.UserEmail = user.UserEmail
		existing.MobilePhone = user.MobilePhone
		existing.UserPassword = user.UserPassword
		existing.PIN = user.PIN
		existing.ConfirmPIN = user.ConfirmPIN
		existing.UserType = user.UserType
		existing.UserRoles = user.UserRoles
		existing.IsActive = user.IsActive
		existing.Username = user.Username
		updatedBy := user.CreatedBy
		existing.UpdatedBy = &updatedBy
		if err := existing.Update(); err != nil {
			return nil, err
		}
		return existing, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		existing = nil
	}
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "no rows") {
		return nil, err
	}

	id, err := user.Create()
	if err != nil {
		return nil, err
	}
	user.ID = id
	return &user, nil
}

func (s *AccountService) UpdateAccount(account models.Account) error {
	acc := s.accountMModel()
	acc.ID = account.ID
	acc.AccountName = account.AccountName
	acc.CompanyName = account.CompanyName
	acc.CompanyLogoURL = account.CompanyLogoURL
	acc.IsActive = account.IsActive
	acc.AccountEmail = account.AccountEmail
	acc.ReferralCode = account.ReferralCode
	acc.Address = account.Address
	acc.Latitude = account.Latitude
	acc.Longitude = account.Longitude
	acc.MobilePhone = account.MobilePhone
	acc.Phone = account.Phone
	acc.OTP = account.OTP
	acc.OTPExpired = account.OTPExpired
	if account.UpdatedBy != nil {
		acc.UpdatedBy = sql.NullString{String: *account.UpdatedBy, Valid: true}
	}
	return acc.Update()
}

func (s *AccountService) CreateUserFromAccount(account models.Account, password string) (int64, error) {
	password = strings.TrimSpace(password)
	if password == "" {
		return 0, errors.New("password user wajib diisi")
	}

	user := models.NewUser(s.DB, models.Auth{
		AccountID: account.ID,
		UserName:  s.Auth.UserName,
	})
	user.AccountID = account.ID
	user.UserEmail = account.AccountEmail
	user.MobilePhone = account.MobilePhone
	user.UserPassword = password
	user.PIN = account.PIN
	user.ConfirmPIN = account.ConfirmPIN
	user.UserType = models.UserTypeSuperAdmin
	user.UserRoles = nil
	user.IsActive = account.IsActive
	user.Username = account.AccountName
	user.CreatedBy = s.Auth.UserName
	if user.CreatedBy == "" {
		user.CreatedBy = account.AccountEmail
	}
	if user.CreatedBy == "" {
		user.CreatedBy = account.AccountName
	}
	if user.CreatedBy == "" {
		user.CreatedBy = "system"
	}
	return user.Create()
}

func (s *AccountService) ActivateAccount(id int64) error {
	if id <= 0 {
		return errors.New("id account tidak valid")
	}

	updatedBy := s.Auth.UserName
	if updatedBy == "" {
		updatedBy = "system"
	}

	sql := `UPDATE accounts
	SET is_active = true,
		updated_at = NOW(),
		updated_by = $1
	WHERE id = $2`

	res, err := s.DB.Exec(sql, updatedBy, id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("account tidak ditemukan")
	}

	return s.activateUsersForAccount(id, updatedBy)
}

func (s *AccountService) VerifyAccountRegistration(otp string) (*models.Account, *models.User, error) {

	otp = strings.TrimSpace(otp)
	if otp == "" {
		return nil, nil, errors.New("otp wajib diisi")
	}

	account, err := s.findAccountByOTP(otp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, errors.New("otp tidak ditemukan atau sudah dipakai")
		}
		return nil, nil, err
	}

	if strings.TrimSpace(account.OTP) == "" {
		return nil, nil, errors.New("otp belum dibuat")
	}
	if strings.TrimSpace(account.OTP) != strings.TrimSpace(otp) {
		return nil, nil, errors.New("otp tidak valid")
	}
	if account.OTPExpired != nil && account.OTPExpired.Before(time.Now().UTC()) {
		return nil, nil, errors.New("otp sudah expired")
	}

	user, err := s.FindUserByAccountAndCredentials(account.AccountEmail, account.MobilePhone)
	if err == nil && user != nil {
		user.DB = s.DB
		user.Auth = s.Auth
		user.IsActive = true
		// ConfirmPIN is not persisted; restore the pair for model validation
		// when activating the existing registration user.
		user.ConfirmPIN = user.PIN
		updatedBy := s.Auth.UserName
		if updatedBy == "" {
			updatedBy = account.AccountEmail
		}
		user.UpdatedBy = &updatedBy
		if err := user.Update(); err != nil {
			return nil, nil, err
		}

		account.IsActive = true
		account.OTP = ""
		account.OTPExpired = nil
		if err := s.UpdateAccount(*account); err != nil {
			return nil, nil, err
		}
		if err := s.activateUsersForAccount(account.ID, updatedBy); err != nil {
			return nil, nil, err
		}
		s.sendRegisterAccountIDEmailAsync(*account)
		return account, user, nil
	}
	if err != nil && !errors.Is(err, errUserNotFound) {
		return nil, nil, err
	}

	account.IsActive = true
	userID, err := s.CreateUserFromAccount(*account, account.Password)
	if err != nil {
		return nil, nil, err
	}

	createdUser := models.NewUser(s.DB, models.Auth{AccountID: account.ID})
	createdUser.ID = userID
	found, err := createdUser.FindByID()
	if err != nil {
		return nil, nil, err
	}

	account.IsActive = true
	account.OTP = ""
	account.OTPExpired = nil
	if err := s.UpdateAccount(*account); err != nil {
		return nil, nil, err
	}
	if err := s.activateUsersForAccount(account.ID, account.AccountEmail); err != nil {
		return nil, nil, err
	}
	s.sendRegisterAccountIDEmailAsync(*account)

	return account, found, nil
}

func (s *AccountService) activateUsersForAccount(accountID int64, updatedBy string) error {
	if accountID <= 0 {
		return errors.New("id account tidak valid")
	}
	if strings.TrimSpace(updatedBy) == "" {
		updatedBy = s.Auth.UserName
	}
	if strings.TrimSpace(updatedBy) == "" {
		updatedBy = "system"
	}

	_, err := s.DB.Exec(`
		UPDATE users
		SET is_active = true,
			updated_at = NOW(),
			updated_by = $1
		WHERE account_id = $2
		  AND is_active = false
	`, updatedBy, accountID)
	return err
}

func (s *AccountService) GenerateAccountOTP(id int64) (string, error) {
	if id <= 0 {
		return "", errors.New("id account tidak valid")
	}

	found, err := s.FindByID(id)
	if err != nil {
		return "", err
	}

	otp, err := s.generateUniqueRegisterOTP(id)
	if err != nil {
		return "", err
	}

	expired := time.Now().UTC().Add(15 * time.Minute)
	found.OTP = otp
	found.OTPExpired = &expired
	if err := s.UpdateAccount(*found); err != nil {
		return "", err
	}

	s.sendRegisterOTPEmailAsync(*found, otp, expired)

	return otp, nil
}

func (s *AccountService) ResendAccountRegistrationOTP(accountID int64) (time.Time, error) {
	if accountID <= 0 {
		return time.Time{}, errors.New("id account tidak valid")
	}

	account, err := s.FindByID(accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return time.Time{}, errors.New("account tidak ditemukan")
		}
		return time.Time{}, err
	}
	if account.IsActive {
		return time.Time{}, errors.New("account sudah diverifikasi")
	}

	_, err = s.GenerateAccountOTP(account.ID)
	if err != nil {
		return time.Time{}, err
	}

	updated, err := s.FindByID(account.ID)
	if err != nil || updated.OTPExpired == nil {
		return time.Time{}, err
	}
	return *updated.OTPExpired, nil
}

func (s *AccountService) Create(account models.Account) (int64, error) {
	return s.CreateAccount(account)
}

func (s *AccountService) Update(account models.Account) error {
	return s.UpdateAccount(account)
}

func (s *AccountService) Delete(id int64) error {
	acc := s.accountMModel()
	acc.ID = id
	return acc.Delete()
}

func (s *AccountService) FindByID(id int64) (*models.Account, error) {
	acc := s.accountMModel()
	acc.ID = id
	found, err := acc.FindByID()
	if err != nil {
		return nil, err
	}
	account := s.accountFromMModel(*found)
	return &account, nil
}

func (s *AccountService) FindByEmail(email string) (*models.Account, error) {
	found, err := s.accountMModel().FindByEmail_(email)
	if err != nil {
		return nil, err
	}
	account := s.accountFromMModel(*found)
	return &account, nil
}

func (s *AccountService) findAccountByOTP(otp string) (*models.Account, error) {
	found, err := s.accountMModel().FindByOTP_(otp)
	if err != nil {
		return nil, err
	}
	account := s.accountFromMModel(*found)
	return &account, nil
}

var errUserNotFound = errors.New("user tidak ditemukan")

func (s *AccountService) FindUserByAccountAndCredentials(email, mobilephone string) (*models.User, error) {
	user := models.NewUser(s.DB, s.Auth)
	user.UserEmail = email
	found, err := user.FindByLoginIdentifierGlobal(email)
	if err == nil && found != nil {
		found.DB = s.DB
		found.Auth = s.Auth
		return found, nil
	}
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "no rows") {
		return nil, err
	}

	user.MobilePhone = mobilephone
	found, err = user.FindByLoginIdentifierGlobal(mobilephone)
	if err == nil && found != nil {
		found.DB = s.DB
		found.Auth = s.Auth
		return found, nil
	}
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "no rows") {
		return nil, errUserNotFound
	}

	return nil, errUserNotFound
}

func (s *AccountService) FindList(search string) ([]models.Account, error) {
	return s.findAccounts(search, "account_name", "ASC", 50)
}

func (s *AccountService) FindAll() ([]models.Account, error) {
	return s.findAccounts("", "account_name", "ASC", 0)
}

func (s *AccountService) FindRecentList(limit int, search string) ([]models.Account, error) {
	if limit <= 0 {
		limit = 5
	}
	return s.findAccounts(search, "created_at,id", "DESC", limit)
}

func (s *AccountService) findAccounts(search, orderBy, direction string, limit int) ([]models.Account, error) {
	model := s.accountMModel()
	rows := make([]mmodels.AccountMModel, 0)
	if err := s.CRUD.FindList_(model, &rows, nil, search,
		[]string{"account_name", "company_name", "account_email", "address", "mobilephone", "phone"}, orderBy, direction, limit); err != nil {
		return nil, err
	}
	accounts := make([]models.Account, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, s.accountFromMModel(row))
	}
	return accounts, nil
}

func validateRegisterPassword(password, confirmPassword string) error {
	password = strings.TrimSpace(password)
	confirmPassword = strings.TrimSpace(confirmPassword)

	if password == "" {
		return errors.New("password wajib diisi")
	}
	if confirmPassword == "" {
		return errors.New("confirm password wajib diisi")
	}
	if password != confirmPassword {
		return errors.New("confirm password tidak cocok")
	}

	return nil
}

func validateRegisterPIN(pin, confirmPIN string) error {
	pin = strings.TrimSpace(pin)
	confirmPIN = strings.TrimSpace(confirmPIN)

	if pin == "" {
		return errors.New("pin wajib diisi")
	}
	if confirmPIN == "" {
		return errors.New("confirm pin wajib diisi")
	}
	if len(pin) != 6 {
		return errors.New("pin harus 6 digit angka")
	}
	for _, char := range pin {
		if char < '0' || char > '9' {
			return errors.New("pin harus 6 digit angka")
		}
	}
	if pin != confirmPIN {
		return errors.New("confirm pin tidak cocok")
	}

	return nil
}

func validateRegisterTerms(accepted bool) error {
	if !accepted {
		return errors.New("syarat dan ketentuan wajib disetujui")
	}
	return nil
}

func validateRegisterAccount(account models.Account) error {
	account.AccountName = strings.TrimSpace(account.AccountName)
	account.AccountEmail = strings.TrimSpace(account.AccountEmail)
	account.Address = strings.TrimSpace(account.Address)
	account.MobilePhone = strings.TrimSpace(account.MobilePhone)
	account.Phone = strings.TrimSpace(account.Phone)

	if account.AccountName == "" {
		return errors.New("account_name wajib diisi")
	}
	if len(account.AccountName) < 3 {
		return errors.New("account_name minimal 3 karakter")
	}
	if account.AccountEmail == "" {
		return errors.New("account_email wajib diisi")
	}
	if _, err := mail.ParseAddress(account.AccountEmail); err != nil {
		return errors.New("account_email tidak valid")
	}
	if account.Address == "" {
		return errors.New("address wajib diisi")
	}
	if account.MobilePhone == "" {
		return errors.New("mobilephone wajib diisi")
	}
	if len(account.MobilePhone) < 6 {
		return errors.New("mobilephone minimal 6 karakter")
	}

	return nil
}

func (s *AccountService) validateReferralCode(code string) error {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return nil
	}
	var exists bool
	if err := s.TransactionDB.Get(&exists, `SELECT EXISTS (SELECT 1 FROM sales_partners WHERE LOWER(referral_code)=LOWER($1) AND status='ACTIVE')`, code); err != nil {
		return err
	}
	if !exists {
		return errors.New("refer code tidak ditemukan atau tidak aktif")
	}
	return nil
}

func (s *AccountService) sendRegisterOTPEmail(account models.Account, otp string, expired time.Time) error {
	cfg := helpers.LoadSMTPConfig()
	if !cfg.Enabled() {
		return nil
	}

	subject, textBody, htmlBody, err := helpers.BuildRegisterOTPMail(account.AccountName, otp, expired)
	if err != nil {
		return err
	}
	message := helpers.WriteHTMLMultipartEmailMessage(cfg.FromHeader(), account.AccountEmail, subject, textBody, htmlBody)

	addr := cfg.Addr()
	var auth smtp.Auth
	if cfg.Username != "" {
		auth = cfg.Auth()
	}

	if err := smtp.SendMail(addr, auth, cfg.FromEmail, []string{account.AccountEmail}, message); err != nil {
		return err
	}
	return nil
}

func (s *AccountService) sendRegisterOTPEmailAsync(account models.Account, otp string, expired time.Time) {
	helpers.SendEmailInBackground("register account otp email", func() error {
		return s.sendRegisterOTPEmail(account, otp, expired)
	})
}

func (s *AccountService) sendRegisterAccountIDEmail(account models.Account) error {
	cfg := helpers.LoadSMTPConfig()
	if !cfg.Enabled() {
		return nil
	}

	subject, body := helpers.BuildRegisterAccountIDMail(account.AccountName, account.ID)
	message := helpers.WriteEmailMessage(cfg.FromHeader(), account.AccountEmail, subject, body)

	addr := cfg.Addr()
	var auth smtp.Auth
	if cfg.Username != "" {
		auth = cfg.Auth()
	}

	if err := smtp.SendMail(addr, auth, cfg.FromEmail, []string{account.AccountEmail}, message); err != nil {
		return err
	}
	return nil
}

func (s *AccountService) sendRegisterAccountIDEmailAsync(account models.Account) {
	helpers.SendEmailInBackground("register account id email", func() error {
		return s.sendRegisterAccountIDEmail(account)
	})
}

func (s *AccountService) generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (s *AccountService) generateUniqueRegisterOTP(excludeAccountID int64) (string, error) {
	for attempt := 0; attempt < 20; attempt++ {
		otp, err := s.generateOTP()
		if err != nil {
			return "", err
		}

		account, err := s.findAccountByOTP(otp)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "no rows") {
				return otp, nil
			}
			if strings.Contains(strings.ToLower(err.Error()), "otp wajib diisi") {
				return otp, nil
			}
			return "", err
		}

		if account != nil && excludeAccountID > 0 && account.ID == excludeAccountID {
			return otp, nil
		}
	}

	return "", errors.New("gagal generate otp unik")
}
