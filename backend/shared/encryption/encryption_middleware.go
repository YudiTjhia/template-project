package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	KeyEndpoint       = "/.well-known/request-encryption-key"
	EncryptedMimeType = "application/vnd.pos.encrypted+json"
	maxEncryptedBody  = 10 << 20
)

type Envelope struct {
	KeyID        string `json:"keyId"`
	EncryptedKey string `json:"encryptedKey"`
	Nonce        string `json:"nonce"`
	Ciphertext   string `json:"ciphertext"`
}

type EncryptionMiddleware struct {
	privateKey *rsa.PrivateKey
	keyID      string
	publicPEM  string
}

func NewFromFile(path string) (*EncryptionMiddleware, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read request encryption private key: %w", err)
	}
	return New(contents)
}

func NewFromBase64(encoded string) (*EncryptionMiddleware, error) {
	privatePEM, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return nil, fmt.Errorf("decode request encryption private key: %w", err)
	}
	return New(privatePEM)
}

func New(privatePEM []byte) (*EncryptionMiddleware, error) {
	block, _ := pem.Decode(privatePEM)
	if block == nil {
		return nil, errors.New("request encryption private key is not PEM")
	}

	var privateKey *rsa.PrivateKey
	if parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		privateKey, _ = parsed.(*rsa.PrivateKey)
	} else if parsed, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		privateKey = parsed
	}
	if privateKey == nil {
		return nil, errors.New("request encryption private key must be RSA PKCS#8 or PKCS#1")
	}
	if err := privateKey.Validate(); err != nil {
		return nil, fmt.Errorf("validate request encryption private key: %w", err)
	}

	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("encode request encryption public key: %w", err)
	}
	fingerprint := sha256.Sum256(publicDER)
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})
	return &EncryptionMiddleware{
		privateKey: privateKey,
		keyID:      hex.EncodeToString(fingerprint[:]),
		publicPEM:  string(publicPEM),
	}, nil
}

func (m *EncryptionMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == KeyEndpoint {
			if req.Method != http.MethodGet {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"keyId":     m.keyID,
				"algorithm": "RSA-OAEP-256+A256GCM",
				"publicKey": m.publicPEM,
			})
			return
		}

		if !strings.EqualFold(strings.TrimSpace(strings.Split(req.Header.Get("Content-Type"), ";")[0]), EncryptedMimeType) {
			next.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodGet || req.Method == http.MethodHead {
			writeError(w, http.StatusBadRequest, "encrypted request body is not supported for this method")
			return
		}

		body, err := io.ReadAll(http.MaxBytesReader(w, req.Body, maxEncryptedBody))
		if err != nil {
			writeError(w, http.StatusRequestEntityTooLarge, "encrypted request body is too large")
			return
		}
		plaintext, err := m.decrypt(body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "encrypted request could not be decrypted")
			return
		}
		req.Body = io.NopCloser(strings.NewReader(string(plaintext)))
		req.ContentLength = int64(len(plaintext))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Del("Content-Encoding")
		next.ServeHTTP(w, req)
	})
}

func (m *EncryptionMiddleware) decrypt(body []byte) ([]byte, error) {
	var envelope Envelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}
	if envelope.KeyID != m.keyID {
		return nil, errors.New("unknown encryption key id")
	}
	decode := base64.StdEncoding.DecodeString
	encryptedKey, err := decode(envelope.EncryptedKey)
	if err != nil {
		return nil, err
	}
	key, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, m.privateKey, encryptedKey, nil)
	if err != nil {
		return nil, err
	}
	if len(key) != 32 {
		return nil, errors.New("invalid AES-256 key length")
	}
	nonce, err := decode(envelope.Nonce)
	if err != nil {
		return nil, err
	}
	ciphertext, err := decode(envelope.Ciphertext)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, errors.New("invalid AES-GCM nonce length")
	}
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message, "message": message})
}
