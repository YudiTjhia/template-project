package services_test

import (
	"testing"
	"time"

	"backend/shared/models"
	"backend/shared/services"
)

func newRealAccountService(t *testing.T) (*services.AccountService, func()) {
	t.Helper()

	if authServiceDB == nil {
		t.Skip("account service db is not initialized")
	}

	service := services.NewAccountService(authServiceDB, models.Auth{UserName: "alice"})
	_, _ = authServiceDB.Exec(`TRUNCATE TABLE users RESTART IDENTITY CASCADE`)
	_, _ = authServiceDB.Exec(`TRUNCATE TABLE accounts RESTART IDENTITY CASCADE`)

	cleanup := func() {
		_, _ = authServiceDB.Exec(`TRUNCATE TABLE users RESTART IDENTITY CASCADE`)
		_, _ = authServiceDB.Exec(`TRUNCATE TABLE accounts RESTART IDENTITY CASCADE`)
	}

	return service, cleanup
}

func TestAccountServiceRegisterAccountCreatesUserWithAccountID(t *testing.T) {
	service, cleanup := newRealAccountService(t)
	defer cleanup()

	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "")
	t.Setenv("SMTP_FROM_NAME", "")
	t.Setenv("SMTP_FROM_EMAIL", "")

	account := models.Account{
		AccountName:     "Tenant A",
		AccountEmail:    "tenant@example.com",
		Address:         "Jl. Melati 1",
		MobilePhone:     "0811111111",
		Phone:           "021222333",
		Password:        "secret123",
		ConfirmPassword: "secret123",
		PIN:             "123456",
		ConfirmPIN:      "123456",
		TermsAccepted:   true,
	}

	id, otp, _, err := service.RegisterAccount(account)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}
	if otp == "" {
		t.Fatalf("expected otp to be generated")
	}

	user := models.NewUser(authServiceDB, models.Auth{AccountID: id})
	user.UserEmail = account.AccountEmail
	found, err := user.FindByEmail()
	if err != nil {
		t.Fatalf("expected user to exist, got %v", err)
	}
	if found == nil {
		t.Fatalf("expected user record")
	}
	if found.AccountID != id {
		t.Fatalf("expected account id %d, got %d", id, found.AccountID)
	}
	if found.UserEmail != account.AccountEmail {
		t.Fatalf("expected email %q, got %q", account.AccountEmail, found.UserEmail)
	}
	if found.PIN != account.PIN {
		t.Fatalf("expected pin to be stored on user")
	}
	if found.IsActive {
		t.Fatalf("expected user to remain inactive before otp verification")
	}
}

func TestAccountServiceVerifyAccountRegistrationActivatesUser(t *testing.T) {
	service, cleanup := newRealAccountService(t)
	defer cleanup()

	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "")
	t.Setenv("SMTP_FROM_NAME", "")
	t.Setenv("SMTP_FROM_EMAIL", "")

	account := models.Account{
		AccountName:     "Tenant B",
		AccountEmail:    "tenantb@example.com",
		Address:         "Jl. Melati 2",
		MobilePhone:     "0811111112",
		Phone:           "021222334",
		Password:        "secret123",
		ConfirmPassword: "secret123",
		PIN:             "654321",
		ConfirmPIN:      "654321",
		TermsAccepted:   true,
	}

	id, otp, _, err := service.RegisterAccount(account)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	verifiedAccount, verifiedUser, err := service.VerifyAccountRegistration(otp)
	if err != nil {
		t.Fatalf("expected no error verifying otp, got %v", err)
	}
	if verifiedAccount == nil || verifiedUser == nil {
		t.Fatalf("expected verified account and user")
	}
	if verifiedAccount.ID != id {
		t.Fatalf("expected account id %d, got %d", id, verifiedAccount.ID)
	}
	if !verifiedAccount.IsActive {
		t.Fatalf("expected account to be active after verification")
	}
	if !verifiedUser.IsActive {
		t.Fatalf("expected user to be active after verification")
	}

	user := models.NewUser(authServiceDB, models.Auth{AccountID: id})
	user.ID = verifiedUser.ID
	found, err := user.FindByID()
	if err != nil {
		t.Fatalf("expected user to exist, got %v", err)
	}
	if !found.IsActive {
		t.Fatalf("expected stored user to be active after verification")
	}
}

