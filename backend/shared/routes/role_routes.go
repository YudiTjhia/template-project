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

type RoleRoutes struct {
	Service *services.RoleService
}

var _ interfaces.ICRUDRouteOperations = (*RoleRoutes)(nil)

func NewRoleRoutes(service *services.RoleService) *RoleRoutes {
	return &RoleRoutes{Service: service}
}

func (r *RoleRoutes) Register(router chi.Router) error {
	if err := impl.NewMRoutes().RegisterRoutes(loggedRouteRegistrar{router}, "/roles", r); err != nil {
		return err
	}
	if err := impl.NewMRoutes().RegisterRoutes(loggedRouteRegistrar{router}, "/roles/actions", roleActionRouteOperations{routes: r}); err != nil {
		return err
	}
	registerLoggedHandler(router, "/roles/recent", r.handleRoleRecent)
	registerLoggedHandler(router, "/roles/name/{name}", r.handleRoleByName)
	return nil
}

type roleActionRouteOperations struct{ routes *RoleRoutes }

var _ interfaces.ICRUDRouteOperations = roleActionRouteOperations{}

func (o roleActionRouteOperations) List(req *http.Request) (any, error) {
	service, err := o.routes.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	includeAll, err := queryBool(req, "all")
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	search := req.URL.Query().Get("search")
	if includeAll {
		return service.FindActionAll(search)
	}
	return service.FindActionList(search)
}

func (o roleActionRouteOperations) Create(req *http.Request, body []byte) (any, error) {
	service, err := o.routes.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	var payload models.RoleAction
	if err := decodeJSONBytes(body, &payload); err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	id, err := service.CreateAction(payload)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	return id, nil
}

func (o roleActionRouteOperations) FindByID(req *http.Request, id int64) (any, error) {
	service, err := o.routes.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	item, err := service.FindActionByID(id)
	if err != nil {
		return nil, err
	}
	return *item, nil
}

func (o roleActionRouteOperations) Update(req *http.Request, id int64, body []byte) error {
	service, err := o.routes.serviceForRequest(req)
	if err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	var payload models.RoleAction
	if err := decodeJSONBytes(body, &payload); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	payload.ID = id
	if err := service.UpdateAction(payload); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	return nil
}

func (o roleActionRouteOperations) Delete(req *http.Request, id int64) error {
	service, err := o.routes.serviceForRequest(req)
	if err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	if err := service.DeleteAction(id); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	return nil
}

func (r *RoleRoutes) List(req *http.Request) (any, error) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	includeAll, err := queryBool(req, "all")
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	if includeAll {
		items, err := service.FindAll()
		return items, err
	}
	items, err := service.FindList(req.URL.Query().Get("search"))
	return items, err
}

func (r *RoleRoutes) Create(req *http.Request, body []byte) (any, error) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	var payload models.Role
	if err := decodeJSONBytes(body, &payload); err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	id, err := service.Create(payload)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	return id, nil
}

func (r *RoleRoutes) FindByID(req *http.Request, id int64) (any, error) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	return service.FindByID(id)
}

func (r *RoleRoutes) Update(req *http.Request, id int64, body []byte) error {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	var payload models.Role
	if err := decodeJSONBytes(body, &payload); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	payload.ID = id
	if err := service.Update(payload); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	return nil
}

func (r *RoleRoutes) Delete(req *http.Request, id int64) error {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	if err := service.Delete(id); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	return nil
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
	limit, err := queryInt(req, "limit", 0)
	if err != nil {
		writeRoleJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, err := service.FindRecentList(limit, req.URL.Query().Get("search"))
	if err != nil {
		writeRoleJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeRoleJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (r *RoleRoutes) handleRoleByName(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeRoleJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	name := routeParam(req, "name")
	if name == "" {
		writeRoleJSONError(w, http.StatusBadRequest, "invalid role name")
		return
	}
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeRoleJSONError(w, http.StatusBadRequest, err.Error())
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
