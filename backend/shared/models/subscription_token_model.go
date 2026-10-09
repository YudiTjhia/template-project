package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"time"
)

type SubscriptionBankInfo struct {
	BankName      string `json:"bankName"`
	AccountNumber string `json:"accountNumber"`
	AccountHolder string `json:"accountHolder"`
	Currency      string `json:"currency"`
}

type SubscriptionPaymentProof struct {
	FileName      string `json:"fileName"`
	FileSize      int64  `json:"fileSize"`
	ContentType   string `json:"contentType"`
	ContentBase64 string `json:"contentBase64,omitempty"`
}

type SubscriptionPaymentTokenRequest struct {
	Months   int                      `json:"months"`
	AppType  string                   `json:"appType"`
	PlanCode string                   `json:"planCode"`
	UserIDs  []int64                  `json:"userIds"`
	Proof    SubscriptionPaymentProof `json:"proof"`
}

type SubscriptionPaymentTokenPayload struct {
	Version       int                      `json:"version"`
	AccountID     int64                    `json:"accountId"`
	AppType       string                   `json:"appType"`
	PlanCode      string                   `json:"planCode"`
	PlanName      string                   `json:"planName"`
	UserIDs       []int64                  `json:"userIds"`
	NumberOfUsers int                      `json:"numberOfUsers"`
	Months        int                      `json:"months"`
	PricePerUser  float64                  `json:"pricePerUser"`
	TotalAmount   float64                  `json:"totalAmount"`
	Currency      string                   `json:"currency"`
	Bank          SubscriptionBankInfo     `json:"bank"`
	Proof         SubscriptionPaymentProof `json:"proof"`
	IssuedAt      time.Time                `json:"issuedAt"`
}

type SubscriptionPaymentTokenResponse struct {
	Token   string                          `json:"token"`
	Payload SubscriptionPaymentTokenPayload `json:"payload"`
}

type SubscriptionApplyTokenRequest struct {
	Token string `json:"token"`
}

func SubscriptionPaymentBankInfo() SubscriptionBankInfo {
	return SubscriptionBankInfo{
		BankName:      subscriptionEnv("SUBSCRIPTION_BANK_NAME", "Bank Transfer"),
		AccountNumber: subscriptionEnv("SUBSCRIPTION_BANK_ACCOUNT_NUMBER", ""),
		AccountHolder: subscriptionEnv("SUBSCRIPTION_BANK_ACCOUNT_HOLDER", ""),
		Currency:      subscriptionEnv("SUBSCRIPTION_CURRENCY", "IDR"),
	}
}

func GenerateSubscriptionPaymentToken(payload SubscriptionPaymentTokenPayload) (string, error) {
	secret := strings.TrimSpace(os.Getenv("SUBSCRIPTION_TOKEN_SECRET"))
	if secret == "" {
		return "", errors.New("SUBSCRIPTION_TOKEN_SECRET belum dikonfigurasi")
	}

	plain, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plain, nil)
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

func DecodeSubscriptionPaymentToken(token string) (*SubscriptionPaymentTokenPayload, error) {
	secret := strings.TrimSpace(os.Getenv("SUBSCRIPTION_TOKEN_SECRET"))
	if secret == "" {
		return nil, errors.New("SUBSCRIPTION_TOKEN_SECRET belum dikonfigurasi")
	}

	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(token))
	if err != nil {
		return nil, errors.New("token subscription tidak valid")
	}

	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(raw) < gcm.NonceSize() {
		return nil, errors.New("token subscription tidak valid")
	}

	nonce := raw[:gcm.NonceSize()]
	ciphertext := raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("token subscription tidak valid")
	}

	var payload SubscriptionPaymentTokenPayload
	if err := json.Unmarshal(plain, &payload); err != nil {
		return nil, errors.New("token subscription tidak valid")
	}
	return &payload, nil
}

func subscriptionEnv(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
