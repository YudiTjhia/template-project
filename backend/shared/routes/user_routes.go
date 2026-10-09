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

type UserRoutes struct {
	Service *services.UserService
}

type userRequest struct {
	models.User
	Password string `json:"user_password"`
}

func NewUserRoutes(service *services.UserService) *UserRoutes {
	return &UserRoutes{Service: service}
}

func (r *UserRoutes) Register(mux *http.ServeMux) {
	mux.HandleFunc("/users", r.handleUsers)
	mux.HandleFunc("/users/", r.handleUserByID)
	mux.HandleFunc("/users/email/", r.handleUserByEmail)
	mux.HandleFunc("/users/username/", r.handleUserByUsername)
	mux.HandleFunc("/users/recent", r.handleUserRecent)
}

func (r *UserRoutes) handleUsers(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		r.handleUserList(w, req)
	case http.MethodPost:
		r.handleUserCreate(w, req)
	default:
		writeUserJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (r *UserRoutes) handleUserRecent(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeUserJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeUserJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
	search := req.URL.Query().Get("search")

	items, err := service.FindRecentList(limit, search)
	if err != nil {
		writeUserJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeUserJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (r *UserRoutes) handleUserList(w http.ResponseWriter, req *http.Request) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeUserJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	search := req.URL.Query().Get("search")
	includeAll := req.URL.Query().Get("all") == "1" || strings.EqualFold(req.URL.Query().Get("all"), "true")

	var items []models.User
	if includeAll {
		items, err = service.FindAll()
	} else {
		items, err = service.FindList(search)
	}
	if err != nil {
		writeUserJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeUserJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (r *UserRoutes) handleUserCreate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeUserJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeUserJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	var request userRequest
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		writeUserJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}
	payload := request.User
	payload.UserPassword = request.Password

	id, err := service.Create(payload)
	if err != nil {
		writeUserJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeUserJSON(w, http.StatusCreated, map[string]interface{}{"id": id})
}

func (r *UserRoutes) handleUserByID(w http.ResponseWriter, req *http.Request) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeUserJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	path := strings.TrimPrefix(req.URL.Path, "/users/")
	if path == "" {
		writeUserJSONError(w, http.StatusNotFound, "not found")
		return
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeUserJSONError(w, http.StatusNotFound, "not found")
		return
	}

	if parts[0] == "email" {
		r.handleUserByEmail(w, req)
		return
	}
	if parts[0] == "username" {
		r.handleUserByUsername(w, req)
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeUserJSONError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if len(parts) > 1 {
		switch parts[1] {
		case "profile":
			r.handleUserProfileUpdate(w, req, id, service)
		case "activate":
			r.handleUserActivate(w, req, id)
		default:
			writeUserJSONError(w, http.StatusNotFound, "not found")
		}
		return
	}

	switch req.Method {
	case http.MethodGet:
		user, err := service.FindByID(id)
		if err != nil {
			writeUserJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		writeUserJSON(w, http.StatusOK, map[string]interface{}{"data": user})
	case http.MethodPut:
		var request userRequest
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			writeUserJSONError(w, http.StatusBadRequest, "invalid json payload")
			return
		}
		payload := request.User
		payload.UserPassword = request.Password
		payload.ID = id
		if err := service.Update(payload); err != nil {
			writeUserJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeUserJSON(w, http.StatusOK, map[string]interface{}{"message": "updated"})
	case http.MethodDelete:
		if err := service.Delete(id); err != nil {
			writeUserJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeUserJSON(w, http.StatusOK, map[string]interface{}{"message": "deleted"})
	default:
		writeUserJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (r *UserRoutes) handleUserProfileUpdate(
	w http.ResponseWriter,
	req *http.Request,
	id int64,
	service *services.UserService,
) {
	if req.Method != http.MethodPatch {
		writeUserJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var payload struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeUserJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	if err := service.UpdateCurrentProfile(id, payload.Username, payload.Email); err != nil {
		writeUserJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeUserJSON(w, http.StatusOK, map[string]interface{}{"message": "profile updated"})
}

func (r *UserRoutes) handleUserActivate(w http.ResponseWriter, req *http.Request, id int64) {
	if req.Method != http.MethodPost {
		writeUserJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeUserJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := service.ActivateUser(id); err != nil {
		writeUserJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeUserJSON(w, http.StatusOK, map[string]interface{}{"message": "activated"})
}

func (r *UserRoutes) handleUserByEmail(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeUserJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeUserJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	email := strings.TrimPrefix(req.URL.Path, "/users/email/")
	email = strings.Trim(email, "/")
	if email == "" {
		writeUserJSONError(w, http.StatusBadRequest, "invalid user email")
		return
	}

	user, err := service.FindByEmail(email)
	if err != nil {
		writeUserJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	writeUserJSON(w, http.StatusOK, map[string]interface{}{"data": user})
}

func (r *UserRoutes) handleUserByUsername(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeUserJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeUserJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	username := strings.TrimPrefix(req.URL.Path, "/users/username/")
	username = strings.Trim(username, "/")
	if username == "" {
		writeUserJSONError(w, http.StatusBadRequest, "invalid username")
		return
	}

	user, err := service.FindByUsername(username)
	if err != nil {
		writeUserJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	writeUserJSON(w, http.StatusOK, map[string]interface{}{"data": user})
}

func writeUserJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeUserJSONError(w http.ResponseWriter, status int, message string) {
	writeUserJSON(w, status, map[string]string{"error": message})
}

func (r *UserRoutes) serviceForRequest(req *http.Request) (*services.UserService, error) {
	if r.Service == nil {
		return nil, errors.New("user service belum dikonfigurasi")
	}

	auth, err := resolveRequestAuth(r.Service.DB, req)
	if err != nil {
		return nil, err
	}

	return services.NewUserService(r.Service.DB, auth), nil
}
