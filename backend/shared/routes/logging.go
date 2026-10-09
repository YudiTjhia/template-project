package routes

import (
	"log"
	"net/http"

	"backend/shared/metadata/interfaces"
	requestmiddleware "backend/shared/middleware"
	"github.com/go-chi/chi/v5"
)

// registerLoggedHandler adds a route with access logging. The middleware is
// idempotent, so applications can also install it once on the parent router.
func registerLoggedHandler(router chi.Router, pattern string, handler http.HandlerFunc) {
	router.Handle(pattern, requestmiddleware.LogMiddleware(log.Default())(handler))
}

type loggedRouteRegistrar struct{ router chi.Router }

func (r loggedRouteRegistrar) Method(method, pattern string, handler http.Handler) {
	r.loggedHandler(method, pattern, handler)
}

func (r loggedRouteRegistrar) loggedHandler(method, pattern string, handler http.Handler) {
	r.router.Method(method, pattern, requestmiddleware.LogMiddleware(log.Default())(handler))
}

var _ interfaces.IRouteRegistrar = loggedRouteRegistrar{}
