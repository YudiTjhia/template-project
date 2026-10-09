package routes

import (
	"encoding/json"
	"errors"
	"net/http"

	"backend/shared/metadata/impl"
	"backend/shared/metadata/interfaces"
	"backend/shared/models"
	"backend/shared/services"
	"github.com/go-chi/chi/v5"
)

type UserRoutes struct {
	Service *services.UserService
}

var _ interfaces.ICRUDRouteOperations = (*UserRoutes)(nil)

type userRequest struct {
	models.User
	Password string `json:"user_password"`
}

func NewUserRoutes(service *services.UserService) *UserRoutes {
	return &UserRoutes{Service: service}
}

func (r *UserRoutes) Register(router chi.Router) error {
	if err := impl.NewMRoutes().RegisterRoutes(loggedRouteRegistrar{router}, "/users", r); err != nil {
		return err
	}
	registerLoggedHandler(router, "/users/recent", r.handleUserRecent)
	registerLoggedHandler(router, "/users/email/{email}", r.handleUserByEmail)
	registerLoggedHandler(router, "/users/username/{username}", r.handleUserByUsername)
	registerLoggedHandler(router, "/users/{id}/profile", r.handleUserProfileRoute)
	registerLoggedHandler(router, "/users/{id}/activate", r.handleUserActivateRoute)
	return nil
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

	limit, err := queryInt(req, "limit", 0)
	if err != nil {
		writeUserJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	search := req.URL.Query().Get("search")

	items, err := service.FindRecentList(limit, search)
	if err != nil {
		writeUserJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeUserJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (r *UserRoutes) List(req *http.Request) (any, error) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}

	search := req.URL.Query().Get("search")
	includeAll, err := queryBool(req, "all")
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}

	var items []models.User
	if includeAll {
		items, err = service.FindAll()
	} else {
		items, err = service.FindList(search)
	}
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *UserRoutes) Create(req *http.Request, body []byte) (any, error) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}

	var request userRequest
	if err := decodeJSONBytes(body, &request); err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	payload := request.User
	payload.UserPassword = request.Password

	id, err := service.Create(payload)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	return id, nil
}

func (r *UserRoutes) FindByID(req *http.Request, id int64) (any, error) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	user, err := service.FindByID(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRoutes) Update(req *http.Request, id int64, body []byte) error {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	var request userRequest
	if err := decodeJSONBytes(body, &request); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	payload := request.User
	payload.UserPassword = request.Password
	payload.ID = id
	if err := service.Update(payload); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	return nil
}

func (r *UserRoutes) Delete(req *http.Request, id int64) error {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	if err := service.Delete(id); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	return nil
}

func userRouteID(req *http.Request) (int64, error) {
	return routeInt64Param(req, "id", "user")
}

func (r *UserRoutes) handleUserProfileRoute(w http.ResponseWriter, req *http.Request) {
	id, err := userRouteID(req)
	if err != nil {
		writeUserJSONError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeUserJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	r.handleUserProfileUpdate(w, req, id, service)
}

func (r *UserRoutes) handleUserActivateRoute(w http.ResponseWriter, req *http.Request) {
	id, err := userRouteID(req)
	if err != nil {
		writeUserJSONError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	r.handleUserActivate(w, req, id)
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
	if err := decodeJSONRequest(req, &payload); err != nil {
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

	email := routeParam(req, "email")
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

	username := routeParam(req, "username")
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