func TestAccountServiceRegisterAccountResendsOtpForInactiveExistingEmail(t *testing.T) {
	service, cleanup := newRealAccountService(t)
	defer cleanup()

	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "")
	t.Setenv("SMTP_FROM_NAME", "")
	t.Setenv("SMTP_FROM_EMAIL", "")

	account := models.Account{
		AccountName:     "Tenant C",
		AccountEmail:    "tenantc@example.com",
		Address:         "Jl. Melati 3",
		MobilePhone:     "0811111113",
		Phone:           "021222335",
		Password:        "secret123",
		ConfirmPassword: "secret123",
		PIN:             "111222",
		ConfirmPIN:      "111222",
		TermsAccepted:   true,
	}

	id, otp, _, err := service.RegisterAccount(account)
	if err != nil {
		t.Fatalf("expected no error on first register, got %v", err)
	}
	if otp == "" {
		t.Fatalf("expected otp to be generated")
	}

	firstAccount, err := service.FindByEmail(account.AccountEmail)
	if err != nil {
		t.Fatalf("expected account to exist, got %v", err)
	}
	if firstAccount.OTPExpired == nil {
		t.Fatalf("expected otp expired timestamp")
	}

	time.Sleep(10 * time.Millisecond)

	updatedAccount := account
	updatedAccount.AccountName = "Tenant C Updated"
	updatedAccount.Address = "Jl. Melati 33"

	id2, otp2, _, err := service.RegisterAccount(updatedAccount)
	if err != nil {
		t.Fatalf("expected no error on re-register inactive account, got %v", err)
	}
	if id2 != id {
		t.Fatalf("expected same account id %d, got %d", id, id2)
	}
	if otp2 == "" {
		t.Fatalf("expected regenerated otp")
	}

	secondAccount, err := service.FindByEmail(account.AccountEmail)
	if err != nil {
		t.Fatalf("expected account to exist after re-register, got %v", err)
	}
	if secondAccount.AccountName != updatedAccount.AccountName {
		t.Fatalf("expected account name to update to %q, got %q", updatedAccount.AccountName, secondAccount.AccountName)
	}
	if secondAccount.OTPExpired == nil || !secondAccount.OTPExpired.After(*firstAccount.OTPExpired) {
		t.Fatalf("expected otp expiry to refresh")
	}
}

func TestAccountServiceRegisterAccountRejectsActiveEmail(t *testing.T) {
	service, cleanup := newRealAccountService(t)
	defer cleanup()

	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "")
	t.Setenv("SMTP_FROM_NAME", "")
	t.Setenv("SMTP_FROM_EMAIL", "")

	account := models.Account{
		AccountName:     "Tenant D",
		AccountEmail:    "tenantd@example.com",
		Address:         "Jl. Melati 4",
		MobilePhone:     "0811111114",
		Phone:           "021222336",
		Password:        "secret123",
		ConfirmPassword: "secret123",
		PIN:             "222333",
		ConfirmPIN:      "222333",
		TermsAccepted:   true,
	}

	_, otp, _, err := service.RegisterAccount(account)
	if err != nil {
		t.Fatalf("expected no error on register, got %v", err)
	}

	verifiedAccount, _, err := service.VerifyAccountRegistration(otp)
	if err != nil {
		t.Fatalf("expected no error verifying otp, got %v", err)
	}
	if verifiedAccount == nil || !verifiedAccount.IsActive {
		t.Fatalf("expected account to be active after verification")
	}

	_, _, _, err = service.RegisterAccount(account)
	if err == nil || err.Error() != "email sudah digunakan" {
		t.Fatalf("expected active email rejection, got %v", err)
	}
}

