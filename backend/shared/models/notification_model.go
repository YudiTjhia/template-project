package models

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Notification struct {
	BaseModel

	ID              int64      `json:"id" db:"id"`
	AccountID       int64      `json:"accountId" db:"account_id"`
	RecipientUserID *int64     `json:"recipientUserId,omitempty" db:"recipient_user_id"`
	RecipientEmail  string     `json:"recipientEmail" db:"recipient_email"`
	Title           string     `json:"title" db:"title"`
	Message         string     `json:"message" db:"message"`
	EntityType      string     `json:"entityType" db:"entity_type"`
	EntityID        int64      `json:"entityId" db:"entity_id"`
	EntityNumber    string     `json:"entityNumber" db:"entity_number"`
	Action          string     `json:"action" db:"action"`
	IsRead          bool       `json:"isRead" db:"is_read"`
	ReadAt          *time.Time `json:"readAt,omitempty" db:"read_at"`
}

const notificationSelectColumns = `n.id, n.account_id, n.recipient_user_id, n.recipient_email, n.title, n.message, n.entity_type, n.entity_id, n.entity_number, n.action, n.is_read, n.read_at,
n.created_at, n.created_by, n.updated_at, n.updated_by`

func NewNotification(db *sqlx.DB, auth Auth) Notification {
	item := Notification{BaseModel: BaseModel{DB: db, Auth: auth}, AccountID: auth.AccountID}
	item.CreatedBy = auth.Email
	return item
}

