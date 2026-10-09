package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"net/smtp"
	"sort"
	"strings"
	"time"

	"backend/shared/helpers"
	"backend/shared/mmodels"
	"backend/shared/models"

	"github.com/jmoiron/sqlx"
)

type AuthService struct {
	DB            *sqlx.DB
	TransactionDB *sqlx.DB
}

type SessionAccess struct {
	UserType    string   `json:"user_type"`
	LoginType   string   `json:"login_type"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

func NewAuthService(db *sqlx.DB) *AuthService {
	return &AuthService{DB: db}
}

func NewAuthServiceWithTransactionDB(db *sqlx.DB, transactionDB *sqlx.DB) *AuthService {
	if transactionDB == nil {
		transactionDB = db
	}
	return &AuthService{DB: db, TransactionDB: transactionDB}
}

func (s *AuthService) transactionDB() *sqlx.DB {
	if s.TransactionDB != nil {
		return s.TransactionDB
	}
	return s.DB
}

func (s *AuthService) NewAuth() models.Auth {
	return models.NewAuth(s.DB)
}

func (s *AuthService) Migrate() {
	if err := s.newAuthMModel().Migrate(); err != nil {
		panic(err)
	}
}

func (s *AuthService) newAuthMModel() *mmodels.AuthMModel {
	return mmodels.NewAuthMModel(context.Background(), s.DB)
}

func authMModelToDTO(record *mmodels.AuthMModel, db *sqlx.DB) *models.Auth {
	if record == nil {
		return nil
	}
	dto := &models.Auth{
		DB: db, ID: record.ID, AccountID: record.AccountID, UserName: record.UserName,
		Email: record.Email, Name: record.Name, Token: record.Token, TokenExpired: record.TokenExpired,
		IsExpired: record.IsExpired, OTP: record.OTP, OTPExpired: record.OTPExpired,
		CreatedAt: record.CreatedAt,
	}
	if record.UpdatedAt.Valid {
		updatedAt := record.UpdatedAt.Time
		dto.UpdatedAt = &updatedAt
	}
	return dto
}

func (s *AuthService) Login(identifier, password string, _ int64, _ string) (string, error) {
	identifier = strings.TrimSpace(identifier)
	password = strings.TrimSpace(password)

	if identifier == "" {
		return "", errors.New("email atau mobilephone wajib diisi")
	}
	if password == "" {
		return "", errors.New("password wajib diisi")
	}

	user := models.NewUser(s.DB, models.Auth{})
	record, err := user.FindByLoginIdentifierGlobal(identifier)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no rows") {
			return "", errors.New("email/mobilephone atau password tidak valid")
		}
		return "", err
	}

	if !helpers.CompareCredential(record.UserPassword, password) {
		return "", errors.New("email/mobilephone atau password tidak valid")
	}
	if !record.IsActive {
		return "", errors.New("user tidak aktif")
	}
	if err := s.ensureAccountActive(record.AccountID); err != nil {
		return "", err
	}
	if token, ok, err := s.findValidSession(record); err != nil {
		return "", err
	} else if ok {
		return token, nil
	}

	return s.issueSessionForUser(record)
}

func (s *AuthService) ensureAccountActive(accountID int64) error {
	if accountID <= 0 {
		return errors.New("account user tidak valid")
	}
	var isActive bool
	err := s.DB.Get(&isActive, `SELECT is_active FROM accounts WHERE id = $1`, accountID)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("account tidak ditemukan")
	}
	if err != nil {
		return err
	}
	if !isActive {
		return errors.New("account tidak aktif")
	}
	return nil
}

// ChangePasswordAndPIN updates the credentials of the user attached to a valid session.
// PIN is optional so users can change only their password.
func (s *AuthService) ChangePasswordAndPIN(token, currentPassword, newPassword, confirmPassword, pin, confirmPIN string) error {
	token = strings.TrimSpace(token)
	currentPassword = strings.TrimSpace(currentPassword)
	newPassword = strings.TrimSpace(newPassword)
	confirmPassword = strings.TrimSpace(confirmPassword)
	pin = strings.TrimSpace(pin)
	confirmPIN = strings.TrimSpace(confirmPIN)
	if token == "" {
		return errors.New("token wajib diisi")
	}
	if currentPassword == "" {
		return errors.New("password saat ini wajib diisi")
	}
	if newPassword == "" {
		return errors.New("password baru wajib diisi")
	}
	if len(newPassword) < 6 {
		return errors.New("password minimal 6 karakter")
	}
	if confirmPassword == "" {
		return errors.New("konfirmasi password baru wajib diisi")
	}
	if newPassword != confirmPassword {
		return errors.New("konfirmasi password baru tidak cocok")
	}
	if pin != "" {
		if !isSixDigitPIN(pin) {
			return errors.New("pin harus 6 digit angka")
		}
		if confirmPIN == "" {
			return errors.New("confirm pin wajib diisi")
		}
		if pin != confirmPIN {
			return errors.New("confirm pin tidak cocok")
		}
	} else if confirmPIN != "" {
		return errors.New("pin wajib diisi jika confirm pin diisi")
	}

	auth := s.newAuthMModel()
	auth.Token = token
	record, err := auth.FindByToken()
	if err != nil {
		return errors.New("session tidak valid")
	}
	if record.IsExpired || (record.TokenExpired != nil && time.Now().UTC().After(*record.TokenExpired)) {
		return errors.New("session sudah expired")
	}

	user := models.NewUser(s.DB, models.Auth{AccountID: record.AccountID})
	user.UserEmail = record.Email
	userRecord, err := user.FindByEmail()
	if err != nil || !helpers.CompareCredential(userRecord.UserPassword, currentPassword) {
		return errors.New("password saat ini tidak valid")
	}

	passwordHash, err := helpers.HashCredential(newPassword)
	if err != nil {
		return err
	}
	pinHash := ""
	if pin != "" {
		pinHash, err = helpers.HashCredential(pin)
		if err != nil {
			return err
		}
	}
	updatedBy := record.Email
	_, err = s.DB.Exec(`UPDATE users
		SET user_password = $1,
			pin = CASE WHEN $2 = '' THEN pin ELSE $2 END,
			updated_at = NOW(),
			updated_by = $3
		WHERE id = $4 AND account_id = $5`, passwordHash, pinHash, updatedBy, userRecord.ID, record.AccountID)
	return err
}

func (s *AuthService) LoginWithPIN(identifier, pin string, _ int64, _ string) (string, error) {
	identifier = strings.TrimSpace(identifier)
	pin = strings.TrimSpace(pin)

	if identifier == "" {
		return "", errors.New("email atau mobilephone wajib diisi")
	}
	if !isSixDigitPIN(pin) {
		return "", errors.New("pin harus 6 digit angka")
	}

	user := models.NewUser(s.DB, models.Auth{})
	errorMessage := "email/mobilephone atau pin tidak valid"
	var record *models.User
	var err error
	if accountID, ok := models.IsNumericIdentifier(identifier); ok {
		record, err = user.FindActiveUserByAccountIDAndPIN(accountID, pin)
		errorMessage = "account id atau pin tidak valid"
	} else {
		record, err = user.FindByLoginIdentifierGlobal(identifier)
	}
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no rows") {
			return "", errors.New(errorMessage)
		}
		return "", err
	}

	if strings.TrimSpace(record.PIN) == "" {
		return "", errors.New("pin user belum dibuat")
	}

	if !helpers.CompareCredential(record.PIN, pin) {
		return "", errors.New(errorMessage)
	}
	if !record.IsActive {
		return "", errors.New("user tidak aktif")
	}
	if err := s.ensureAccountActive(record.AccountID); err != nil {
		return "", err
	}
	if token, ok, err := s.findValidSession(record); err != nil {
		return "", err
	} else if ok {
		return token, nil
	}

	return s.issueSessionForUser(record)
}

func (s *AuthService) issueSessionForUser(record *models.User) (string, error) {
	if record == nil {
		return "", errors.New("user tidak ditemukan")
	}

	email := strings.ToLower(strings.TrimSpace(record.UserEmail))
	name := strings.TrimSpace(record.Username)
	if record.AccountID <= 0 || email == "" || name == "" {
		return "", errors.New("data user tidak lengkap untuk membuat session")
	}

	token := s.mustGenerateToken()
	tokenExpired := time.Now().UTC().Add(24 * time.Hour)

	tx, err := s.DB.Beginx()
	if err != nil {
		return "", err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	authRecord := s.newAuthMModel()
	authRecord.AccountID = record.AccountID
	authRecord.Email = email
	authRecord.Name = name
	authRecord.Token = token
	authRecord.TokenExpired = &tokenExpired
	if err := authRecord.CreateSessionTx(tx, true); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	tx = nil

	return token, nil
}

func (s *AuthService) findValidSession(record *models.User) (string, bool, error) {
	if record == nil || record.AccountID <= 0 || strings.TrimSpace(record.UserEmail) == "" {
		return "", false, nil
	}

	existing := s.newAuthMModel()
	existing.AccountID = record.AccountID
	existing.Email = record.UserEmail
	existing, err := existing.FindReusableSession()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	if strings.TrimSpace(existing.Token) == "" {
		return "", false, nil
	}
	return existing.Token, true, nil
}

func (s *AuthService) recordDeviceLogin(record *models.User, deviceID string) {
	if record == nil {
		return
	}
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return
	}

	service := NewAccountDeviceService(s.DB, models.Auth{
		AccountID: record.AccountID,
		UserName:  strings.TrimSpace(record.Username),
		Email:     record.UserEmail,
	})
	if strings.TrimSpace(service.Auth.UserName) == "" {
		service.Auth.UserName = strings.TrimSpace(record.UserEmail)
	}
	if _, err := service.UpsertLogin(deviceID); err != nil {
		fmt.Println("[WARN] account device login upsert failed:", err)
	}
}

func (s *AuthService) ensureLocationCashDrawerAccess(record *models.User, locationID int64, deviceID string) error {
	if record == nil {
		return nil
	}

	deviceID = strings.TrimSpace(deviceID)
	if deviceID != "" {
		var deviceSession struct {
			ID          int64  `db:"id"`
			CashierID   int64  `db:"cashier_id"`
			CashierName string `db:"cashier_name"`
			SessionNo   string `db:"session_no"`
			DeviceID    string `db:"device_id"`
		}
		err := s.transactionDB().Get(&deviceSession, `
			SELECT id, cashier_id, cashier_name, session_no, device_id
			FROM cash_drawer_sessions
			WHERE account_id = $1
				AND device_id = $2
				AND status = 'OPEN'
			ORDER BY opened_at DESC, id DESC
			LIMIT 1`,
			record.AccountID,
			deviceID,
		)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if err == nil && deviceSession.ID > 0 && deviceSession.CashierID > 0 && deviceSession.CashierID != record.ID {
			sessionName := strings.TrimSpace(deviceSession.CashierName)
			if sessionName == "" {
				sessionName = fmt.Sprintf("user ID %d", deviceSession.CashierID)
			}
			return fmt.Errorf("login ditolak: browser ini masih terikat cash drawer session OPEN milik %s. Tutup session dulu atau gunakan user yang sama", sessionName)
		}
	}

	if locationID <= 0 {
		return nil
	}

	var session struct {
		ID          int64  `db:"id"`
		CashierID   int64  `db:"cashier_id"`
		CashierName string `db:"cashier_name"`
		SessionNo   string `db:"session_no"`
		DeviceID    string `db:"device_id"`
	}
	err := s.transactionDB().Get(&session, `
		SELECT id, cashier_id, cashier_name, session_no, device_id
		FROM cash_drawer_sessions
		WHERE account_id = $1
			AND location_id = $2
			AND status = 'OPEN'
		ORDER BY opened_at DESC, id DESC
		LIMIT 1`,
		record.AccountID,
		locationID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	if session.CashierID <= 0 || session.CashierID == record.ID {
		return nil
	}

	sessionName := strings.TrimSpace(session.CashierName)
	if sessionName == "" {
		sessionName = fmt.Sprintf("user ID %d", session.CashierID)
	}

	if strings.TrimSpace(session.DeviceID) != "" && deviceID != "" && strings.TrimSpace(session.DeviceID) != deviceID {
		return fmt.Errorf("login ditolak: lokasi ini masih punya cash drawer session OPEN dari browser lain milik %s. Login dengan user yang sama atau close session dulu", sessionName)
	}

	return fmt.Errorf("login ditolak: cash drawer session OPEN di lokasi ini masih dipakai %s. Login dengan user yang sama atau close session dulu", sessionName)
}

func (s *AuthService) ExpireSession(token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("token wajib diisi")
	}

	if err := s.ensureLogoutAllowed(token); err != nil {
		return err
	}

	auth := s.newAuthMModel()
	auth.Token = token
	return auth.ExpireByToken()
}

func (s *AuthService) ensureLogoutAllowed(token string) error {
	auth := s.newAuthMModel()
	auth.Token = token
	record, err := auth.FindByToken()
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no rows") {
			return nil
		}
		return err
	}

	user := models.NewUser(s.DB, models.Auth{AccountID: record.AccountID})
	user.UserEmail = record.Email
	cashier, err := user.FindByEmail()
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no rows") {
			return nil
		}
		return err
	}

	var sessionNo string
	err = s.transactionDB().Get(&sessionNo, `
		SELECT session_no
		FROM cash_drawer_sessions
		WHERE account_id = $1
			AND cashier_id = $2
			AND status = 'OPEN'
		ORDER BY opened_at DESC, id DESC
		LIMIT 1`, record.AccountID, cashier.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s.ensureSalesLogoutAllowed(record.AccountID, cashier.ID)
		}
		return err
	}

	if err := fmt.Errorf("logout ditolak: cash drawer session OPEN (%s) masih aktif. Close session terlebih dahulu atau gunakan Lock", sessionNo); err != nil {
		return err
	}
	return nil
}

func (s *AuthService) ensureSalesLogoutAllowed(accountID, cashierID int64) error {
	var pending struct {
		SalesNo       string `db:"sales_no"`
		OrderStatus   string `db:"order_status"`
		PaymentStatus string `db:"payment_status"`
	}

	err := s.transactionDB().Get(&pending, `
		SELECT sales_no, order_status, payment_status
		FROM sales_orders
		WHERE account_id = $1
			AND cashier_id = $2
			AND (
				order_status = 'DRAFT'
				OR payment_status IN ('UNPAID', 'PARTIAL')
			)
		ORDER BY created_at DESC, id DESC
		LIMIT 1`, accountID, cashierID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	return fmt.Errorf("logout ditolak: transaksi %s masih %s dan belum lunas. Selesaikan atau batalkan transaksi terlebih dahulu, atau gunakan Lock", pending.SalesNo, strings.ToLower(strings.TrimSpace(pending.PaymentStatus)))
}

func (s *AuthService) RefreshSession(token string) (*models.Auth, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("token wajib diisi")
	}

	auth := s.newAuthMModel()
	auth.Token = token
	record, err := auth.FindByToken()
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no rows") {
			return nil, errors.New("session tidak valid")
		}
		return nil, err
	}
	if record.IsExpired {
		return nil, errors.New("session sudah expired")
	}
	if record.TokenExpired != nil && time.Now().UTC().After(*record.TokenExpired) {
		return nil, errors.New("session sudah expired")
	}

	nextExpired := time.Now().UTC().Add(24 * time.Hour)
	record.TokenExpired = &nextExpired
	record.IsExpired = false
	if err := record.ExtendByToken(nextExpired); err != nil {
		return nil, err
	}

	return authMModelToDTO(record, s.DB), nil
}

// ResolveSessionAccess reads the current roles from the database. It is meant
// to be called when a session is issued/refreshed, not on every page request.
func (s *AuthService) ResolveSessionAccess(auth *models.Auth) (*SessionAccess, error) {
	if auth == nil || auth.AccountID <= 0 || strings.TrimSpace(auth.Email) == "" {
		return nil, errors.New("session tidak valid untuk mengambil permission")
	}

	userModel := models.NewUser(s.DB, *auth)
	userModel.UserEmail = auth.Email
	user, err := userModel.FindByEmail()
	if err != nil {
		return nil, err
	}

	roles := make([]string, 0, len(user.UserRoles))
	for _, roleName := range user.UserRoles {
		roleName = strings.TrimSpace(roleName)
		if roleName != "" {
			roles = append(roles, roleName)
		}
	}
	sort.Strings(roles)

	loginType := strings.ToLower(strings.TrimSpace(user.LoginType))
	if loginType == "" {
		loginType = models.LoginTypeCashier
	}

	if strings.EqualFold(strings.TrimSpace(user.UserType), models.UserTypeSuperAdmin) {
		return &SessionAccess{UserType: models.UserTypeSuperAdmin, LoginType: loginType, Roles: roles, Permissions: []string{"*"}}, nil
	}

	permissionSet := make(map[string]struct{})
	roleModel := models.NewRole(s.DB, *auth)
	for _, roleName := range roles {
		role, findErr := roleModel.FindByName(roleName)
		if findErr != nil {
			if errors.Is(findErr, sql.ErrNoRows) {
				continue
			}
			return nil, findErr
		}
		for _, action := range role.Actions {
			action = strings.ToLower(strings.TrimSpace(action))
			if action != "" {
				permissionSet[action] = struct{}{}
			}
		}
	}

	permissions := make([]string, 0, len(permissionSet))
	for permission := range permissionSet {
		permissions = append(permissions, permission)
	}
	sort.Strings(permissions)

	return &SessionAccess{UserType: user.UserType, LoginType: loginType, Roles: roles, Permissions: permissions}, nil
}

func isCustomerTableUser(user *models.User) bool {
	if user == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(user.LoginType), models.LoginTypeCustomerTable)
}

func (s *AuthService) GenerateLoginOTP(identifier string) (int64, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return 0, errors.New("email atau mobilephone wajib diisi")
	}

	user := models.NewUser(s.DB, models.Auth{})
	userRecord, err := user.FindByLoginIdentifierGlobal(identifier)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no rows") {
			return 0, errors.New("user tidak ditemukan")
		}
		return 0, err
	}

	if !userRecord.IsActive {
		return 0, errors.New("user tidak aktif")
	}

	otp, err := s.generateOTP()
	if err != nil {
		return 0, err
	}

	expired := time.Now().UTC().Add(10 * time.Minute)
	authRecord := s.newAuthMModel()
	authRecord.AccountID = userRecord.AccountID
	authRecord.Email = userRecord.UserEmail
	authRecord.Name = userRecord.Username
	authRecord.OTP = otp
	authRecord.OTPExpired = &expired

	id, err := authRecord.UpsertOTP()
	if err != nil {
		return 0, err
	}

	s.sendLoginOTPEmailAsync(models.Auth{
		AccountID: userRecord.AccountID,
		Email:     userRecord.UserEmail,
		Name:      userRecord.Username,
	}, otp, expired)

	return id, nil
}

func (s *AuthService) VerifyLoginOTP(identifier, otp string) (*models.Auth, error) {
	identifier = strings.TrimSpace(identifier)
	otp = strings.TrimSpace(otp)
	if identifier == "" {
		return nil, errors.New("email atau mobilephone wajib diisi")
	}
	if otp == "" {
		return nil, errors.New("otp wajib diisi")
	}

	user := models.NewUser(s.DB, models.Auth{})
	userRecord, err := user.FindByLoginIdentifierGlobal(identifier)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no rows") {
			return nil, errors.New("user tidak ditemukan")
		}
		return nil, err
	}

	auth := s.newAuthMModel()
	auth.AccountID = userRecord.AccountID
	auth.Email = userRecord.UserEmail

	authRecord, err := auth.FindByEmail()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(authRecord.OTP) == "" {
		return nil, errors.New("otp belum dibuat")
	}

	if strings.TrimSpace(authRecord.OTP) != otp {
		return nil, errors.New("otp tidak valid")
	}

	if authRecord.OTPExpired != nil && time.Now().UTC().After(*authRecord.OTPExpired) {
		return nil, errors.New("otp sudah expired")
	}

	authRecord.OTP = ""
	authRecord.OTPExpired = nil
	if err := authRecord.Update(); err != nil {
		return nil, err
	}

	return authMModelToDTO(authRecord, s.DB), nil
}

func (s *AuthService) VerifyPIN(token, pin string) (*models.Auth, error) {
	token = strings.TrimSpace(token)
	pin = strings.TrimSpace(pin)
	if token == "" {
		return nil, errors.New("token wajib diisi")
	}
	if !isSixDigitPIN(pin) {
		return nil, errors.New("pin harus 6 digit angka")
	}

	authRecord, err := s.RefreshSession(token)
	if err != nil {
		return nil, err
	}

	user := models.NewUser(s.DB, models.Auth{})
	userRecord, err := user.FindByLoginIdentifierGlobal(authRecord.Email)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no rows") {
			return nil, errors.New("user tidak ditemukan")
		}
		return nil, err
	}
	if !userRecord.IsActive {
		return nil, errors.New("user tidak aktif")
	}
	if strings.TrimSpace(userRecord.PIN) == "" {
		return nil, errors.New("pin user belum dibuat")
	}
	if !helpers.CompareCredential(userRecord.PIN, pin) {
		return nil, errors.New("pin tidak valid")
	}

	return authRecord, nil
}

func (s *AuthService) PINStatus(token string) (*models.Auth, bool, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, false, errors.New("token wajib diisi")
	}

	authRecord, err := s.RefreshSession(token)
	if err != nil {
		return nil, false, err
	}

	user := models.NewUser(s.DB, models.Auth{})
	userRecord, err := user.FindByLoginIdentifierGlobal(authRecord.Email)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no rows") {
			return nil, false, errors.New("user tidak ditemukan")
		}
		return nil, false, err
	}
	if !userRecord.IsActive {
		return nil, false, errors.New("user tidak aktif")
	}

	return authRecord, strings.TrimSpace(userRecord.PIN) != "", nil
}

func (s *AuthService) generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}

func isSixDigitPIN(pin string) bool {
	if len(pin) != 6 {
		return false
	}
	for _, char := range pin {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func (s *AuthService) mustGenerateToken() string {
	token, err := s.generateToken()
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	}

	return token
}

func (s *AuthService) generateToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (s *AuthService) sendLoginOTPEmail(record models.Auth, otp string, expired time.Time) error {
	cfg := helpers.LoadSMTPConfig()
	if !cfg.Enabled() {
		return nil
	}

	subject, textBody, htmlBody, err := helpers.BuildLoginOTPMail(record.Name, otp, expired)
	if err != nil {
		return err
	}
	message := helpers.WriteHTMLMultipartEmailMessage(cfg.FromHeader(), record.Email, subject, textBody, htmlBody)

	addr := cfg.Addr()
	var auth smtp.Auth
	if cfg.Username != "" {
		auth = cfg.Auth()
	}

	if err := smtp.SendMail(addr, auth, cfg.FromEmail, []string{record.Email}, message); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) sendLoginOTPEmailAsync(record models.Auth, otp string, expired time.Time) {
	helpers.SendEmailInBackground("login otp email", func() error {
		return s.sendLoginOTPEmail(record, otp, expired)
	})
}
