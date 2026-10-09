package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
	"strings"
)

type ErrorMiddleware struct {
	ServiceName string
}

func NewErrorMiddleware(serviceName string) *ErrorMiddleware {
	return &ErrorMiddleware{ServiceName: strings.TrimSpace(serviceName)}
}

func (m *ErrorMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		tracked := &TrackedResponseWriter{ResponseWriter: w}
		defer func() {
			if recovered := recover(); recovered != nil {
				serviceName := "handler"
				if m != nil && m.ServiceName != "" {
					serviceName = m.ServiceName
				}
				log.Printf("%s panic %s %q: %v\n%s", serviceName, req.Method, req.URL.Path, recovered, debug.Stack())
				if !tracked.WroteHeader {
					WriteJSONError(tracked, http.StatusInternalServerError, "internal server error")
				}
			}
		}()
		next.ServeHTTP(tracked, req)
	})
}

type TrackedResponseWriter struct {
	http.ResponseWriter
	WroteHeader bool
}

func (w *TrackedResponseWriter) WriteHeader(status int) {
	if w.WroteHeader {
		return
	}
	w.WroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *TrackedResponseWriter) Write(body []byte) (int, error) {
	if !w.WroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

func (w *TrackedResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
