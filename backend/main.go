package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"backend/shared/connections"
	"backend/shared/encryption"
	appmiddleware "backend/shared/middleware"
	"backend/shared/models"
	"backend/shared/routes"
	"backend/shared/services"
	"github.com/go-chi/chi/v5"
)

type apiResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	allowedOrigin := os.Getenv("CORS_ALLOW_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "*"
	}
	db, err := connections.NewMasterConnection()
	if err != nil {
		log.Fatalf("connect master database: %v", err)
	}
	defer db.Close()

	router := chi.NewRouter()
	router.Use(appmiddleware.LogMiddleware(log.Default()))
	router.Use(func(next http.Handler) http.Handler {
		return cors(next, allowedOrigin)
	})
	requestEncryption, err := newRequestEncryptionMiddlewareFromEnv()
	if err != nil {
		log.Fatalf("configure request encryption: %v", err)
	}
	if requestEncryption != nil {
		router.Use(func(next http.Handler) http.Handler {
			return requestEncryption.Wrap(next)
		})
		log.Printf("request encryption enabled; public key endpoint: %s", encryption.KeyEndpoint)
	} else {
		log.Printf("request encryption disabled; no private key configured")
	}
	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, apiResponse{Status: "ok"})
	})

	authMiddleware := appmiddleware.NewAuthMiddleware(func(ctx context.Context, token string, _ *http.Request) (*appmiddleware.AuthIdentity, error) {
		var session models.Auth
		err := db.GetContext(ctx, &session, `
			SELECT id, account_id, email, name, token, token_expired, is_expired
			FROM auths
			WHERE token = $1`, token)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("session tidak valid")
		}
		if err != nil {
			return nil, err
		}
		if session.IsExpired || (session.TokenExpired != nil && time.Now().UTC().After(*session.TokenExpired)) {
			return nil, errors.New("session sudah expired")
		}
		if session.AccountID <= 0 {
			return nil, errors.New("session tidak valid")
		}
		name := strings.TrimSpace(session.Name)
		if name == "" {
			name = strings.TrimSpace(session.Email)
		}
		return &appmiddleware.AuthIdentity{
			ID: session.ID, AccountID: session.AccountID, Name: name, Email: session.Email,
		}, nil
	})
	errorMiddleware := appmiddleware.NewErrorMiddleware("api")
	accountService := services.NewAccountService(db, models.Auth{})
	userService := services.NewUserService(db, models.Auth{})
	authService := services.NewAuthService(db)
	notificationService := services.NewNotificationService(db, models.Auth{})
	salesPartnerService := services.NewSalesPartnerService(db, models.Auth{})
	subscriptionService := services.NewSubscriptionService(db, models.Auth{})
	roleService := services.NewRoleService(db, models.Auth{})

	if err := accountService.Migrate(); err != nil {
		log.Fatalf("migrate accounts: %v", err)
	}
	userService.Migrate()
	authService.Migrate()
	notificationService.Migrate()
	if err := salesPartnerService.Migrate(); err != nil {
		log.Fatalf("migrate sales partners: %v", err)
	}
	subscriptionService.Migrate()
	if err := roleService.Migrate(); err != nil {
		log.Fatalf("migrate roles and role actions: %v", err)
	}

	accountRoutes := routes.NewAccountRoutes(accountService)
	authRoutes := routes.NewAuthRoutes(accountService, authService)
	authRoutes.RegisterPublic(router)
	userRoutes := routes.NewUserRoutes(userService)
	roleRoutes := routes.NewRoleRoutes(roleService)
	var routeRegistrationErr error
	router.Group(func(protected chi.Router) {
		protected.Use(func(next http.Handler) http.Handler { return errorMiddleware.Wrap(next) })
		protected.Use(func(next http.Handler) http.Handler { return authMiddleware.Wrap(next) })
		routeRegistrationErr = accountRoutes.RegisterProtected(protected)
		if routeRegistrationErr != nil {
			return
		}
		authRoutes.RegisterProtected(protected)
		routeRegistrationErr = userRoutes.Register(protected)
		if routeRegistrationErr != nil {
			return
		}
		routeRegistrationErr = roleRoutes.Register(protected)
	})
	if routeRegistrationErr != nil {
		log.Fatalf("register protected routes: %v", routeRegistrationErr)
	}

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, apiResponse{Status: "error", Message: "route not found"})
	})
	router.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusMethodNotAllowed, apiResponse{Status: "error", Message: "method not allowed"})
	})

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("backend listening on :%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

}

func newRequestEncryptionMiddlewareFromEnv() (*encryption.EncryptionMiddleware, error) {
	encodedKey := strings.TrimSpace(os.Getenv("REQUEST_ENCRYPTION_PRIVATE_KEY_B64"))
	if encodedKey == "" {
		customerKey := strings.TrimSpace(os.Getenv("CUSTOMER_ENCRYPTION_PRIVATE_KEY_B64"))
		cashierKey := strings.TrimSpace(os.Getenv("CASHIER_ENCRYPTION_PRIVATE_KEY_B64"))
		role := strings.ToLower(strings.TrimSpace(os.Getenv("REQUEST_ENCRYPTION_KEY_ROLE")))

		switch role {
		case "customer":
			encodedKey = customerKey
		case "cashier":
			encodedKey = cashierKey
		case "":
			switch {
			case customerKey != "" && cashierKey == "":
				encodedKey = customerKey
			case cashierKey != "" && customerKey == "":
				encodedKey = cashierKey
			case customerKey != "" && cashierKey != "":
				return nil, errors.New("set REQUEST_ENCRYPTION_KEY_ROLE to customer or cashier when both role keys are configured")
			}
		default:
			return nil, errors.New("REQUEST_ENCRYPTION_KEY_ROLE must be customer or cashier")
		}
		if role != "" && encodedKey == "" {
			return nil, fmt.Errorf("encryption private key for role %q is not configured", role)
		}
	}

	if encodedKey == "" {
		return nil, nil
	}
	return encryption.NewFromBase64(encodedKey)
}

func cors(next http.Handler, allowedOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigin == "*" || origin == allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func notImplemented(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, apiResponse{
		Status:  "error",
		Message: strings.TrimSpace(r.Method + " " + r.URL.Path + " API belum diimplementasikan"),
	})
}

func writeJSON(w http.ResponseWriter, status int, payload apiResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("encode response: %v", err)
	}
}
