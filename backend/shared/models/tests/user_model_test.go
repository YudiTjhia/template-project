package models_test

import (
	"fmt"
	"testing"
	"time"

	"backend/shared/models"
)

func newRealUserDB(t *testing.T) (*models.User, func()) {
	t.Helper()

	if accountDB == nil {
		t.Skip("user db is not initialized")
	}

	auth := models.Auth{AccountID: 7, UserName: "alice"}
	user := models.NewUser(accountDB, auth)
	_, _ = accountDB.Exec(`TRUNCATE TABLE users RESTART IDENTITY CASCADE`)

	cleanup := func() {
		_, _ = accountDB.Exec(`TRUNCATE TABLE users RESTART IDENTITY CASCADE`)
	}

	return &user, cleanup
}

func TestNewUser(t *testing.T) {
	user, cleanup := newRealUserDB(t)
	defer cleanup()

	if user.DB == nil {
		t.Fatalf("expected db to be set")
	}
	if user.AccountID != 7 {
		t.Fatalf("expected AccountID 7, got %d", user.AccountID)
	}
	if user.CreatedBy != "alice" {
		t.Fatalf("expected CreatedBy alice, got %q", user.CreatedBy)
	}
	if !user.IsActive {
		t.Fatalf("expected IsActive true")
	}
	if user.UserType != models.UserTypeSuperAdmin {
		t.Fatalf("expected UserType SUPER_ADMIN, got %q", user.UserType)
	}
	if len(user.UserRoles) != 0 {
		t.Fatalf("expected empty UserRoles, got %#v", user.UserRoles)
	}
}

func TestUserTableName(t *testing.T) {
	var user models.User
	if got := user.TableName(); got != "users" {
		t.Fatalf("expected table users, got %q", got)
	}
}

