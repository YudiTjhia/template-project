package services

import (
	"backend/shared/mmodels"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend/shared/models"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type approverContext struct {
	userID    int64
	roleIDs   []int64
	roleNames []string
}

type ApprovalService struct {
	DB            *sqlx.DB
	TransactionDB *sqlx.DB
	Auth          models.Auth
}

type ApprovalSubmitInput struct {
	EntityType   string
	EntityID     int64
	EntityNumber string
	Amount       float64
	DepartmentID *int64
	CostCenterID *int64
}

type ApprovalActionInput struct {
	RequestID int64
	Notes     string
}

func NewApprovalService(db *sqlx.DB, auth models.Auth) *ApprovalService {
	return &ApprovalService{DB: db, Auth: auth}
}

func NewApprovalServiceWithTransactionDB(db *sqlx.DB, transactionDB *sqlx.DB, auth models.Auth) *ApprovalService {
	if transactionDB == nil {
		transactionDB = db
	}
	return &ApprovalService{DB: db, TransactionDB: transactionDB, Auth: auth}
}

func (s *ApprovalService) transactionDB() *sqlx.DB {
	if s.TransactionDB != nil {
		return s.TransactionDB
	}
	return s.DB
}

func (s *ApprovalService) Migrate() {
	models.NewApprovalRule(s.DB, s.Auth).Migrate()
	models.NewApprovalRuleStep(s.DB, s.Auth).Migrate()
	models.NewApprovalRequest(s.DB, s.Auth).Migrate()
	models.NewApprovalRequestStep(s.DB, s.Auth).Migrate()
}

func (s *ApprovalService) SaveRule(rule models.ApprovalRule) (int64, error) {
	if rule.ID > 0 {
		if err := s.updateRule(rule); err != nil {
			return 0, err
		}
		return rule.ID, nil
	}
	return s.createRule(rule)
}

func (s *ApprovalService) createRule(rule models.ApprovalRule) (int64, error) {
	tx, err := s.DB.BeginTxx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	rule.Auth = s.Auth
	rule.EntityType = strings.ToUpper(strings.TrimSpace(rule.EntityType))
	rule.RuleName = strings.TrimSpace(rule.RuleName)
	rule.Active = true
	if err := rule.ValidateCreate(); err != nil {
		return 0, err
	}

	createdBy := rule.CreatedBy
	if createdBy == "" {
		createdBy = s.Auth.Email
	}
	if createdBy == "" {
		return 0, errors.New("created_by wajib diisi")
	}

	var id int64
	err = tx.QueryRow(
		`INSERT INTO `+mmodels.ApprovalRuleTable+`
		(account_id, entity_type, rule_name, department_id, cost_center_id, min_amount, max_amount, priority, active, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id`,
		s.Auth.AccountID,
		rule.EntityType,
		rule.RuleName,
		rule.DepartmentID,
		rule.CostCenterID,
		rule.MinAmount,
		rule.MaxAmount,
		rule.Priority,
		rule.Active,
		createdBy,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	if err = s.replaceRuleStepsTx(tx, id, rule.Steps, createdBy); err != nil {
		return 0, err
	}
	rule.ID = id
	if err = s.syncPendingRequestsForRuleTx(tx, rule, createdBy); err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}
	tx = nil
	return id, nil
}

func (s *ApprovalService) updateRule(rule models.ApprovalRule) error {
	tx, err := s.DB.BeginTxx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	rule.Auth = s.Auth
	rule.EntityType = strings.ToUpper(strings.TrimSpace(rule.EntityType))
	rule.RuleName = strings.TrimSpace(rule.RuleName)
	if err := rule.ValidateCreate(); err != nil {
		return err
	}

	updatedBy := rule.UpdatedBy
	if updatedBy == nil || *updatedBy == "" {
		if s.Auth.Email != "" {
			updatedBy = &s.Auth.Email
		}
	}

	_, err = tx.Exec(
		`UPDATE `+mmodels.ApprovalRuleTable+`
		SET entity_type = $1,
			rule_name = $2,
			department_id = $3,
			cost_center_id = $4,
			min_amount = $5,
			max_amount = $6,
			priority = $7,
			active = $8,
			updated_at = NOW(),
			updated_by = $9
		WHERE id = $10 AND account_id = $11`,
		rule.EntityType,
		rule.RuleName,
		rule.DepartmentID,
		rule.CostCenterID,
		rule.MinAmount,
		rule.MaxAmount,
		rule.Priority,
		rule.Active,
		updatedBy,
		rule.ID,
		s.Auth.AccountID,
	)
	if err != nil {
		return err
	}

	if err = s.replaceRuleStepsTx(tx, rule.ID, rule.Steps, pointerString(updatedBy)); err != nil {
		return err
	}
	if err = s.syncPendingRequestsForRuleTx(tx, rule, pointerString(updatedBy)); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	tx = nil
	return nil
}

