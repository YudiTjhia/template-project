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

type RoleRoutes struct {
	Service *services.RoleService
}

func NewRoleRoutes(service *services.RoleService) *RoleRoutes {
	return &RoleRoutes{Service: service}
}

func (r *RoleRoutes) Register(mux *http.ServeMux) {
	mux.HandleFunc("/roles/actions", r.handleRoleActions)
	mux.HandleFunc("/roles/actions/", r.handleRoleActionByID)
	mux.HandleFunc("/roles", r.handleRoles)
	mux.HandleFunc("/roles/", r.handleRoleByID)
	mux.HandleFunc("/roles/recent", r.handleRoleRecent)
	mux.HandleFunc("/roles/name/", r.handleRoleByName)
}

func (r *RoleRoutes) handleRoleActions(w http.ResponseWriter, req *http.Request) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeRoleJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	switch req.Method {
	case http.MethodGet:
		search := req.URL.Query().Get("search")
		includeAll := req.URL.Query().Get("all") == "1" || strings.EqualFold(req.URL.Query().Get("all"), "true")

		var items []models.RoleAction
		if includeAll {
			items, err = service.FindActionAll(search)
		} else {
			items, err = service.FindActionList(search)
		}
		if err != nil {
			writeRoleJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeRoleJSON(w, http.StatusOK, map[string]interface{}{"data": serializeRoleActions(items)})
	case http.MethodPost:
		var payload models.RoleAction
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			writeRoleJSONError(w, http.StatusBadRequest, "invalid json payload")
			return
		}

		id, err := service.CreateAction(payload)
		if err != nil {
			writeRoleJSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeRoleJSON(w, http.StatusCreated, map[string]interface{}{"id": id})
	default:
		writeRoleJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (r *RoleRoutes) handleRoleActionByID(w http.ResponseWriter, req *http.Request) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeRoleJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	path := strings.TrimPrefix(req.URL.Path, "/roles/actions/")
	path = strings.Trim(path, "/")
	if path == "" {
		writeRoleJSONError(w, http.StatusNotFound, "not found")
		return
	}

	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		writeRoleJSONError(w, http.StatusBadRequest, "invalid role action id")
		return
	}

	switch req.Method {
	case http.MethodGet:
		item, err := service.FindActionByID(id)
		if err != nil {
			writeRoleJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		writeRoleJSON(w, http.StatusOK, map[string]interface{}{"data": serializeRoleAction(*item)})
	case http.MethodPut:
		var payload models.RoleAction
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			writeRoleJSONError(w, http.StatusBadRequest, "invalid json payload")
			return
		}
		payload.ID = id
		if err := service.UpdateAction(payload); err != nil {
			writeRoleJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeRoleJSON(w, http.StatusOK, map[string]interface{}{"message": "updated"})
	case http.MethodDelete:
		if err := service.DeleteAction(id); err != nil {
			writeRoleJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeRoleJSON(w, http.StatusOK, map[string]interface{}{"message": "deleted"})
	default:
		writeRoleJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

type roleActionResponse struct {
	ID        int64  `json:"id"`
	Value     string `json:"value"`
	Label     string `json:"label"`
	SortOrder int    `json:"sort_order"`
	Active    bool   `json:"active"`
}

func serializeRoleAction(item models.RoleAction) roleActionResponse {
	return roleActionResponse{
		ID:        item.ID,
		Value:     item.Code,
		Label:     item.Label,
		SortOrder: item.SortOrder,
		Active:    item.Active,
	}
}

func serializeRoleActions(items []models.RoleAction) []roleActionResponse {
	response := make([]roleActionResponse, 0, len(items))
	for _, item := range items {
		response = append(response, serializeRoleAction(item))
	}
	return response
}

func (r *RoleRoutes) handleRoles(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		r.handleRoleList(w, req)
	case http.MethodPost:
		r.handleRoleCreate(w, req)
	default:
		writeRoleJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (r *RoleRoutes) handleRoleRecent(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeRoleJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeRoleJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
	search := req.URL.Query().Get("search")
	items, err := service.FindRecentList(limit, search)
	if err != nil {
		writeRoleJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeRoleJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (r *RoleRoutes) handleRoleList(w http.ResponseWriter, req *http.Request) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeRoleJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	search := req.URL.Query().Get("search")
	includeAll := req.URL.Query().Get("all") == "1" || strings.EqualFold(req.URL.Query().Get("all"), "true")

	var items []models.Role
	if includeAll {
		items, err = service.FindAll()
	} else {
		items, err = service.FindList(search)
	}
	if err != nil {
		writeRoleJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeRoleJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (r *RoleRoutes) handleRoleCreate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeRoleJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeRoleJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	var payload models.Role
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeRoleJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	id, err := service.Create(payload)
	if err != nil {
		writeRoleJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeRoleJSON(w, http.StatusCreated, map[string]interface{}{"id": id})
}

func (r *RoleRoutes) handleRoleByID(w http.ResponseWriter, req *http.Request) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeRoleJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	path := strings.TrimPrefix(req.URL.Path, "/roles/")
	if path == "" {
		writeRoleJSONError(w, http.StatusNotFound, "not found")
		return
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeRoleJSONError(w, http.StatusNotFound, "not found")
		return
	}

	if parts[0] == "name" {
		r.handleRoleByName(w, req)
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeRoleJSONError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	switch req.Method {
	case http.MethodGet:
		item, err := service.FindByID(id)
		if err != nil {
			writeRoleJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		writeRoleJSON(w, http.StatusOK, map[string]interface{}{"data": item})
	case http.MethodPut:
		var payload models.Role
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			writeRoleJSONError(w, http.StatusBadRequest, "invalid json payload")
			return
		}
		payload.ID = id
		if err := service.Update(payload); err != nil {
			writeRoleJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeRoleJSON(w, http.StatusOK, map[string]interface{}{"message": "updated"})
	case http.MethodDelete:
		if err := service.Delete(id); err != nil {
			writeRoleJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeRoleJSON(w, http.StatusOK, map[string]interface{}{"message": "deleted"})
	default:
		writeRoleJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (r *RoleRoutes) handleRoleByName(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeRoleJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeRoleJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	name := strings.TrimPrefix(req.URL.Path, "/roles/name/")
	name = strings.Trim(name, "/")
	if name == "" {
		writeRoleJSONError(w, http.StatusBadRequest, "invalid role name")
		return
	}

	item, err := service.FindByName(name)
	if err != nil {
		writeRoleJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	writeRoleJSON(w, http.StatusOK, map[string]interface{}{"data": item})
}

func writeRoleJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeRoleJSONError(w http.ResponseWriter, status int, message string) {
	writeRoleJSON(w, status, map[string]string{"error": message})
}

func (r *RoleRoutes) serviceForRequest(req *http.Request) (*services.RoleService, error) {
	if r.Service == nil {
		return nil, errors.New("role service belum dikonfigurasi")
	}

	auth, err := resolveRequestAuth(r.Service.DB, req)
	if err != nil {
		return nil, err
	}

	return services.NewRoleService(r.Service.DB, auth), nil
}
