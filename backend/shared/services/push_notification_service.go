package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

const defaultFirebaseCredentialsFile = "market-pos-93cc1-firebase-adminsdk-fbsvc-8ab6d0454d.json"

type PushNotification struct {
	Title string
	Body  string
	Data  map[string]string
}

type PushNotificationService struct {
	client *messaging.Client
}

func NewPushNotificationService(ctx context.Context) (*PushNotificationService, error) {
	credentialsFile := strings.TrimSpace(os.Getenv("FIREBASE_CREDENTIALS_FILE"))
	if credentialsFile == "" {
		credentialsFile = defaultFirebaseCredentialsFile
	}
	if _, err := os.Stat(credentialsFile); err != nil {
		return nil, fmt.Errorf("firebase credentials file: %w", err)
	}
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("initialize firebase app: %w", err)
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize firebase messaging: %w", err)
	}
	return &PushNotificationService{client: client}, nil
}

func (s *PushNotificationService) SendToToken(ctx context.Context, token string, notification PushNotification) (string, error) {
	if s == nil || s.client == nil {
		return "", errors.New("push notification service belum diinisialisasi")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", errors.New("firebase device token wajib diisi")
	}
	if strings.TrimSpace(notification.Title) == "" || strings.TrimSpace(notification.Body) == "" {
		return "", errors.New("title dan body notification wajib diisi")
	}
	return s.client.Send(ctx, &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: notification.Title,
			Body:  notification.Body,
		},
		Data: notification.Data,
	})
}

func (s *PushNotificationService) SendToTokens(ctx context.Context, tokens []string, notification PushNotification) (*messaging.BatchResponse, error) {
	if s == nil || s.client == nil {
		return nil, errors.New("push notification service belum diinisialisasi")
	}
	if len(tokens) == 0 {
		return nil, errors.New("firebase device token wajib diisi")
	}
	if strings.TrimSpace(notification.Title) == "" || strings.TrimSpace(notification.Body) == "" {
		return nil, errors.New("title dan body notification wajib diisi")
	}
	messages := make([]*messaging.Message, 0, len(tokens))
	for _, token := range tokens {
		if strings.TrimSpace(token) == "" {
			continue
		}
		messages = append(messages, &messaging.Message{
			Token:        token,
			Notification: &messaging.Notification{Title: notification.Title, Body: notification.Body},
			Data:         notification.Data,
		})
	}
	if len(messages) == 0 {
		return nil, errors.New("firebase device token wajib diisi")
	}
	return s.client.SendEach(ctx, messages)
}

func (s *PushNotificationService) SendToTopic(ctx context.Context, topic string, notification PushNotification) (string, error) {
	if s == nil || s.client == nil {
		return "", errors.New("push notification service belum diinisialisasi")
	}
	if strings.TrimSpace(topic) == "" {
		return "", errors.New("firebase topic wajib diisi")
	}
	if strings.TrimSpace(notification.Title) == "" || strings.TrimSpace(notification.Body) == "" {
		return "", errors.New("title dan body notification wajib diisi")
	}
	return s.client.Send(ctx, &messaging.Message{
		Topic:        topic,
		Notification: &messaging.Notification{Title: notification.Title, Body: notification.Body},
		Data:         notification.Data,
	})
}
