package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type requestLogContextKey struct{}

// LogMiddleware logs each incoming request and its completed response once.
// It records the matched route pattern rather than path values or query data.
func LogMiddleware(logger *log.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = log.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Context().Value(requestLogContextKey{}) == true {
				next.ServeHTTP(w, req)
				return
			}

			started := time.Now()
			ctx := context.WithValue(req.Context(), requestLogContextKey{}, true)
			req = req.WithContext(ctx)
			response := chimiddleware.NewWrapResponseWriter(w, req.ProtoMajor)
			next.ServeHTTP(response, req)

			status := response.Status()
			if status == 0 {
				status = http.StatusOK
			}
			var route string
			if routeContext := chi.RouteContext(req.Context()); routeContext != nil {
				route = routeContext.RoutePattern()
			}
			if route == "" {
				route = "<unmatched>"
			}
			logger.Printf("http request method=%s route=%q status=%d bytes=%d duration=%s",
				req.Method,
				route,
				status,
				response.BytesWritten(),
				time.Since(started),
			)
		})
	}
}
