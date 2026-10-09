package routes

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"backend/shared/models"
	"backend/shared/services"
	"github.com/go-chi/chi/v5"
)

type AuthRoutes struct {
	AccountService *services.AccountService
	AuthService    *services.AuthService
}

func NewAuthRoutes(accountService *services.AccountService, authService *services.AuthService) *AuthRoutes {
	return &AuthRoutes{
		AccountService: accountService,
		AuthService:    authService,
	}
}

func (r *AuthRoutes) Register(router chi.Router) {
	registerLoggedHandler(router, "/auth/*", r.handleAuth)
}

func (r *AuthRoutes) handleAuth(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/auth/*")
	path = strings.Trim(path, "/")
	switch path {
	case "login":
		r.handleLogin(w, req)
	case "login/pin":
		r.handleLoginPIN(w, req)
	case "account-devices/location":
		r.handleAccountDeviceLocation(w, req)
	case "refresh":
		r.handleRefresh(w, req)
	case "logout":
		r.handleLogout(w, req)
	case "change-password-pin":
		r.handleChangePasswordAndPIN(w, req)
	case "accounts/register":
		r.handleAccountRegister(w, req)
	case "accounts/verify":
		r.handleAccountVerify(w, req)
	case "accounts/resend-otp":
		r.handleAccountResendOTP(w, req)
	case "users/otp":
		r.handleUserOTP(w, req)
	case "users/verify":
		r.handleUserVerify(w, req)
	case "users/pin/status":
		r.handleUserPINStatus(w, req)
	case "users/pin/verify":
		r.handleUserPINVerify(w, req)
	default:
		writeAuthJSONError(w, http.StatusNotFound, "not found")
	}
}

func (r *AuthRoutes) handleChangePasswordAndPIN(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeAuthJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if r.AuthService == nil {
		writeAuthJSONError(w, http.StatusServiceUnavailable, "auth service belum dikonfigurasi")
		return
	}
	var payload struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
		ConfirmPassword string `json:"confirmPassword"`
		PIN             string `json:"pin"`
		ConfirmPIN      string `json:"confirmPin"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeAuthJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	token := strings.TrimSpace(req.Header.Get("X-Session-Token"))
	if token == "" {
		authorization := strings.TrimSpace(req.Header.Get("Authorization"))
		token = strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
	}
	if err := r.AuthService.ChangePasswordAndPIN(token, payload.CurrentPassword, payload.NewPassword, payload.ConfirmPassword, payload.PIN, payload.ConfirmPIN); err != nil {
		writeAuthJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeAuthJSON(w, http.StatusOK, map[string]string{"message": "password dan PIN berhasil diubah"})
}

func (r *AuthRoutes) handleAccountRegister(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeAuthJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if r.AccountService == nil {
		writeAuthJSONError(w, http.StatusServiceUnavailable, "account service belum dikonfigurasi")
		return
	}

	var payload models.Account
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeAuthJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	id, _, expired, err := r.AccountService.RegisterAccount(payload)
	if err != nil {
		writeAuthJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeAuthJSON(w, http.StatusCreated, map[string]interface{}{
		"id":          id,
		"otp_expired": expired,
		"is_active":   false,
		"message":     "OTP sudah dikirim ke email. Silakan ambil OTP dari email untuk verifikasi.",
	})
}

func (r *AuthRoutes) handleAccountVerify(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeAuthJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if r.AccountService == nil {
		writeAuthJSONError(w, http.StatusServiceUnavailable, "account service belum dikonfigurasi")
		return
	}

	var payload struct {
		OTP string `json:"otp"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeAuthJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	account, user, err := r.AccountService.VerifyAccountRegistration(payload.OTP)
	if err != nil {
		writeAuthJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeAuthJSON(w, http.StatusOK, map[string]interface{}{
		"message": "verified",
		"account": account,
		"user":    user,
	})
}

func (r *AuthRoutes) handleAccountResendOTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeAuthJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if r.AccountService == nil {
		writeAuthJSONError(w, http.StatusServiceUnavailable, "account service belum dikonfigurasi")
		return
	}

	var payload struct {
		AccountID int64 `json:"account_id"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeAuthJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	expired, err := r.AccountService.ResendAccountRegistrationOTP(payload.AccountID)
	if err != nil {
		writeAuthJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeAuthJSON(w, http.StatusOK, map[string]interface{}{
		"otp_expired": expired,
		"message":     "OTP baru sudah dikirim ke email.",
	})
}

func (r *AuthRoutes) handleUserOTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeAuthJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if r.AuthService == nil {
		writeAuthJSONError(w, http.StatusServiceUnavailable, "auth service belum dikonfigurasi")
		return
	}

	var payload struct {
		Identifier string `json:"identifier"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeAuthJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	id, err := r.AuthService.GenerateLoginOTP(payload.Identifier)
	if err != nil {
		writeAuthJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeAuthJSON(w, http.StatusCreated, map[string]interface{}{
		"id":      id,
		"message": "otp login dikirim ke email terdaftar",
	})
}

func (r *AuthRoutes) handleUserVerify(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeAuthJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if r.AuthService == nil {
		writeAuthJSONError(w, http.StatusServiceUnavailable, "auth service belum dikonfigurasi")
		return
	}

	var payload struct {
		Identifier string `json:"identifier"`
		OTP        string `json:"otp"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeAuthJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	record, err := r.AuthService.VerifyLoginOTP(payload.Identifier, payload.OTP)
	if err != nil {
		writeAuthJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeAuthJSON(w, http.StatusOK, map[string]interface{}{"data": record})
}

func (r *AuthRoutes) handleUserPINVerify(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeAuthJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if r.AuthService == nil {
		writeAuthJSONError(w, http.StatusServiceUnavailable, "auth service belum dikonfigurasi")
		return
	}

	var payload struct {
		PIN string `json:"pin"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeAuthJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	token := bearerToken(req.Header.Get("Authorization"))
	if token == "" {
		token = strings.TrimSpace(req.Header.Get("X-Session-Token"))
	}
	if token == "" {
		writeAuthJSONError(w, http.StatusUnauthorized, "session buyer wajib diisi")
		return
	}

	record, err := r.AuthService.VerifyPIN(token, payload.PIN)
	if err != nil {
		status := http.StatusBadRequest
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "expired") || strings.Contains(lower, "tidak valid") || strings.Contains(lower, "session") {
			status = http.StatusUnauthorized
		}
		writeAuthJSONError(w, status, err.Error())
		return
	}

	writeAuthJSON(w, http.StatusOK, map[string]interface{}{
		"ok":            true,
		"id":            record.ID,
		"token":         record.Token,
		"token_expired": record.TokenExpired,
		"account_id":    record.AccountID,
		"email":         record.Email,
		"name":          record.Name,
	})
}

func (r *AuthRoutes) handleUserPINStatus(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet && req.Method != http.MethodPost {
		writeAuthJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if r.AuthService == nil {
		writeAuthJSONError(w, http.StatusServiceUnavailable, "auth service belum dikonfigurasi")
		return
	}

	token := bearerToken(req.Header.Get("Authorization"))
	if token == "" {
		token = strings.TrimSpace(req.Header.Get("X-Session-Token"))
	}
	if token == "" {
		writeAuthJSONError(w, http.StatusUnauthorized, "session buyer wajib diisi")
		return
	}

	record, hasPIN, err := r.AuthService.PINStatus(token)
	if err != nil {
		status := http.StatusBadRequest
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "expired") || strings.Contains(lower, "tidak valid") || strings.Contains(lower, "session") {
			status = http.StatusUnauthorized
		}
		writeAuthJSONError(w, status, err.Error())
		return
	}

	writeAuthJSON(w, http.StatusOK, map[string]interface{}{
		"ok":            true,
		"has_pin":       hasPIN,
		"token":         record.Token,
		"token_expired": record.TokenExpired,
		"account_id":    record.AccountID,
		"email":         record.Email,
		"name":          record.Name,
	})
}

func (r *AuthRoutes) handleLogin(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeAuthJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if r.AuthService == nil {
		writeAuthJSONError(w, http.StatusServiceUnavailable, "auth service belum dikonfigurasi")
		return
	}

	var payload struct {
		Identifier string `json:"identifier"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		LocationID int64  `json:"locationId"`
		DeviceID   string `json:"deviceId"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeAuthJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	identifier := payload.Identifier
	if strings.TrimSpace(identifier) == "" {
		identifier = payload.Email
	}

	token, err := r.AuthService.Login(identifier, payload.Password, payload.LocationID, payload.DeviceID)
	if err != nil {
		writeAuthJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	record, err := r.AuthService.RefreshSession(token)
	if err != nil {
		writeAuthJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	access, err := r.AuthService.ResolveSessionAccess(record)
	if err != nil {
		writeAuthJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	userID, err := r.findUserID(record.AccountID, record.Email)
	if err != nil {
		writeAuthJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	customerID, err := r.findCustomerID(record.Email)
	if err != nil {
		writeAuthJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeAuthJSON(w, http.StatusOK, map[string]interface{}{
		"token": token, "user_type": access.UserType, "login_type": access.LoginType, "roles": access.Roles, "permissions": access.Permissions,
		"account_id": record.AccountID,
		"user_id":    userID,
		"customerId": customerID,
	})
}

func (r *AuthRoutes) handleLoginPIN(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeAuthJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if r.AuthService == nil {
		writeAuthJSONError(w, http.StatusServiceUnavailable, "auth service belum dikonfigurasi")
		return
	}

	var payload struct {
		Identifier string `json:"identifier"`
		PIN        string `json:"pin"`
		LocationID int64  `json:"locationId"`
		DeviceID   string `json:"deviceId"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeAuthJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	token, err := r.AuthService.LoginWithPIN(payload.Identifier, payload.PIN, payload.LocationID, payload.DeviceID)
	if err != nil {
		writeAuthJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	record, err := r.AuthService.RefreshSession(token)
	if err != nil {
		writeAuthJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	access, err := r.AuthService.ResolveSessionAccess(record)
	if err != nil {
		writeAuthJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeAuthJSON(w, http.StatusOK, map[string]interface{}{
		"token": token, "user_type": access.UserType, "login_type": access.LoginType, "roles": access.Roles, "permissions": access.Permissions,
	})
}

func (r *AuthRoutes) handleLogout(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeAuthJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if r.AuthService == nil {
		writeAuthJSONError(w, http.StatusServiceUnavailable, "auth service belum dikonfigurasi")
		return
	}

	var payload struct {
		Token string `json:"token"`
	}
	if req.Body != nil {
		_ = json.NewDecoder(req.Body).Decode(&payload)
	}

	token := strings.TrimSpace(payload.Token)
	if token == "" {
		token = bearerToken(req.Header.Get("Authorization"))
	}
	if token == "" {
		writeAuthJSONError(w, http.StatusBadRequest, "token wajib diisi")
		return
	}

	if err := r.AuthService.ExpireSession(token); err != nil && !errorsIsNoRows(err) {
		writeAuthJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeAuthJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "session expired",
	})
}

func (r *AuthRoutes) handleAccountDeviceLocation(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeAuthJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	auth, err := authFromRequest(req)
	if err != nil {
		writeAuthJSONError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if r.AuthService == nil || r.AuthService.DB == nil {
		writeAuthJSONError(w, http.StatusServiceUnavailable, "auth service belum dikonfigurasi")
		return
	}

	var payload struct {
		DeviceID   string `json:"deviceId"`
		LocationID int64  `json:"locationId"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeAuthJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	service := services.NewAccountDeviceServiceWithContext(req.Context(), r.AuthService.DB, auth)
	item, err := service.UpdateLocation(strings.TrimSpace(payload.DeviceID), payload.LocationID)
	if err != nil {
		writeAuthJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeAuthJSON(w, http.StatusOK, map[string]interface{}{
		"message": "updated",
		"data":    item,
	})
}

func (r *AuthRoutes) handleRefresh(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeAuthJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if r.AuthService == nil {
		writeAuthJSONError(w, http.StatusServiceUnavailable, "auth service belum dikonfigurasi")
		return
	}

	var payload struct {
		Token string `json:"token"`
	}
	if req.Body != nil {
		_ = json.NewDecoder(req.Body).Decode(&payload)
	}

	token := strings.TrimSpace(payload.Token)
	if token == "" {
		token = bearerToken(req.Header.Get("Authorization"))
	}
	if token == "" {
		writeAuthJSONError(w, http.StatusBadRequest, "token wajib diisi")
		return
	}

	record, err := r.AuthService.RefreshSession(token)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "expired") || strings.Contains(strings.ToLower(err.Error()), "tidak valid") {
			status = http.StatusUnauthorized
		}
		writeAuthJSONError(w, status, err.Error())
		return
	}
	access, err := r.AuthService.ResolveSessionAccess(record)
	if err != nil {
		writeAuthJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	customerID, err := r.findCustomerID(record.Email)
	if err != nil {
		writeAuthJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeAuthJSON(w, http.StatusOK, map[string]interface{}{
		"ok":            true,
		"token":         record.Token,
		"token_expired": record.TokenExpired,
		"account_id":    record.AccountID,
		"customerId":    customerID,
		"email":         record.Email,
		"name":          record.Name,
		"user_type":     access.UserType,
		"login_type":    access.LoginType,
		"roles":         access.Roles,
		"permissions":   access.Permissions,
	})
}

func (r *AuthRoutes) findCustomerID(email string) (int64, error) {
	var customerID int64
	err := r.AuthService.DB.Get(
		&customerID,
		`SELECT id FROM customer_accounts WHERE LOWER(email) = LOWER($1) LIMIT 1`,
		strings.TrimSpace(email),
	)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return customerID, err
}

func (r *AuthRoutes) findUserID(accountID int64, email string) (int64, error) {
	var userID int64
	err := r.AuthService.DB.Get(
		&userID,
		`SELECT id FROM users WHERE account_id = $1 AND LOWER(user_email) = LOWER($2) LIMIT 1`,
		accountID,
		strings.TrimSpace(email),
	)
	return userID, err
}

func bearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}

	prefix := "bearer "
	if !strings.HasPrefix(strings.ToLower(header), prefix) {
		return ""
	}

	return strings.TrimSpace(header[len(prefix):])
}

func errorsIsNoRows(err error) bool {
	return err == sql.ErrNoRows || strings.Contains(strings.ToLower(err.Error()), "no rows")
}

func writeAuthJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeAuthJSONError(w http.ResponseWriter, status int, message string) {
	writeAuthJSON(w, status, map[string]string{"error": message})
}
