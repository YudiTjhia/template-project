package services_test

import (
	"testing"
	"time"

	"backend/shared/models"
	"backend/shared/services"
)

func newRealUserService(t *testing.T, auth models.Auth) (*services.UserService, func()) {
	t.Helper()

	if authServiceDB == nil {
		t.Skip("user service db is not initialized")
	}

	models.NewUser(authServiceDB, auth).Migrate()
	service := services.NewUserService(authServiceDB, auth)
	_, _ = authServiceDB.Exec(`TRUNCATE TABLE users RESTART IDENTITY CASCADE`)

	cleanup := func() {
		_, _ = authServiceDB.Exec(`TRUNCATE TABLE users RESTART IDENTITY CASCADE`)
	}

	return service, cleanup
}

func TestUserServiceCRUD(t *testing.T) {
	service, cleanup := newRealUserService(t, models.Auth{AccountID: 7, UserName: "alice"})
	defer cleanup()

	id, err := service.Create(models.User{
		UserEmail:    "user@example.com",
		MobilePhone:  "0811111111",
		UserPassword: "secret",
		UserType:     models.UserTypeUser,
		UserRoles:    models.StringArray{"buyer"},
		IsActive:     true,
		Username:     "alice",
	})
	if err != nil {
		t.Fatalf("expected no error creating user, got %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	found, err := service.FindByID(id)
	if err != nil {
		t.Fatalf("expected no error finding user, got %v", err)
	}
	if found == nil || found.UserEmail != "user@example.com" || found.MobilePhone != "0811111111" {
		t.Fatalf("unexpected user: %+v", found)
	}

	updatedBy := "admin"
	if err := service.Update(models.User{
		ID:           id,
		UserEmail:    "updated@example.com",
		MobilePhone:  "0822222222",
		UserPassword: "secret2",
		UserType:     models.UserTypeSuperAdmin,
		IsActive:     false,
		Username:     "alice-updated",
		BaseModel: models.BaseModel{
			UpdatedBy: &updatedBy,
		},
	}); err != nil {
		t.Fatalf("expected no error updating user, got %v", err)
	}

	found, err = service.FindByEmail("UPDATED@example.com")
	if err != nil {
		t.Fatalf("expected no error finding by email, got %v", err)
	}
	if found == nil || found.Username != "alice-updated" || found.MobilePhone != "0822222222" || found.IsActive {
		t.Fatalf("unexpected updated user: %+v", found)
	}

	found, err = service.FindByUsername("alice-updated")
	if err != nil {
		t.Fatalf("expected no error finding by username, got %v", err)
	}
	if found == nil || found.ID != id {
		t.Fatalf("unexpected user by username: %+v", found)
	}

	items, err := service.FindList("updated")
	if err != nil {
		t.Fatalf("expected no error listing users, got %v", err)
	}
	if len(items) != 1 || items[0].ID != id {
		t.Fatalf("unexpected list result: %+v", items)
	}

	all, err := service.FindAll()
	if err != nil {
		t.Fatalf("expected no error finding all users, got %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 user, got %d", len(all))
	}

	if err := service.Delete(id); err != nil {
		t.Fatalf("expected no error deleting user, got %v", err)
	}
	if _, err := service.FindByID(id); err == nil {
		t.Fatalf("expected error after delete")
	}
}

func TestUserServiceDeleteBlocksSuperAdmin(t *testing.T) {
	service, cleanup := newRealUserService(t, models.Auth{AccountID: 7, UserName: "alice"})
	defer cleanup()

	id, err := service.Create(models.User{
		UserEmail:    "superadmin@example.com",
		MobilePhone:  "0811111119",
		UserPassword: "secret",
		UserType:     models.UserTypeSuperAdmin,
		IsActive:     true,
		Username:     "superadmin",
	})
	if err != nil {
		t.Fatalf("expected no error creating user, got %v", err)
	}

	if err := service.Delete(id); err == nil || err.Error() != "user super admin tidak bisa dihapus" {
		t.Fatalf("expected super admin delete to be blocked, got %v", err)
	}
}

func TestUserServiceActivateUser(t *testing.T) {
	service, cleanup := newRealUserService(t, models.Auth{AccountID: 7, UserName: "alice"})
	defer cleanup()

	id, err := service.Create(models.User{
		UserEmail:    "inactive@example.com",
		MobilePhone:  "0811111111",
		UserPassword: "secret",
		UserType:     models.UserTypeSuperAdmin,
		IsActive:     false,
		Username:     "inactive",
	})
	if err != nil {
		t.Fatalf("expected no error creating user, got %v", err)
	}

	if err := service.ActivateUser(id); err != nil {
		t.Fatalf("expected no error activating user, got %v", err)
	}

	found, err := service.FindByID(id)
	if err != nil {
		t.Fatalf("expected no error finding activated user, got %v", err)
	}
	if !found.IsActive {
		t.Fatalf("expected user to be active")
	}
}

func TestUserServiceActivateUserValidatesID(t *testing.T) {
	service, cleanup := newRealUserService(t, models.Auth{AccountID: 7, UserName: "alice"})
	defer cleanup()

	if err := service.ActivateUser(0); err == nil || err.Error() != "id user tidak valid" {
		t.Fatalf("expected invalid id error, got %v", err)
	}
}

func TestUserServiceKeepsAccountScope(t *testing.T) {
	service, cleanup := newRealUserService(t, models.Auth{AccountID: 7, UserName: "alice"})
	defer cleanup()

	id, err := service.Create(models.User{
		UserEmail:    "scoped@example.com",
		MobilePhone:  "0811111111",
		UserPassword: "secret",
		UserType:     models.UserTypeSuperAdmin,
		IsActive:     true,
		Username:     "scoped",
	})
	if err != nil {
		t.Fatalf("expected no error creating user, got %v", err)
	}

	otherService := services.NewUserService(authServiceDB, models.Auth{AccountID: 8, UserName: "bob"})
	if _, err := otherService.FindByID(id); err == nil {
		t.Fatalf("expected account-scoped lookup to miss user")
	}
	if err := otherService.ActivateUser(id); err == nil || err.Error() != "user tidak ditemukan" {
		t.Fatalf("expected account-scoped activate error, got %v", err)
	}
}

func TestUserServiceEnforcesGlobalEmailAndMobilephone(t *testing.T) {
	service, cleanup := newRealUserService(t, models.Auth{AccountID: 7, UserName: "alice"})
	defer cleanup()

	if _, err := service.Create(models.User{
		UserEmail:    "global@example.com",
		MobilePhone:  "0811111111",
		UserPassword: "secret",
		UserType:     models.UserTypeSuperAdmin,
		IsActive:     true,
		Username:     "alice",
	}); err != nil {
		t.Fatalf("expected no error creating first user, got %v", err)
	}

	otherService := services.NewUserService(authServiceDB, models.Auth{AccountID: 8, UserName: "bob"})
	if _, err := otherService.Create(models.User{
		UserEmail:    "GLOBAL@example.com",
		MobilePhone:  "0822222222",
		UserPassword: "secret",
		UserType:     models.UserTypeSuperAdmin,
		IsActive:     true,
		Username:     "bob",
	}); err == nil || err.Error() != "email user sudah digunakan" {
		t.Fatalf("expected duplicate global email error, got %v", err)
	}

	if _, err := otherService.Create(models.User{
		UserEmail:    "bob@example.com",
		MobilePhone:  "0811111111",
		UserPassword: "secret",
		UserType:     models.UserTypeSuperAdmin,
		IsActive:     true,
		Username:     "bob",
	}); err == nil || err.Error() != "mobilephone user sudah digunakan" {
		t.Fatalf("expected duplicate global mobilephone error, got %v", err)
	}
}

func TestUserServiceFindRecentList(t *testing.T) {
	service, cleanup := newRealUserService(t, models.Auth{AccountID: 7, UserName: "alice"})
	defer cleanup()

	rows := []struct {
		email string
		phone string
		name  string
	}{
		{"first@example.com", "0811111111", "first"},
		{"second@example.com", "0822222222", "second"},
		{"third@example.com", "0833333333", "third"},
	}

	for _, row := range rows {
		if _, err := service.Create(models.User{
			UserEmail:    row.email,
			MobilePhone:  row.phone,
			UserPassword: "secret",
			UserType:     models.UserTypeSuperAdmin,
			IsActive:     true,
			Username:     row.name,
		}); err != nil {
			t.Fatalf("expected no error creating user, got %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	items, err := service.FindRecentList(2, "")
	if err != nil {
		t.Fatalf("expected no error finding recent users, got %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 users, got %d", len(items))
	}
	if items[0].Username != "third" {
		t.Fatalf("expected newest user first, got %q", items[0].Username)
	}
}
