package services_test

import (
	"testing"

	"backend/shared/models"
	"backend/shared/services"
)

func newRealRoleService(t *testing.T) (*services.RoleService, func()) {
	t.Helper()

	if authServiceDB == nil {
		t.Skip("role service db is not initialized")
	}

	models.NewRole(authServiceDB, models.Auth{AccountID: 7, UserName: "alice", Email: "alice@example.com"}).Migrate()
	service := services.NewRoleService(authServiceDB, models.Auth{AccountID: 7, UserName: "alice", Email: "alice@example.com"})
	_, _ = authServiceDB.Exec(`TRUNCATE TABLE roles RESTART IDENTITY CASCADE`)

	cleanup := func() {
		_, _ = authServiceDB.Exec(`TRUNCATE TABLE roles RESTART IDENTITY CASCADE`)
	}

	return service, cleanup
}

func TestRoleServiceCRUD(t *testing.T) {
	service, cleanup := newRealRoleService(t)
	defer cleanup()

	id, err := service.Create(models.Role{
		RoleName: "Admin",
		Actions:  models.StringArray{"read", "update"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	found, err := service.FindByID(id)
	if err != nil {
		t.Fatalf("expected no error finding role, got %v", err)
	}
	if found == nil || found.RoleName != "Admin" {
		t.Fatalf("unexpected role: %+v", found)
	}

	updatedBy := "bob@example.com"
	if err := service.Update(models.Role{
		ID:       id,
		RoleName: "Admin Updated",
		Actions:  models.StringArray{"create", "delete"},
		BaseModel: models.BaseModel{
			UpdatedBy: &updatedBy,
		},
	}); err != nil {
		t.Fatalf("expected no error updating, got %v", err)
	}

	found, err = service.FindByName("Admin Updated")
	if err != nil {
		t.Fatalf("expected no error finding by name, got %v", err)
	}
	if found == nil || found.RoleName != "Admin Updated" {
		t.Fatalf("unexpected updated role: %+v", found)
	}

	items, err := service.FindList("read")
	if err != nil {
		t.Fatalf("expected no error listing, got %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 role, got %d", len(items))
	}

	if err := service.Delete(id); err != nil {
		t.Fatalf("expected no error deleting, got %v", err)
	}
}
