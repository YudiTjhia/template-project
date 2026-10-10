package models

import (
	"backend/shared/mmodels"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	ApprovalEntityBudgetAdjustment = "BUDGET_ADJUSTMENT"
	ApprovalEntityBudgetControl    = "BUDGET_CONTROL"
)

const (
	ApprovalApproverTypeUser          = "USER"
	ApprovalApproverTypeRole          = "ROLE"
	ApprovalApproverTypeDepartmentMgr = "DEPARTMENT_MANAGER"
	ApprovalApproverTypeCostCenterMgr = "COST_CENTER_MANAGER"
)

const (
	ApprovalRequiredActionAny = "ANY"
	ApprovalRequiredActionAll = "ALL"
)

const (
	ApprovalRequestStatusPending   = "Pending"
	ApprovalRequestStatusApproved  = "Approved"
	ApprovalRequestStatusRejected  = "Rejected"
	ApprovalRequestStatusRevision  = "Revision"
	ApprovalRequestStatusCancelled = "Cancelled"
)

const (
	ApprovalStepStatusPending  = "Pending"
	ApprovalStepStatusApproved = "Approved"
	ApprovalStepStatusRejected = "Rejected"
	ApprovalStepStatusRevision = "Revision"
)

type ApprovalRule struct {
	BaseModel

	ID           int64   `json:"id" db:"id"`
	AccountID    int64   `json:"account_id" db:"account_id"`
	EntityType   string  `json:"entityType" db:"entity_type"`
	RuleName     string  `json:"ruleName" db:"rule_name"`
	DepartmentID *int64  `json:"departmentId,omitempty" db:"department_id"`
	CostCenterID *int64  `json:"costCenterId,omitempty" db:"cost_center_id"`
	MinAmount    float64 `json:"minAmount" db:"min_amount"`
	MaxAmount    float64 `json:"maxAmount" db:"max_amount"`
	Priority     int     `json:"priority" db:"priority"`
	Active       bool    `json:"active" db:"active"`

	Steps []ApprovalRuleStep `json:"steps,omitempty" db:"-"`
}

type ApprovalRuleStep struct {
	BaseModel

	ID             int64   `json:"id" db:"id"`
	AccountID      int64   `json:"account_id" db:"account_id"`
	ApprovalRuleID int64   `json:"approvalRuleId" db:"approval_rule_id"`
	Level          int     `json:"level" db:"level"`
	ApproverType   string  `json:"approverType" db:"approver_type"`
	ApproverUserID *int64  `json:"approverUserId,omitempty" db:"approver_user_id"`
	ApproverName   *string `json:"approverName,omitempty" db:"approver_name"`
	ApproverRoleID *int64  `json:"approverRoleId,omitempty" db:"approver_role_id"`
	ApproverRole   *string `json:"approverRole,omitempty" db:"approver_role"`
	RequiredAction string  `json:"requiredAction" db:"required_action"`
	Active         bool    `json:"active" db:"active"`
}

type ApprovalRequest struct {
	BaseModel

	ID            int64      `json:"id" db:"id"`
	AccountID     int64      `json:"account_id" db:"account_id"`
	EntityType    string     `json:"entityType" db:"entity_type"`
	EntityID      int64      `json:"entityId" db:"entity_id"`
	EntityNumber  string     `json:"entityNumber" db:"entity_number"`
	DepartmentID  *int64     `json:"departmentId,omitempty" db:"department_id"`
	CostCenterID  *int64     `json:"costCenterId,omitempty" db:"cost_center_id"`
	Amount        float64    `json:"amount" db:"amount"`
	Status        string     `json:"status" db:"status"`
	CurrentLevel  int        `json:"currentLevel" db:"current_level"`
	SubmittedAt   *time.Time `json:"submittedAt,omitempty" db:"submitted_at"`
	ApprovedAt    *time.Time `json:"approvedAt,omitempty" db:"approved_at"`
	RejectedAt    *time.Time `json:"rejectedAt,omitempty" db:"rejected_at"`
	RejectionNote string     `json:"rejectionNote" db:"rejection_note"`

	Steps []ApprovalRequestStep `json:"steps,omitempty" db:"-"`
}

