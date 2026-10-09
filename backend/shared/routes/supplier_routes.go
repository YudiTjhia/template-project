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

type SupplierRoutes struct {
	Service *services.SupplierService
}

func NewSupplierRoutes(service *services.SupplierService) *SupplierRoutes {
	return &SupplierRoutes{Service: service}
}

func (r *SupplierRoutes) Register(mux *http.ServeMux) {
	mux.HandleFunc("/suppliers", r.handleSuppliers)
	mux.HandleFunc("/suppliers/", r.handleSupplierByID)
	mux.HandleFunc("/suppliers/recent", r.handleSupplierRecent)
	mux.HandleFunc("/suppliers/code/", r.handleSupplierByCode)
	mux.HandleFunc("/suppliers/email/", r.handleSupplierByEmail)
	mux.HandleFunc("/suppliers/phone-number/", r.handleSupplierByPhoneNumber)
}

func (r *SupplierRoutes) handleSuppliers(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		r.handleSupplierList(w, req)
	case http.MethodPost:
		r.handleSupplierCreate(w, req)
	default:
		writeSupplierJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (r *SupplierRoutes) handleSupplierRecent(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeSupplierJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSupplierJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
	search := req.URL.Query().Get("search")
	items, err := service.FindRecentList(limit, search)
	if err != nil {
		writeSupplierJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSupplierJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (r *SupplierRoutes) handleSupplierList(w http.ResponseWriter, req *http.Request) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSupplierJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	search := req.URL.Query().Get("search")
	includeAll := req.URL.Query().Get("all") == "1" || strings.EqualFold(req.URL.Query().Get("all"), "true")

	var items []models.Supplier
	if includeAll {
		items, err = service.FindAll()
	} else {
		items, err = service.FindList(search)
	}
	if err != nil {
		writeSupplierJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSupplierJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (r *SupplierRoutes) handleSupplierCreate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeSupplierJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSupplierJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	var payload models.Supplier
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeSupplierJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	id, err := service.Create(payload)
	if err != nil {
		writeSupplierJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSupplierJSON(w, http.StatusCreated, map[string]interface{}{"id": id})
}

func (r *SupplierRoutes) handleSupplierByID(w http.ResponseWriter, req *http.Request) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSupplierJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	path := strings.TrimPrefix(req.URL.Path, "/suppliers/")
	if path == "" {
		writeSupplierJSONError(w, http.StatusNotFound, "not found")
		return
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeSupplierJSONError(w, http.StatusNotFound, "not found")
		return
	}

	switch parts[0] {
	case "recent":
		r.handleSupplierRecent(w, req)
		return
	case "code":
		r.handleSupplierByCode(w, req)
		return
	case "email":
		r.handleSupplierByEmail(w, req)
		return
	case "phone-number":
		r.handleSupplierByPhoneNumber(w, req)
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeSupplierJSONError(w, http.StatusBadRequest, "invalid supplier id")
		return
	}

	switch req.Method {
	case http.MethodGet:
		item, err := service.FindByID(id)
		if err != nil {
			writeSupplierJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		writeSupplierJSON(w, http.StatusOK, map[string]interface{}{"data": item})
	case http.MethodPut:
		var payload models.Supplier
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			writeSupplierJSONError(w, http.StatusBadRequest, "invalid json payload")
			return
		}
		payload.ID = id
		if err := service.Update(payload); err != nil {
			writeSupplierJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeSupplierJSON(w, http.StatusOK, map[string]interface{}{"message": "updated"})
	case http.MethodDelete:
		if err := service.Delete(id); err != nil {
			writeSupplierJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeSupplierJSON(w, http.StatusOK, map[string]interface{}{"message": "deleted"})
	default:
		writeSupplierJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (r *SupplierRoutes) handleSupplierByCode(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeSupplierJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSupplierJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	code := strings.TrimPrefix(req.URL.Path, "/suppliers/code/")
	code = strings.Trim(code, "/")
	if code == "" {
		writeSupplierJSONError(w, http.StatusBadRequest, "invalid supplier code")
		return
	}

	item, err := service.FindByCode(code)
	if err != nil {
		writeSupplierJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	writeSupplierJSON(w, http.StatusOK, map[string]interface{}{"data": item})
}

func (r *SupplierRoutes) handleSupplierByEmail(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeSupplierJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSupplierJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	email := strings.TrimPrefix(req.URL.Path, "/suppliers/email/")
	email = strings.Trim(email, "/")
	if email == "" {
		writeSupplierJSONError(w, http.StatusBadRequest, "invalid supplier email")
		return
	}

	item, err := service.FindByEmail(email)
	if err != nil {
		writeSupplierJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	writeSupplierJSON(w, http.StatusOK, map[string]interface{}{"data": item})
}

func (r *SupplierRoutes) handleSupplierByPhoneNumber(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeSupplierJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSupplierJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	phoneNumber := strings.TrimPrefix(req.URL.Path, "/suppliers/phone-number/")
	phoneNumber = strings.Trim(phoneNumber, "/")
	if phoneNumber == "" {
		writeSupplierJSONError(w, http.StatusBadRequest, "invalid supplier phone number")
		return
	}

	item, err := service.FindByPhoneNumber(phoneNumber)
	if err != nil {
		writeSupplierJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	writeSupplierJSON(w, http.StatusOK, map[string]interface{}{"data": item})
}

func writeSupplierJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeSupplierJSONError(w http.ResponseWriter, status int, message string) {
	writeSupplierJSON(w, status, map[string]string{"error": message})
}

func (r *SupplierRoutes) serviceForRequest(req *http.Request) (*services.SupplierService, error) {
	if r.Service == nil {
		return nil, errors.New("supplier service belum dikonfigurasi")
	}

	auth, err := authFromRequest(req)
	if err != nil {
		return nil, err
	}

	return services.NewSupplierService(r.Service.DB, auth), nil
}
