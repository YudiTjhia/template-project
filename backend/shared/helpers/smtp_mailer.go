package helpers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"backend/shared/templates"
)

type SMTPConfig struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromName  string
	FromEmail string
}

func LoadSMTPConfig() SMTPConfig {
	port, _ := strconv.Atoi(strings.TrimSpace(os.Getenv("SMTP_PORT")))
	return SMTPConfig{
		Host:      strings.TrimSpace(os.Getenv("SMTP_HOST")),
		Port:      port,
		Username:  strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		Password:  strings.TrimSpace(os.Getenv("SMTP_PASSWORD")),
		FromName:  strings.TrimSpace(os.Getenv("SMTP_FROM_NAME")),
		FromEmail: strings.TrimSpace(os.Getenv("SMTP_FROM_EMAIL")),
	}
}

func (c SMTPConfig) Enabled() bool {
	return c.Host != "" && c.Port > 0 && c.FromEmail != ""
}

func (c SMTPConfig) FromHeader() string {
	if c.FromName == "" {
		return c.FromEmail
	}

	return fmt.Sprintf("%s <%s>", c.FromName, c.FromEmail)
}

func (c SMTPConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c SMTPConfig) Auth() smtp.Auth {
	if c.Username == "" {
		return nil
	}

	return smtp.PlainAuth("", c.Username, c.Password, c.Host)
}

func BuildRegisterOTPMail(accountName, otp string, expired time.Time) (string, string, string, error) {
	subject := "OTP registrasi account POS Cashier"
	textBody := fmt.Sprintf(
		"Hallo %s,\n\nOTP registrasi account kamu adalah: %s\nBerlaku sampai: %s\n\nOTP ini berlaku selama 15 menit.\nJika kamu tidak merasa mendaftar, abaikan email ini.\n",
		accountName,
		otp,
		expired.Format(time.RFC3339),
	)
	htmlBody, err := templates.RenderAccountOTP(accountName, otp, expired)
	return subject, textBody, htmlBody, err
}

func BuildRegisterAccountIDMail(accountName string, accountID int64) (string, string) {
	subject := "Account ID POS Cashier kamu"
	body := fmt.Sprintf(
		"Hallo %s,\n\nRegistrasi account kamu sudah berhasil diverifikasi.\nAccount ID kamu adalah: %d\n\nSimpan account ID ini untuk login ke POS Cashier.\nJika kamu lupa, gunakan email registrasi ini untuk bantuan admin.\n",
		accountName,
		accountID,
	)

	return subject, body
}

func BuildLoginOTPMail(username, otp string, expired time.Time) (string, string, string, error) {
	subject := "OTP login buyer"
	textBody := fmt.Sprintf(
		"Hallo %s,\n\nOTP login kamu adalah: %s\nBerlaku sampai: %s\n\nOTP ini berlaku selama 10 menit.\nJika kamu tidak merasa mencoba login, abaikan email ini.\n",
		username,
		otp,
		expired.Format(time.RFC3339),
	)
	htmlBody, err := templates.RenderLoginOTP(username, otp, expired)
	return subject, textBody, htmlBody, err
}

func WriteEmailMessage(from, to, subject, body string) []byte {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", to))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(body)
	return buf.Bytes()
}

func WriteHTMLMultipartEmailMessage(from, to, subject, textBody, htmlBody string) []byte {
	boundary := fmt.Sprintf("procurement-%d", time.Now().UnixNano())
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", to))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%q\r\n", boundary))
	buf.WriteString("\r\n")

	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	buf.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(textBody)
	buf.WriteString("\r\n")

	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buf.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(htmlBody)
	buf.WriteString("\r\n")

	buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	return buf.Bytes()
}

func WriteAttachmentEmailMessage(from, to, subject, body, fileName, contentType string, content []byte) []byte {
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}
	fileName = strings.NewReplacer("\r", "", "\n", "", "\"", "").Replace(strings.TrimSpace(fileName))
	if fileName == "" {
		fileName = "payment-proof"
	}
	boundary := fmt.Sprintf("pos-kasir-attachment-%d", time.Now().UnixNano())
	encoded := base64.StdEncoding.EncodeToString(content)
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", to))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%q\r\n", boundary))
	buf.WriteString("\r\n")
	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	buf.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
	buf.WriteString(body)
	buf.WriteString("\r\n\r\n")
	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString(fmt.Sprintf("Content-Type: %s\r\n", contentType))
	buf.WriteString("Content-Transfer-Encoding: base64\r\n")
	buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", fileName))
	for start := 0; start < len(encoded); start += 76 {
		end := start + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		buf.WriteString(encoded[start:end])
		buf.WriteString("\r\n")
	}
	buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	return buf.Bytes()
}

func SendEmailInBackground(task string, sendFn func() error) {
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("[WARN] %s panic: %v", task, recovered)
			}
		}()

		if err := sendFn(); err != nil {
			log.Printf("[WARN] %s failed: %v", task, err)
		}
	}()
}
