package models_test

import (
	"testing"

	"backend/shared/models"
)

func newRealRoleDB(t *testing.T) (*models.Role, func()) {
	t.Helper()

	if accountDB == nil {
		t.Skip("role db is not initialized")
	}

	auth := models.Auth{AccountID: 7, UserName: "alice", Email: "alice@example.com"}
	role := models.NewRole(accountDB, auth)
	_, _ = accountDB.Exec(`TRUNCATE TABLE roles RESTART IDENTITY CASCADE`)

	cleanup := func() {
		_, _ = accountDB.Exec(`TRUNCATE TABLE roles RESTART IDENTITY CASCADE`)
	}

	return &role, cleanup
}

func TestRoleTableName(t *testing.T) {
	var role models.Role
	if got := role.TableName(); got != "roles" {
		t.Fatalf("expected table roles, got %q", got)
	}
}

func TestStringArrayScanValue(t *testing.T) {
	arr := models.StringArray{"create", "update"}
	if got, err := arr.Value(); err != nil || got == nil {
		t.Fatalf("expected value to work, got %v %v", got, err)
	}

	var scanned models.StringArray
	if err := (&scanned).Scan(`{create,update}`); err != nil {
		t.Fatalf("expected scan to work, got %v", err)
	}
	if len(scanned) != 2 || scanned[0] != "create" || scanned[1] != "update" {
		t.Fatalf("unexpected scanned array: %#v", scanned)
	}
}

func TestRoleValidateCreate(t *testing.T) {
	tests := []struct {
		name    string
		role    models.Role
		wantErr string
	}{
		{
			name:    "missing account",
			role:    models.Role{RoleName: "Admin"},
			wantErr: "id akun wajib diisi",
		},
		{
			name:    "missing name",
			role:    models.Role{BaseModel: models.BaseModel{Auth: models.Auth{AccountID: 1}}, Actions: models.StringArray{"read"}},
			wantErr: "nama role wajib diisi",
		},
		{
			name: "valid",
			role: models.Role{
				BaseModel: models.BaseModel{Auth: models.Auth{AccountID: 1}},
				RoleName:  "Admin",
				Actions:   models.StringArray{"read"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.role.ValidateCreate()
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

func TestRoleCreateUpdateDelete(t *testing.T) {
	role, cleanup := newRealRoleDB(t)
	defer cleanup()

	role.RoleName = "Admin"
	role.Actions = models.StringArray{"read", "update"}

	id, err := role.Create()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	role.ID = id
	found, err := role.FindByID()
	if err != nil {
		t.Fatalf("expected no error reading back, got %v", err)
	}
	if found == nil || found.RoleName != "Admin" || len(found.Actions) != 2 {
		t.Fatalf("unexpected role: %+v", found)
	}

	updatedBy := "bob@example.com"
	role.RoleName = "Admin Updated"
	role.Actions = models.StringArray{"create", "delete"}
	role.UpdatedBy = &updatedBy
	if err := role.Update(); err != nil {
		t.Fatalf("expected no error updating, got %v", err)
	}

	found, err = role.FindByID()
	if err != nil {
		t.Fatalf("expected no error reading back after update, got %v", err)
	}
	if found.RoleName != "Admin Updated" || len(found.Actions) != 2 {
		t.Fatalf("unexpected updated role: %+v", found)
	}

	if err := role.Delete(); err != nil {
		t.Fatalf("expected no error deleting, got %v", err)
	}
	if _, err := role.FindByID(); err == nil {
		t.Fatalf("expected error after delete")
	}
}

func TestRoleFindListAndName(t *testing.T) {
	role, cleanup := newRealRoleDB(t)
	defer cleanup()

	rows := []struct {
		name    string
		actions []string
	}{
		{"Admin", []string{"read", "update"}},
		{"Operator", []string{"read"}},
	}

	for _, row := range rows {
		role.RoleName = row.name
		role.Actions = models.StringArray(row.actions)
		if _, err := role.Create(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}

	found, err := role.FindByName("Admin")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found == nil || found.RoleName != "Admin" {
		t.Fatalf("unexpected role by name: %+v", found)
	}

	items, err := role.FindList("read")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}
