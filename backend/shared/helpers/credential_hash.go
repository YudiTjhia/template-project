package helpers

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const credentialHashPrefix = "$2"

func HashCredential(value string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func IsCredentialHash(value string) bool {
	return strings.HasPrefix(strings.TrimSpace(value), credentialHashPrefix)
}

func CompareCredential(stored, plain string) bool {
	stored = strings.TrimSpace(stored)
	plain = strings.TrimSpace(plain)
	if stored == "" || plain == "" {
		return false
	}
	if IsCredentialHash(stored) {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(plain)) == nil
	}
	// Compatibility for legacy plaintext rows; new writes are always hashed.
	return stored == plain
}