func TestAccountServiceRegisterAccountValidatesFields(t *testing.T) {
	service, cleanup := newRealAccountService(t)
	defer cleanup()

	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "")
	t.Setenv("SMTP_FROM_NAME", "")
	t.Setenv("SMTP_FROM_EMAIL", "")

	tests := []struct {
		name    string
		account models.Account
		wantErr string
	}{
		{
			name: "missing account name",
			account: models.Account{
				AccountEmail:    "tenant@example.com",
				Address:         "Jl. Melati 1",
				MobilePhone:     "0811111111",
				Phone:           "021222333",
				Password:        "secret123",
				ConfirmPassword: "secret123",
			},
			wantErr: "account_name wajib diisi",
		},
		{
			name: "password too short",
			account: models.Account{
				AccountName:     "Tenant A",
				AccountEmail:    "tenant@example.com",
				Address:         "Jl. Melati 1",
				MobilePhone:     "0811111111",
				Phone:           "021222333",
				Password:        "12345",
				ConfirmPassword: "12345",
			},
			wantErr: "password minimal 6 karakter",
		},
		{
			name: "confirm password mismatch",
			account: models.Account{
				AccountName:     "Tenant A",
				AccountEmail:    "tenant@example.com",
				Address:         "Jl. Melati 1",
				MobilePhone:     "0811111111",
				Phone:           "021222333",
				Password:        "secret123",
				ConfirmPassword: "secret124",
			},
			wantErr: "confirm password tidak cocok",
		},
		{
			name: "phone optional",
			account: models.Account{
				AccountName:     "Tenant A",
				AccountEmail:    "tenant@example.com",
				Address:         "Jl. Melati 1",
				MobilePhone:     "0811111111",
				Password:        "secret123",
				ConfirmPassword: "secret123",
				PIN:             "123456",
				ConfirmPIN:      "123456",
				TermsAccepted:   true,
			},
		},
		{
			name: "missing terms acceptance",
			account: models.Account{
				AccountName:     "Tenant A",
				AccountEmail:    "tenant@example.com",
				Address:         "Jl. Melati 1",
				MobilePhone:     "0811111111",
				Password:        "secret123",
				ConfirmPassword: "secret123",
				PIN:             "123456",
				ConfirmPIN:      "123456",
			},
			wantErr: "syarat dan ketentuan wajib disetujui",
		},
		{
			name: "missing pin",
			account: models.Account{
				AccountName:     "Tenant A",
				AccountEmail:    "tenant@example.com",
				Address:         "Jl. Melati 1",
				MobilePhone:     "0811111111",
				Password:        "secret123",
				ConfirmPassword: "secret123",
			},
			wantErr: "pin wajib diisi",
		},
		{
			name: "invalid pin",
			account: models.Account{
				AccountName:     "Tenant A",
				AccountEmail:    "tenant@example.com",
				Address:         "Jl. Melati 1",
				MobilePhone:     "0811111111",
				Password:        "secret123",
				ConfirmPassword: "secret123",
				PIN:             "12ab56",
				ConfirmPIN:      "12ab56",
			},
			wantErr: "pin harus 6 digit angka",
		},
		{
			name: "confirm pin mismatch",
			account: models.Account{
				AccountName:     "Tenant A",
				AccountEmail:    "tenant@example.com",
				Address:         "Jl. Melati 1",
				MobilePhone:     "0811111111",
				Password:        "secret123",
				ConfirmPassword: "secret123",
				PIN:             "123456",
				ConfirmPIN:      "654321",
			},
			wantErr: "confirm pin tidak cocok",
		},
		{
			name: "invalid email",
			account: models.Account{
				AccountName:     "Tenant A",
				AccountEmail:    "invalid-email",
				Address:         "Jl. Melati 1",
				MobilePhone:     "0811111111",
				Phone:           "021222333",
				Password:        "secret123",
				ConfirmPassword: "secret123",
			},
			wantErr: "account_email tidak valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := service.RegisterAccount(tt.account)
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
