package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"backend/shared/connections"
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
	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, apiResponse{Status: "ok"})
	})
	router.Post("/account/register", notImplemented)
	router.Post("/auth/login", notImplemented)

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
	userRoutes := routes.NewUserRoutes(services.NewUserService(db, models.Auth{}))
	roleRoutes := routes.NewRoleRoutes(services.NewRoleService(db, models.Auth{}))
	var routeRegistrationErr error
	router.Group(func(protected chi.Router) {
		protected.Use(func(next http.Handler) http.Handler { return errorMiddleware.Wrap(next) })
		protected.Use(func(next http.Handler) http.Handler { return authMiddleware.Wrap(next) })
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
