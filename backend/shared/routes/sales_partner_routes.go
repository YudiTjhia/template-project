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

type SalesPartnerRoutes struct{ Service *services.SalesPartnerService }

var _ interfaces.ICRUDRouteOperations = (*SalesPartnerRoutes)(nil)

func NewSalesPartnerRoutes(service *services.SalesPartnerService) *SalesPartnerRoutes {
	return &SalesPartnerRoutes{Service: service}
}

func (r *SalesPartnerRoutes) Register(router chi.Router) {
	_ = impl.NewMRoutes().RegisterRoutes(loggedRouteRegistrar{router}, "/sales-partners", r)
	registerLoggedHandler(router, "/sales-partners/{id}/accounts", r.handleAccountsRoute)
	registerLoggedHandler(router, "/sales-partners/{id}/accounts/{accountID}", r.handleAccountsRoute)
}

func (r *SalesPartnerRoutes) List(req *http.Request) (any, error) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	items, err := service.FindList(req.URL.Query().Get("search"))
	if err != nil {
		return nil, err
	}
	return impl.RouteResponse{
		Status: http.StatusOK,
		Body: map[string]interface{}{
			"data": items, "commissionPercent": models.SalesPartnerCommissionPercent(),
		},
	}, nil
}

func (r *SalesPartnerRoutes) Create(req *http.Request, body []byte) (any, error) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	var payload models.SalesPartner
	if err := decodeJSONBytes(body, &payload); err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	return service.Create(payload)
}

func (r *SalesPartnerRoutes) FindByID(req *http.Request, id int64) (any, error) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return nil, impl.NewRouteError(http.StatusBadRequest, err)
	}
	return service.FindByID(id)
}

func (r *SalesPartnerRoutes) Update(req *http.Request, id int64, body []byte) error {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	var payload models.SalesPartner
	if err := decodeJSONBytes(body, &payload); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	payload.ID = id
	if err := service.Update(payload); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	return nil
}

func (r *SalesPartnerRoutes) Delete(req *http.Request, id int64) error {
	service, err := r.serviceForRequest(req)
	if err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	if err := service.Delete(id); err != nil {
		return impl.NewRouteError(http.StatusBadRequest, err)
	}
	return nil
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

func (r *SalesPartnerRoutes) handleAccountsRoute(w http.ResponseWriter, req *http.Request) {
	service, err := r.serviceForRequest(req)
	if err != nil {
		writeSalesPartnerError(w, http.StatusBadRequest, err.Error())
		return
	}
	partnerID, err := routeInt64Param(req, "id", "sales partner")
	if err != nil {
		writeSalesPartnerError(w, http.StatusBadRequest, err.Error())
		return
	}
	accountID := int64(0)
	if routeParam(req, "accountID") != "" {
		accountID, err = routeInt64Param(req, "accountID", "account")
		if err != nil {
			writeSalesPartnerError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	r.handleAccounts(w, req, service, partnerID, accountID)
}

func (r *SalesPartnerRoutes) handleAccounts(w http.ResponseWriter, req *http.Request, service *services.SalesPartnerService, partnerID, accountID int64) {
	if accountID == 0 {
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
	var err error
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
