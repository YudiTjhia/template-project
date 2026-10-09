package routes

import (
	"backend/shared/helpers"
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

type ApprovalRoutes struct {
	Service *services.ApprovalService
}

var _ interfaces.ICRUDRouteOperations = (*ApprovalRoutes)(nil)

func NewApprovalRoutes(service *services.ApprovalService) *ApprovalRoutes {
	return &ApprovalRoutes{Service: service}
}

func (r *ApprovalRoutes) Register(router chi.Router) {
	_ = impl.NewMRoutes().RegisterRoutes(loggedRouteRegistrar{router}, "/approval-rules", r)
	registerLoggedHandler(router, "/approval-requests", r.handleApprovalRequests)
	registerLoggedHandler(router, "/approval-requests/*", r.handleApprovalRequestByID)
}

func (r *ApprovalRoutes) List(req *http.Request) (any, error) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	includeAll, err := helpers.QueryBool(req, "all")
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	if includeAll {
		return service.FindAllRules()
	}
	return service.FindRules(req.URL.Query().Get("search"))
}

func (r *ApprovalRoutes) Create(req *http.Request, body []byte) (any, error) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	var payload models.ApprovalRule
	if err := helpers.DecodeJSONBytes(body, &payload); err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	id, err := service.SaveRule(payload)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	return id, nil
}

func (r *ApprovalRoutes) FindByID(req *http.Request, id int64) (any, error) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	return service.FindRuleByID(id)
}

func (r *ApprovalRoutes) Update(req *http.Request, id int64, body []byte) error {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	var payload models.ApprovalRule
	if err := helpers.DecodeJSONBytes(body, &payload); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	payload.ID = id
	if _, err := service.SaveRule(payload); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	return nil
}

func (r *ApprovalRoutes) Delete(req *http.Request, id int64) error {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	if err := service.DeleteRule(id); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	return nil
}

func (r *ApprovalRoutes) handleApprovalRequests(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		r.handleApprovalRequestList(w, req)
	default:
		writeApprovalJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (r *ApprovalRoutes) handleApprovalRequestList(w http.ResponseWriter, req *http.Request) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeApprovalJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if strings.EqualFold(req.URL.Query().Get("count"), "unread") || strings.EqualFold(req.URL.Query().Get("count"), "pending") {
		count, err := service.CountPendingRequests(req.URL.Query().Get("search"))
		if err != nil {
			writeApprovalJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeApprovalJSON(w, http.StatusOK, map[string]interface{}{"unreadCount": count})
		return
	}

	items, err := service.FindRequests(req.URL.Query().Get("search"))
	if err != nil {
		writeApprovalJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeApprovalJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (r *ApprovalRoutes) handleApprovalRequestByID(w http.ResponseWriter, req *http.Request) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeApprovalJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	path := strings.TrimPrefix(req.URL.Path, "/approval-requests/*")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeApprovalJSONError(w, http.StatusNotFound, "not found")
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeApprovalJSONError(w, http.StatusBadRequest, "invalid approval request id")
		return
	}

	if len(parts) >= 2 {
		switch parts[1] {
		case "approve":
			r.handleApprovalRequestAction(w, req, service.ApproveRequest, id)
			return
		case "reject":
			r.handleApprovalRequestAction(w, req, service.RejectRequest, id)
			return
		case "revision":
			r.handleApprovalRequestAction(w, req, service.RequestRevision, id)
			return
		case "steps":
			if req.Method != http.MethodGet {
				writeApprovalJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			items, err := service.FindRequestStepsByRequestID(id)
			if err != nil {
				writeApprovalJSONError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeApprovalJSON(w, http.StatusOK, map[string]interface{}{"data": items})
			return
		}
	}

	if req.Method != http.MethodGet {
		writeApprovalJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	item, err := service.FindRequestByID(id)
	if err != nil {
		writeApprovalJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	writeApprovalJSON(w, http.StatusOK, map[string]interface{}{"data": item})
}

func (r *ApprovalRoutes) handleApprovalRequestAction(w http.ResponseWriter, req *http.Request, action func(services.ApprovalActionInput) error, requestID int64) {
	if req.Method != http.MethodPost {
		writeApprovalJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeApprovalJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	var payload struct {
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil && err.Error() != "EOF" {
		writeApprovalJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	if err := action(services.ApprovalActionInput{RequestID: requestID, Notes: payload.Notes}); err != nil {
		writeApprovalJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	if refreshed, err := service.FindRequestByID(requestID); err == nil {
		writeApprovalJSON(w, http.StatusOK, map[string]interface{}{"data": refreshed})
		return
	}
	writeApprovalJSON(w, http.StatusOK, map[string]interface{}{"message": "ok"})
}

func writeApprovalJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeApprovalJSONError(w http.ResponseWriter, status int, message string) {
	writeApprovalJSON(w, status, map[string]string{"error": message})
}

func (r *ApprovalRoutes) serviceForRequest(req *http.Request) (*services.ApprovalService, error) {
	if r.Service == nil {
		return nil, errors.New("approval service belum dikonfigurasi")
	}
	auth, err := authFromRequest(req)
	if err != nil {
		return nil, err
	}
	return services.NewApprovalServiceWithTransactionDB(r.Service.DB, r.Service.TransactionDB, auth), nil
}
