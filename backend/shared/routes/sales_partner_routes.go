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

type SalesPartnerRoutes struct{ Service *services.SalesPartnerService }

func NewSalesPartnerRoutes(service *services.SalesPartnerService) *SalesPartnerRoutes {
	return &SalesPartnerRoutes{Service: service}
}

func (r *SalesPartnerRoutes) Register(mux *http.ServeMux) {
	mux.HandleFunc("/sales-partners", r.handleListOrCreate)
	mux.HandleFunc("/sales-partners/", r.handleByPath)
}

func (r *SalesPartnerRoutes) serviceForRequest(req *http.Request) (*services.SalesPartnerService, error) {
	if r.Service == nil {
		return nil, errors.New("sales partner service belum dikonfigurasi")
	}
	auth, err := authFromRequest(req)
	if err != nil {
		return nil, err
	}
	return services.NewSalesPartnerServiceWithDatabases(r.Service.DB, r.Service.MasterDB, auth), nil
}

func (r *SalesPartnerRoutes) handleListOrCreate(w http.ResponseWriter, req *http.Request) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSalesPartnerError(w, http.StatusBadRequest, err.Error())
		return
	}
	switch req.Method {
	case http.MethodGet:
		items, err := service.FindList(req.URL.Query().Get("search"))
		if err != nil {
			writeSalesPartnerError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeSalesPartnerJSON(w, http.StatusOK, map[string]interface{}{"data": items, "commissionPercent": models.SalesPartnerCommissionPercent()})
	case http.MethodPost:
		var payload models.SalesPartner
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			writeSalesPartnerError(w, http.StatusBadRequest, "invalid json payload")
			return
		}
		id, err := service.Create(payload)
		if err != nil {
			writeSalesPartnerError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeSalesPartnerJSON(w, http.StatusCreated, map[string]interface{}{"id": id})
	default:
		writeSalesPartnerError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (r *SalesPartnerRoutes) handleByPath(w http.ResponseWriter, req *http.Request) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSalesPartnerError(w, http.StatusBadRequest, err.Error())
		return
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(req.URL.Path, "/sales-partners/"), "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeSalesPartnerError(w, http.StatusNotFound, "not found")
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeSalesPartnerError(w, http.StatusBadRequest, "invalid sales partner id")
		return
	}
	if len(parts) > 1 && parts[1] == "accounts" {
		r.handleAccounts(w, req, service, id, parts[2:])
		return
	}
	switch req.Method {
	case http.MethodGet:
		item, err := service.FindByID(id)
		if err != nil {
			writeSalesPartnerError(w, http.StatusNotFound, err.Error())
			return
		}
		writeSalesPartnerJSON(w, http.StatusOK, map[string]interface{}{"data": item})
	case http.MethodPut:
		var payload models.SalesPartner
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			writeSalesPartnerError(w, http.StatusBadRequest, "invalid json payload")
			return
		}
		payload.ID = id
		if err := service.Update(payload); err != nil {
			writeSalesPartnerError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeSalesPartnerJSON(w, http.StatusOK, map[string]interface{}{"message": "updated"})
	case http.MethodDelete:
		if err := service.Delete(id); err != nil {
			writeSalesPartnerError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeSalesPartnerJSON(w, http.StatusOK, map[string]interface{}{"message": "deleted"})
	default:
		writeSalesPartnerError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (r *SalesPartnerRoutes) handleAccounts(w http.ResponseWriter, req *http.Request, service *services.SalesPartnerService, partnerID int64, rest []string) {
	if len(rest) == 0 {
		if req.Method != http.MethodGet {
			writeSalesPartnerError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		items, err := service.FindPartnerAccounts(partnerID)
		if err != nil {
			writeSalesPartnerError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeSalesPartnerJSON(w, http.StatusOK, map[string]interface{}{"data": items})
		return
	}
	accountID, err := strconv.ParseInt(rest[0], 10, 64)
	if err != nil {
		writeSalesPartnerError(w, http.StatusBadRequest, "invalid account id")
		return
	}
	switch req.Method {
	case http.MethodPost:
		err = service.LinkAccount(partnerID, accountID)
	case http.MethodDelete:
		err = service.UnlinkAccount(partnerID, accountID)
	default:
		writeSalesPartnerError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err != nil {
		writeSalesPartnerError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeSalesPartnerJSON(w, http.StatusOK, map[string]interface{}{"message": "updated"})
}

func writeSalesPartnerJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
func writeSalesPartnerError(w http.ResponseWriter, status int, message string) {
	writeSalesPartnerJSON(w, status, map[string]string{"error": message})
}
