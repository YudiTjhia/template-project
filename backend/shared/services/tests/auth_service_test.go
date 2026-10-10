package services_test

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"backend/shared/connections"
	"backend/shared/models"
	"backend/shared/services"
	"github.com/jmoiron/sqlx"
)

const authServiceTestSuiteLockID int64 = 91384716

var (
	authServiceDB *sqlx.DB
	authLockConn  *sql.Conn
)

func TestMain(m *testing.M) {
	_ = loadEnvFile(".env", "../../../../.env")

	db, err := connections.NewMasterConnection()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		code := m.Run()
		os.Exit(code)
	}
	authServiceDB = db

	authLockConn, err = db.Conn(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		authServiceDB = nil
		code := m.Run()
		_ = db.Close()
		os.Exit(code)
	}
	if _, err := authLockConn.ExecContext(context.Background(), `SELECT pg_advisory_lock($1)`, authServiceTestSuiteLockID); err != nil {
		fmt.Fprintln(os.Stderr, err)
		_ = authLockConn.Close()
		_ = db.Close()
		os.Exit(1)
	}

	models.NewUser(authServiceDB, models.Auth{AccountID: 7, UserName: "alice"}).Migrate()
	models.NewAuth(authServiceDB).Migrate()
	models.NewAccountDevice(authServiceDB, models.Auth{AccountID: 7, UserName: "alice", Email: "alice@example.com"}).Migrate()
	if _, err := authServiceDB.Exec(`
		CREATE TABLE IF NOT EXISTS locations (
			id BIGSERIAL PRIMARY KEY,
			account_id BIGINT NOT NULL,
			code VARCHAR(100) NOT NULL,
			name VARCHAR(255) NOT NULL
		)`); err != nil {
		fmt.Fprintln(os.Stderr, "create auth service location test schema:", err)
		_ = db.Close()
		os.Exit(1)
	}
	_, _ = authServiceDB.Exec(`TRUNCATE TABLE users RESTART IDENTITY CASCADE`)
	_, _ = authServiceDB.Exec(`TRUNCATE TABLE auths RESTART IDENTITY CASCADE`)
	_, _ = authServiceDB.Exec(`TRUNCATE TABLE account_devices RESTART IDENTITY CASCADE`)
	_, _ = authServiceDB.Exec(`TRUNCATE TABLE locations RESTART IDENTITY CASCADE`)

	code := m.Run()

	_, _ = authLockConn.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, authServiceTestSuiteLockID)
	_ = authLockConn.Close()
	_ = db.Close()
	os.Exit(code)
}

func loadEnvFile(paths ...string) error {
	for _, name := range paths {
		path := filepath.Clean(name)
		file, err := os.Open(path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			key, value, found := strings.Cut(line, "=")
			if !found {
				continue
			}

			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			if strings.HasSuffix(key, "_DSN") {
				value = strings.ReplaceAll(value, "host.docker.internal", "127.0.0.1")
			}
			_ = os.Setenv(key, value)
		}

		_ = file.Close()
		return scanner.Err()
	}

	return nil
}

func newRealAuthService(t *testing.T) (*services.AuthService, func()) {
	t.Helper()

	if authServiceDB == nil {
		t.Skip("auth service db is not initialized")
	}

	service := services.NewAuthService(authServiceDB)
	_, _ = authServiceDB.Exec(`TRUNCATE TABLE users RESTART IDENTITY CASCADE`)
	_, _ = authServiceDB.Exec(`TRUNCATE TABLE auths RESTART IDENTITY CASCADE`)
	_, _ = authServiceDB.Exec(`TRUNCATE TABLE account_devices RESTART IDENTITY CASCADE`)
	_, _ = authServiceDB.Exec(`TRUNCATE TABLE locations RESTART IDENTITY CASCADE`)

	cleanup := func() {
		_, _ = authServiceDB.Exec(`TRUNCATE TABLE users RESTART IDENTITY CASCADE`)
		_, _ = authServiceDB.Exec(`TRUNCATE TABLE auths RESTART IDENTITY CASCADE`)
		_, _ = authServiceDB.Exec(`TRUNCATE TABLE locations RESTART IDENTITY CASCADE`)
	}

	return service, cleanup
}