func (s *ApprovalService) DeleteRule(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM `+mmodels.ApprovalRuleTable+` WHERE id = $1 AND account_id = $2`, id, s.Auth.AccountID)
	return err
}

func (s *ApprovalService) FindRuleByID(id int64) (*models.ApprovalRule, error) {
	var rule models.ApprovalRule
	if err := s.DB.Get(&rule, `SELECT id, account_id, entity_type, rule_name, department_id, cost_center_id, min_amount, max_amount, priority, active, created_at, created_by, updated_at, updated_by FROM `+mmodels.ApprovalRuleTable+` WHERE id = $1 AND account_id = $2`, id, s.Auth.AccountID); err != nil {
		return nil, err
	}
	steps, err := s.FindRuleStepsByRuleID(id)
	if err != nil {
		return nil, err
	}
	rule.Steps = steps
	return &rule, nil
}

func (s *ApprovalService) FindRuleStepsByRuleID(ruleID int64) ([]models.ApprovalRuleStep, error) {
	var steps []models.ApprovalRuleStep
	err := s.DB.Select(&steps, `SELECT ars.id, ars.account_id, ars.approval_rule_id, ars.level, ars.approver_type, ars.approver_user_id,
			CASE
				WHEN ars.approver_user_id IS NOT NULL THEN COALESCE(NULLIF(u.username, ''), NULLIF(u.user_email, ''), ars.approver_name)
				ELSE ars.approver_name
			END AS approver_name,
			ars.approver_role_id, ars.approver_role, ars.required_action, ars.active, ars.created_at, ars.created_by, ars.updated_at, ars.updated_by
		FROM `+mmodels.ApprovalRuleStepTable+` ars
		LEFT JOIN `+mmodels.UserTable+` u ON u.account_id = ars.account_id AND u.id = ars.approver_user_id
		WHERE ars.approval_rule_id = $1 AND ars.account_id = $2
		ORDER BY ars.level ASC, ars.id ASC`, ruleID, s.Auth.AccountID)
	return steps, err
}

func (s *ApprovalService) FindRules(search string) ([]models.ApprovalRule, error) {
	sql := `SELECT id, account_id, entity_type, rule_name, department_id, cost_center_id, min_amount, max_amount, priority, active, created_at, created_by, updated_at, updated_by
		FROM ` + mmodels.ApprovalRuleTable + ` WHERE account_id = $1`
	args := []interface{}{s.Auth.AccountID}
	if strings.TrimSpace(search) != "" {
		args = append(args, search)
		ph := fmt.Sprintf("$%d", len(args))
		sql += ` AND (entity_type ILIKE '%' || ` + ph + ` || '%' OR rule_name ILIKE '%' || ` + ph + ` || '%')`
	}
	sql += ` ORDER BY priority ASC, id DESC LIMIT 10`
	var items []models.ApprovalRule
	err := s.DB.Select(&items, sql, args...)
	return items, err
}

func (s *ApprovalService) FindAllRules() ([]models.ApprovalRule, error) {
	var items []models.ApprovalRule
	err := s.DB.Select(&items, `SELECT id, account_id, entity_type, rule_name, department_id, cost_center_id, min_amount, max_amount, priority, active, created_at, created_by, updated_at, updated_by
		FROM `+mmodels.ApprovalRuleTable+` WHERE account_id = $1 ORDER BY priority ASC, id DESC LIMIT 10`, s.Auth.AccountID)
	return items, err
}

func (s *ApprovalService) pendingApprovalRequestFilter(search string) (string, []interface{}, error) {
	ctx, err := s.resolveApproverContext()
	if err != nil {
		return "", nil, err
	}

	sql := ` FROM ` + mmodels.ApprovalRequestTable + ` ar
		WHERE ar.account_id = $1
		  AND ar.status = $2
		  AND EXISTS (
			SELECT 1
			FROM ` + mmodels.ApprovalRequestStepTable + ` aqs
			WHERE aqs.account_id = ar.account_id
			  AND aqs.approval_request_id = ar.id
			  AND aqs.status = $3
			  AND (
				(aqs.approver_type = $4 AND aqs.approver_user_id = $5)
				OR (
					aqs.approver_type = $6
					AND (
						(aqs.approver_role_id IS NOT NULL AND aqs.approver_role_id = ANY($7::BIGINT[]))
						OR (COALESCE(aqs.approver_role, '') <> '' AND LOWER(aqs.approver_role) = ANY($8::TEXT[]))
					)
				)
			  )
		  )`
	args := []interface{}{
		s.Auth.AccountID,
		models.ApprovalRequestStatusPending,
		models.ApprovalStepStatusPending,
		models.ApprovalApproverTypeUser,
		ctx.userID,
		models.ApprovalApproverTypeRole,
		pq.Array(ctx.roleIDs),
		pq.Array(ctx.roleNames),
	}
	if strings.TrimSpace(search) != "" {
		args = append(args, search)
		ph := fmt.Sprintf("$%d", len(args))
		sql += ` AND (ar.entity_type ILIKE '%' || ` + ph + ` || '%' OR ar.entity_number ILIKE '%' || ` + ph + ` || '%' OR ar.status ILIKE '%' || ` + ph + ` || '%')`
	}

	return sql, args, nil
}

func (s *ApprovalService) FindRequests(search string) ([]models.ApprovalRequest, error) {
	filterSQL, args, err := s.pendingApprovalRequestFilter(search)
	if err != nil {
		return nil, err
	}

	sql := `SELECT ar.id, ar.account_id, ar.entity_type, ar.entity_id, ar.entity_number, ar.department_id, ar.cost_center_id, ar.amount, ar.status, ar.current_level, ar.submitted_at, ar.approved_at, ar.rejected_at, ar.rejection_note, ar.created_at, ar.created_by, ar.updated_at, ar.updated_by` +
		filterSQL +
		` ORDER BY ar.created_at DESC, ar.id DESC LIMIT 10`
	items := make([]models.ApprovalRequest, 0)
	err = s.DB.Select(&items, sql, args...)
	return items, err
}

func (s *ApprovalService) CountPendingRequests(search string) (int64, error) {
	filterSQL, args, err := s.pendingApprovalRequestFilter(search)
	if err != nil {
		return 0, err
	}

	var count int64
	err = s.DB.Get(&count, `SELECT COUNT(*)`+filterSQL, args...)
	return count, err
}

func (s *ApprovalService) FindRequestByID(id int64) (*models.ApprovalRequest, error) {
	var req models.ApprovalRequest
	if err := s.DB.Get(&req, `SELECT id, account_id, entity_type, entity_id, entity_number, department_id, cost_center_id, amount, status, current_level, submitted_at, approved_at, rejected_at, rejection_note, created_at, created_by, updated_at, updated_by
		FROM `+mmodels.ApprovalRequestTable+` WHERE id = $1 AND account_id = $2`, id, s.Auth.AccountID); err != nil {
		return nil, err
	}
	steps, err := s.FindRequestStepsByRequestID(id)
	if err != nil {
		return nil, err
	}
	req.Steps = steps
	return &req, nil
}

func (s *ApprovalService) FindRequestStepsByRequestID(requestID int64) ([]models.ApprovalRequestStep, error) {
	var steps []models.ApprovalRequestStep
	err := s.DB.Select(&steps, `SELECT aqs.id, aqs.account_id, aqs.approval_request_id, aqs.level, aqs.approver_type, aqs.approver_user_id,
			CASE
				WHEN aqs.approver_user_id IS NOT NULL THEN COALESCE(NULLIF(u.username, ''), NULLIF(u.user_email, ''), aqs.approver_name)
				ELSE aqs.approver_name
			END AS approver_name,
			aqs.approver_role_id, aqs.approver_role, aqs.required_action, aqs.status, aqs.action_by, aqs.action_at, aqs.notes, aqs.created_at, aqs.created_by, aqs.updated_at, aqs.updated_by
		FROM `+mmodels.ApprovalRequestStepTable+` aqs
		LEFT JOIN `+mmodels.UserTable+` u ON u.account_id = aqs.account_id AND u.id = aqs.approver_user_id
		WHERE aqs.approval_request_id = $1 AND aqs.account_id = $2
		ORDER BY aqs.level ASC, aqs.id ASC`, requestID, s.Auth.AccountID)
	return steps, err
}

func (s *ApprovalService) SubmitDocument(input ApprovalSubmitInput) (*models.ApprovalRequest, error) {
	input.EntityType = strings.ToUpper(strings.TrimSpace(input.EntityType))
	if input.EntityType == "" {
		return nil, errors.New("entity type wajib diisi")
	}
	if input.EntityID <= 0 {
		return nil, errors.New("entity id wajib diisi")
	}
	if strings.TrimSpace(input.EntityNumber) == "" {
		return nil, errors.New("entity number wajib diisi")
	}

	rule, err := s.findBestRule(input)
	if err != nil {
		return nil, err
	}

	tx, err := s.DB.BeginTxx(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	createdBy := s.Auth.Email
	if createdBy == "" {
		createdBy = s.Auth.UserName
	}
	if createdBy == "" {
		createdBy = "system"
	}

	status := models.ApprovalRequestStatusPending
	if len(rule.Steps) == 0 {
		status = models.ApprovalRequestStatusApproved
	}

	var approvalID int64
	if err = tx.QueryRow(
		`INSERT INTO `+mmodels.ApprovalRequestTable+`
		(account_id, entity_type, entity_id, entity_number, department_id, cost_center_id, amount, status, current_level, submitted_at, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),$10)
		RETURNING id`,
		s.Auth.AccountID,
		input.EntityType,
		input.EntityID,
		input.EntityNumber,
		input.DepartmentID,
		input.CostCenterID,
		input.Amount,
		status,
		0,
		createdBy,
	).Scan(&approvalID); err != nil {
		return nil, err
	}

	if err = s.insertRequestStepsTx(tx, approvalID, rule.Steps, createdBy); err != nil {
		return nil, err
	}

	targetStatus := models.ApprovalRequestStatusPending
	if status == models.ApprovalRequestStatusApproved {
		targetStatus = models.ApprovalRequestStatusApproved
		now := time.Now().UTC()
		if _, err = tx.Exec(`UPDATE `+mmodels.ApprovalRequestTable+` SET approved_at = $1, current_level = 0, updated_at = NOW(), updated_by = $2 WHERE id = $3 AND account_id = $4`, now, createdBy, approvalID, s.Auth.AccountID); err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil

	if err = s.syncDocumentStatus(input.EntityType, input.EntityID, "", targetStatus, createdBy); err != nil {
		return nil, err
	}

	req, err := s.FindRequestByID(approvalID)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (s *ApprovalService) ApproveRequest(input ApprovalActionInput) error {
	return s.applyAction(input, models.ApprovalStepStatusApproved)
}

func (s *ApprovalService) RejectRequest(input ApprovalActionInput) error {
	return s.applyAction(input, models.ApprovalStepStatusRejected)
}

func (s *ApprovalService) RequestRevision(input ApprovalActionInput) error {
	return s.applyAction(input, models.ApprovalStepStatusRevision)
}

func (s *ApprovalService) applyAction(input ApprovalActionInput, action string) error {
	tx, err := s.DB.BeginTxx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	request, err := s.getRequestTx(tx, input.RequestID)
	if err != nil {
		return err
	}

	step, err := s.findPendingStepTx(tx, request.ID)
	if err != nil {
		return err
	}
	if step == nil {
		return errors.New("tidak ada step approval yang pending")
	}

	if !s.isStepAssignableToCurrentUser(tx, request.ID, step) {
		return errors.New("request ini bukan untuk user yang sedang login")
	}

	now := time.Now().UTC()
	actor := s.Auth.Email
	if actor == "" {
		actor = s.Auth.UserName
	}
	if actor == "" {
		actor = "system"
	}

	if _, err = tx.Exec(
		`UPDATE `+mmodels.ApprovalRequestStepTable+`
		SET status = $1, action_by = $2, action_at = $3, notes = $4, updated_at = NOW(), updated_by = $2
		WHERE id = $5 AND account_id = $6`,
		action,
		actor,
		now,
		strings.TrimSpace(input.Notes),
		step.ID,
		s.Auth.AccountID,
	); err != nil {
		return err
	}

	finalApprovalStatus := ""
	switch action {
	case models.ApprovalStepStatusRejected:
		finalApprovalStatus = models.ApprovalRequestStatusRejected
		_, err = tx.Exec(`UPDATE `+mmodels.ApprovalRequestTable+` SET status = $1, rejected_at = $2, rejection_note = $3, updated_at = NOW(), updated_by = $4 WHERE id = $5 AND account_id = $6`,
			models.ApprovalRequestStatusRejected, now, strings.TrimSpace(input.Notes), actor, request.ID, s.Auth.AccountID)
		if err != nil {
			return err
		}
	case models.ApprovalStepStatusRevision:
		finalApprovalStatus = models.ApprovalRequestStatusRevision
		_, err = tx.Exec(`UPDATE `+mmodels.ApprovalRequestTable+` SET status = $1, updated_at = NOW(), updated_by = $2 WHERE id = $3 AND account_id = $4`,
			models.ApprovalRequestStatusRevision, actor, request.ID, s.Auth.AccountID)
		if err != nil {
			return err
		}
	case models.ApprovalStepStatusApproved:
		nextPending, err := s.findNextPendingStepTx(tx, request.ID, step.Level)
		if err != nil {
			return err
		}
		if nextPending == nil {
			finalApprovalStatus = models.ApprovalRequestStatusApproved
			_, err = tx.Exec(`UPDATE `+mmodels.ApprovalRequestTable+` SET status = $1, approved_at = $2, current_level = $3, updated_at = NOW(), updated_by = $4 WHERE id = $5 AND account_id = $6`,
				models.ApprovalRequestStatusApproved, now, step.Level, actor, request.ID, s.Auth.AccountID)
			if err != nil {
				return err
			}
		} else {
			finalApprovalStatus = models.ApprovalRequestStatusPending
			_, err = tx.Exec(`UPDATE `+mmodels.ApprovalRequestTable+` SET status = $1, current_level = $2, updated_at = NOW(), updated_by = $3 WHERE id = $4 AND account_id = $5`,
				models.ApprovalRequestStatusPending, nextPending.Level, actor, request.ID, s.Auth.AccountID)
			if err != nil {
				return err
			}
		}
	default:
		return errors.New("aksi approval tidak valid")
	}

	if finalApprovalStatus != "" {
		if err := s.syncDocumentStatusTx(tx, request.EntityType, request.EntityID, finalApprovalStatus, actor); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	tx = nil

	if finalApprovalStatus != "" {
		s.notifyApprovalChange(*request, finalApprovalStatus, actor)
	}
	return nil
}

func (s *ApprovalService) findBestRule(input ApprovalSubmitInput) (*models.ApprovalRule, error) {
	var rule models.ApprovalRule
	err := s.DB.Get(&rule, `
		SELECT id, account_id, entity_type, rule_name, department_id, cost_center_id, min_amount, max_amount, priority, active, created_at, created_by, updated_at, updated_by
		FROM `+mmodels.ApprovalRuleTable+`
		WHERE account_id = $1
		  AND active = true
		  AND entity_type = $2
		  AND (department_id IS NULL OR department_id = $3)
		  AND (cost_center_id IS NULL OR cost_center_id = $4)
		  AND (min_amount = 0 OR min_amount <= $5)
		  AND (max_amount = 0 OR max_amount >= $5)
		ORDER BY priority ASC, id ASC
		LIMIT 1`,
		s.Auth.AccountID,
		input.EntityType,
		input.DepartmentID,
		input.CostCenterID,
		input.Amount,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no rows") {
			return nil, s.approvalRuleMissingError(input)
		}
		return nil, err
	}

	steps, err := s.FindRuleStepsByRuleID(rule.ID)
	if err != nil {
		return nil, err
	}
	rule.Steps = steps
	return &rule, nil
}

func (s *ApprovalService) approvalRuleMissingError(input ApprovalSubmitInput) error {
	var entityRuleCount int
	if err := s.DB.Get(&entityRuleCount, `
		SELECT COUNT(*)
		FROM `+mmodels.ApprovalRuleTable+`
		WHERE account_id = $1
		  AND active = true
		  AND entity_type = $2`,
		s.Auth.AccountID,
		input.EntityType,
	); err != nil {
		return err
	}
	if entityRuleCount == 0 {
		return fmt.Errorf("approval rule tidak ditemukan untuk %s", input.EntityType)
	}

	if input.DepartmentID != nil {
		var departmentRuleCount int
		if err := s.DB.Get(&departmentRuleCount, `
			SELECT COUNT(*)
			FROM `+mmodels.ApprovalRuleTable+`
			WHERE account_id = $1
			  AND active = true
			  AND entity_type = $2
			  AND (department_id IS NULL OR department_id = $3)`,
			s.Auth.AccountID,
			input.EntityType,
			input.DepartmentID,
		); err != nil {
			return err
		}
		if departmentRuleCount == 0 {
			return fmt.Errorf("approval rule untuk %s department ID %d belum ada", input.EntityType, *input.DepartmentID)
		}
	}

	if input.CostCenterID != nil {
		var costCenterRuleCount int
		if err := s.DB.Get(&costCenterRuleCount, `
			SELECT COUNT(*)
			FROM `+mmodels.ApprovalRuleTable+`
			WHERE account_id = $1
			  AND active = true
			  AND entity_type = $2
			  AND (department_id IS NULL OR department_id = $3)
			  AND (cost_center_id IS NULL OR cost_center_id = $4)`,
			s.Auth.AccountID,
			input.EntityType,
			input.DepartmentID,
			input.CostCenterID,
		); err != nil {
			return err
		}
		if costCenterRuleCount == 0 {
			return fmt.Errorf("approval rule untuk %s cost center ID %d belum ada", input.EntityType, *input.CostCenterID)
		}
	}

	return fmt.Errorf("approval rule untuk %s belum ada yang sesuai amount %.2f", input.EntityType, input.Amount)
}

func (s *ApprovalService) replaceRuleStepsTx(tx *sqlx.Tx, ruleID int64, steps []models.ApprovalRuleStep, createdBy string) error {
	if len(steps) == 0 {
		return errors.New("approval rule harus memiliki minimal 1 approval step")
	}
	if _, err := tx.Exec(`DELETE FROM `+mmodels.ApprovalRuleStepTable+` WHERE approval_rule_id = $1 AND account_id = $2`, ruleID, s.Auth.AccountID); err != nil {
		return err
	}
	for idx := range steps {
		step := steps[idx]
		step.Auth = s.Auth
		step.AccountID = s.Auth.AccountID
		step.ApprovalRuleID = ruleID
		if step.Level <= 0 {
			step.Level = idx + 1
		}
		if strings.TrimSpace(step.RequiredAction) == "" {
			step.RequiredAction = models.ApprovalRequiredActionAny
		}
		if err := step.ValidateCreate(); err != nil {
			return err
		}
		if _, err := tx.Exec(
			`INSERT INTO `+mmodels.ApprovalRuleStepTable+`
			(account_id, approval_rule_id, level, approver_type, approver_user_id, approver_name, approver_role_id, approver_role, required_action, active, created_by)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
			s.Auth.AccountID,
			ruleID,
			step.Level,
			step.ApproverType,
			step.ApproverUserID,
			step.ApproverName,
			step.ApproverRoleID,
			step.ApproverRole,
			step.RequiredAction,
			step.Active,
			createdBy,
		); err != nil {
			return err
		}
	}
	return nil
}

