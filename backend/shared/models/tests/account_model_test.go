package models_test

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
	"github.com/jmoiron/sqlx"
)

const testSuiteLockID int64 = 91384713

var (
	accountDB *sqlx.DB
	lockConn  *sql.Conn
)

func TestMain(m *testing.M) {
	_ = loadEnvFile(".env", "../../../.env")

	db, err := connections.NewMasterConnection()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		code := m.Run()
		os.Exit(code)
	}
	accountDB = db

	lockConn, err = db.Conn(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		accountDB = nil
		code := m.Run()
		_ = db.Close()
		os.Exit(code)
	}
	if _, err := lockConn.ExecContext(context.Background(), `SELECT pg_advisory_lock($1)`, testSuiteLockID); err != nil {
		fmt.Fprintln(os.Stderr, err)
		_ = lockConn.Close()
		_ = db.Close()
		os.Exit(1)
	}

	auth := models.Auth{UserName: "alice"}
	models.NewAccount(accountDB, auth).Migrate()
	models.NewUser(accountDB, models.Auth{AccountID: 7, UserName: "alice"}).Migrate()
	models.NewRole(accountDB, models.Auth{AccountID: 7, UserName: "alice", Email: "alice@example.com"}).Migrate()
	models.NewAuth(accountDB).Migrate()
	models.NewSupplier(accountDB, models.Auth{AccountID: 7, UserName: "alice", Email: "alice@example.com"}).Migrate()
	_, _ = accountDB.Exec(`TRUNCATE TABLE accounts RESTART IDENTITY CASCADE`)
	_, _ = accountDB.Exec(`TRUNCATE TABLE users RESTART IDENTITY CASCADE`)
	_, _ = accountDB.Exec(`TRUNCATE TABLE roles RESTART IDENTITY CASCADE`)
	_, _ = accountDB.Exec(`TRUNCATE TABLE auths RESTART IDENTITY CASCADE`)
	_, _ = accountDB.Exec(`TRUNCATE TABLE suppliers RESTART IDENTITY CASCADE`)

	code := m.Run()

	_, _ = lockConn.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, testSuiteLockID)
	_ = lockConn.Close()
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

func newRealAccountDB(t *testing.T) (*models.Account, func()) {
	t.Helper()

	if accountDB == nil {
		t.Skip("account db is not initialized")
	}

	auth := models.Auth{UserName: "alice"}
	account := models.NewAccount(accountDB, auth)
	_, _ = accountDB.Exec(`TRUNCATE TABLE accounts RESTART IDENTITY CASCADE`)

	cleanup := func() {
		_, _ = accountDB.Exec(`TRUNCATE TABLE accounts RESTART IDENTITY CASCADE`)
	}

	return &account, cleanup
}

func TestNewAccount(t *testing.T) {
	account, cleanup := newRealAccountDB(t)
	defer cleanup()

	if account.DB == nil {
		t.Fatalf("expected db to be set")
	}
	if account.CreatedBy != "alice" {
		t.Fatalf("expected CreatedBy alice, got %q", account.CreatedBy)
	}
	if !account.IsActive {
		t.Fatalf("expected IsActive true")
	}
}

func TestAccountTableName(t *testing.T) {
	var account models.Account
	if got := account.TableName(); got != "accounts" {
		t.Fatalf("expected table accounts, got %q", got)
	}
}

