package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"backend/shared/models"
	"backend/shared/services"
)

type AccountRoutes struct {
	Service *services.AccountService
}

func NewAccountRoutes(service *services.AccountService) *AccountRoutes {
	return &AccountRoutes{Service: service}
}

func (r *AccountRoutes) Register(mux *http.ServeMux) {
	mux.HandleFunc("/accounts", r.handleAccounts)
	mux.HandleFunc("/accounts/", r.handleAccountByID)
	mux.HandleFunc("/accounts/email/", r.handleAccountByEmail)
	mux.HandleFunc("/accounts/recent", r.handleAccountRecent)
}

func (r *AccountRoutes) handleAccounts(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		r.handleAccountList(w, req)
	case http.MethodPost:
		r.handleAccountCreate(w, req)
	default:
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (r *AccountRoutes) handleAccountRecent(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
	search := req.URL.Query().Get("search")

	items, err := service.FindRecentList(limit, search)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (r *AccountRoutes) handleAccountList(w http.ResponseWriter, req *http.Request) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	search := req.URL.Query().Get("search")
	includeAll := req.URL.Query().Get("all") == "1" || strings.EqualFold(req.URL.Query().Get("all"), "true")

	var items []models.Account
	if includeAll {
		items, err = service.FindAll()
	} else {
		items, err = service.FindList(search)
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (r *AccountRoutes) handleAccountCreate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	var payload models.Account
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	id, err := service.Create(payload)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"id": id})
}

func (r *AccountRoutes) handleAccountByID(w http.ResponseWriter, req *http.Request) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	path := strings.TrimPrefix(req.URL.Path, "/accounts/")
	if path == "" {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}

	if parts[0] == "email" {
		r.handleAccountByEmail(w, req)
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid account id")
		return
	}

	if len(parts) > 1 {
		switch parts[1] {
		case "activate":
			r.handleAccountActivate(w, req, id)
		case "otp":
			r.handleAccountOTP(w, req, id)
		default:
			writeJSONError(w, http.StatusNotFound, "not found")
		}
		return
	}

	switch req.Method {
	case http.MethodGet:
		acc, err := service.FindByID(id)
		if err != nil {
			writeJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"data": acc})
	case http.MethodPut:
		var payload models.Account
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json payload")
			return
		}
		payload.ID = id
		if err := service.Update(payload); err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"message": "updated"})
	case http.MethodDelete:
		if err := service.Delete(id); err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"message": "deleted"})
	default:
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (r *AccountRoutes) handleAccountActivate(w http.ResponseWriter, req *http.Request, id int64) {
	if req.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := service.ActivateAccount(id); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"message": "activated"})
}

func (r *AccountRoutes) handleAccountOTP(w http.ResponseWriter, req *http.Request, id int64) {
	if req.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	otp, err := service.GenerateAccountOTP(id)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"otp": otp})
}

func (r *AccountRoutes) handleAccountByEmail(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	email := strings.TrimPrefix(req.URL.Path, "/accounts/email/")
	email = strings.Trim(email, "/")
	if email == "" {
		writeJSONError(w, http.StatusBadRequest, "invalid account email")
		return
	}

	acc, err := service.FindByEmail(email)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"data": acc})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func authFromRequest(req *http.Request) (models.Auth, error) {
	userName := strings.TrimSpace(req.Header.Get("X-User-Name"))
	if userName == "" {
		return models.Auth{}, errors.New("X-User-Name wajib diisi")
	}

	accountID, _ := strconv.ParseInt(req.Header.Get("X-Account-ID"), 10, 64)
	return models.Auth{
		AccountID: accountID,
		UserName:  userName,
		Email:     strings.TrimSpace(req.Header.Get("X-User-Email")),
	}, nil
}

func (r *AccountRoutes) serviceForRequest(req *http.Request) (*services.AccountService, error) {
	if r.Service == nil {
		return nil, errors.New("account service belum dikonfigurasi")
	}

	auth, err := authFromRequest(req)
	if err != nil {
		return nil, err
	}

	return services.NewAccountService(r.Service.DB, auth), nil
}
