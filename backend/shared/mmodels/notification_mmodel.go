package mmodels

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend/shared/metadata/impl"
	"backend/shared/metadata/interfaces"
	"github.com/jmoiron/sqlx"
)

// NotificationMModel owns notification persistence and account scoping.
type NotificationMModel struct {
	impl.MModel

	ID              int64          `json:"id" db:"id"`
	AccountID       int64          `json:"accountId" db:"account_id"`
	RecipientUserID *int64         `json:"recipientUserId,omitempty" db:"recipient_user_id"`
	RecipientEmail  string         `json:"recipientEmail" db:"recipient_email"`
	Title           string         `json:"title" db:"title"`
	Message         string         `json:"message" db:"message"`
	EntityType      string         `json:"entityType" db:"entity_type"`
	EntityID        int64          `json:"entityId" db:"entity_id"`
	EntityNumber    string         `json:"entityNumber" db:"entity_number"`
	Action          string         `json:"action" db:"action"`
	IsRead          bool           `json:"isRead" db:"is_read"`
	ReadAt          *time.Time     `json:"readAt,omitempty" db:"read_at"`
	CreatedAt       time.Time      `json:"-" db:"created_at"`
	CreatedBy       string         `json:"-" db:"created_by"`
	UpdatedAt       sql.NullTime   `json:"-" db:"updated_at"`
	UpdatedBy       sql.NullString `json:"-" db:"updated_by"`
	AuthAccountID   int64          `json:"-" db:"-"`
	AuthEmail       string         `json:"-" db:"-"`
	AuthUserName    string         `json:"-" db:"-"`
}

func NewNotificationMModel(ctx context.Context, db *sqlx.DB) *NotificationMModel {
	return &NotificationMModel{
		MModel: impl.MModel{
			Context: ctx, Name: "Notification", TableName: "notifications", DB: db,
			Fields: []interfaces.IMField{
				impl.MField{Type: "int64", DbType: "INTEGER", JsonTag: "id", DbTag: "id", IsPrimaryKey: true, IsIdentity: true, Required: true, IsNotUpdateField: true},
				impl.MField{Type: "int64", DbType: "BIGINT", JsonTag: "accountId", DbTag: "account_id", Required: true, IsNotUpdateField: true},
				impl.MField{Type: "int64", DbType: "BIGINT", JsonTag: "recipientUserId", DbTag: "recipient_user_id", IsNullable: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "recipientEmail", DbTag: "recipient_email", MaxLength: 255, Default: "''", Required: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "title", DbTag: "title", MaxLength: 255, Required: true},
				impl.MField{Type: "string", DbType: "TEXT", JsonTag: "message", DbTag: "message", Required: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "entityType", DbTag: "entity_type", MaxLength: 100, Default: "''", Required: true},
				impl.MField{Type: "int64", DbType: "BIGINT", JsonTag: "entityId", DbTag: "entity_id", Default: "0", Required: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "entityNumber", DbTag: "entity_number", MaxLength: 100, Default: "''", Required: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "action", DbTag: "action", MaxLength: 50, Default: "''", Required: true},
				impl.MField{Type: "bool", DbType: "BOOLEAN", JsonTag: "isRead", DbTag: "is_read", Default: "false", Required: true},
				impl.MField{Type: "time.Time", DbType: "TIMESTAMPTZ", JsonTag: "readAt", DbTag: "read_at", IsNullable: true},
				impl.MField{Type: "time.Time", DbType: "TIMESTAMPTZ", JsonTag: "-", DbTag: "created_at", Default: "NOW()", Required: true, IsNotUpdateField: true},
				impl.MField{Type: "string", DbType: "VARCHAR", JsonTag: "-", DbTag: "created_by", MaxLength: 255, Default: "''", Required: true, IsNotUpdateField: true},
				impl.MField{Type: "sql.NullTime", DbType: "TIMESTAMPTZ", JsonTag: "-", DbTag: "updated_at", IsNullable: true, IsNotCreateField: true},
				impl.MField{Type: "sql.NullString", DbType: "VARCHAR", JsonTag: "-", DbTag: "updated_by", MaxLength: 255, IsNullable: true, IsNotCreateField: true},
			},
		},
	}
}

func (m *NotificationMModel) Migrate() error {
	if m == nil {
		return errors.New("notification metadata model is nil")
	}
	if m.DB == nil {
		return errors.New("notification metadata database is nil")
	}
	ctx := m.ContextOrBackground()
	tx, err := m.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin notifications migration: %w", err)
	}
	defer tx.Rollback()

	migrator := impl.NewMMigrator(&m.MModel, tx)
	migrator.SetContext(ctx)
	if err := migrator.Migrate(); err != nil {
		return err
	}
	statements := []string{
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS account_id BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS recipient_user_id BIGINT`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS recipient_email VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS title VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS message TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS entity_type VARCHAR(100) NOT NULL DEFAULT ''`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS entity_id BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS entity_number VARCHAR(100) NOT NULL DEFAULT ''`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS action VARCHAR(50) NOT NULL DEFAULT ''`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS is_read BOOLEAN NOT NULL DEFAULT false`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS read_at TIMESTAMPTZ`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS created_by VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS updated_by VARCHAR(255)`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_account_recipient ON notifications (account_id, recipient_user_id, is_read, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_account_email ON notifications (account_id, LOWER(recipient_email), is_read, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_entity ON notifications (account_id, entity_type, entity_id)`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("run notifications migration statement %q: %w", statement, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit notifications migration: %w", err)
	}
	return nil
}