func TestAuthServiceGenerateAndVerifyLoginOTP(t *testing.T) {
	service, cleanup := newRealAuthService(t)
	defer cleanup()

	user := models.NewUser(authServiceDB, models.Auth{AccountID: 7, UserName: "alice"})
	user.UserEmail = "user@example.com"
	user.MobilePhone = "0811111111"
	user.UserPassword = "secret"
	user.IsActive = true
	user.Username = "Alice"
	if _, err := user.Create(); err != nil {
		t.Fatalf("expected no error creating user, got %v", err)
	}

	id, err := service.GenerateLoginOTP("user@example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id <= 0 {
		t.Fatalf("unexpected otp id result: id=%d", id)
	}

	authRecord := models.NewAuth(authServiceDB)
	authRecord.AccountID = 7
	authRecord.Email = "user@example.com"
	found, err := authRecord.FindByEmail()
	if err != nil {
		t.Fatalf("expected auth record to exist, got %v", err)
	}
	if found == nil || len(found.OTP) != 6 {
		t.Fatalf("unexpected otp record: %+v", found)
	}

	record, err := service.VerifyLoginOTP("user@example.com", found.OTP)
	if err != nil {
		t.Fatalf("expected no error verifying otp, got %v", err)
	}
	if record == nil || record.Email != "user@example.com" {
		t.Fatalf("unexpected record: %+v", record)
	}

	id2, err := service.GenerateLoginOTP("user@example.com")
	if err != nil {
		t.Fatalf("expected no error generating second otp, got %v", err)
	}
	if id2 <= 0 {
		t.Fatalf("expected positive id, got %d", id2)
	}
}

func TestAuthServiceVerifyExpiredOTP(t *testing.T) {
	service, cleanup := newRealAuthService(t)
	defer cleanup()

	expired := time.Now().UTC().Add(-1 * time.Minute)
	auth := models.NewAuth(authServiceDB)
	auth.AccountID = 7
	auth.Email = "user@example.com"
	auth.Name = "Alice"
	auth.OTP = "123456"
	auth.OTPExpired = &expired
	if _, err := auth.Create(); err != nil {
		t.Fatalf("expected no error creating auth, got %v", err)
	}

	if _, err := service.VerifyLoginOTP("user@example.com", "123456"); err == nil {
		t.Fatalf("expected error for expired otp")
	}
}

func TestAuthServiceLoginExpiresOldSessionAndCreatesNewToken(t *testing.T) {
	service, cleanup := newRealAuthService(t)
	defer cleanup()

	user := models.NewUser(authServiceDB, models.Auth{AccountID: 7, UserName: "alice"})
	user.UserEmail = "user@example.com"
	user.MobilePhone = "0811111111"
	user.UserPassword = "secret"
	user.IsActive = true
	user.Username = "Alice"
	if _, err := user.Create(); err != nil {
		t.Fatalf("expected no error creating user, got %v", err)
	}

	expired := time.Now().UTC().Add(24 * time.Hour)
	auth := models.NewAuth(authServiceDB)
	auth.AccountID = 7
	auth.Email = "user@example.com"
	auth.Name = "Alice"
	auth.Token = "existing-token"
	auth.TokenExpired = &expired
	if _, err := auth.Create(); err != nil {
		t.Fatalf("expected no error creating auth, got %v", err)
	}

	token, err := service.Login("user@example.com", "secret", 0, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" || token == "existing-token" {
		t.Fatalf("expected a new token, got %q", token)
	}

	auth.Token = "existing-token"
	oldSession, err := auth.FindByToken()
	if err != nil {
		t.Fatalf("expected old token to still be stored, got %v", err)
	}
	if !oldSession.IsExpired {
		t.Fatalf("expected old session to be expired")
	}

	if _, err := service.RefreshSession("existing-token"); err == nil || err.Error() != "session sudah expired" {
		t.Fatalf("expected old token refresh to be rejected, got %v", err)
	}

	auth.Token = token
	newSession, err := auth.FindByToken()
	if err != nil {
		t.Fatalf("expected new token to be stored, got %v", err)
	}
	if newSession.IsExpired {
		t.Fatalf("expected new session to be active")
	}
}

