package middleware

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type PermissionMiddleware struct {
	DB               *sqlx.DB
	ActionForRequest func(*http.Request) string
	BeforeAction     func(http.ResponseWriter, *http.Request, PermissionUser) (handled bool)
	SkipAction       func(*http.Request, PermissionUser) bool
}

type PermissionUser struct {
	ID        int64          `db:"id"`
	AccountID int64          `db:"account_id"`
	Email     string         `db:"user_email"`
	UserType  string         `db:"user_type"`
	LoginType string         `db:"login_type"`
	Roles     pq.StringArray `db:"user_roles"`
	IsActive  bool           `db:"is_active"`
}

func NewPermissionMiddleware(db *sqlx.DB, actionForRequest func(*http.Request) string) *PermissionMiddleware {
	return &PermissionMiddleware{DB: db, ActionForRequest: actionForRequest}
}

func (m *PermissionMiddleware) Wrap(next http.Handler, exemptPaths ...string) http.Handler {
	exempt := ExemptPathSet(exemptPaths)
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if IsExemptPath(req.URL.Path, exempt) {
			next.ServeHTTP(w, req)
			return
		}
		if m == nil || m.DB == nil {
			WriteJSONError(w, http.StatusInternalServerError, "permission middleware belum dikonfigurasi")
			return
		}

		accountID, err := strconv.ParseInt(strings.TrimSpace(req.Header.Get("X-Account-ID")), 10, 64)
		if err != nil || accountID <= 0 {
			WriteJSONError(w, http.StatusUnauthorized, "account session tidak valid")
			return
		}
		email := strings.ToLower(strings.TrimSpace(req.Header.Get("X-User-Email")))
		if email == "" {
			WriteJSONError(w, http.StatusUnauthorized, "user session tidak valid")
			return
		}

		user, err := m.findUser(accountID, email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				WriteJSONError(w, http.StatusForbidden, "user tidak terdaftar pada account session")
				return
			}
			WriteJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !user.IsActive || user.AccountID != accountID {
			WriteJSONError(w, http.StatusForbidden, "user tidak aktif atau account tidak sesuai")
			return
		}

		req = req.Clone(req.Context())
		req.Header.Set("X-User-ID", strconv.FormatInt(user.ID, 10))
		req.Header.Set("X-Account-ID", strconv.FormatInt(user.AccountID, 10))
		req.Header.Set("X-Login-Type", strings.ToLower(strings.TrimSpace(user.LoginType)))

		if m.BeforeAction != nil && m.BeforeAction(w, req, *user) {
			return
		}

		action := ""
		if m.ActionForRequest != nil && (m.SkipAction == nil || !m.SkipAction(req, *user)) {
			action = m.ActionForRequest(req)
		}
		if action != "" && !IsSuperAdmin(user.UserType) {
			allowed, err := m.hasAction(user.AccountID, []string(user.Roles), action)
			if err != nil {
				WriteJSONError(w, http.StatusInternalServerError, err.Error())
				return
			}
			if !allowed {
				WriteJSONError(w, http.StatusForbidden, "permission ditolak untuk action "+action)
				return
			}
		}

		next.ServeHTTP(w, req)
	})
}

func (m *PermissionMiddleware) findUser(accountID int64, email string) (*PermissionUser, error) {
	var user PermissionUser
	err := m.DB.Get(&user, `
		SELECT id, account_id, user_email, user_type, login_type, user_roles, is_active
		FROM users
		WHERE account_id = $1 AND LOWER(user_email) = LOWER($2)
		LIMIT 1`, accountID, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (m *PermissionMiddleware) hasAction(accountID int64, roleNames []string, action string) (bool, error) {
	if len(roleNames) == 0 {
		return false, nil
	}
	var allowed bool
	err := m.DB.Get(&allowed, `
		SELECT EXISTS (
			SELECT 1
			FROM roles r
			WHERE r.account_id = $1
				AND LOWER(r.role_name) = ANY($2::text[])
				AND $3 = ANY(r.actions)
		)`, accountID, pq.Array(LowerStrings(roleNames)), action)
	return allowed, err
}

func IsSuperAdmin(userType string) bool {
	return strings.EqualFold(strings.TrimSpace(userType), "SUPER_ADMIN")
}

func LowerStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}
