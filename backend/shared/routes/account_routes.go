package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"backend/shared/metadata/impl"
	"backend/shared/metadata/interfaces"
	"backend/shared/models"
	"backend/shared/services"
	"github.com/go-chi/chi/v5"
)

type AccountRoutes struct {
	Service *services.AccountService
}

var _ interfaces.ICRUDRouteOperations = (*AccountRoutes)(nil)

func NewAccountRoutes(service *services.AccountService) *AccountRoutes {
	return &AccountRoutes{Service: service}
}

func (r *AccountRoutes) Register(router chi.Router) error {
	if err := impl.NewMRoutes().RegisterRoutes(loggedRouteRegistrar{router}, "/accounts", r); err != nil {
		return err
	}
	registerLoggedHandler(router, "/accounts/recent", r.handleAccountRecent)
	registerLoggedHandler(router, "/accounts/email/{email}", r.handleAccountByEmail)
	registerLoggedHandler(router, "/accounts/{id}/activate", r.handleAccountActivate)
	registerLoggedHandler(router, "/accounts/{id}/otp", r.handleAccountOTP)
	return nil
}

func (r *AccountRoutes) List(req *http.Request) (any, error) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	includeAll, err := queryBool(req, "all")
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	if includeAll {
		return service.FindAll()
	}
	return service.FindList(req.URL.Query().Get("search"))
}

func (r *AccountRoutes) Create(req *http.Request, body []byte) (any, error) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	var payload models.Account
	if err := decodeJSONBytes(body, &payload); err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	id, err := service.Create(payload)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	return id, nil
}

func (r *AccountRoutes) FindByID(req *http.Request, id int64) (any, error) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	return service.FindByID(id)
}

func (r *AccountRoutes) Update(req *http.Request, id int64, body []byte) error {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	var payload models.Account
	if err := decodeJSONBytes(body, &payload); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	payload.ID = id
	if err := service.Update(payload); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	return nil
}

func (r *AccountRoutes) Delete(req *http.Request, id int64) error {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	if err := service.Delete(id); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	return nil
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
	limit, err := queryInt(req, "limit", 0)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, err := service.FindRecentList(limit, req.URL.Query().Get("search"))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (r *AccountRoutes) handleAccountActivate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id, err := routeInt64Param(req, "id", "account")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
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

func (r *AccountRoutes) handleAccountOTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id, err := routeInt64Param(req, "id", "account")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
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
	email := routeParam(req, "email")
	if email == "" {
		writeJSONError(w, http.StatusBadRequest, "invalid account email")
		return
	}
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	account, err := service.FindByEmail(email)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"data": account})
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