func TestUserValidateCreate(t *testing.T) {
	tests := []struct {
		name    string
		user    models.User
		wantErr string
	}{
		{
			name:    "missing account",
			user:    models.User{UserEmail: "user@example.com", UserPassword: "secret", Username: "alice"},
			wantErr: "id akun wajib diisi",
		},
		{
			name:    "missing email",
			user:    models.User{BaseModel: models.BaseModel{Auth: models.Auth{AccountID: 1}}, MobilePhone: "0811111111", UserPassword: "secret", Username: "alice"},
			wantErr: "email user wajib diisi",
		},
		{
			name:    "missing mobilephone",
			user:    models.User{BaseModel: models.BaseModel{Auth: models.Auth{AccountID: 1}}, UserEmail: "user@example.com", UserPassword: "secret", Username: "alice"},
			wantErr: "mobilephone user wajib diisi",
		},
		{
			name:    "missing password",
			user:    models.User{BaseModel: models.BaseModel{Auth: models.Auth{AccountID: 1}}, UserEmail: "user@example.com", MobilePhone: "0811111111", Username: "alice"},
			wantErr: "password user wajib diisi",
		},
		{
			name:    "missing username",
			user:    models.User{BaseModel: models.BaseModel{Auth: models.Auth{AccountID: 1}}, UserEmail: "user@example.com", MobilePhone: "0811111111", UserPassword: "secret"},
			wantErr: "username user wajib diisi",
		},
		{
			name:    "invalid user type",
			user:    models.User{BaseModel: models.BaseModel{Auth: models.Auth{AccountID: 1}}, UserEmail: "user@example.com", MobilePhone: "0811111111", UserPassword: "secret", Username: "alice", UserType: "GUEST"},
			wantErr: "user_type user tidak valid",
		},
		{
			name:    "user without roles",
			user:    models.User{BaseModel: models.BaseModel{Auth: models.Auth{AccountID: 1}}, UserEmail: "user@example.com", MobilePhone: "0811111111", UserPassword: "secret", Username: "alice", UserType: models.UserTypeUser},
			wantErr: "user_roles wajib minimal 1 untuk user_type USER",
		},
		{
			name:    "invalid pin",
			user:    models.User{BaseModel: models.BaseModel{Auth: models.Auth{AccountID: 1}}, UserEmail: "user@example.com", MobilePhone: "0811111111", UserPassword: "secret", Username: "alice", PIN: "12ab56"},
			wantErr: "pin harus 6 digit angka",
		},
		{
			name:    "confirm pin mismatch",
			user:    models.User{BaseModel: models.BaseModel{Auth: models.Auth{AccountID: 1}}, UserEmail: "user@example.com", MobilePhone: "0811111111", UserPassword: "secret", Username: "alice", PIN: "123456", ConfirmPIN: "654321"},
			wantErr: "confirm pin tidak cocok",
		},
		{
			name: "valid",
			user: models.User{
				BaseModel:    models.BaseModel{Auth: models.Auth{AccountID: 1}},
				UserEmail:    "user@example.com",
				MobilePhone:  "0811111111",
				UserPassword: "secret",
				PIN:          "123456",
				ConfirmPIN:   "123456",
				UserType:     models.UserTypeUser,
				UserRoles:    models.StringArray{"cashier"},
				Username:     "alice",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.ValidateCreate()
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

func TestUserValidateID(t *testing.T) {
	if err := (models.User{}).ValidateID(); err == nil {
		t.Fatalf("expected error for empty id")
	}

	if err := (models.User{ID: 10}).ValidateID(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUserCreate(t *testing.T) {
	user, cleanup := newRealUserDB(t)
	defer cleanup()

	user.UserEmail = "user@example.com"
	user.MobilePhone = "0811111111"
	user.UserPassword = "secret"
	user.IsActive = true
	user.Username = "alice"

	id, err := user.Create()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	user.ID = id
	found, err := user.FindByID()
	if err != nil {
		t.Fatalf("expected no error when reading back, got %v", err)
	}
	if found == nil || found.UserEmail != "user@example.com" || found.Username != "alice" {
		t.Fatalf("unexpected user: %+v", found)
	}
}

func TestUserEmailUnique(t *testing.T) {
	user, cleanup := newRealUserDB(t)
	defer cleanup()

	user.UserEmail = "user@example.com"
	user.MobilePhone = "0811111111"
	user.UserPassword = "secret"
	user.IsActive = true
	user.Username = "alice"

	if _, err := user.Create(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	otherUser := models.NewUser(accountDB, models.Auth{AccountID: 8, UserName: "bob"})
	otherUser.UserEmail = "USER@example.com"
	otherUser.MobilePhone = "0822222222"
	otherUser.UserPassword = "secret2"
	otherUser.Username = "bob"
	if _, err := otherUser.Create(); err == nil || err.Error() != "email user sudah digunakan" {
		t.Fatalf("expected duplicate email error, got %v", err)
	}
}

func TestUserMobilephoneUnique(t *testing.T) {
	user, cleanup := newRealUserDB(t)
	defer cleanup()

	user.UserEmail = "user@example.com"
	user.MobilePhone = "0811111111"
	user.UserPassword = "secret"
	user.IsActive = true
	user.Username = "alice"

	if _, err := user.Create(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	otherUser := models.NewUser(accountDB, models.Auth{AccountID: 8, UserName: "bob"})
	otherUser.UserEmail = "user2@example.com"
	otherUser.MobilePhone = "0811111111"
	otherUser.UserPassword = "secret2"
	otherUser.Username = "bob"
	if _, err := otherUser.Create(); err == nil || err.Error() != "mobilephone user sudah digunakan" {
		t.Fatalf("expected duplicate mobilephone error, got %v", err)
	}
}

func TestUserDeleteBlocksSuperAdmin(t *testing.T) {
	user, cleanup := newRealUserDB(t)
	defer cleanup()

	user.UserEmail = "superadmin@example.com"
	user.MobilePhone = "0811111119"
	user.UserPassword = "secret"
	user.IsActive = true
	user.Username = "superadmin"

	id, err := user.Create()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	user.ID = id
	if err := user.Delete(); err == nil || err.Error() != "user super admin tidak bisa dihapus" {
		t.Fatalf("expected super admin delete to be blocked, got %v", err)
	}
}

func TestUserUpdate(t *testing.T) {
	user, cleanup := newRealUserDB(t)
	defer cleanup()

	user.UserEmail = "user@example.com"
	user.MobilePhone = "0811111111"
	user.UserPassword = "secret"
	user.IsActive = true
	user.Username = "alice"

	id, err := user.Create()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updatedBy := "bob"
	user.ID = id
	user.UserEmail = "user2@example.com"
	user.MobilePhone = "0822222222"
	user.UserPassword = "secret2"
	user.IsActive = false
	user.Username = "bob"
	user.UpdatedBy = &updatedBy

	if err := user.Update(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	found, err := user.FindByID()
	if err != nil {
		t.Fatalf("expected no error when reading back, got %v", err)
	}
	if found.UserEmail != "user2@example.com" || found.Username != "bob" || found.IsActive {
		t.Fatalf("unexpected updated user: %+v", found)
	}
}

func TestUserUpdateKeepsPasswordWhenEmpty(t *testing.T) {
	user, cleanup := newRealUserDB(t)
	defer cleanup()

	user.UserEmail = "user@example.com"
	user.MobilePhone = "0811111111"
	user.UserPassword = "secret"
	user.IsActive = true
	user.Username = "alice"

	id, err := user.Create()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updatedBy := "bob"
	user.ID = id
	user.UserEmail = "user2@example.com"
	user.MobilePhone = "0822222222"
	user.UserPassword = ""
	user.IsActive = false
	user.Username = "bob"
	user.UpdatedBy = &updatedBy

	if err := user.Update(); err != nil {
		t.Fatalf("expected no error updating with empty password, got %v", err)
	}

	found, err := user.FindByID()
	if err != nil {
		t.Fatalf("expected no error when reading back, got %v", err)
	}
	if found.UserPassword != "secret" {
		t.Fatalf("expected existing password to stay unchanged, got %q", found.UserPassword)
	}
}

func TestUserDelete(t *testing.T) {
	user, cleanup := newRealUserDB(t)
	defer cleanup()

	user.UserEmail = "user@example.com"
	user.MobilePhone = "0811111111"
	user.UserPassword = "secret"
	user.IsActive = true
	user.Username = "alice"

	id, err := user.Create()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	user.ID = id
	if err := user.Delete(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = user.FindByID()
	if err == nil {
		t.Fatalf("expected error after delete")
	}
}

func TestUserFindByEmail(t *testing.T) {
	user, cleanup := newRealUserDB(t)
	defer cleanup()

	user.UserEmail = "user@example.com"
	user.MobilePhone = "0811111111"
	user.UserPassword = "secret"
	user.IsActive = true
	user.Username = "alice"

	if _, err := user.Create(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	user.UserEmail = "USER@example.com"
	user.MobilePhone = "0811111111"
	found, err := user.FindByEmail()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found == nil || found.UserEmail != "user@example.com" {
		t.Fatalf("unexpected user: %+v", found)
	}
}

func TestUserFindByUsername(t *testing.T) {
	user, cleanup := newRealUserDB(t)
	defer cleanup()

	user.UserEmail = "user@example.com"
	user.MobilePhone = "0811111111"
	user.UserPassword = "secret"
	user.IsActive = true
	user.Username = "alice"

	if _, err := user.Create(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	user.Username = "alice"
	found, err := user.FindByUsername()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found == nil || found.Username != "alice" {
		t.Fatalf("unexpected user: %+v", found)
	}
}

func TestUserFindList(t *testing.T) {
	user, cleanup := newRealUserDB(t)
	defer cleanup()

	rows := []struct {
		email string
		name  string
	}{
		{"user1@example.com", "alice"},
		{"user2@example.com", "bob"},
	}

	for _, row := range rows {
		user.UserEmail = row.email
		user.MobilePhone = "0811111111" + row.name
		user.UserPassword = "secret"
		user.IsActive = true
		user.Username = row.name
		if _, err := user.Create(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}

	items, err := user.FindList("alice")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 1 || items[0].Username != "alice" {
		t.Fatalf("unexpected items: %+v", items)
	}
}

func TestUserFindRecentList(t *testing.T) {
	user, cleanup := newRealUserDB(t)
	defer cleanup()

	for i := 1; i <= 3; i++ {
		user.UserEmail = fmt.Sprintf("user%d@example.com", i)
		user.MobilePhone = fmt.Sprintf("08111111%d", i)
		user.UserPassword = "secret"
		user.IsActive = true
		user.Username = fmt.Sprintf("user-%d", i)
		if _, err := user.Create(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	items, err := user.FindRecentList(2, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Username != "user-3" {
		t.Fatalf("expected most recent username user-3, got %q", items[0].Username)
	}
}
