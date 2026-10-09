package routes

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

func decodeJSONRequest(req *http.Request, destination any) error {
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

func decodeJSONBytes(body []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("invalid json payload")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("invalid json payload")
	}
	return nil
}

func routeParam(req *http.Request, name string) string {
	if req == nil {
		return ""
	}
	return strings.TrimSpace(chi.URLParam(req, name))
}

func routeInt64Param(req *http.Request, name, label string) (int64, error) {
	value := routeParam(req, name)
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid %s id", label)
	}
	return id, nil
}

func queryInt(req *http.Request, name string, fallback int) (int, error) {
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

func queryBool(req *http.Request, name string) (bool, error) {
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
