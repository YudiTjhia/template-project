package routes

import (
	"backend/shared/helpers"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"backend/shared/metadata/impl"
	"backend/shared/metadata/interfaces"
	"backend/shared/models"
	"backend/shared/services"
	"github.com/go-chi/chi/v5"
)

type SubscriptionRoutes struct {
	Service *services.SubscriptionService
}

var _ interfaces.ICRUDRouteOperations = (*SubscriptionRoutes)(nil)

func NewSubscriptionRoutes(service *services.SubscriptionService) *SubscriptionRoutes {
	return &SubscriptionRoutes{Service: service}
}

func (r *SubscriptionRoutes) Register(router chi.Router) {
	_ = impl.NewMRoutes().RegisterRoutes(loggedRouteRegistrar{router}, "/subscriptions", r)
	registerLoggedHandler(router, "/subscriptions/current", r.handleSubscriptionCurrent)
	registerLoggedHandler(router, "/subscriptions/status", r.handleSubscriptionStatus)
	registerLoggedHandler(router, "/subscriptions/payment-info", r.handleSubscriptionPaymentInfo)
	registerLoggedHandler(router, "/subscriptions/payments", r.handleSubscriptionPayments)
	registerLoggedHandler(router, "/subscriptions/generate-token", r.handleSubscriptionGenerateToken)
	registerLoggedHandler(router, "/subscriptions/apply-token", r.handleSubscriptionApplyToken)
	registerLoggedHandler(router, "/subscriptions/pay", r.handleSubscriptionPay)
	registerLoggedHandler(router, "/subscriptions/{id}/adjust-users", r.handleSubscriptionAdjustUsersRoute)
	registerLoggedHandler(router, "/subscriptions/{id}/histories", r.handleSubscriptionHistoriesRoute)
	registerLoggedHandler(router, "/subscription-histories", r.handleSubscriptionHistories)
}

func (r *SubscriptionRoutes) List(req *http.Request) (any, error) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	return service.FindList()
}

func (r *SubscriptionRoutes) Create(req *http.Request, body []byte) (any, error) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	var payload models.Subscription
	if err := helpers.DecodeJSONBytes(body, &payload); err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	id, err := service.Create(payload)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	return id, nil
}

func (r *SubscriptionRoutes) FindByID(req *http.Request, id int64) (any, error) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	return service.FindByID(id)
}

func (r *SubscriptionRoutes) Update(req *http.Request, id int64, body []byte) error {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	var payload models.Subscription
	if err := helpers.DecodeJSONBytes(body, &payload); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	payload.ID = id
	if err := service.Update(payload); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	return nil
}

func (r *SubscriptionRoutes) Delete(req *http.Request, id int64) error {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	if err := service.Delete(id); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	return nil
}

func (r *SubscriptionRoutes) handleSubscriptionAdjustUsersRoute(w http.ResponseWriter, req *http.Request) {
	id, err := helpers.RouteInt64Param(req, "id", "subscription")
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	r.handleSubscriptionAdjustUsers(w, req, id)
}

func (r *SubscriptionRoutes) handleSubscriptionHistoriesRoute(w http.ResponseWriter, req *http.Request) {
	id, err := helpers.RouteInt64Param(req, "id", "subscription")
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	r.handleSubscriptionHistoriesByID(w, req, id)
}

func (r *SubscriptionRoutes) handleSubscriptionCurrent(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeSubscriptionJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := service.FindCurrent()
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	writeSubscriptionJSON(w, http.StatusOK, map[string]interface{}{"data": item})
}

func (r *SubscriptionRoutes) handleSubscriptionStatus(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeSubscriptionJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	status, err := service.StatusForCurrentAccount()
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSubscriptionJSON(w, http.StatusOK, status)
}

func (r *SubscriptionRoutes) handleSubscriptionPaymentInfo(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeSubscriptionJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	info, err := service.PaymentInfo()
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSubscriptionJSON(w, http.StatusOK, map[string]interface{}{"data": info})
}

func (r *SubscriptionRoutes) handleSubscriptionPayments(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeSubscriptionJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := service.FindPayments(req.URL.Query().Get("search"))
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSubscriptionJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (r *SubscriptionRoutes) handleSubscriptionGenerateToken(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeSubscriptionJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	var payload models.SubscriptionPaymentTokenRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	result, err := service.GeneratePaymentToken(payload)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSubscriptionJSON(w, http.StatusOK, map[string]interface{}{"data": result})
}

func (r *SubscriptionRoutes) handleSubscriptionApplyToken(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeSubscriptionJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	var payload models.SubscriptionApplyTokenRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	item, err := service.ApplyPaymentToken(payload.Token)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSubscriptionJSON(w, http.StatusOK, map[string]interface{}{"data": item, "message": "subscription token applied"})
}

func (r *SubscriptionRoutes) handleSubscriptionPay(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeSubscriptionJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	var payload models.SubscriptionPaymentTokenRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	item, err := service.Pay(payload)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSubscriptionJSON(w, http.StatusOK, map[string]interface{}{"data": item, "message": "subscription payment applied"})
}

func (r *SubscriptionRoutes) handleSubscriptionAdjustUsers(w http.ResponseWriter, req *http.Request, id int64) {
	if req.Method != http.MethodPost {
		writeSubscriptionJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	var payload struct {
		NumberOfUsers int    `json:"numberOfUsers"`
		Notes         string `json:"notes"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	item, err := service.AdjustUsers(id, payload.NumberOfUsers, payload.Notes)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSubscriptionJSON(w, http.StatusOK, map[string]interface{}{"data": item})
}

func (r *SubscriptionRoutes) handleSubscriptionHistories(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeSubscriptionJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	subscriptionID, _ := strconv.ParseInt(req.URL.Query().Get("subscriptionId"), 10, 64)
	if subscriptionID <= 0 {
		subscriptionID, _ = strconv.ParseInt(req.URL.Query().Get("subscription_id"), 10, 64)
	}
	if subscriptionID <= 0 {
		writeSubscriptionJSONError(w, http.StatusBadRequest, "subscription_id wajib diisi")
		return
	}

	items, err := service.FindHistories(subscriptionID)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSubscriptionJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (r *SubscriptionRoutes) handleSubscriptionHistoriesByID(w http.ResponseWriter, req *http.Request, id int64) {
	if req.Method != http.MethodGet {
		writeSubscriptionJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := service.FindHistories(id)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSubscriptionJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (r *SubscriptionRoutes) serviceForRequest(req *http.Request) (*services.SubscriptionService, error) {
	if r.Service == nil {
		return nil, errors.New("subscription service belum dikonfigurasi")
	}

	auth, err := authFromRequest(req)
	if err != nil {
		return nil, err
	}

	return services.NewSubscriptionServiceWithDatabases(r.Service.DB, r.Service.TransactionDB, auth), nil
}

func writeSubscriptionJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeSubscriptionJSONError(w http.ResponseWriter, status int, message string) {
	writeSubscriptionJSON(w, status, map[string]string{"error": message})
}

func normalizeSubscriptionDate(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	normalized := models.StartOfDay(*value)
	return &normalized
}
