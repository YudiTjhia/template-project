package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"backend/shared/services"
)

type NotificationRoutes struct {
	Service *services.NotificationService
}

func NewNotificationRoutes(service *services.NotificationService) *NotificationRoutes {
	return &NotificationRoutes{Service: service}
}

func (r *NotificationRoutes) Register(mux *http.ServeMux) {
	mux.HandleFunc("/notifications", r.handleNotifications)
	mux.HandleFunc("/notifications/", r.handleNotificationByID)
}

func (r *NotificationRoutes) handleNotifications(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeNotificationJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeNotificationJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if strings.EqualFold(req.URL.Query().Get("count"), "unread") {
		count, err := service.CountUnread()
		if err != nil {
			writeNotificationJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeNotificationJSON(w, http.StatusOK, map[string]interface{}{"unreadCount": count})
		return
	}

	limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
	items, err := service.FindList(req.URL.Query().Get("search"), limit)
	if err != nil {
		writeNotificationJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeNotificationJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (r *NotificationRoutes) handleNotificationByID(w http.ResponseWriter, req *http.Request) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeNotificationJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	path := strings.TrimPrefix(req.URL.Path, "/notifications/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] != "read" {
		writeNotificationJSONError(w, http.StatusNotFound, "not found")
		return
	}
	if req.Method != http.MethodPost {
		writeNotificationJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeNotificationJSONError(w, http.StatusBadRequest, "invalid notification id")
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
