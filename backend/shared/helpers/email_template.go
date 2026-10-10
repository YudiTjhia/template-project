package helpers

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type EmailTemplateField struct {
	Label string
	Value string
}

type TransactionNotificationEmailData struct {
	Subject        string
	Title          string
	Message        string
	RecipientName  string
	RecipientEmail string
	EntityType     string
	EntityLabel    string
	EntityNumber   string
	Action         string
	ActionLabel    string
	Actor          string
	SystemURL      string
	DocumentURL    string
	GeneratedAt    string
	Fields         []EmailTemplateField
}

const fallbackTransactionNotificationTemplate = `<!doctype html>
<html lang="id">
  <body style="margin:0;background:#e2e8f0;color:#0f172a;font-family:Arial,Helvetica,sans-serif;">
    <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background:#e2e8f0;margin:0;padding:10px 6px;">
      <tr><td align="center">
        <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="width:100%;max-width:640px;background:#ffffff;border:1px solid #cbd5e1;border-radius:12px;overflow:hidden;">
          <tr><td style="padding:16px 16px 12px;border-bottom:1px solid #e2e8f0;">
            <div style="font-size:20px;line-height:1.25;font-weight:900;">{{.EntityLabel}}</div>
            <div style="margin-top:2px;font-size:12px;color:#64748b;">Document Notification</div>
            <div style="margin-top:8px;font-size:14px;line-height:1.3;font-weight:900;word-break:break-word;overflow-wrap:anywhere;">{{if .EntityNumber}}{{.EntityNumber}}{{else}}-{{end}}</div>
          </td></tr>
          <tr><td style="padding:14px 16px 6px;">
            <div style="font-size:18px;line-height:1.3;font-weight:900;">{{.Title}}</div>
            <div style="margin-top:6px;font-size:13px;line-height:1.5;color:#334155;word-break:break-word;overflow-wrap:anywhere;">{{.Message}}</div>
          </td></tr>
          <tr><td style="padding:10px 16px;">
            <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="border:1px solid #cbd5e1;border-radius:10px;overflow:hidden;table-layout:fixed;">
              <tr><td style="padding:8px 10px;background:#f8fafc;border-bottom:1px solid #cbd5e1;font-size:13px;font-weight:900;">Informasi Dokumen</td></tr>
              {{range .Fields}}<tr>
                <td style="padding:8px 10px;border-bottom:1px solid #e2e8f0;">
                  <div style="font-size:11px;line-height:1.25;color:#64748b;">{{.Label}}</div>
                  <div style="margin-top:3px;font-size:13px;line-height:1.35;font-weight:700;word-break:break-word;overflow-wrap:anywhere;">{{.Value}}</div>
                </td>
              </tr>{{end}}
            </table>
          </td></tr>
          {{if or .SystemURL .DocumentURL}}<tr><td style="padding:4px 16px 14px;">
            {{if .SystemURL}}<a href="{{.SystemURL}}" style="display:inline-block;background:#1565c0;color:#ffffff;text-decoration:none;border-radius:999px;padding:9px 14px;font-size:13px;font-weight:900;">Kembali ke System</a>{{end}}
            {{if .DocumentURL}}<a href="{{.DocumentURL}}" style="display:inline-block;margin:6px 0 0 6px;background:#ffffff;color:#1565c0;text-decoration:none;border:1px solid #93c5fd;border-radius:999px;padding:8px 13px;font-size:13px;font-weight:900;">Buka Dokumen</a>{{end}}
            <div style="margin-top:6px;font-size:11px;line-height:1.45;color:#64748b;word-break:break-word;overflow-wrap:anywhere;">{{if .SystemURL}}System: {{.SystemURL}}{{else}}Dokumen: {{.DocumentURL}}{{end}}</div>
          </td></tr>{{end}}
          <tr><td style="padding:12px 16px 16px;border-top:1px solid #e2e8f0;background:#f8fafc;font-size:11px;line-height:1.5;color:#64748b;">Email ini dikirim otomatis oleh Buyer Services pada {{.GeneratedAt}}.</td></tr>
        </table>
      </td></tr>
    </table>
  </body>
</html>`

