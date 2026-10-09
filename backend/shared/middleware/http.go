package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
)

func BearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" || !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return ""
	}
	return strings.TrimSpace(header[len("bearer "):])
}

func IsExemptPath(path string, exempt map[string]struct{}) bool {
	if _, ok := exempt[path]; ok {
		return true
	}
	for prefix := range exempt {
		if strings.HasSuffix(prefix, "/") && strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func ExemptPathSet(paths []string) map[string]struct{} {
	exempt := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		exempt[path] = struct{}{}
	}
	return exempt
}

func WriteJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":      false,
		"error":   message,
		"message": message,
	})
}
