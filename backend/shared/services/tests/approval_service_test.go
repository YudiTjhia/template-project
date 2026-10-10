package services_test

import (
	"fmt"
	"testing"
	"time"

	"backend/shared/models"
	"backend/shared/services"
)

func newRealApprovalService(t *testing.T, auth models.Auth) (*services.ApprovalService, func()) {
	t.Helper()

	if authServiceDB == nil {
		t.Skip("approval service db is not initialized")
	}

	models.NewUser(authServiceDB, auth).Migrate()
	models.NewRole(authServiceDB, auth).Migrate()
	models.NewApprovalRule(authServiceDB, auth).Migrate()
	models.NewApprovalRuleStep(authServiceDB, auth).Migrate()
	models.NewApprovalRequest(authServiceDB, auth).Migrate()
	models.NewApprovalRequestStep(authServiceDB, auth).Migrate()

	service := services.NewApprovalService(authServiceDB, auth)
	_, _ = authServiceDB.Exec(`TRUNCATE TABLE approval_request_steps RESTART IDENTITY CASCADE`)
	_, _ = authServiceDB.Exec(`TRUNCATE TABLE approval_requests RESTART IDENTITY CASCADE`)
	_, _ = authServiceDB.Exec(`TRUNCATE TABLE approval_rule_steps RESTART IDENTITY CASCADE`)
	_, _ = authServiceDB.Exec(`TRUNCATE TABLE approval_rules RESTART IDENTITY CASCADE`)
	_, _ = authServiceDB.Exec(`TRUNCATE TABLE users RESTART IDENTITY CASCADE`)
	_, _ = authServiceDB.Exec(`TRUNCATE TABLE roles RESTART IDENTITY CASCADE`)

	cleanup := func() {
		_, _ = authServiceDB.Exec(`TRUNCATE TABLE approval_request_steps RESTART IDENTITY CASCADE`)
		_, _ = authServiceDB.Exec(`TRUNCATE TABLE approval_requests RESTART IDENTITY CASCADE`)
		_, _ = authServiceDB.Exec(`TRUNCATE TABLE approval_rule_steps RESTART IDENTITY CASCADE`)
		_, _ = authServiceDB.Exec(`TRUNCATE TABLE approval_rules RESTART IDENTITY CASCADE`)
		_, _ = authServiceDB.Exec(`TRUNCATE TABLE users RESTART IDENTITY CASCADE`)
		_, _ = authServiceDB.Exec(`TRUNCATE TABLE roles RESTART IDENTITY CASCADE`)
	}

	return service, cleanup
}

