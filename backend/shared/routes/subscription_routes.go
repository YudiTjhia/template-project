package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend/shared/models"
	"backend/shared/services"
)

type SubscriptionRoutes struct {
	Service *services.SubscriptionService
}

func NewSubscriptionRoutes(service *services.SubscriptionService) *SubscriptionRoutes {
	return &SubscriptionRoutes{Service: service}
}

func (r *SubscriptionRoutes) Register(mux *http.ServeMux) {
	mux.HandleFunc("/subscriptions", r.handleSubscriptions)
	mux.HandleFunc("/subscriptions/", r.handleSubscriptionByID)
	mux.HandleFunc("/subscription-histories", r.handleSubscriptionHistories)
}

func (r *SubscriptionRoutes) handleSubscriptions(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		r.handleSubscriptionList(w, req)
	case http.MethodPost:
		r.handleSubscriptionCreate(w, req)
	default:
		writeSubscriptionJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (r *SubscriptionRoutes) handleSubscriptionList(w http.ResponseWriter, req *http.Request) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := service.FindList()
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSubscriptionJSON(w, http.StatusOK, map[string]interface{}{"data": items})
}

func (r *SubscriptionRoutes) handleSubscriptionCreate(w http.ResponseWriter, req *http.Request) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	var payload models.Subscription
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	id, err := service.Create(payload)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSubscriptionJSON(w, http.StatusCreated, map[string]interface{}{"id": id})
}

func (r *SubscriptionRoutes) handleSubscriptionByID(w http.ResponseWriter, req *http.Request) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	path := strings.TrimPrefix(req.URL.Path, "/subscriptions/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeSubscriptionJSONError(w, http.StatusNotFound, "not found")
		return
	}

	if parts[0] == "current" {
		r.handleSubscriptionCurrent(w, req)
		return
	}
	if parts[0] == "status" {
		r.handleSubscriptionStatus(w, req)
		return
	}
	if parts[0] == "payment-info" {
		r.handleSubscriptionPaymentInfo(w, req)
		return
	}
	if parts[0] == "payments" {
		r.handleSubscriptionPayments(w, req)
		return
	}
	if parts[0] == "generate-token" {
		r.handleSubscriptionGenerateToken(w, req)
		return
	}
	if parts[0] == "apply-token" {
		r.handleSubscriptionApplyToken(w, req)
		return
	}
	if parts[0] == "pay" {
		r.handleSubscriptionPay(w, req)
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeSubscriptionJSONError(w, http.StatusBadRequest, "invalid subscription id")
		return
	}

	if len(parts) > 1 {
		switch parts[1] {
		case "adjust-users":
			r.handleSubscriptionAdjustUsers(w, req, id)
		case "histories":
			r.handleSubscriptionHistoriesByID(w, req, id)
		default:
			writeSubscriptionJSONError(w, http.StatusNotFound, "not found")
		}
		return
	}

	switch req.Method {
	case http.MethodGet:
		item, err := service.FindByID(id)
		if err != nil {
			writeSubscriptionJSONError(w, http.StatusNotFound, err.Error())
			return
		}
		writeSubscriptionJSON(w, http.StatusOK, map[string]interface{}{"data": item})
	case http.MethodPut:
		var payload models.Subscription
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			writeSubscriptionJSONError(w, http.StatusBadRequest, "invalid json payload")
			return
		}
		payload.ID = id
		if err := service.Update(payload); err != nil {
			writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeSubscriptionJSON(w, http.StatusOK, map[string]interface{}{"message": "updated"})
	case http.MethodDelete:
		if err := service.Delete(id); err != nil {
			writeSubscriptionJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeSubscriptionJSON(w, http.StatusOK, map[string]interface{}{"message": "deleted"})
	default:
		writeSubscriptionJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
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