func TestAuthServiceCustomerTableLoginReusesActiveSession(t *testing.T) {
	service, cleanup := newRealAuthService(t)
	defer cleanup()

	user := models.NewUser(authServiceDB, models.Auth{AccountID: 7, UserName: "alice"})
	user.UserEmail = "table@example.com"
	user.MobilePhone = "0811555010"
	user.UserPassword = "secret"
	user.PIN = "445566"
	user.IsActive = true
	user.Username = "Customer Table"
	user.UserType = models.UserTypeUser
	user.LoginType = models.LoginTypeCustomerTable
	user.UserRoles = models.StringArray{"buyer"}
	if _, err := user.Create(); err != nil {
		t.Fatalf("expected no error creating user, got %v", err)
	}

	firstToken, err := service.Login("table@example.com", "secret", 0, "table-device-1")
	if err != nil {
		t.Fatalf("expected no error on first login, got %v", err)
	}
	if strings.TrimSpace(firstToken) == "" {
		t.Fatalf("expected first token")
	}

	secondToken, err := service.LoginWithPIN("table@example.com", "445566", 0, "table-device-2")
	if err != nil {
		t.Fatalf("expected no error on second login, got %v", err)
	}
	if secondToken != firstToken {
		t.Fatalf("expected customer_table login to reuse active token, got first=%q second=%q", firstToken, secondToken)
	}

	var activeCount int
	if err := authServiceDB.Get(&activeCount, `SELECT COUNT(*) FROM auths WHERE account_id = $1 AND LOWER(email) = LOWER($2) AND is_expired = false`, 7, "table@example.com"); err != nil {
		t.Fatalf("expected no error counting sessions, got %v", err)
	}
	if activeCount != 1 {
		t.Fatalf("expected one reused active session, got %d", activeCount)
	}

	access, err := service.ResolveSessionAccess(&models.Auth{AccountID: 7, Email: "table@example.com"})
	if err != nil {
		t.Fatalf("expected no error resolving access, got %v", err)
	}
	if access.LoginType != models.LoginTypeCustomerTable {
		t.Fatalf("expected customer_table login type, got %q", access.LoginType)
	}
}

