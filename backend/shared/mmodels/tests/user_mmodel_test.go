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

// TestUserMModelCRUD uses the local PostgreSQL dev database configured in
// infra/dev/db and migrates users before exercising CRUD.
func TestUserMModelCRUD(t *testing.T) {
	const dsn = "postgres://template_project:local-dev-password@localhost:31008/template_project_master?sslmode=disable"

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	model := mmodels.NewUserMModel(context.Background(), db)
	if err := model.Migrate(); err != nil {
		t.Fatalf("migrate users table: %v", err)
	}

	var id int64
	t.Cleanup(func() {
		if id > 0 {
			if _, err := db.Exec(`DELETE FROM users WHERE id = $1`, id); err != nil {
				t.Errorf("clean up test user: %v", err)
			}
		}
	})
	user := &mmodels.UserMModel{
		AccountID:    81234,
		UserEmail:    fmt.Sprintf("crud-%d@example.test", time.Now().UnixNano()),
		MobilePhone:  fmt.Sprintf("08%010d", time.Now().UnixNano()%10_000_000_000),
		UserPassword: "hashed-password-placeholder",
		UserType:     "USER",
		LoginType:    "cashier",
		UserRoles:    pq.StringArray{"operator"},
		IsActive:     true,
		Username:     fmt.Sprintf("crud-%d", time.Now().UnixNano()),
		CreatedBy:    "metadata-test",
	}

	id, err = model.Create_(user)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if id <= 0 {
		t.Fatalf("created user id = %d, want positive id", id)
	}

	found := &mmodels.UserMModel{ID: id}
	if err := model.FindByID_(found); err != nil {
		t.Fatalf("find created user: %v", err)
	}
	if found.UserEmail != user.UserEmail || found.Username != user.Username {
		t.Fatalf("found user = email %q username %q, want %q / %q", found.UserEmail, found.Username, user.UserEmail, user.Username)
	}
	var persistedRoles string
	if err := db.Get(&persistedRoles, fmt.Sprintf(`SELECT user_roles::text FROM %q WHERE id = $1`, model.TableName), id); err != nil {
		t.Fatalf("read persisted roles: %v", err)
	}
	if persistedRoles != "{operator}" {
		t.Fatalf("persisted roles = %q, want %q", persistedRoles, "{operator}")
	}

	found.Username = found.Username + "-updated"
	found.UserRoles = user.UserRoles
	if err := model.Update_(found); err != nil {
		t.Fatalf("update user: %v", err)
	}

	updated := &mmodels.UserMModel{ID: id}
	if err := model.FindByID_(updated); err != nil {
		t.Fatalf("find updated user: %v", err)
	}
	if updated.Username != found.Username {
		t.Fatalf("updated username = %q, want %q", updated.Username, found.Username)
	}

	list := make([]mmodels.UserMModel, 0)
	if err := model.FindListOrdered(&list, map[string]any{"account_id": user.AccountID}, user.UserEmail, []string{"user_email"}, "id", "ASC", 10); err != nil {
		t.Fatalf("list users from database: %v", err)
	}
	if len(list) != 1 || list[0].ID != id {
		t.Fatalf("list returned %d rows; want one row with id %d", len(list), id)
	}

	if err := model.Delete_(updated); err != nil {
		t.Fatalf("delete user: %v", err)
	}
	deleted := &mmodels.UserMModel{ID: id}
	if err := model.FindByID_(deleted); err == nil {
		t.Fatal("FindByID_ succeeded after deleting the user")
	}
}
