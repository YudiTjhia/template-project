package middleware

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type AuthIdentity struct {
	ID        int64
	AccountID int64
	Name      string
	Email     string
	Extra     map[string]string
}

type AuthResolver func(ctx context.Context, token string, req *http.Request) (*AuthIdentity, error)

type StatusError interface {
	error
	HTTPStatus() int
}

type AuthMiddleware struct {
	Resolver          AuthResolver
	MissingTokenError string
	ConfigError       string
	SessionCookieName string
}

func NewAuthMiddleware(resolver AuthResolver) *AuthMiddleware {
	return &AuthMiddleware{
		Resolver:          resolver,
		MissingTokenError: "token sesi wajib diisi",
		ConfigError:       "auth middleware belum dikonfigurasi",
	}
}

func (m *AuthMiddleware) Wrap(next http.Handler, exemptPaths ...string) http.Handler {
	exempt := ExemptPathSet(exemptPaths)
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if IsExemptPath(req.URL.Path, exempt) {
			next.ServeHTTP(w, req)
			return
		}
		if m == nil || m.Resolver == nil {
			WriteJSONError(w, http.StatusInternalServerError, configError(m))
			return
		}

		token := BearerToken(req.Header.Get("Authorization"))
		if token == "" {
			token = strings.TrimSpace(req.Header.Get("X-Session-Token"))
		}
		if token == "" && strings.TrimSpace(m.SessionCookieName) != "" {
			if cookie, err := req.Cookie(m.SessionCookieName); err == nil {
				token = strings.TrimSpace(cookie.Value)
			}
		}
		if token == "" {
			WriteJSONError(w, http.StatusUnauthorized, missingTokenError(m))
			return
		}

		identity, err := m.Resolver(req.Context(), token, req)
		if err != nil {
			status := http.StatusUnauthorized
			if errors.Is(err, ErrAuthConfig) {
				status = http.StatusInternalServerError
			} else if statusErr, ok := err.(StatusError); ok && statusErr.HTTPStatus() > 0 {
				status = statusErr.HTTPStatus()
			}
			WriteJSONError(w, status, err.Error())
			return
		}
		if identity == nil {
			WriteJSONError(w, http.StatusUnauthorized, "session tidak valid")
			return
		}

		req = req.Clone(req.Context())
		if identity.ID > 0 {
			req.Header.Set("X-Auth-ID", strconv.FormatInt(identity.ID, 10))
		}
		if identity.AccountID > 0 {
			req.Header.Set("X-Account-ID", strconv.FormatInt(identity.AccountID, 10))
		}
		req.Header.Set("X-User-Name", identity.Name)
		req.Header.Set("X-User-Email", identity.Email)
		for key, value := range identity.Extra {
			req.Header.Set(key, value)
		}

		next.ServeHTTP(w, req)
	})
}

var ErrAuthConfig = errors.New("auth middleware belum dikonfigurasi")

func configError(m *AuthMiddleware) string {
	if m != nil && strings.TrimSpace(m.ConfigError) != "" {
		return m.ConfigError
	}
	return ErrAuthConfig.Error()
}

func missingTokenError(m *AuthMiddleware) string {
	if m != nil && strings.TrimSpace(m.MissingTokenError) != "" {
		return m.MissingTokenError
	}
	return "token sesi wajib diisi"
}