func (m Notification) Migrate() {
	sql := `CREATE TABLE IF NOT EXISTS notifications (
	id SERIAL PRIMARY KEY,
	account_id BIGINT NOT NULL DEFAULT 0,
	recipient_user_id BIGINT,
	recipient_email VARCHAR(255) NOT NULL DEFAULT '',
	title VARCHAR(255) NOT NULL DEFAULT '',
	message TEXT NOT NULL DEFAULT '',
	entity_type VARCHAR(100) NOT NULL DEFAULT '',
	entity_id BIGINT NOT NULL DEFAULT 0,
	entity_number VARCHAR(100) NOT NULL DEFAULT '',
	action VARCHAR(50) NOT NULL DEFAULT '',
	is_read BOOLEAN NOT NULL DEFAULT false,
	read_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_by VARCHAR(255) NOT NULL DEFAULT '',
	updated_at TIMESTAMPTZ,
	updated_by VARCHAR(255)
);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)

	statements := []string{
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS account_id BIGINT NOT NULL DEFAULT 0;`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS recipient_user_id BIGINT;`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS recipient_email VARCHAR(255) NOT NULL DEFAULT '';`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS title VARCHAR(255) NOT NULL DEFAULT '';`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS message TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS entity_type VARCHAR(100) NOT NULL DEFAULT '';`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS entity_id BIGINT NOT NULL DEFAULT 0;`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS entity_number VARCHAR(100) NOT NULL DEFAULT '';`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS action VARCHAR(50) NOT NULL DEFAULT '';`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS is_read BOOLEAN NOT NULL DEFAULT false;`,
		`ALTER TABLE notifications ADD COLUMN IF NOT EXISTS read_at TIMESTAMPTZ;`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_account_recipient ON notifications (account_id, recipient_user_id, is_read, created_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_account_email ON notifications (account_id, LOWER(recipient_email), is_read, created_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_entity ON notifications (account_id, entity_type, entity_id);`,
	}
	for _, statement := range statements {
		log.Println("[SQL]=", statement)
		m.DB.MustExec(statement)
	}
}

func (m Notification) Create() (int64, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return 0, err
	}
	m.Normalize()
	if m.Title == "" {
		return 0, errors.New("title notification wajib diisi")
	}
	if m.Message == "" {
		return 0, errors.New("message notification wajib diisi")
	}
	if m.RecipientUserID == nil && m.RecipientEmail == "" {
		return 0, errors.New("recipient notification wajib diisi")
	}
	createdBy := m.CreatedBy
	if createdBy == "" {
		createdBy = m.Auth.Email
	}
	if createdBy == "" {
		createdBy = m.Auth.UserName
	}
	if createdBy == "" {
		createdBy = "system"
	}

	var id int64
	err = m.DB.QueryRow(
		`INSERT INTO notifications
		(account_id, recipient_user_id, recipient_email, title, message, entity_type, entity_id, entity_number, action, is_read, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,false,$10)
		RETURNING id`,
		accountID,
		m.RecipientUserID,
		m.RecipientEmail,
		m.Title,
		m.Message,
		m.EntityType,
		m.EntityID,
		m.EntityNumber,
		m.Action,
		createdBy,
	).Scan(&id)
	return id, err
}

func (m Notification) FindList(recipientUserID int64, recipientEmail string, search string, limit int) ([]Notification, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	sql := `SELECT ` + notificationSelectColumns + ` FROM notifications n WHERE n.account_id = $1`
	args := []interface{}{accountID}
	if recipientUserID > 0 {
		args = append(args, recipientUserID)
		sql += ` AND n.recipient_user_id = $2`
	} else if strings.TrimSpace(recipientEmail) != "" {
		args = append(args, strings.ToLower(strings.TrimSpace(recipientEmail)))
		sql += ` AND LOWER(n.recipient_email) = $2`
	} else {
		return []Notification{}, nil
	}
	if strings.TrimSpace(search) != "" {
		args = append(args, search)
		placeholder := nextSQLPlaceholder(args)
		sql += ` AND (n.title ILIKE '%' || ` + placeholder + ` || '%' OR n.message ILIKE '%' || ` + placeholder + ` || '%' OR n.entity_number ILIKE '%' || ` + placeholder + ` || '%')`
	}
	sql += ` ORDER BY n.created_at DESC, n.id DESC LIMIT ` + nextSQLPlaceholder(append(args, limit))
	args = append(args, limit)

	var items []Notification
	err = m.DB.Select(&items, sql, args...)
	return items, err
}

func (m Notification) CountUnread(recipientUserID int64, recipientEmail string) (int64, error) {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return 0, err
	}
	sql := `SELECT COUNT(*) FROM notifications WHERE account_id = $1 AND is_read = false`
	args := []interface{}{accountID}
	if recipientUserID > 0 {
		args = append(args, recipientUserID)
		sql += ` AND recipient_user_id = $2`
	} else if strings.TrimSpace(recipientEmail) != "" {
		args = append(args, strings.ToLower(strings.TrimSpace(recipientEmail)))
		sql += ` AND LOWER(recipient_email) = $2`
	} else {
		return 0, nil
	}

	var count int64
	err = m.DB.Get(&count, sql, args...)
	return count, err
}

func (m Notification) MarkRead(id int64) error {
	accountID, err := m.resolveAccountID()
	if err != nil {
		return err
	}
	if id <= 0 {
		return errors.New("id notification tidak valid")
	}
	updatedBy := m.Auth.Email
	if updatedBy == "" {
		updatedBy = m.Auth.UserName
	}
	_, err = m.DB.Exec(`UPDATE notifications SET is_read = true, read_at = NOW(), updated_at = NOW(), updated_by = $1 WHERE id = $2 AND account_id = $3`, updatedBy, id, accountID)
	return err
}

func (m Notification) Normalize() {
	m.RecipientEmail = strings.ToLower(strings.TrimSpace(m.RecipientEmail))
	m.Title = strings.TrimSpace(m.Title)
	m.Message = strings.TrimSpace(m.Message)
	m.EntityType = strings.ToUpper(strings.TrimSpace(m.EntityType))
	m.EntityNumber = strings.TrimSpace(m.EntityNumber)
	m.Action = strings.ToUpper(strings.TrimSpace(m.Action))
}

func (m Notification) resolveAccountID() (int64, error) {
	if m.Auth.AccountID > 0 {
		return m.Auth.AccountID, nil
	}
	if m.AccountID > 0 {
		return m.AccountID, nil
	}
	return 0, errors.New("id akun wajib diisi")
}
