package routes

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"backend/shared/models"
	"github.com/jmoiron/sqlx"
)

func resolveRequestAuth(db *sqlx.DB, req *http.Request) (models.Auth, error) {
	accountID, _ := strconv.ParseInt(strings.TrimSpace(req.Header.Get("X-Account-ID")), 10, 64)
	userName := strings.TrimSpace(req.Header.Get("X-User-Name"))
	userEmail := strings.TrimSpace(req.Header.Get("X-User-Email"))

	if accountID > 0 && userName != "" {
		return models.Auth{
			AccountID: accountID,
			UserName:  userName,
			Email:     userEmail,
		}, nil
	}

	token := bearerTokenAuthRequest(req.Header.Get("Authorization"))
	if token == "" {
		token = strings.TrimSpace(req.Header.Get("X-Session-Token"))
	}
	if token == "" {
		return models.Auth{}, errors.New("session buyer wajib diisi")
	}
	if db == nil {
		return models.Auth{}, errors.New("auth request belum dikonfigurasi")
	}

	auth := models.NewAuth(db)
	auth.Token = token
	record, err := auth.FindByToken()
	if err != nil {
		return models.Auth{}, err
	}
	if record.IsExpired {
		return models.Auth{}, errors.New("session sudah expired")
	}
	return models.Auth{
		AccountID: record.AccountID,
		UserName:  record.Name,
		Email:     record.Email,
	}, nil
}

func bearerTokenAuthRequest(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}

	if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return ""
	}

	return strings.TrimSpace(header[len("bearer "):])
}