type ApprovalRequestStep struct {
	BaseModel

	ID                int64      `json:"id" db:"id"`
	AccountID         int64      `json:"account_id" db:"account_id"`
	ApprovalRequestID int64      `json:"approvalRequestId" db:"approval_request_id"`
	Level             int        `json:"level" db:"level"`
	ApproverType      string     `json:"approverType" db:"approver_type"`
	ApproverUserID    *int64     `json:"approverUserId,omitempty" db:"approver_user_id"`
	ApproverName      *string    `json:"approverName,omitempty" db:"approver_name"`
	ApproverRoleID    *int64     `json:"approverRoleId,omitempty" db:"approver_role_id"`
	ApproverRole      *string    `json:"approverRole,omitempty" db:"approver_role"`
	RequiredAction    string     `json:"requiredAction" db:"required_action"`
	Status            string     `json:"status" db:"status"`
	ActionBy          *string    `json:"actionBy,omitempty" db:"action_by"`
	ActionAt          *time.Time `json:"actionAt,omitempty" db:"action_at"`
	Notes             string     `json:"notes" db:"notes"`
}

const approvalRuleSelectColumns = `ar.id, ar.account_id, ar.entity_type, ar.rule_name, ar.department_id, ar.cost_center_id, ar.min_amount, ar.max_amount, ar.priority, ar.active,
ar.created_at, ar.created_by, ar.updated_at, ar.updated_by`

const approvalRuleStepSelectColumns = `ars.id, ars.account_id, ars.approval_rule_id, ars.level, ars.approver_type, ars.approver_user_id, ars.approver_name, ars.approver_role_id, ars.approver_role, ars.required_action, ars.active,
ars.created_at, ars.created_by, ars.updated_at, ars.updated_by`

const approvalRequestSelectColumns = `aq.id, aq.account_id, aq.entity_type, aq.entity_id, aq.entity_number, aq.department_id, aq.cost_center_id, aq.amount, aq.status, aq.current_level, aq.submitted_at, aq.approved_at, aq.rejected_at, aq.rejection_note,
aq.created_at, aq.created_by, aq.updated_at, aq.updated_by`

const approvalRequestStepSelectColumns = `aqs.id, aqs.account_id, aqs.approval_request_id, aqs.level, aqs.approver_type, aqs.approver_user_id, aqs.approver_name, aqs.approver_role_id, aqs.approver_role, aqs.required_action, aqs.status, aqs.action_by, aqs.action_at, aqs.notes,
aqs.created_at, aqs.created_by, aqs.updated_at, aqs.updated_by`

func NewApprovalRule(db *sqlx.DB, auth Auth) ApprovalRule {
	item := ApprovalRule{BaseModel: BaseModel{DB: db, Auth: auth}}
	item.AccountID = auth.AccountID
	item.CreatedBy = auth.Email
	item.Active = true
	return item
}

func NewApprovalRuleStep(db *sqlx.DB, auth Auth) ApprovalRuleStep {
	item := ApprovalRuleStep{BaseModel: BaseModel{DB: db, Auth: auth}}
	item.AccountID = auth.AccountID
	item.CreatedBy = auth.Email
	item.Active = true
	item.RequiredAction = ApprovalRequiredActionAny
	return item
}

func NewApprovalRequest(db *sqlx.DB, auth Auth) ApprovalRequest {
	item := ApprovalRequest{BaseModel: BaseModel{DB: db, Auth: auth}}
	item.AccountID = auth.AccountID
	item.CreatedBy = auth.Email
	item.Status = ApprovalRequestStatusPending
	return item
}

func NewApprovalRequestStep(db *sqlx.DB, auth Auth) ApprovalRequestStep {
	item := ApprovalRequestStep{BaseModel: BaseModel{DB: db, Auth: auth}}
	item.AccountID = auth.AccountID
	item.CreatedBy = auth.Email
	item.Status = ApprovalStepStatusPending
	item.RequiredAction = ApprovalRequiredActionAny
	return item
}