func (m *NotificationMModel) Normalize() {
	if m == nil {
		return
	}
	m.RecipientEmail = strings.ToLower(strings.TrimSpace(m.RecipientEmail))
	m.Title = strings.TrimSpace(m.Title)
	m.Message = strings.TrimSpace(m.Message)
	m.EntityType = strings.ToUpper(strings.TrimSpace(m.EntityType))
	m.EntityNumber = strings.TrimSpace(m.EntityNumber)
	m.Action = strings.ToUpper(strings.TrimSpace(m.Action))
}

func (m *NotificationMModel) ValidateCreate() error {
	if m == nil {
		return errors.New("notification metadata model is nil")
	}
	if m.DB == nil {
		return errors.New("notification metadata database is nil")
	}
	accountID, err := m.MModel.ValidateAccountID_(m.AuthAccountID, m.AccountID)
	if err != nil {
		return err
	}
	m.AccountID = accountID
	m.Normalize()
	if m.Title == "" {
		return errors.New("title notification wajib diisi")
	}
	if m.Message == "" {
		return errors.New("message notification wajib diisi")
	}
	if m.RecipientUserID == nil && m.RecipientEmail == "" {
		return errors.New("recipient notification wajib diisi")
	}
	return m.MModel.ValidateCreate_(m)
}

func (m *NotificationMModel) Create() (int64, error) {
	if err := m.ValidateCreate(); err != nil {
		return 0, err
	}
	createdBy, err := m.MModel.ResolveCreatedBy_(m.CreatedBy, firstNonBlank(m.AuthEmail, m.AuthUserName, "system"))
	if err != nil {
		return 0, err
	}
	m.CreatedBy = createdBy
	m.IsRead = false
	return m.MModel.Create_(m)
}

func (m *NotificationMModel) FindList(recipientUserID int64, recipientEmail, search string, limit int) ([]NotificationMModel, error) {
	if m == nil {
		return nil, errors.New("notification metadata model is nil")
	}
	if m.DB == nil {
		return nil, errors.New("notification metadata database is nil")
	}
	accountID, err := m.MModel.ValidateAccountID_(m.AuthAccountID, m.AccountID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	query := `SELECT id, account_id, recipient_user_id, recipient_email, title, message, entity_type, entity_id, entity_number, action, is_read, read_at, created_at, created_by, updated_at, updated_by FROM notifications WHERE account_id = $1`
	args := []any{accountID}
	if recipientUserID > 0 {
		args = append(args, recipientUserID)
		query += ` AND recipient_user_id = $2`
	} else if recipientEmail = strings.ToLower(strings.TrimSpace(recipientEmail)); recipientEmail != "" {
		args = append(args, recipientEmail)
		query += ` AND LOWER(recipient_email) = $2`
	} else {
		return []NotificationMModel{}, nil
	}
	if search = strings.TrimSpace(search); search != "" {
		args = append(args, search)
		placeholder := fmt.Sprintf("$%d", len(args))
		query += ` AND (title ILIKE '%' || ` + placeholder + ` || '%' OR message ILIKE '%' || ` + placeholder + ` || '%' OR entity_number ILIKE '%' || ` + placeholder + ` || '%')`
	}
	args = append(args, limit)
	query += fmt.Sprintf(` ORDER BY created_at DESC, id DESC LIMIT $%d`, len(args))
	items := []NotificationMModel{}
	if err := m.DB.SelectContext(m.ContextOrBackground(), &items, query, args...); err != nil {
		return nil, err
	}
	return items, nil
}

func (m *NotificationMModel) CountUnread(recipientUserID int64, recipientEmail string) (int64, error) {
	if m == nil {
		return 0, errors.New("notification metadata model is nil")
	}
	if m.DB == nil {
		return 0, errors.New("notification metadata database is nil")
	}
	accountID, err := m.MModel.ValidateAccountID_(m.AuthAccountID, m.AccountID)
	if err != nil {
		return 0, err
	}
	query := `SELECT COUNT(*) FROM notifications WHERE account_id = $1 AND is_read = false`
	args := []any{accountID}
	if recipientUserID > 0 {
		args = append(args, recipientUserID)
		query += ` AND recipient_user_id = $2`
	} else if recipientEmail = strings.ToLower(strings.TrimSpace(recipientEmail)); recipientEmail != "" {
		args = append(args, recipientEmail)
		query += ` AND LOWER(recipient_email) = $2`
	} else {
		return 0, nil
	}
	var count int64
	if err := m.DB.GetContext(m.ContextOrBackground(), &count, query, args...); err != nil {
		return 0, err
	}
	return count, nil
}

func (m *NotificationMModel) MarkRead(id int64) error {
	if m == nil {
		return errors.New("notification metadata model is nil")
	}
	if m.DB == nil {
		return errors.New("notification metadata database is nil")
	}
	if id <= 0 {
		return errors.New("id notification tidak valid")
	}
	accountID, err := m.MModel.ValidateAccountID_(m.AuthAccountID, m.AccountID)
	if err != nil {
		return err
	}
	updatedBy := firstNonBlank(m.AuthEmail, m.AuthUserName)
	_, err = m.DB.ExecContext(m.ContextOrBackground(), `
		UPDATE notifications SET is_read = true, read_at = NOW(), updated_at = NOW(), updated_by = NULLIF($1, '')
		WHERE id = $2 AND account_id = $3`, updatedBy, id, accountID)
	return err
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
