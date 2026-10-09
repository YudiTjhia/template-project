package models_test

import (
	"testing"
	"time"

	"backend/shared/models"
)

func newRealAuthDB(t *testing.T) (*models.Auth, func()) {
	t.Helper()

	if accountDB == nil {
		t.Skip("auth db is not initialized")
	}

	auth := models.NewAuth(accountDB)
	_, _ = accountDB.Exec(`TRUNCATE TABLE auths RESTART IDENTITY CASCADE`)

	cleanup := func() {
		_, _ = accountDB.Exec(`TRUNCATE TABLE auths RESTART IDENTITY CASCADE`)
	}

	return &auth, cleanup
}

func TestAuthTableName(t *testing.T) {
	var auth models.Auth
	if got := auth.TableName(); got != "auths" {
		t.Fatalf("expected table auths, got %q", got)
	}
}

func TestAuthValidateCreate(t *testing.T) {
	tests := []struct {
		name    string
		auth    models.Auth
		wantErr string
	}{
		{
			name:    "missing account",
			auth:    models.Auth{Email: "user@example.com", Name: "Alice"},
			wantErr: "id akun wajib diisi",
		},
		{
			name:    "missing email",
			auth:    models.Auth{AccountID: 1, Name: "Alice"},
			wantErr: "email wajib diisi",
		},
		{
			name:    "missing name",
			auth:    models.Auth{AccountID: 1, Email: "user@example.com"},
			wantErr: "nama wajib diisi",
		},
		{
			name: "valid",
			auth: models.Auth{
				AccountID: 1,
				Email:     "user@example.com",
				Name:      "Alice",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.auth.ValidateCreate()
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

func TestAuthCreateFindUpdateDelete(t *testing.T) {
	auth, cleanup := newRealAuthDB(t)
	defer cleanup()

	expired := time.Now().UTC().Add(10 * time.Minute)
	auth.AccountID = 7
	auth.Email = "user@example.com"
	auth.Name = "Alice"
	auth.OTP = "123456"
	auth.OTPExpired = &expired

	id, err := auth.Create()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	auth.ID = id
	found, err := auth.FindByID()
	if err != nil {
		t.Fatalf("expected no error when reading back, got %v", err)
	}
	if found == nil || found.Email != "user@example.com" || found.OTP != "123456" {
		t.Fatalf("unexpected auth: %+v", found)
	}
	if found.IsExpired {
		t.Fatalf("expected new auth session to be active")
	}

	auth.Email = "USER@example.com"
	found, err = auth.FindByEmail()
	if err != nil {
		t.Fatalf("expected no error when finding by email, got %v", err)
	}
	if found == nil || found.Name != "Alice" {
		t.Fatalf("unexpected auth by email: %+v", found)
	}

	auth.Name = "Alice Updated"
	auth.OTP = "654321"
	auth.Token = "active-token"
	if err := auth.Update(); err != nil {
		t.Fatalf("expected no error updating, got %v", err)
	}

	found, err = auth.FindByID()
	if err != nil {
		t.Fatalf("expected no error after update, got %v", err)
	}
	if found.Name != "Alice Updated" || found.OTP != "654321" {
		t.Fatalf("unexpected updated auth: %+v", found)
	}

	auth.Token = "active-token"
	if err := auth.ExpireByToken(); err != nil {
		t.Fatalf("expected no error expiring token, got %v", err)
	}

	found, err = auth.FindByToken()
	if err != nil {
		t.Fatalf("expected no error finding by token, got %v", err)
	}
	if !found.IsExpired {
		t.Fatalf("expected token to be expired")
	}

	if err := auth.Delete(); err != nil {
		t.Fatalf("expected no error deleting, got %v", err)
	}
	if _, err := auth.FindByID(); err == nil {
		t.Fatalf("expected error after delete")
	}
}

func TestAuthExpireActiveByEmailAllowsNewActiveSession(t *testing.T) {
	auth, cleanup := newRealAuthDB(t)
	defer cleanup()

	expired := time.Now().UTC().Add(1 * time.Hour)
	auth.AccountID = 7
	auth.Email = "USER@example.com"
	auth.Name = "Alice"
	auth.Token = "old-token"
	auth.TokenExpired = &expired
	if _, err := auth.Create(); err != nil {
		t.Fatalf("expected no error creating first auth, got %v", err)
	}

	auth.Email = "user@example.com"
	if err := auth.ExpireActiveByEmail(); err != nil {
		t.Fatalf("expected no error expiring active auth by email, got %v", err)
	}

	auth.Token = "old-token"
	oldSession, err := auth.FindByToken()
	if err != nil {
		t.Fatalf("expected old token to be found, got %v", err)
	}
	if !oldSession.IsExpired {
		t.Fatalf("expected old session to be expired")
	}

	nextExpired := time.Now().UTC().Add(24 * time.Hour)
	newAuth := models.NewAuth(accountDB)
	newAuth.AccountID = 7
	newAuth.Email = "user@example.com"
	newAuth.Name = "Alice"
	newAuth.Token = "new-token"
	newAuth.TokenExpired = &nextExpired
	if _, err := newAuth.Create(); err != nil {
		t.Fatalf("expected new active auth to be allowed, got %v", err)
	}

	found, err := newAuth.FindByEmail()
	if err != nil {
		t.Fatalf("expected latest active auth by email, got %v", err)
	}
	if found.Token != "new-token" || found.IsExpired {
		t.Fatalf("expected active new token, got %+v", found)
	}
}