func (m ApprovalRule) Migrate() {
	sql := `CREATE TABLE IF NOT EXISTS ` + mmodels.ApprovalRuleTable + ` (
	id SERIAL PRIMARY KEY,
	account_id BIGINT NOT NULL DEFAULT 0,
	entity_type VARCHAR(50) NOT NULL DEFAULT '',
	rule_name VARCHAR(255) NOT NULL DEFAULT '',
	department_id BIGINT,
	cost_center_id BIGINT,
	min_amount NUMERIC(18,2) NOT NULL DEFAULT 0,
	max_amount NUMERIC(18,2) NOT NULL DEFAULT 0,
	priority INT NOT NULL DEFAULT 0,
	active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_by VARCHAR(255) NOT NULL,
	updated_at TIMESTAMPTZ,
	updated_by VARCHAR(255)
);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `CREATE INDEX IF NOT EXISTS idx_approval_rules_account_id ON ` + mmodels.ApprovalRuleTable + ` (account_id);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `CREATE INDEX IF NOT EXISTS idx_approval_rules_entity ON ` + mmodels.ApprovalRuleTable + ` (account_id, entity_type, active);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRuleTable + ` ADD COLUMN IF NOT EXISTS account_id BIGINT NOT NULL DEFAULT 0;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRuleTable + ` ADD COLUMN IF NOT EXISTS entity_type VARCHAR(50) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRuleTable + ` ADD COLUMN IF NOT EXISTS rule_name VARCHAR(255) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRuleTable + ` ADD COLUMN IF NOT EXISTS department_id BIGINT;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRuleTable + ` ADD COLUMN IF NOT EXISTS cost_center_id BIGINT;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRuleTable + ` ADD COLUMN IF NOT EXISTS min_amount NUMERIC(18,2) NOT NULL DEFAULT 0;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRuleTable + ` ADD COLUMN IF NOT EXISTS max_amount NUMERIC(18,2) NOT NULL DEFAULT 0;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRuleTable + ` ADD COLUMN IF NOT EXISTS priority INT NOT NULL DEFAULT 0;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRuleTable + ` ADD COLUMN IF NOT EXISTS active BOOLEAN NOT NULL DEFAULT true;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
}

func (m ApprovalRuleStep) Migrate() {
	sql := `CREATE TABLE IF NOT EXISTS ` + mmodels.ApprovalRuleStepTable + ` (
	id SERIAL PRIMARY KEY,
	account_id BIGINT NOT NULL DEFAULT 0,
	approval_rule_id BIGINT NOT NULL,
	level INT NOT NULL DEFAULT 1,
	approver_type VARCHAR(50) NOT NULL DEFAULT '',
	approver_user_id BIGINT,
	approver_name VARCHAR(255),
	approver_role_id BIGINT,
	approver_role VARCHAR(255),
	required_action VARCHAR(20) NOT NULL DEFAULT 'ANY',
	active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_by VARCHAR(255) NOT NULL,
	updated_at TIMESTAMPTZ,
	updated_by VARCHAR(255)
);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `CREATE INDEX IF NOT EXISTS idx_approval_rule_steps_rule_id ON ` + mmodels.ApprovalRuleStepTable + ` (account_id, approval_rule_id, level);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRuleStepTable + ` ADD COLUMN IF NOT EXISTS account_id BIGINT NOT NULL DEFAULT 0;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRuleStepTable + ` ADD COLUMN IF NOT EXISTS approval_rule_id BIGINT NOT NULL;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRuleStepTable + ` ADD COLUMN IF NOT EXISTS level INT NOT NULL DEFAULT 1;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRuleStepTable + ` ADD COLUMN IF NOT EXISTS approver_type VARCHAR(50) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRuleStepTable + ` ADD COLUMN IF NOT EXISTS approver_user_id BIGINT;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRuleStepTable + ` ADD COLUMN IF NOT EXISTS approver_name VARCHAR(255);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRuleStepTable + ` ADD COLUMN IF NOT EXISTS approver_role_id BIGINT;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRuleStepTable + ` ADD COLUMN IF NOT EXISTS approver_role VARCHAR(255);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRuleStepTable + ` ADD COLUMN IF NOT EXISTS required_action VARCHAR(20) NOT NULL DEFAULT 'ANY';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRuleStepTable + ` ADD COLUMN IF NOT EXISTS active BOOLEAN NOT NULL DEFAULT true;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
}

