package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func DecodeJSONRequest(req *http.Request, destination any) error {
	if req == nil || req.Body == nil {
		return fmt.Errorf("invalid json payload")
	}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("invalid json payload")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("invalid json payload")
	}
	return nil
}

func DecodeJSONBytes(body []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("invalid json payload")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("invalid json payload")
	}
	return nil
}

func RouteParam(req *http.Request, name string) string {
	if req == nil {
		return ""
	}
	return strings.TrimSpace(chi.URLParam(req, name))
}

func RouteInt64Param(req *http.Request, name, label string) (int64, error) {
	value := RouteParam(req, name)
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid %s id", label)
	}
	return id, nil
}

func QueryInt(req *http.Request, name string, fallback int) (int, error) {
	value := strings.TrimSpace(req.URL.Query().Get(name))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s", name)
	}
	return parsed, nil
}

func QueryBool(req *http.Request, name string) (bool, error) {
	value := strings.TrimSpace(req.URL.Query().Get(name))
	if value == "" {
		return false, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid %s", name)
	}
	return parsed, nil
}

func RequireMethod(w http.ResponseWriter, req *http.Request, method string) bool {
	if req != nil && req.Method == method {
		return true
	}
	WriteJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	return false
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteJSONError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

func WriteDataJSON(w http.ResponseWriter, status int, data any) {
	WriteJSON(w, status, map[string]any{"data": data})
}

func WriteMessageJSON(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"message": message})
}
