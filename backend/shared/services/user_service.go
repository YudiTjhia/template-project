package services

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"backend/shared/models"

	"github.com/jmoiron/sqlx"
)

type UserService struct {
	DB            *sqlx.DB
	TransactionDB *sqlx.DB
	Auth          models.Auth
}

func NewUserService(db *sqlx.DB, auth models.Auth) *UserService {
	return NewUserServiceWithDatabases(db, db, auth)
}

func NewUserServiceWithDatabases(db, transactionDB *sqlx.DB, auth models.Auth) *UserService {
	if transactionDB == nil {
		transactionDB = db
	}
	return &UserService{DB: db, TransactionDB: transactionDB, Auth: auth}
}

func (s *UserService) NewUser() models.User {
	return models.NewUser(s.DB, s.Auth)
}

func (s *UserService) Migrate() {
	s.NewUser().Migrate()
}

func (s *UserService) CreateUser(user models.User) (int64, error) {
	accountID := user.AccountID
	if accountID <= 0 {
		accountID = s.Auth.AccountID
	}
	if accountID > 0 {
		plan, err := s.currentSubscriptionPlan(accountID)
		if err != nil {
			return 0, err
		}
		if plan.Limits.Users > 0 {
			count, err := s.CountActiveUsers(accountID)
			if err != nil {
				return 0, err
			}
			if count >= plan.Limits.Users {
				return 0, fmt.Errorf("limit paket %s untuk users sudah tercapai (%d/%d)", plan.Name, count, plan.Limits.Users)
			}
		}
	}
	u := s.NewUser()
	u.AccountID = user.AccountID
	u.UserEmail = user.UserEmail
	u.MobilePhone = user.MobilePhone
	u.UserPassword = user.UserPassword
	u.PIN = user.PIN
	u.ConfirmPIN = user.ConfirmPIN
	u.UserType = user.UserType
	u.LoginType = user.LoginType
	u.UserRoles = user.UserRoles
	u.IsActive = user.IsActive
	u.Username = user.Username
	u.CreatedBy = user.CreatedBy
	id, err := u.Create()
	if err != nil {
		return 0, err
	}

	accountID = u.AccountID
	if accountID <= 0 {
		accountID = s.Auth.AccountID
	}
	if err := s.syncSubscriptionForAccount(accountID, "sync subscription after user create"); err != nil {
		return 0, err
	}

	return id, nil
}

func (s *UserService) UpdateUser(user models.User) error {
	u := s.NewUser()
	u.ID = user.ID
	u.AccountID = user.AccountID
	u.UserEmail = user.UserEmail
	u.MobilePhone = user.MobilePhone
	u.UserPassword = user.UserPassword
	u.PIN = user.PIN
	u.ConfirmPIN = user.ConfirmPIN
	u.UserType = user.UserType
	u.LoginType = user.LoginType
	u.UserRoles = user.UserRoles
	u.IsActive = user.IsActive
	u.Username = user.Username
	u.UpdatedBy = user.UpdatedBy
	if err := u.Update(); err != nil {
		return err
	}

	accountID := u.AccountID
	if accountID <= 0 {
		accountID = s.Auth.AccountID
	}
	return s.syncSubscriptionForAccount(accountID, "sync subscription after user update")
}