func TestAuthServiceLoginRefreshesExpiredSession(t *testing.T) {
	service, cleanup := newRealAuthService(t)
	defer cleanup()

	user := models.NewUser(authServiceDB, models.Auth{AccountID: 7, UserName: "alice"})
	user.UserEmail = "user@example.com"
	user.MobilePhone = "0811111111"
	user.UserPassword = "secret"
	user.IsActive = true
	user.Username = "Alice"
	if _, err := user.Create(); err != nil {
		t.Fatalf("expected no error creating user, got %v", err)
	}

	expired := time.Now().UTC().Add(24 * time.Hour)
	auth := models.NewAuth(authServiceDB)
	auth.AccountID = 7
	auth.Email = "user@example.com"
	auth.Name = "Alice"
	auth.Token = "expired-session-token"
	auth.TokenExpired = &expired
	auth.IsExpired = true
	if _, err := auth.Create(); err != nil {
		t.Fatalf("expected no error creating auth, got %v", err)
	}

	token, err := service.Login("user@example.com", "secret", 0, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "expired-session-token" {
		t.Fatalf("expected expired session token to be refreshed")
	}

	auth.Token = token
	found, err := auth.FindByToken()
	if err != nil {
		t.Fatalf("expected new token to be persisted, got %v", err)
	}
	if found.IsExpired {
		t.Fatalf("expected new token to be active")
	}
}

func TestAuthServiceVerifyPIN(t *testing.T) {
	service, cleanup := newRealAuthService(t)
	defer cleanup()

	user := models.NewUser(authServiceDB, models.Auth{AccountID: 7, UserName: "alice"})
	user.UserEmail = "pin@example.com"
	user.MobilePhone = "0811111111"
	user.UserPassword = "secret"
	user.PIN = "123456"
	user.IsActive = true
	user.Username = "Alice"
	if _, err := user.Create(); err != nil {
		t.Fatalf("expected no error creating user, got %v", err)
	}

	token, err := service.Login("pin@example.com", "secret", 0, "")
	if err != nil {
		t.Fatalf("expected no error logging in, got %v", err)
	}

	record, err := service.VerifyPIN(token, "123456")
	if err != nil {
		t.Fatalf("expected no error verifying pin, got %v", err)
	}
	if record == nil || record.Email != "pin@example.com" {
		t.Fatalf("unexpected auth record: %+v", record)
	}

	if _, err := service.VerifyPIN(token, "654321"); err == nil || err.Error() != "pin tidak valid" {
		t.Fatalf("expected invalid pin error, got %v", err)
	}
}

func TestAuthServiceLoginWithPIN(t *testing.T) {
	service, cleanup := newRealAuthService(t)
	defer cleanup()

	user := models.NewUser(authServiceDB, models.Auth{AccountID: 7, UserName: "alice"})
	user.UserEmail = "pin-login@example.com"
	user.MobilePhone = "0812222222"
	user.UserPassword = "secret"
	user.PIN = "654321"
	user.IsActive = true
	user.Username = "Alice"
	if _, err := user.Create(); err != nil {
		t.Fatalf("expected no error creating user, got %v", err)
	}

	token, err := service.LoginWithPIN("pin-login@example.com", "654321", 0, "")
	if err != nil {
		t.Fatalf("expected no error logging in with pin, got %v", err)
	}
	if strings.TrimSpace(token) == "" {
		t.Fatalf("expected token to be returned")
	}

	if _, err := service.LoginWithPIN("pin-login@example.com", "000000", 0, ""); err == nil || err.Error() != "email/mobilephone atau pin tidak valid" {
		t.Fatalf("expected invalid pin error, got %v", err)
	}
}

func TestAuthServiceLoginWithAccountIDAndPIN(t *testing.T) {
	service, cleanup := newRealAuthService(t)
	defer cleanup()

	user := models.NewUser(authServiceDB, models.Auth{AccountID: 77, UserName: "cashier"})
	user.UserEmail = "account-pin-login@example.com"
	user.MobilePhone = "0812333444"
	user.UserPassword = "secret"
	user.PIN = "112233"
	user.IsActive = true
	user.Username = "Cashier"
	if _, err := user.Create(); err != nil {
		t.Fatalf("expected no error creating user, got %v", err)
	}

	token, err := service.LoginWithPIN("77", "112233", 0, "")
	if err != nil {
		t.Fatalf("expected no error logging in with account id and pin, got %v", err)
	}
	if strings.TrimSpace(token) == "" {
		t.Fatalf("expected token to be returned")
	}

	if _, err := service.LoginWithPIN("77", "000000", 0, ""); err == nil || err.Error() != "account id atau pin tidak valid" {
		t.Fatalf("expected invalid account id pin error, got %v", err)
	}
}

func TestAuthServiceExpireSession(t *testing.T) {
	service, cleanup := newRealAuthService(t)
	defer cleanup()

	expired := time.Now().UTC().Add(24 * time.Hour)
	auth := models.NewAuth(authServiceDB)
	auth.AccountID = 7
	auth.Email = "user@example.com"
	auth.Name = "Alice"
	auth.Token = "logout-token"
	auth.TokenExpired = &expired
	if _, err := auth.Create(); err != nil {
		t.Fatalf("expected no error creating auth, got %v", err)
	}

	if err := service.ExpireSession("logout-token"); err != nil {
		t.Fatalf("expected no error expiring session, got %v", err)
	}

	auth.Token = "logout-token"
	found, err := auth.FindByToken()
	if err != nil {
		t.Fatalf("expected no error finding auth token, got %v", err)
	}
	if !found.IsExpired {
		t.Fatalf("expected session to be expired")
	}
}

func TestAuthServiceExpireSessionRequiresToken(t *testing.T) {
	service := services.NewAuthService(authServiceDB)

	if err := service.ExpireSession(" "); err == nil || err.Error() != "token wajib diisi" {
		t.Fatalf("expected token validation error, got %v", err)
	}
}

func TestAuthServiceRefreshSessionExtendsValidToken(t *testing.T) {
	service, cleanup := newRealAuthService(t)
	defer cleanup()

	expired := time.Now().UTC().Add(1 * time.Hour)
	auth := models.NewAuth(authServiceDB)
	auth.AccountID = 7
	auth.Email = "user@example.com"
	auth.Name = "Alice"
	auth.Token = "refresh-token"
	auth.TokenExpired = &expired
	if _, err := auth.Create(); err != nil {
		t.Fatalf("expected no error creating auth, got %v", err)
	}

	refreshed, err := service.RefreshSession("refresh-token")
	if err != nil {
		t.Fatalf("expected no error refreshing session, got %v", err)
	}
	if refreshed == nil || refreshed.Token != "refresh-token" || refreshed.TokenExpired == nil {
		t.Fatalf("unexpected refreshed auth: %+v", refreshed)
	}
	if !refreshed.TokenExpired.After(expired) {
		t.Fatalf("expected token expiry to be extended, got %v before %v", refreshed.TokenExpired, expired)
	}
}

func TestAuthServiceRefreshSessionRejectsExpiredToken(t *testing.T) {
	service, cleanup := newRealAuthService(t)
	defer cleanup()

	expired := time.Now().UTC().Add(-1 * time.Minute)
	auth := models.NewAuth(authServiceDB)
	auth.AccountID = 7
	auth.Email = "user@example.com"
	auth.Name = "Alice"
	auth.Token = "already-expired-token"
	auth.TokenExpired = &expired
	if _, err := auth.Create(); err != nil {
		t.Fatalf("expected no error creating auth, got %v", err)
	}

	if _, err := service.RefreshSession("already-expired-token"); err == nil || err.Error() != "session sudah expired" {
		t.Fatalf("expected expired session error, got %v", err)
	}
}

func TestAuthServiceLoginRefreshesExpiredToken(t *testing.T) {
	service, cleanup := newRealAuthService(t)
	defer cleanup()

	user := models.NewUser(authServiceDB, models.Auth{AccountID: 7, UserName: "alice"})
	user.UserEmail = "user@example.com"
	user.MobilePhone = "0811111111"
	user.UserPassword = "secret"
	user.IsActive = true
	user.Username = "Alice"
	if _, err := user.Create(); err != nil {
		t.Fatalf("expected no error creating user, got %v", err)
	}

	expired := time.Now().UTC().Add(-1 * time.Minute)
	auth := models.NewAuth(authServiceDB)
	auth.AccountID = 7
	auth.Email = "user@example.com"
	auth.Name = "Alice"
	auth.Token = "expired-token"
	auth.TokenExpired = &expired
	if _, err := auth.Create(); err != nil {
		t.Fatalf("expected no error creating auth, got %v", err)
	}

	token, err := service.Login("user@example.com", "secret", 0, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "expired-token" {
		t.Fatalf("expected token to be refreshed")
	}
}

func TestAuthServiceLoginCreatesNewToken(t *testing.T) {
	service, cleanup := newRealAuthService(t)
	defer cleanup()

	user := models.NewUser(authServiceDB, models.Auth{AccountID: 7, UserName: "alice"})
	user.UserEmail = "new@example.com"
	user.MobilePhone = "0811111111"
	user.UserPassword = "secret"
	user.IsActive = true
	user.Username = "Alice"
	if _, err := user.Create(); err != nil {
		t.Fatalf("expected no error creating user, got %v", err)
	}

	token, err := service.Login("new@example.com", "secret", 0, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatalf("expected non-empty token")
	}

	found, err := (models.Auth{AccountID: 7, Email: "new@example.com"}).FindByEmail()
	if err == nil {
		t.Fatalf("expected find by email without account to fail, got %+v", found)
	}
}

func TestAuthServiceLoginRejectsWrongPassword(t *testing.T) {
	service, cleanup := newRealAuthService(t)
	defer cleanup()

	user := models.NewUser(authServiceDB, models.Auth{AccountID: 7, UserName: "alice"})
	user.UserEmail = "user@example.com"
	user.MobilePhone = "0811111111"
	user.UserPassword = "secret"
	user.IsActive = true
	user.Username = "Alice"
	if _, err := user.Create(); err != nil {
		t.Fatalf("expected no error creating user, got %v", err)
	}

	if _, err := service.Login("user@example.com", "wrong", 0, ""); err == nil {
		t.Fatalf("expected error for wrong password")
	}
}

func TestAuthServiceLoginByMobilephone(t *testing.T) {
	service, cleanup := newRealAuthService(t)
	defer cleanup()

	user := models.NewUser(authServiceDB, models.Auth{AccountID: 7, UserName: "alice"})
	user.UserEmail = "user@example.com"
	user.MobilePhone = "0811111111"
	user.UserPassword = "secret"
	user.IsActive = true
	user.Username = "Alice"
	if _, err := user.Create(); err != nil {
		t.Fatalf("expected no error creating user, got %v", err)
	}

	token, err := service.Login("0811111111", "secret", 0, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatalf("expected non-empty token")
	}
}
