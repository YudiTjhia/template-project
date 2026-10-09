package templates

import (
	"bytes"
	"embed"
	"html/template"
	"time"
)

//go:embed account_otp.html login_otp.html
var files embed.FS

type OTPEmailData struct {
	Name      string
	OTP       string
	ExpiresAt string
}

func RenderAccountOTP(name, otp string, expires time.Time) (string, error) {
	return renderOTP("account_otp.html", name, otp, expires)
}

func RenderLoginOTP(name, otp string, expires time.Time) (string, error) {
	return renderOTP("login_otp.html", name, otp, expires)
}

func renderOTP(file, name, otp string, expires time.Time) (string, error) {
	tmpl, err := template.ParseFS(files, file)
	if err != nil {
		return "", err
	}

	var output bytes.Buffer
	err = tmpl.Execute(&output, OTPEmailData{
		Name:      name,
		OTP:       otp,
		ExpiresAt: expires.Format("02 Jan 2006 15:04 MST"),
	})
	if err != nil {
		return "", err
	}
	return output.String(), nil
}
