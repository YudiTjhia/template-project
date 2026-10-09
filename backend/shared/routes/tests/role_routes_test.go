package routes_test

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"backend/shared/connections"
	"backend/shared/models"
	"backend/shared/routes"
	"backend/shared/services"
	"github.com/jmoiron/sqlx"
)

const roleRoutesTestSuiteLockID int64 = 91384719

var (
	roleRoutesDB *sqlx.DB
	roleLockConn *sql.Conn
)

func TestMain(m *testing.M) {
	_ = loadEnvFile(".env", "../../../../.env")

	db, err := connections.NewMasterConnection()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		code := m.Run()
		os.Exit(code)
	}
	roleRoutesDB = db

	roleLockConn, err = db.Conn(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		roleRoutesDB = nil
		code := m.Run()
		_ = db.Close()
		os.Exit(code)
	}
	if _, err := roleLockConn.ExecContext(context.Background(), `SELECT pg_advisory_lock($1)`, roleRoutesTestSuiteLockID); err != nil {
		fmt.Fprintln(os.Stderr, err)
		_ = roleLockConn.Close()
		_ = db.Close()
		os.Exit(1)
	}

	models.NewRole(roleRoutesDB, models.Auth{AccountID: 7, UserName: "alice", Email: "alice@example.com"}).Migrate()
	models.NewRoleAction(roleRoutesDB, models.Auth{}).Migrate()
	_, _ = roleRoutesDB.Exec(`TRUNCATE TABLE roles RESTART IDENTITY CASCADE`)
	_, _ = roleRoutesDB.Exec(`TRUNCATE TABLE role_actions RESTART IDENTITY CASCADE`)
	_, _ = roleRoutesDB.Exec(`
		INSERT INTO role_actions (code, label, sort_order, active) VALUES
		('create_user', 'Create User', 10, TRUE),
		('update_role', 'Update Role', 20, TRUE),
		('delete_role', 'Delete Role', 30, TRUE)
		ON CONFLICT (code) DO UPDATE SET
		  label = EXCLUDED.label,
		  sort_order = EXCLUDED.sort_order,
		  active = EXCLUDED.active
	`)

	code := m.Run()

	_, _ = roleLockConn.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, roleRoutesTestSuiteLockID)
	_ = roleLockConn.Close()
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

func TestRoleRoutesRegisterAndCreate(t *testing.T) {
	if roleRoutesDB == nil {
		t.Skip("role routes db is not initialized")
	}

	_, _ = roleRoutesDB.Exec(`TRUNCATE TABLE roles RESTART IDENTITY CASCADE`)

	service := services.NewRoleService(roleRoutesDB, models.Auth{AccountID: 7, UserName: "alice", Email: "alice@example.com"})
	handler := routes.NewRoleRoutes(service)
	mux := http.NewServeMux()
	handler.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(`{"role_name":"Admin","actions":["read","update"]}`))
	req.Header.Set("X-Account-ID", "7")
	req.Header.Set("X-User-Name", "alice")
	req.Header.Set("X-User-Email", "alice@example.com")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestRoleRoutesActionsLookup(t *testing.T) {
	if roleRoutesDB == nil {
		t.Skip("role routes db is not initialized")
	}

	service := services.NewRoleService(roleRoutesDB, models.Auth{AccountID: 7, UserName: "alice", Email: "alice@example.com"})
	handler := routes.NewRoleRoutes(service)
	mux := http.NewServeMux()
	handler.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/roles/actions?search=role", nil)
	req.Header.Set("X-Account-ID", "7")
	req.Header.Set("X-User-Name", "alice")
	req.Header.Set("X-User-Email", "alice@example.com")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	if !strings.Contains(rr.Body.String(), "update_role") {
		t.Fatalf("expected action catalog in response, got %s", rr.Body.String())
	}
}

func TestRoleRoutesActionCRUD(t *testing.T) {
	if roleRoutesDB == nil {
		t.Skip("role routes db is not initialized")
	}

	_, _ = roleRoutesDB.Exec(`TRUNCATE TABLE role_actions RESTART IDENTITY CASCADE`)

	service := services.NewRoleService(roleRoutesDB, models.Auth{AccountID: 7, UserName: "alice", Email: "alice@example.com"})
	handler := routes.NewRoleRoutes(service)
	mux := http.NewServeMux()
	handler.Register(mux)

	createReq := httptest.NewRequest(http.MethodPost, "/roles/actions", strings.NewReader(`{"value":"export_report","label":"Export Report","sort_order":90,"active":true}`))
	createReq.Header.Set("X-Account-ID", "7")
	createReq.Header.Set("X-User-Name", "alice")
	createReq.Header.Set("X-User-Email", "alice@example.com")
	createRR := httptest.NewRecorder()
	mux.ServeHTTP(createRR, createReq)

	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", createRR.Code, createRR.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/roles/actions?all=1", nil)
	listReq.Header.Set("X-Account-ID", "7")
	listReq.Header.Set("X-User-Name", "alice")
	listReq.Header.Set("X-User-Email", "alice@example.com")
	listRR := httptest.NewRecorder()
	mux.ServeHTTP(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", listRR.Code, listRR.Body.String())
	}

	if !strings.Contains(listRR.Body.String(), "export_report") {
		t.Fatalf("expected created action in response, got %s", listRR.Body.String())
	}
}