func (m ApprovalRequest) Migrate() {
	sql := `CREATE TABLE IF NOT EXISTS ` + mmodels.ApprovalRequestTable + ` (
	id SERIAL PRIMARY KEY,
	account_id BIGINT NOT NULL DEFAULT 0,
	entity_type VARCHAR(50) NOT NULL DEFAULT '',
	entity_id BIGINT NOT NULL DEFAULT 0,
	entity_number VARCHAR(100) NOT NULL DEFAULT '',
	department_id BIGINT,
	cost_center_id BIGINT,
	amount NUMERIC(18,2) NOT NULL DEFAULT 0,
	status VARCHAR(50) NOT NULL DEFAULT 'Pending',
	current_level INT NOT NULL DEFAULT 0,
	submitted_at TIMESTAMPTZ,
	approved_at TIMESTAMPTZ,
	rejected_at TIMESTAMPTZ,
	rejection_note TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_by VARCHAR(255) NOT NULL,
	updated_at TIMESTAMPTZ,
	updated_by VARCHAR(255)
);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `CREATE INDEX IF NOT EXISTS idx_approval_requests_account_id ON ` + mmodels.ApprovalRequestTable + ` (account_id);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `CREATE INDEX IF NOT EXISTS idx_approval_requests_entity ON ` + mmodels.ApprovalRequestTable + ` (account_id, entity_type, entity_id);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestTable + ` ADD COLUMN IF NOT EXISTS account_id BIGINT NOT NULL DEFAULT 0;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestTable + ` ADD COLUMN IF NOT EXISTS entity_type VARCHAR(50) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestTable + ` ADD COLUMN IF NOT EXISTS entity_id BIGINT NOT NULL DEFAULT 0;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestTable + ` ADD COLUMN IF NOT EXISTS entity_number VARCHAR(100) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestTable + ` ADD COLUMN IF NOT EXISTS department_id BIGINT;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestTable + ` ADD COLUMN IF NOT EXISTS cost_center_id BIGINT;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestTable + ` ADD COLUMN IF NOT EXISTS amount NUMERIC(18,2) NOT NULL DEFAULT 0;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestTable + ` ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'Pending';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestTable + ` ADD COLUMN IF NOT EXISTS current_level INT NOT NULL DEFAULT 0;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestTable + ` ADD COLUMN IF NOT EXISTS submitted_at TIMESTAMPTZ;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestTable + ` ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestTable + ` ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMPTZ;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestTable + ` ADD COLUMN IF NOT EXISTS rejection_note TEXT NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
}

func (m ApprovalRequestStep) Migrate() {
	sql := `CREATE TABLE IF NOT EXISTS ` + mmodels.ApprovalRequestStepTable + ` (
	id SERIAL PRIMARY KEY,
	account_id BIGINT NOT NULL DEFAULT 0,
	approval_request_id BIGINT NOT NULL,
	level INT NOT NULL DEFAULT 1,
	approver_type VARCHAR(50) NOT NULL DEFAULT '',
	approver_user_id BIGINT,
	approver_name VARCHAR(255),
	approver_role_id BIGINT,
	approver_role VARCHAR(255),
	required_action VARCHAR(20) NOT NULL DEFAULT 'ANY',
	status VARCHAR(50) NOT NULL DEFAULT 'Pending',
	action_by VARCHAR(255),
	action_at TIMESTAMPTZ,
	notes TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_by VARCHAR(255) NOT NULL,
	updated_at TIMESTAMPTZ,
	updated_by VARCHAR(255)
);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `CREATE INDEX IF NOT EXISTS idx_approval_request_steps_request_id ON ` + mmodels.ApprovalRequestStepTable + ` (account_id, approval_request_id, level);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestStepTable + ` ADD COLUMN IF NOT EXISTS account_id BIGINT NOT NULL DEFAULT 0;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestStepTable + ` ADD COLUMN IF NOT EXISTS approval_request_id BIGINT NOT NULL;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestStepTable + ` ADD COLUMN IF NOT EXISTS level INT NOT NULL DEFAULT 1;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestStepTable + ` ADD COLUMN IF NOT EXISTS approver_type VARCHAR(50) NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestStepTable + ` ADD COLUMN IF NOT EXISTS approver_user_id BIGINT;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestStepTable + ` ADD COLUMN IF NOT EXISTS approver_name VARCHAR(255);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestStepTable + ` ADD COLUMN IF NOT EXISTS approver_role_id BIGINT;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestStepTable + ` ADD COLUMN IF NOT EXISTS approver_role VARCHAR(255);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestStepTable + ` ADD COLUMN IF NOT EXISTS required_action VARCHAR(20) NOT NULL DEFAULT 'ANY';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestStepTable + ` ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'Pending';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestStepTable + ` ADD COLUMN IF NOT EXISTS action_by VARCHAR(255);`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestStepTable + ` ADD COLUMN IF NOT EXISTS action_at TIMESTAMPTZ;`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
	sql = `ALTER TABLE ` + mmodels.ApprovalRequestStepTable + ` ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT '';`
	log.Println("[SQL]=", sql)
	m.DB.MustExec(sql)
}

