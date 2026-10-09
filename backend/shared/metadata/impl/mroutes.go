package impl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"backend/shared/metadata/interfaces"
)

// MRoutes provides the standard HTTP behavior for collection and item CRUD.
// Domain-specific work is delegated to ICRUDRouteOperations.
type MRoutes struct{}

// RouteResponse lets a route operation return a special top-level payload
// while still using the shared list handler.
type RouteResponse struct {
	Status int
	Body   any
}

type RouteError struct {
	Status int
	Err    error
}

func (e *RouteError) Error() string   { return e.Err.Error() }
func (e *RouteError) Unwrap() error   { return e.Err }
func (e *RouteError) HTTPStatus() int { return e.Status }

func NewRouteError(status int, err error) error {
	if err == nil {
		return nil
	}
	return &RouteError{Status: status, Err: err}
}

func NewMRoutes() *MRoutes { return &MRoutes{} }

func (m *MRoutes) RegisterRoutes(router interfaces.IRouteRegistrar, basePath string, operations interfaces.ICRUDRouteOperations) error {
	if err := validateRouteRegistration(router, basePath, operations); err != nil {
		return err
	}
	basePath = normalizeRouteBasePath(basePath)
	m.registerCollectionRoutes(router, basePath, operations)
	m.registerItemRoutes(router, basePath, operations)
	return nil
}

// RegisterItemRoutes registers the standard operations on /{id} while leaving
// collection routes available for domain-specific responses and behavior.
func (m *MRoutes) RegisterItemRoutes(router interfaces.IRouteRegistrar, basePath string, operations interfaces.ICRUDRouteOperations) error {
	if err := validateRouteRegistration(router, basePath, operations); err != nil {
		return err
	}
	m.registerItemRoutes(router, normalizeRouteBasePath(basePath), operations)
	return nil
}

func (m *MRoutes) RegisterListRoute(router interfaces.IRouteRegistrar, basePath string, operations interfaces.IListRouteOperations) error {
	if router == nil {
		return errors.New("route registrar is nil")
	}
	if operations == nil {
		return errors.New("list route operations are nil")
	}
	basePath = normalizeRouteBasePath(basePath)
	if basePath == "/" {
		return errors.New("route base path is required")
	}
	router.Method(http.MethodGet, basePath, http.HandlerFunc(m.list(operations)))
	return nil
}

func validateRouteRegistration(router interfaces.IRouteRegistrar, basePath string, operations interfaces.ICRUDRouteOperations) error {
	if router == nil {
		return errors.New("route registrar is nil")
	}
	if operations == nil {
		return errors.New("CRUD route operations are nil")
	}
	basePath = "/" + strings.Trim(strings.TrimSpace(basePath), "/")
	if basePath == "/" {
		return errors.New("route base path is required")
	}
	return nil
}

func normalizeRouteBasePath(basePath string) string {
	return "/" + strings.Trim(strings.TrimSpace(basePath), "/")
}

func (m *MRoutes) registerCollectionRoutes(router interfaces.IRouteRegistrar, basePath string, operations interfaces.ICRUDRouteOperations) {
	router.Method(http.MethodGet, basePath, http.HandlerFunc(m.list(operations)))
	router.Method(http.MethodPost, basePath, http.HandlerFunc(m.create(operations)))
}

func (m *MRoutes) registerItemRoutes(router interfaces.IRouteRegistrar, basePath string, operations interfaces.ICRUDRouteOperations) {
	itemPath := fmt.Sprintf("%s/{id}", basePath)
	router.Method(http.MethodGet, itemPath, http.HandlerFunc(m.findByID(basePath, operations)))
	router.Method(http.MethodPut, itemPath, http.HandlerFunc(m.update(basePath, operations)))
	router.Method(http.MethodDelete, itemPath, http.HandlerFunc(m.delete(basePath, operations)))
}

func (m *MRoutes) list(operations interfaces.IListRouteOperations) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		result, err := operations.List(req)
		if err != nil {
			writeRouteError(w, http.StatusInternalServerError, err)
			return
		}
		if response, ok := result.(RouteResponse); ok {
			status := response.Status
			if status == 0 {
				status = http.StatusOK
			}
			writeRouteJSON(w, status, response.Body)
			return
		}
		writeRouteJSON(w, http.StatusOK, map[string]any{"data": result})
	}
}

func (m *MRoutes) create(operations interfaces.ICRUDRouteOperations) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			writeRouteError(w, http.StatusBadRequest, err)
			return
		}
		id, err := operations.Create(req, body)
		if err != nil {
			writeRouteError(w, http.StatusBadRequest, err)
			return
		}
		writeRouteJSON(w, http.StatusCreated, map[string]any{"id": id})
	}
}

func (m *MRoutes) findByID(basePath string, operations interfaces.ICRUDRouteOperations) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		id, err := routeItemID(req, basePath)
		if err != nil {
			writeRouteError(w, http.StatusBadRequest, err)
			return
		}
		result, err := operations.FindByID(req, id)
		if err != nil {
			writeRouteError(w, http.StatusNotFound, err)
			return
		}
		writeRouteJSON(w, http.StatusOK, map[string]any{"data": result})
	}
}

func (m *MRoutes) update(basePath string, operations interfaces.ICRUDRouteOperations) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		id, err := routeItemID(req, basePath)
		if err != nil {
			writeRouteError(w, http.StatusBadRequest, err)
			return
		}
		body, err := io.ReadAll(req.Body)
		if err == nil {
			err = operations.Update(req, id, body)
		}
		if err != nil {
			writeRouteError(w, http.StatusBadRequest, err)
			return
		}
		writeRouteJSON(w, http.StatusOK, map[string]any{"message": "updated"})
	}
}

func (m *MRoutes) delete(basePath string, operations interfaces.ICRUDRouteOperations) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		id, err := routeItemID(req, basePath)
		if err == nil {
			err = operations.Delete(req, id)
		}
		if err != nil {
			writeRouteError(w, http.StatusBadRequest, err)
			return
		}
		writeRouteJSON(w, http.StatusOK, map[string]any{"message": "deleted"})
	}
}

func routeItemID(req *http.Request, basePath string) (int64, error) {
	path := strings.Trim(strings.TrimPrefix(req.URL.Path, basePath+"/"), "/")
	if path == "" || strings.Contains(path, "/") {
		return 0, errors.New("invalid item ID path")
	}
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid item ID")
	}
	return id, nil
}

func writeRouteError(w http.ResponseWriter, fallbackStatus int, err error) {
	status := fallbackStatus
	var routeErr interfaces.IRouteError
	if errors.As(err, &routeErr) && routeErr.HTTPStatus() >= 400 && routeErr.HTTPStatus() <= 599 {
		status = routeErr.HTTPStatus()
	}
	writeRouteJSON(w, status, map[string]string{"error": err.Error()})
}

func writeRouteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

var _ interfaces.IMRoutes = (*MRoutes)(nil)
