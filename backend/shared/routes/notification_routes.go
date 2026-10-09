package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"backend/shared/metadata/impl"
	"backend/shared/metadata/interfaces"
	"backend/shared/services"
	"github.com/go-chi/chi/v5"
)

type NotificationRoutes struct {
	Service *services.NotificationService
}

var _ interfaces.IListRouteOperations = (*NotificationRoutes)(nil)

func NewNotificationRoutes(service *services.NotificationService) *NotificationRoutes {
	return &NotificationRoutes{Service: service}
}

func (r *NotificationRoutes) Register(router chi.Router) {
	_ = impl.NewMRoutes().RegisterListRoute(loggedRouteRegistrar{router}, "/notifications", r)
	registerLoggedHandler(router, "/notifications/{id}/read", r.handleMarkRead)
}

func (r *NotificationRoutes) List(req *http.Request) (any, error) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	if strings.EqualFold(req.URL.Query().Get("count"), "unread") {
		count, err := service.CountUnread()
		if err != nil {
			return nil, err
		}
		return impl.RouteResponse{
			Status: http.StatusOK,
			Body:   map[string]interface{}{"unreadCount": count},
		}, nil
	}
	limit, err := queryInt(req, "limit", 0)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	return service.FindList(req.URL.Query().Get("search"), limit)
}

func (r *NotificationRoutes) handleMarkRead(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeNotificationJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id, err := routeInt64Param(req, "id", "notification")
	if err != nil {
		writeNotificationJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeNotificationJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := service.MarkRead(id); err != nil {
		writeNotificationJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeNotificationJSON(w, http.StatusOK, map[string]interface{}{"message": "read"})
}

func (r *NotificationRoutes) serviceForRequest(req *http.Request) (*services.NotificationService, error) {
	if r.Service == nil {
		return nil, errors.New("notification service belum dikonfigurasi")
	}
	auth, err := authFromRequest(req)
	if err != nil {
		return nil, err
	}
	return services.NewNotificationService(r.Service.DB, auth), nil
}

func writeNotificationJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeNotificationJSONError(w http.ResponseWriter, status int, message string) {
	writeNotificationJSON(w, status, map[string]string{"error": message})
}