func (s *UserService) UpdateCurrentProfile(id int64, username, email string) error {
	username = strings.TrimSpace(username)
	email = strings.ToLower(strings.TrimSpace(email))
	currentEmail := strings.ToLower(strings.TrimSpace(s.Auth.Email))
	if id <= 0 || currentEmail == "" {
		return errors.New("sesi user tidak valid")
	}
	if username == "" {
		return errors.New("username wajib diisi")
	}
	if email == "" || !strings.Contains(email, "@") {
		return errors.New("email tidak valid")
	}

	tx, err := s.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.Exec(`
		UPDATE users
		SET username = $1, user_email = $2, updated_at = NOW(), updated_by = $3
		WHERE id = $4 AND account_id = $5 AND LOWER(user_email) = LOWER($6)`,
		username, email, currentEmail, id, s.Auth.AccountID, currentEmail,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("profil user tidak ditemukan atau bukan milik sesi ini")
	}
	if _, err := tx.Exec(`
		UPDATE auths SET email = $1, updated_at = NOW()
		WHERE account_id = $2 AND LOWER(email) = LOWER($3)`,
		email, s.Auth.AccountID, currentEmail,
	); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *UserService) ActivateUser(id int64) error {
	if id <= 0 {
		return errors.New("id user tidak valid")
	}

	updatedBy := s.Auth.UserName
	if updatedBy == "" {
		updatedBy = "system"
	}

	sql := `UPDATE users
	SET is_active = true,
		updated_at = NOW(),
		updated_by = $1
	WHERE id = $2 AND account_id = $3`

	res, err := s.DB.Exec(sql, updatedBy, id, s.Auth.AccountID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("user tidak ditemukan")
	}

	return s.syncSubscriptionForAccount(s.Auth.AccountID, "sync subscription after user activate")
}

func (s *UserService) Create(user models.User) (int64, error) {
	return s.CreateUser(user)
}

func (s *UserService) Update(user models.User) error {
	return s.UpdateUser(user)
}

func (s *UserService) Delete(id int64) error {
	u := s.NewUser()
	u.ID = id
	if err := u.Delete(); err != nil {
		return err
	}

	accountID := u.AccountID
	if accountID <= 0 {
		accountID = s.Auth.AccountID
	}
	return s.syncSubscriptionForAccount(accountID, "sync subscription after user delete")
}

func (s *UserService) FindByID(id int64) (*models.User, error) {
	u := s.NewUser()
	u.ID = id
	return u.FindByID()
}

func (s *UserService) FindByEmail(email string) (*models.User, error) {
	u := s.NewUser()
	u.UserEmail = email
	return u.FindByEmail()
}

func (s *UserService) FindByUsername(username string) (*models.User, error) {
	u := s.NewUser()
	u.Username = username
	return u.FindByUsername()
}

func (s *UserService) FindList(search string) ([]models.User, error) {
	return s.NewUser().FindList(search)
}

func (s *UserService) FindAll() ([]models.User, error) {
	return s.NewUser().FindAll()
}

func (s *UserService) FindRecentList(limit int, search string) ([]models.User, error) {
	return s.NewUser().FindRecentList(limit, search)
}

func (s *UserService) CountActiveUsers(accountID int64) (int, error) {
	if accountID <= 0 {
		accountID = s.Auth.AccountID
	}
	var count int
	err := s.DB.Get(&count, `SELECT COUNT(*) FROM users WHERE account_id = $1 AND is_active = true`, accountID)
	return count, err
}

func (s *UserService) currentSubscriptionPlan(accountID int64) (models.SubscriptionPlan, error) {
	var code string
	err := s.DB.Get(&code, `SELECT plan_code FROM subscriptions WHERE account_id = $1`, accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || isMissingUserSubscriptionPlanColumn(err) {
			return models.SubscriptionPlanByCode("STARTER"), nil
		}
		return models.SubscriptionPlan{}, err
	}
	return models.SubscriptionPlanByCode(code), nil
}

func isMissingUserSubscriptionPlanColumn(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "plan_code") || strings.Contains(message, "subscriptions") && strings.Contains(message, "does not exist") || strings.Contains(message, "no such table")
}

func (s *UserService) syncSubscriptionForAccount(accountID int64, notes string) error {
	if accountID <= 0 {
		accountID = s.Auth.AccountID
	}
	if accountID <= 0 {
		return nil
	}

	service := NewSubscriptionServiceWithDatabases(s.DB, s.TransactionDB, models.Auth{
		AccountID: accountID,
		UserName:  s.Auth.UserName,
		Email:     s.Auth.Email,
	})
	if err := service.SyncActiveUsersForAccount(accountID, notes); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	return nil
}
