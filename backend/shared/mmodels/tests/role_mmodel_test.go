package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"backend/shared/mmodels"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// TestRoleMModelCRUD migrates the roles table before exercising its CRUD methods.
func TestRoleMModelCRUD(t *testing.T) {
	const dsn = "postgres://template_project:local-dev-password@localhost:31008/template_project_master?sslmode=disable"

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	model := mmodels.NewRoleMModel(context.Background(), db)
	if err := model.Migrate(); err != nil {
		t.Fatalf("migrate roles table: %v", err)
	}

	accountID := time.Now().UnixNano()
	roleName := fmt.Sprintf("metadata-test-%d", accountID)
	model.AccountID = accountID
	model.AuthAccountID = accountID
	model.AuthEmail = "metadata-test@example.test"
	model.CreatedBy = model.AuthEmail

	var id int64
	t.Cleanup(func() {
		if id > 0 {
			if _, err := db.Exec(`DELETE FROM roles WHERE id = $1 AND account_id = $2`, id, accountID); err != nil {
				t.Errorf("clean up test role: %v", err)
			}
		}
	})

	model.RoleName = roleName
	model.Actions = pq.StringArray{"read", "update"}
	id, err = model.Create()
	if err != nil {
		t.Fatalf("create role: %v", err)
	}
	if id <= 0 {
		t.Fatalf("created role id = %d, want positive id", id)
	}

	foundModel := mmodels.NewRoleMModel(context.Background(), db)
	foundModel.AccountID = accountID
	foundModel.AuthAccountID = accountID
	foundModel.ID = id
	found, err := foundModel.FindByID()
	if err != nil {
		t.Fatalf("find created role: %v", err)
	}
	if found.RoleName != roleName || len(found.Actions) != 2 {
		t.Fatalf("found role = name %q actions %v, want %q and two actions", found.RoleName, found.Actions, roleName)
	}

	found.RoleName += "-updated"
	found.Actions = pq.StringArray{"create", "delete"}
	if err := found.Update(); err != nil {
		t.Fatalf("update role: %v", err)
	}

	byName, err := found.FindByName(found.RoleName)
	if err != nil {
		t.Fatalf("find updated role by name: %v", err)
	}
	if byName.ID != id || byName.RoleName != found.RoleName {
		t.Fatalf("role found by name = %+v, want id %d and name %q", byName, id, found.RoleName)
	}

	items, err := found.FindList("delete")
	if err != nil {
		t.Fatalf("list roles by action: %v", err)
	}
	if len(items) != 1 || items[0].ID != id {
		t.Fatalf("role list = %v, want one row with id %d", items, id)
	}

	if err := found.Delete(); err != nil {
		t.Fatalf("delete role: %v", err)
	}
	deleted, err := found.FindByID()
	if err == nil {
		t.Fatalf("FindByID returned role %+v after delete", deleted)
	}
	id = 0
}