func seedApprovalStepRequest(t *testing.T, accountID int64, entityNumber string, approverType string, approverUserID *int64, approverRoleID *int64, approverRole string) int64 {
	t.Helper()

	now := time.Now().UTC()
	var requestID int64
	if err := authServiceDB.QueryRow(
		`INSERT INTO approval_requests
		(account_id, entity_type, entity_id, entity_number, amount, status, current_level, submitted_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`,
		accountID,
		"TEST_DOCUMENT",
		time.Now().UnixNano(),
		entityNumber,
		1000,
		models.ApprovalRequestStatusPending,
		1,
		now,
		"seed",
	).Scan(&requestID); err != nil {
		t.Fatalf("insert approval request: %v", err)
	}

	if _, err := authServiceDB.Exec(
		`INSERT INTO approval_request_steps
		(account_id, approval_request_id, level, approver_type, approver_user_id, approver_role_id, approver_role, required_action, status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		accountID,
		requestID,
		1,
		approverType,
		approverUserID,
		approverRoleID,
		approverRole,
		models.ApprovalRequiredActionAny,
		models.ApprovalStepStatusPending,
		"seed",
	); err != nil {
		t.Fatalf("insert approval request step: %v", err)
	}

	return requestID
}

func TestApprovalServiceFindRequestsFiltersByCurrentApprover(t *testing.T) {
	auth := models.Auth{AccountID: 7, Email: "bob@example.com", UserName: "bob", Name: "Bob"}
	service, cleanup := newRealApprovalService(t, auth)
	defer cleanup()

	role := models.NewRole(authServiceDB, auth)
	role.RoleName = "Approver"
	role.Actions = models.StringArray{"approve"}
	roleID, err := role.Create()
	if err != nil {
		t.Fatalf("create role: %v", err)
	}

	user := models.NewUser(authServiceDB, auth)
	user.UserEmail = auth.Email
	user.MobilePhone = "0811111111"
	user.UserPassword = "secret"
	user.UserType = models.UserTypeUser
	user.UserRoles = models.StringArray{"Approver"}
	user.IsActive = true
	user.Username = auth.UserName
	if _, err := user.Create(); err != nil {
		t.Fatalf("create user: %v", err)
	}

	otherRole := models.NewRole(authServiceDB, auth)
	otherRole.RoleName = "Requester"
	otherRole.Actions = models.StringArray{"read"}
	otherRoleID, err := otherRole.Create()
	if err != nil {
		t.Fatalf("create other role: %v", err)
	}

	seedApprovalStepRequest(t, auth.AccountID, "REQ-APPROVER", models.ApprovalApproverTypeRole, nil, &roleID, "Approver")
	seedApprovalStepRequest(t, auth.AccountID, "REQ-OTHER", models.ApprovalApproverTypeRole, nil, &otherRoleID, "Requester")

	items, err := service.FindRequests("")
	if err != nil {
		t.Fatalf("find requests: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 request, got %d: %+v", len(items), items)
	}
	if items[0].EntityNumber != "REQ-APPROVER" {
		t.Fatalf("unexpected request returned: %+v", items[0])
	}
}

func TestApprovalServiceSaveRuleSyncsPendingRequestStepsToLatestApprover(t *testing.T) {
	auth := models.Auth{AccountID: 7, Email: "admin@example.com", UserName: "admin", Name: "Admin"}
	service, cleanup := newRealApprovalService(t, auth)
	defer cleanup()

	oldUser := models.NewUser(authServiceDB, auth)
	oldUser.UserEmail = "old-approver@example.com"
	oldUser.MobilePhone = "0811111111"
	oldUser.UserPassword = "secret"
	oldUser.UserType = models.UserTypeUser
	oldUser.UserRoles = models.StringArray{"Approver"}
	oldUser.IsActive = true
	oldUser.Username = "old-approver"
	oldUserID, err := oldUser.Create()
	if err != nil {
		t.Fatalf("create old approver: %v", err)
	}

	newUser := models.NewUser(authServiceDB, auth)
	newUser.UserEmail = "new-approver@example.com"
	newUser.MobilePhone = "0822222222"
	newUser.UserPassword = "secret"
	newUser.UserType = models.UserTypeUser
	newUser.UserRoles = models.StringArray{"Approver"}
	newUser.IsActive = true
	newUser.Username = "new-approver"
	newUserID, err := newUser.Create()
	if err != nil {
		t.Fatalf("create new approver: %v", err)
	}

	requestID := seedApprovalStepRequest(t, auth.AccountID, "REQ-SYNC", models.ApprovalApproverTypeUser, &oldUserID, nil, "")
	approverName := "new-approver"
	_, err = service.SaveRule(models.ApprovalRule{
		EntityType: "TEST_DOCUMENT",
		RuleName:   "Test document approval",
		Priority:   1,
		Active:     true,
		Steps: []models.ApprovalRuleStep{
			{
				Level:          1,
				ApproverType:   models.ApprovalApproverTypeUser,
				ApproverUserID: &newUserID,
				ApproverName:   &approverName,
				RequiredAction: models.ApprovalRequiredActionAny,
				Active:         true,
			},
		},
	})
	if err != nil {
		t.Fatalf("save rule: %v", err)
	}

	var syncedStep models.ApprovalRequestStep
	if err := authServiceDB.Get(&syncedStep, `SELECT id, account_id, approval_request_id, level, approver_type, approver_user_id, approver_name, approver_role_id, approver_role, required_action, status, action_by, action_at, notes, created_at, created_by, updated_at, updated_by
		FROM approval_request_steps
		WHERE account_id = $1 AND approval_request_id = $2`, auth.AccountID, requestID); err != nil {
		t.Fatalf("find synced request step: %v", err)
	}
	if syncedStep.ApproverUserID == nil || *syncedStep.ApproverUserID != newUserID {
		t.Fatalf("expected request step approver user %d, got %+v", newUserID, syncedStep.ApproverUserID)
	}
	if syncedStep.ApproverName == nil || *syncedStep.ApproverName != approverName {
		t.Fatalf("expected request step approver name %q, got %+v", approverName, syncedStep.ApproverName)
	}

	newApproverService := services.NewApprovalService(authServiceDB, models.Auth{
		AccountID: auth.AccountID,
		Email:     newUser.UserEmail,
		UserName:  newUser.Username,
		Name:      "New Approver",
	})
	items, err := newApproverService.FindRequests("")
	if err != nil {
		t.Fatalf("find requests as new approver: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 synced request for new approver, got %d: %+v", len(items), items)
	}
	if items[0].EntityNumber != "REQ-SYNC" {
		t.Fatalf("unexpected synced request returned: %+v", items[0])
	}
}

func TestApprovalServiceRejectsActionForNonApprover(t *testing.T) {
	approverAuth := models.Auth{AccountID: 7, Email: "bob@example.com", UserName: "bob", Name: "Bob"}
	approverService, cleanup := newRealApprovalService(t, approverAuth)
	defer cleanup()

	role := models.NewRole(authServiceDB, approverAuth)
	role.RoleName = "Approver"
	role.Actions = models.StringArray{"approve"}
	roleID, err := role.Create()
	if err != nil {
		t.Fatalf("create role: %v", err)
	}

	approverUser := models.NewUser(authServiceDB, approverAuth)
	approverUser.UserEmail = approverAuth.Email
	approverUser.MobilePhone = "0811111111"
	approverUser.UserPassword = "secret"
	approverUser.UserType = models.UserTypeUser
	approverUser.UserRoles = models.StringArray{"Approver"}
	approverUser.IsActive = true
	approverUser.Username = approverAuth.UserName
	if _, err := approverUser.Create(); err != nil {
		t.Fatalf("create approver user: %v", err)
	}

	requestID := seedApprovalStepRequest(t, approverAuth.AccountID, "REQ-ACTION", models.ApprovalApproverTypeRole, nil, &roleID, "Approver")

	otherAuth := models.Auth{AccountID: 7, Email: "charlie@example.com", UserName: "charlie", Name: "Charlie"}
	otherUser := models.NewUser(authServiceDB, otherAuth)
	otherUser.UserEmail = otherAuth.Email
	otherUser.MobilePhone = "0822222222"
	otherUser.UserPassword = "secret"
	otherUser.UserType = models.UserTypeUser
	otherUser.UserRoles = models.StringArray{"Requester"}
	otherUser.IsActive = true
	otherUser.Username = otherAuth.UserName
	if _, err := otherUser.Create(); err != nil {
		t.Fatalf("create non approver user: %v", err)
	}

	otherRole := models.NewRole(authServiceDB, approverAuth)
	otherRole.RoleName = "Requester"
	otherRole.Actions = models.StringArray{"read"}
	if _, err := otherRole.Create(); err != nil {
		t.Fatalf("create requester role: %v", err)
	}

	nonApproverService := services.NewApprovalService(authServiceDB, otherAuth)
	err = nonApproverService.ApproveRequest(services.ApprovalActionInput{
		RequestID: requestID,
		Notes:     fmt.Sprintf("attempted by %s", otherAuth.UserName),
	})
	if err == nil {
		t.Fatalf("expected action to be rejected for non approver")
	}
	if err.Error() != "request ini bukan untuk user yang sedang login" {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = approverService
}