func (m ApprovalRule) TableName() string        { return mmodels.ApprovalRuleTable }
func (m ApprovalRuleStep) TableName() string    { return mmodels.ApprovalRuleStepTable }
func (m ApprovalRequest) TableName() string     { return mmodels.ApprovalRequestTable }
func (m ApprovalRequestStep) TableName() string { return mmodels.ApprovalRequestStepTable }

func (m ApprovalRule) ValidateCreate() error {
	if _, err := m.resolveAccountID(); err != nil {
		return err
	}
	if strings.TrimSpace(m.EntityType) == "" {
		return errors.New("entity type wajib diisi")
	}
	if strings.TrimSpace(m.RuleName) == "" {
		return errors.New("rule name wajib diisi")
	}
	if m.MinAmount < 0 || m.MaxAmount < 0 {
		return errors.New("amount harus valid")
	}
	if m.MaxAmount > 0 && m.MaxAmount < m.MinAmount {
		return errors.New("max amount harus lebih besar dari min amount")
	}
	return nil
}

func (m ApprovalRuleStep) ValidateCreate() error {
	if _, err := m.resolveAccountID(); err != nil {
		return err
	}
	if m.ApprovalRuleID <= 0 {
		return errors.New("approval rule id wajib diisi")
	}
	if m.Level <= 0 {
		return errors.New("level wajib diisi")
	}
	if strings.TrimSpace(m.ApproverType) == "" {
		return errors.New("approver type wajib diisi")
	}
	return nil
}

func (m ApprovalRequest) ValidateCreate() error {
	if _, err := m.resolveAccountID(); err != nil {
		return err
	}
	if strings.TrimSpace(m.EntityType) == "" {
		return errors.New("entity type wajib diisi")
	}
	if m.EntityID <= 0 {
		return errors.New("entity id wajib diisi")
	}
	if strings.TrimSpace(m.EntityNumber) == "" {
		return errors.New("entity number wajib diisi")
	}
	return nil
}

func (m ApprovalRequestStep) ValidateCreate() error {
	if _, err := m.resolveAccountID(); err != nil {
		return err
	}
	if m.ApprovalRequestID <= 0 {
		return errors.New("approval request id wajib diisi")
	}
	if m.Level <= 0 {
		return errors.New("level wajib diisi")
	}
	if strings.TrimSpace(m.ApproverType) == "" {
		return errors.New("approver type wajib diisi")
	}
	return nil
}

func (m ApprovalRule) ValidateID() error {
	if m.ID <= 0 {
		return errors.New("id approval rule tidak valid")
	}
	return nil
}

func (m ApprovalRuleStep) ValidateID() error {
	if m.ID <= 0 {
		return errors.New("id approval rule step tidak valid")
	}
	return nil
}

func (m ApprovalRequest) ValidateID() error {
	if m.ID <= 0 {
		return errors.New("id approval request tidak valid")
	}
	return nil
}

func (m ApprovalRequestStep) ValidateID() error {
	if m.ID <= 0 {
		return errors.New("id approval request step tidak valid")
	}
	return nil
}

func (m ApprovalRule) resolveAccountID() (int64, error) {
	return resolveGenericAccountID(m.AccountID, m.Auth.AccountID)
}
func (m ApprovalRuleStep) resolveAccountID() (int64, error) {
	return resolveGenericAccountID(m.AccountID, m.Auth.AccountID)
}
func (m ApprovalRequest) resolveAccountID() (int64, error) {
	return resolveGenericAccountID(m.AccountID, m.Auth.AccountID)
}
func (m ApprovalRequestStep) resolveAccountID() (int64, error) {
	return resolveGenericAccountID(m.AccountID, m.Auth.AccountID)
}

func resolveGenericAccountID(current, authAccountID int64) (int64, error) {
	if authAccountID > 0 {
		return authAccountID, nil
	}
	if current > 0 {
		return current, nil
	}
	return 0, fmt.Errorf("id akun wajib diisi")
}