func (s *ApprovalService) insertRequestStepsTx(tx *sqlx.Tx, requestID int64, steps []models.ApprovalRuleStep, createdBy string) error {
	if len(steps) == 0 {
		return nil
	}
	for idx := range steps {
		step := steps[idx]
		if _, err := tx.Exec(
			`INSERT INTO `+mmodels.ApprovalRequestStepTable+`
			(account_id, approval_request_id, level, approver_type, approver_user_id, approver_name, approver_role_id, approver_role, required_action, status, created_by)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
			s.Auth.AccountID,
			requestID,
			step.Level,
			step.ApproverType,
			step.ApproverUserID,
			step.ApproverName,
			step.ApproverRoleID,
			step.ApproverRole,
			step.RequiredAction,
			models.ApprovalStepStatusPending,
			createdBy,
		); err != nil {
			return err
		}
	}
	return nil
}

func (s *ApprovalService) syncPendingRequestsForRuleTx(tx *sqlx.Tx, rule models.ApprovalRule, updatedBy string) error {
	if !rule.Active || strings.TrimSpace(rule.EntityType) == "" || len(rule.Steps) == 0 {
		return nil
	}

	var requestIDs []int64
	err := tx.Select(&requestIDs, `
		SELECT ar.id
		FROM `+mmodels.ApprovalRequestTable+` ar
		WHERE ar.account_id = $1
		  AND ar.status = $2
		  AND ar.entity_type = $3
		  AND ($4::BIGINT IS NULL OR ar.department_id = $4)
		  AND ($5::BIGINT IS NULL OR ar.cost_center_id = $5)
		  AND ($6::NUMERIC = 0 OR ar.amount >= $6)
		  AND ($7::NUMERIC = 0 OR ar.amount <= $7)
		  AND NOT EXISTS (
			SELECT 1
			FROM `+mmodels.ApprovalRequestStepTable+` aqs
			WHERE aqs.account_id = ar.account_id
			  AND aqs.approval_request_id = ar.id
			  AND aqs.status <> $8
		  )`,
		s.Auth.AccountID,
		models.ApprovalRequestStatusPending,
		rule.EntityType,
		rule.DepartmentID,
		rule.CostCenterID,
		rule.MinAmount,
		rule.MaxAmount,
		models.ApprovalStepStatusPending,
	)
	if err != nil {
		return err
	}

	for _, requestID := range requestIDs {
		if _, err := tx.Exec(`DELETE FROM `+mmodels.ApprovalRequestStepTable+` WHERE account_id = $1 AND approval_request_id = $2`, s.Auth.AccountID, requestID); err != nil {
			return err
		}
		if err := s.insertRequestStepsTx(tx, requestID, rule.Steps, updatedBy); err != nil {
			return err
		}
		if _, err := tx.Exec(`UPDATE `+mmodels.ApprovalRequestTable+` SET current_level = 0, updated_at = NOW(), updated_by = $1 WHERE id = $2 AND account_id = $3`, updatedBy, requestID, s.Auth.AccountID); err != nil {
			return err
		}
	}
	return nil
}

func (s *ApprovalService) getRequestTx(tx *sqlx.Tx, requestID int64) (*models.ApprovalRequest, error) {
	var req models.ApprovalRequest
	if err := tx.Get(&req, `SELECT id, account_id, entity_type, entity_id, entity_number, department_id, cost_center_id, amount, status, current_level, submitted_at, approved_at, rejected_at, rejection_note, created_at, created_by, updated_at, updated_by
		FROM `+mmodels.ApprovalRequestTable+` WHERE id = $1 AND account_id = $2`, requestID, s.Auth.AccountID); err != nil {
		return nil, err
	}
	return &req, nil
}

func (s *ApprovalService) findPendingStepTx(tx *sqlx.Tx, requestID int64) (*models.ApprovalRequestStep, error) {
	var step models.ApprovalRequestStep
	err := tx.Get(&step, `SELECT id, account_id, approval_request_id, level, approver_type, approver_user_id, approver_name, approver_role_id, approver_role, required_action, status, action_by, action_at, notes, created_at, created_by, updated_at, updated_by
		FROM `+mmodels.ApprovalRequestStepTable+`
		WHERE approval_request_id = $1 AND account_id = $2 AND status = $3
		ORDER BY level ASC, id ASC
		LIMIT 1`, requestID, s.Auth.AccountID, models.ApprovalStepStatusPending)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no rows") {
			return nil, nil
		}
		return nil, err
	}
	return &step, nil
}

func (s *ApprovalService) isStepAssignableToCurrentUser(tx *sqlx.Tx, requestID int64, step *models.ApprovalRequestStep) bool {
	if step == nil {
		return false
	}

	ctx, err := s.resolveApproverContext()
	if err != nil {
		return false
	}

	var matched models.ApprovalRequestStep
	err = tx.Get(&matched, `SELECT id, account_id, approval_request_id, level, approver_type, approver_user_id, approver_name, approver_role_id, approver_role, required_action, status, action_by, action_at, notes, created_at, created_by, updated_at, updated_by
		FROM `+mmodels.ApprovalRequestStepTable+`
		WHERE approval_request_id = $1
		  AND account_id = $2
		  AND id = $3
		  AND status = $4
		  AND (
			(approver_type = $5 AND approver_user_id = $6)
			OR (
				approver_type = $7
				AND (
					(approver_role_id IS NOT NULL AND approver_role_id = ANY($8::BIGINT[]))
					OR (COALESCE(approver_role, '') <> '' AND LOWER(approver_role) = ANY($9::TEXT[]))
				)
			)
		  )
		LIMIT 1`,
		requestID,
		s.Auth.AccountID,
		step.ID,
		models.ApprovalStepStatusPending,
		models.ApprovalApproverTypeUser,
		ctx.userID,
		models.ApprovalApproverTypeRole,
		pq.Array(ctx.roleIDs),
		pq.Array(ctx.roleNames),
	)
	return err == nil
}

func (s *ApprovalService) resolveApproverContext() (*approverContext, error) {
	user, err := s.resolveCurrentUser()
	if err != nil {
		return nil, err
	}

	roleNames := make([]string, 0, len(user.UserRoles))
	seenNames := map[string]struct{}{}
	for _, roleName := range user.UserRoles {
		normalized := strings.ToLower(strings.TrimSpace(roleName))
		if normalized == "" {
			continue
		}
		if _, ok := seenNames[normalized]; ok {
			continue
		}
		seenNames[normalized] = struct{}{}
		roleNames = append(roleNames, normalized)
	}

	ctx := &approverContext{userID: user.ID, roleNames: roleNames}
	if len(roleNames) == 0 {
		return ctx, nil
	}

	var roleRows []struct {
		ID       int64  `db:"id"`
		RoleName string `db:"role_name"`
	}
	if err := s.DB.Select(&roleRows, `SELECT id, role_name FROM `+mmodels.RoleTable+` WHERE account_id = $1 AND LOWER(role_name) = ANY($2::TEXT[])`, s.Auth.AccountID, pq.Array(roleNames)); err != nil {
		return nil, err
	}

	ctx.roleIDs = make([]int64, 0, len(roleRows))
	seenIDs := map[int64]struct{}{}
	for _, row := range roleRows {
		if _, ok := seenIDs[row.ID]; ok {
			continue
		}
		seenIDs[row.ID] = struct{}{}
		ctx.roleIDs = append(ctx.roleIDs, row.ID)
	}

	return ctx, nil
}

func (s *ApprovalService) resolveCurrentUser() (*models.User, error) {
	userModel := models.NewUser(s.DB, s.Auth)
	if strings.TrimSpace(s.Auth.Email) != "" {
		userModel.UserEmail = s.Auth.Email
		if user, err := userModel.FindByEmail(); err == nil {
			return user, nil
		}
	}
	if strings.TrimSpace(s.Auth.UserName) != "" {
		userModel.Username = s.Auth.UserName
		if user, err := userModel.FindByUsername(); err == nil {
			return user, nil
		}
	}
	if strings.TrimSpace(s.Auth.Email) == "" && strings.TrimSpace(s.Auth.UserName) == "" {
		return nil, errors.New("email auth wajib diisi")
	}
	return nil, errors.New("user aktif tidak ditemukan")
}

func (s *ApprovalService) findNextPendingStepTx(tx *sqlx.Tx, requestID int64, currentLevel int) (*models.ApprovalRequestStep, error) {
	var step models.ApprovalRequestStep
	err := tx.Get(&step, `SELECT id, account_id, approval_request_id, level, approver_type, approver_user_id, approver_name, approver_role_id, approver_role, required_action, status, action_by, action_at, notes, created_at, created_by, updated_at, updated_by
		FROM `+mmodels.ApprovalRequestStepTable+`
		WHERE approval_request_id = $1 AND account_id = $2 AND status = $3 AND level > $4
		ORDER BY level ASC, id ASC
		LIMIT 1`, requestID, s.Auth.AccountID, models.ApprovalStepStatusPending, currentLevel)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no rows") {
			return nil, nil
		}
		return nil, err
	}
	return &step, nil
}

func (s *ApprovalService) syncDocumentStatus(entityType string, entityID int64, fromStatus string, approvalStatus string, actor string) error {
	_ = fromStatus
	entityType = strings.ToUpper(strings.TrimSpace(entityType))
	if entityType == models.ApprovalEntityBudgetControl || entityType == models.ApprovalEntityBudgetAdjustment {
		return s.syncApprovalEntityStatus(approvalEntityTableName(entityType), entityID, approvalStatus, actor)
	}
	return nil
}

func (s *ApprovalService) syncDocumentStatusTx(tx *sqlx.Tx, entityType string, entityID int64, approvalStatus string, actor string) error {
	entityType = strings.ToUpper(strings.TrimSpace(entityType))
	if entityType == models.ApprovalEntityBudgetControl || entityType == models.ApprovalEntityBudgetAdjustment {
		return s.syncApprovalEntityStatusTx(tx, approvalEntityTableName(entityType), entityID, approvalStatus, actor)
	}
	return nil
}

func (s *ApprovalService) syncApprovalEntityStatus(tableName string, entityID int64, approvalStatus string, actor string) error {
	if tableName == "" {
		return nil
	}
	status := map[string]string{
		models.ApprovalRequestStatusApproved: "Approved",
		models.ApprovalRequestStatusRejected: "Rejected",
		models.ApprovalRequestStatusRevision: "Revision",
		models.ApprovalRequestStatusPending:  "WaitingApproval",
	}[approvalStatus]
	if status == "" {
		status = approvalStatus
	}
	now := time.Now().UTC()
	var approvedAt *time.Time
	if approvalStatus == models.ApprovalRequestStatusApproved {
		approvedAt = &now
	}
	_, err := s.transactionDB().Exec(`UPDATE `+tableName+` SET status = $1, approved_at = COALESCE($2, approved_at), approved_by = CASE WHEN $2 IS NULL THEN approved_by ELSE $3 END, updated_at = NOW(), updated_by = $3 WHERE id = $4 AND account_id = $5`,
		status, approvedAt, actor, entityID, s.Auth.AccountID)
	if err != nil {
		return err
	}
	return nil
}

func (s *ApprovalService) syncApprovalEntityStatusTx(tx *sqlx.Tx, tableName string, entityID int64, approvalStatus string, actor string) error {
	if tableName == "" {
		return nil
	}
	status := map[string]string{
		models.ApprovalRequestStatusApproved: "Approved",
		models.ApprovalRequestStatusRejected: "Rejected",
		models.ApprovalRequestStatusRevision: "Revision",
		models.ApprovalRequestStatusPending:  "WaitingApproval",
	}[approvalStatus]
	if status == "" {
		status = approvalStatus
	}
	now := time.Now().UTC()
	var approvedAt *time.Time
	if approvalStatus == models.ApprovalRequestStatusApproved {
		approvedAt = &now
	}
	_, err := tx.Exec(`UPDATE `+tableName+` SET status = $1, approved_at = COALESCE($2, approved_at), approved_by = CASE WHEN $2 IS NULL THEN approved_by ELSE $3 END, updated_at = NOW(), updated_by = $3 WHERE id = $4 AND account_id = $5`,
		status, approvedAt, actor, entityID, s.Auth.AccountID)
	if err != nil {
		return err
	}
	return nil
}

func approvalEntityTableName(entityType string) string {
	switch strings.ToUpper(strings.TrimSpace(entityType)) {
	case models.ApprovalEntityBudgetControl, models.ApprovalEntityBudgetAdjustment:
		return "budget_controls"
	default:
		return ""
	}
}

func (s *ApprovalService) notifyApprovalChange(request models.ApprovalRequest, approvalStatus string, actor string) {
	entityType := strings.ToUpper(strings.TrimSpace(request.EntityType))
	statusLabel := approvalStatusLabel(approvalStatus)
	entityLabel := approvalEntityLabel(entityType)
	entityNumber := strings.TrimSpace(request.EntityNumber)

	title := fmt.Sprintf("%s approval %s", entityLabel, statusLabel)
	if entityNumber != "" {
		title = fmt.Sprintf("%s %s approval %s", entityLabel, entityNumber, statusLabel)
	}
	message := fmt.Sprintf("%s approval %s oleh %s.", entityLabel, statusLabel, actor)
	if entityNumber != "" {
		message = fmt.Sprintf("%s %s approval %s oleh %s.", entityLabel, entityNumber, statusLabel, actor)
	}

	NewNotificationService(s.DB, s.Auth).NotifyAccountUsers(NotificationEvent{
		EntityType:   entityType,
		EntityID:     request.EntityID,
		EntityNumber: entityNumber,
		Action:       "APPROVAL_" + strings.ToUpper(strings.TrimSpace(approvalStatus)),
		Title:        title,
		Message:      message,
		Actor:        actor,
	})
}

func approvalStatusLabel(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case strings.ToUpper(models.ApprovalRequestStatusApproved):
		return "approved"
	case strings.ToUpper(models.ApprovalRequestStatusRejected):
		return "rejected"
	case strings.ToUpper(models.ApprovalRequestStatusRevision):
		return "revision requested"
	case strings.ToUpper(models.ApprovalRequestStatusPending):
		return "moved to next approver"
	default:
		return strings.ToLower(strings.TrimSpace(status))
	}
}

func approvalEntityLabel(entityType string) string {
	switch strings.ToUpper(strings.TrimSpace(entityType)) {
	case models.ApprovalEntityBudgetControl:
		return "Budget Control"
	case models.ApprovalEntityBudgetAdjustment:
		return "Budget Adjustment"
	default:
		return strings.Title(strings.ToLower(strings.ReplaceAll(strings.TrimSpace(entityType), "_", " ")))
	}
}

func pointerString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