func RenderTransactionNotificationEmail(data TransactionNotificationEmailData) (string, error) {
	data = normalizeTransactionNotificationEmailData(data)

	tmpl, err := parseTransactionNotificationTemplate()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func normalizeTransactionNotificationEmailData(data TransactionNotificationEmailData) TransactionNotificationEmailData {
	data.Title = strings.TrimSpace(data.Title)
	data.Subject = strings.TrimSpace(data.Subject)
	if data.Subject == "" {
		data.Subject = data.Title
	}
	data.Message = strings.TrimSpace(data.Message)
	data.EntityType = strings.ToUpper(strings.TrimSpace(data.EntityType))
	data.EntityLabel = strings.TrimSpace(data.EntityLabel)
	if data.EntityLabel == "" {
		data.EntityLabel = titleizeUnderscore(data.EntityType)
	}
	data.EntityNumber = strings.TrimSpace(data.EntityNumber)
	data.Action = strings.ToUpper(strings.TrimSpace(data.Action))
	data.ActionLabel = strings.TrimSpace(data.ActionLabel)
	if data.ActionLabel == "" {
		data.ActionLabel = titleizeUnderscore(data.Action)
	}
	data.Actor = strings.TrimSpace(data.Actor)
	data.SystemURL = strings.TrimSpace(data.SystemURL)
	data.DocumentURL = strings.TrimSpace(data.DocumentURL)
	data.RecipientName = strings.TrimSpace(data.RecipientName)
	data.RecipientEmail = strings.TrimSpace(data.RecipientEmail)
	data.GeneratedAt = strings.TrimSpace(data.GeneratedAt)
	if data.GeneratedAt == "" {
		data.GeneratedAt = time.Now().Format("02 Jan 2006 15:04 MST")
	}
	if len(data.Fields) == 0 {
		data.Fields = []EmailTemplateField{
			{Label: "Tipe Dokumen", Value: data.EntityLabel},
			{Label: "Nomor", Value: fallbackDash(data.EntityNumber)},
			{Label: "Action", Value: fallbackDash(data.ActionLabel)},
			{Label: "Actor", Value: fallbackDash(data.Actor)},
		}
	}
	return data
}

func parseTransactionNotificationTemplate() (*template.Template, error) {
	for _, candidate := range transactionNotificationTemplateCandidates() {
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			return template.ParseFiles(candidate)
		}
	}
	return template.New("transaction_notification").Parse(fallbackTransactionNotificationTemplate)
}

func transactionNotificationTemplateCandidates() []string {
	templateDir := strings.TrimSpace(os.Getenv("EMAIL_TEMPLATE_DIR"))
	backendTemplateDir := strings.TrimSpace(os.Getenv("BACKEND_TEMPLATE_DIR"))

	candidates := []string{}
	if templateDir != "" {
		candidates = append(candidates, filepath.Join(templateDir, "transaction_notification.html"))
	}
	if backendTemplateDir != "" {
		candidates = append(candidates, filepath.Join(backendTemplateDir, "email", "transaction_notification.html"))
	}
	candidates = append(candidates,
		filepath.Join("templates", "email", "transaction_notification.html"),
		filepath.Join("backend", "templates", "email", "transaction_notification.html"),
		filepath.Join("/app", "templates", "email", "transaction_notification.html"),
	)
	return candidates
}

func titleizeUnderscore(value string) string {
	parts := strings.Fields(strings.ReplaceAll(strings.ToLower(strings.TrimSpace(value)), "_", " "))
	for index, part := range parts {
		if part == "" {
			continue
		}
		parts[index] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func fallbackDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return strings.TrimSpace(value)
}