func TestAccountValidateCreate(t *testing.T) {
	tests := []struct {
		name    string
		account models.Account
		wantErr string
	}{
		{
			name:    "missing name",
			account: models.Account{AccountEmail: "tenant@example.com"},
			wantErr: "nama account wajib diisi",
		},
		{
			name:    "missing email",
			account: models.Account{AccountName: "Tenant A"},
			wantErr: "email account wajib diisi",
		},
		{
			name:    "missing mobilephone",
			account: models.Account{AccountName: "Tenant A", AccountEmail: "tenant@example.com"},
			wantErr: "mobilephone account wajib diisi",
		},
		{
			name: "valid",
			account: models.Account{
				AccountName:  "Tenant A",
				AccountEmail: "tenant@example.com",
				MobilePhone:  "0811111111",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.account.ValidateCreate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}

			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("expected %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestAccountValidateID(t *testing.T) {
	if err := (models.Account{}).ValidateID(); err == nil {
		t.Fatalf("expected error for empty id")
	}

	if err := (models.Account{ID: 10}).ValidateID(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestAccountOtpFields(t *testing.T) {
	expired := time.Now().Add(15 * time.Minute)
	account := models.Account{
		AccountName:  "Tenant A",
		AccountEmail: "tenant@example.com",
		OTP:          "123456",
		OTPExpired:   &expired,
	}

	if account.OTP != "123456" {
		t.Fatalf("expected otp to be set")
	}
	if account.OTPExpired == nil {
		t.Fatalf("expected otp_expired to be set")
	}
}

func TestAccountCreate(t *testing.T) {
	account, cleanup := newRealAccountDB(t)
	defer cleanup()

	account.AccountName = "Tenant A"
	account.IsActive = true
	account.AccountEmail = "tenant@example.com"
	account.Address = "Jl. Contoh 1"
	account.MobilePhone = "0811111111"
	account.Phone = "0822222222"

	id, err := account.Create()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	account.ID = id
	found, err := account.FindByID()
	if err != nil {
		t.Fatalf("expected no error when reading back, got %v", err)
	}
	if found == nil || found.AccountName != "Tenant A" || found.AccountEmail != "tenant@example.com" {
		t.Fatalf("unexpected account: %+v", found)
	}
}

func TestAccountEmailUnique(t *testing.T) {
	account, cleanup := newRealAccountDB(t)
	defer cleanup()

	account.AccountName = "Tenant A"
	account.IsActive = true
	account.AccountEmail = "tenant@example.com"
	account.Address = "Jl. Contoh 1"
	account.MobilePhone = "0811111111"
	account.Phone = "0822222222"

	if _, err := account.Create(); err != nil {
		t.Fatalf("expected first insert to succeed, got %v", err)
	}

	account.AccountName = "Tenant B"
	account.AccountEmail = "TENANT@example.com"
	if _, err := account.Create(); err == nil || err.Error() != "email account sudah digunakan" {
		t.Fatalf("expected duplicate email error, got %v", err)
	}
}

func TestAccountMobilephoneUnique(t *testing.T) {
	account, cleanup := newRealAccountDB(t)
	defer cleanup()

	account.AccountName = "Tenant A"
	account.IsActive = true
	account.AccountEmail = "tenant@example.com"
	account.MobilePhone = "0811111111"
	account.Phone = "0822222222"

	if _, err := account.Create(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	account.AccountName = "Tenant B"
	account.AccountEmail = "tenant2@example.com"
	account.MobilePhone = "0811111111"
	if _, err := account.Create(); err == nil || err.Error() != "mobilephone account sudah digunakan" {
		t.Fatalf("expected duplicate mobilephone error, got %v", err)
	}
}

func TestAccountUpdate(t *testing.T) {
	account, cleanup := newRealAccountDB(t)
	defer cleanup()

	account.AccountName = "Tenant A"
	account.IsActive = true
	account.AccountEmail = "tenant@example.com"
	account.Address = "Jl. Contoh 1"
	account.MobilePhone = "0811111111"
	account.Phone = "0822222222"

	id, err := account.Create()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updatedBy := "bob"
	account.ID = id
	account.AccountName = "Tenant B"
	account.IsActive = false
	account.AccountEmail = "tenant2@example.com"
	account.Address = "Jl. Contoh 2"
	account.MobilePhone = "0833333333"
	account.Phone = "0844444444"
	account.UpdatedBy = &updatedBy

	if err := account.Update(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	found, err := account.FindByID()
	if err != nil {
		t.Fatalf("expected no error when reading back, got %v", err)
	}
	if found.AccountName != "Tenant B" || found.AccountEmail != "tenant2@example.com" || found.IsActive {
		t.Fatalf("unexpected updated account: %+v", found)
	}
}

func TestAccountUpdateEmailUnique(t *testing.T) {
	account, cleanup := newRealAccountDB(t)
	defer cleanup()

	account.AccountName = "Tenant A"
	account.IsActive = true
	account.AccountEmail = "tenant@example.com"
	account.Address = "Jl. Contoh 1"
	account.MobilePhone = "0811111111"
	account.Phone = "0822222222"

	firstID, err := account.Create()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	account.AccountName = "Tenant B"
	account.AccountEmail = "tenant2@example.com"
	secondID, err := account.Create()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	account.ID = secondID
	account.AccountName = "Tenant B"
	account.AccountEmail = "TENANT@example.com"
	account.IsActive = false
	if err := account.Update(); err == nil || err.Error() != "email account sudah digunakan" {
		t.Fatalf("expected duplicate email error on update, got %v", err)
	}

	account.ID = firstID
	account.AccountEmail = "tenant@example.com"
	if err := account.Update(); err != nil {
		t.Fatalf("expected same email update to succeed, got %v", err)
	}
}

func TestAccountDelete(t *testing.T) {
	account, cleanup := newRealAccountDB(t)
	defer cleanup()

	account.AccountName = "Tenant A"
	account.IsActive = true
	account.AccountEmail = "tenant@example.com"
	account.Address = "Jl. Contoh 1"
	account.MobilePhone = "0811111111"
	account.Phone = "0822222222"

	id, err := account.Create()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	account.ID = id
	if err := account.Delete(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = account.FindByID()
	if err == nil {
		t.Fatalf("expected error after delete")
	}
}

func TestAccountFindByEmail(t *testing.T) {
	account, cleanup := newRealAccountDB(t)
	defer cleanup()

	account.AccountName = "Tenant A"
	account.IsActive = true
	account.AccountEmail = "tenant@example.com"
	account.Address = "Jl. Contoh 1"
	account.MobilePhone = "0811111111"
	account.Phone = "0822222222"

	if _, err := account.Create(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	account.AccountEmail = "tenant@example.com"
	found, err := account.FindByEmail()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found == nil || found.AccountEmail != "tenant@example.com" {
		t.Fatalf("unexpected account: %+v", found)
	}
}

func TestAccountFindList(t *testing.T) {
	account, cleanup := newRealAccountDB(t)
	defer cleanup()

	rows := []struct {
		name  string
		email string
	}{
		{"Tenant A", "tenant@example.com"},
		{"Tenant B", "tenant2@example.com"},
	}

	for _, row := range rows {
		account.AccountName = row.name
		account.AccountEmail = row.email
		account.IsActive = true
		if _, err := account.Create(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}

	items, err := account.FindList("tenant")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("unexpected items: %+v", items)
	}
}

func TestAccountFindRecentList(t *testing.T) {
	account, cleanup := newRealAccountDB(t)
	defer cleanup()

	for i := 1; i <= 3; i++ {
		account.AccountName = fmt.Sprintf("Tenant %d", i)
		account.AccountEmail = fmt.Sprintf("tenant%d@example.com", i)
		account.Address = fmt.Sprintf("Jl. Contoh %d", i)
		account.MobilePhone = fmt.Sprintf("08111%d", i)
		account.Phone = fmt.Sprintf("08222%d", i)
		account.IsActive = true
		if _, err := account.Create(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	items, err := account.FindRecentList(2, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].AccountName != "Tenant 3" {
		t.Fatalf("expected most recent account name Tenant 3, got %q", items[0].AccountName)
	}
}
