package routes

import (
	"backend/shared/helpers"
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
	if !helpers.RequireMethod(w, req, http.MethodGet) {
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		helpers.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	limit, err := helpers.QueryInt(req, "limit", 0)
	if err != nil {
		helpers.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	search := req.URL.Query().Get("search")

	items, err := service.FindRecentList(limit, search)
	if err != nil {
		helpers.WriteJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.WriteDataJSON(w, http.StatusOK, items)
}

func (r *UserRoutes) List(req *http.Request) (any, error) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}

	search := req.URL.Query().Get("search")
	includeAll, err := helpers.QueryBool(req, "all")
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

	payload, err := decodeUserPayload(body)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}

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
	payload, err := decodeUserPayload(body)
	if err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
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

func (r *UserRoutes) handleUserProfileRoute(w http.ResponseWriter, req *http.Request) {
	id, err := helpers.RouteInt64Param(req, "id", "user")
	if err != nil {
		helpers.WriteJSONError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	service, err := r.serviceForRequest(req)
	if err != nil {
		helpers.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	r.handleUserProfileUpdate(w, req, id, service)
}

func (r *UserRoutes) handleUserActivateRoute(w http.ResponseWriter, req *http.Request) {
	id, err := helpers.RouteInt64Param(req, "id", "user")
	if err != nil {
		helpers.WriteJSONError(w, http.StatusBadRequest, "invalid user id")
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
	if !helpers.RequireMethod(w, req, http.MethodPatch) {
		return
	}

	var payload struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := helpers.DecodeJSONRequest(req, &payload); err != nil {
		helpers.WriteJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	if err := service.UpdateCurrentProfile(id, payload.Username, payload.Email); err != nil {
		helpers.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	helpers.WriteMessageJSON(w, http.StatusOK, "profile updated")
}

func (r *UserRoutes) handleUserActivate(w http.ResponseWriter, req *http.Request, id int64) {
	if !helpers.RequireMethod(w, req, http.MethodPost) {
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		helpers.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := service.ActivateUser(id); err != nil {
		helpers.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	helpers.WriteMessageJSON(w, http.StatusOK, "activated")
}

func (r *UserRoutes) handleUserByEmail(w http.ResponseWriter, req *http.Request) {
	r.handleUserLookup(w, req, "email", "invalid user email", (*services.UserService).FindByEmail)
}

func (r *UserRoutes) handleUserByUsername(w http.ResponseWriter, req *http.Request) {
	r.handleUserLookup(w, req, "username", "invalid username", (*services.UserService).FindByUsername)
}

func (r *UserRoutes) handleUserLookup(
	w http.ResponseWriter,
	req *http.Request,
	paramName string,
	invalidValueMessage string,
	find func(*services.UserService, string) (*models.User, error),
) {
	if !helpers.RequireMethod(w, req, http.MethodGet) {
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		helpers.WriteJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	value := helpers.RouteParam(req, paramName)
	if value == "" {
		helpers.WriteJSONError(w, http.StatusBadRequest, invalidValueMessage)
		return
	}

	user, err := find(service, value)
	if err != nil {
		helpers.WriteJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	helpers.WriteDataJSON(w, http.StatusOK, user)
}

func decodeUserPayload(body []byte) (models.User, error) {
	var request userRequest
	if err := helpers.DecodeJSONBytes(body, &request); err != nil {
		return models.User{}, err
	}
	payload := request.User
	payload.UserPassword = request.Password
	return payload, nil
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
