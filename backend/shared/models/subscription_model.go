package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend/shared/mmodels"
	"github.com/jmoiron/sqlx"
)

const (
	SubscriptionStatusActive    = "ACTIVE"
	SubscriptionStatusExpired   = "EXPIRED"
	SubscriptionStatusSuspended = "SUSPENDED"

	SubscriptionAppTypePOSKasir      = "POS Kasir"
	SubscriptionAppTypeHumanResource = "Human Resources"
)

type Subscription struct {
	BaseModel

	ID                   int64                  `json:"id" db:"id"`
	AccountID            int64                  `json:"account_id" db:"account_id"`
	AccountName          string                 `json:"accountName,omitempty" db:"account_name"`
	CompanyName          string                 `json:"companyName,omitempty" db:"company_name"`
	AppType              string                 `json:"appType" db:"app_type"`
	PlanCode             string                 `json:"planCode" db:"plan_code"`
	PlanName             string                 `json:"planName" db:"plan_name"`
	PlanLimitsJSON       string                 `json:"-" db:"plan_limits"`
	PlanLimits           SubscriptionPlanLimits `json:"planLimits" db:"-"`
	NumberOfUsers        int                    `json:"numberOfUsers" db:"number_of_users"`
	PricePerUser         float64                `json:"pricePerUser" db:"price_per_user"`
	Amount               float64                `json:"amount" db:"amount"`
	CurrentStartDate     time.Time              `json:"currentStartDate" db:"current_start_date"`
	CurrentEndDate       time.Time              `json:"currentEndDate" db:"current_end_date"`
	PendingNumberOfUsers *int                   `json:"pendingNumberOfUsers,omitempty" db:"pending_number_of_users"`
	PendingEffectiveDate *time.Time             `json:"pendingEffectiveDate,omitempty" db:"pending_effective_date"`
	Status               string                 `json:"status" db:"status"`
}

type SubscriptionStatus struct {
	IsPaid          bool          `json:"isPaid"`
	HasSubscription bool          `json:"hasSubscription"`
	Message         string        `json:"message"`
	Subscription    *Subscription `json:"subscription,omitempty"`
}

func NewSubscription(db *sqlx.DB, auth Auth) Subscription {
	subscription := Subscription{
		BaseModel: BaseModel{DB: db, Auth: auth},
	}
	subscription.AccountID = auth.AccountID
	subscription.CreatedBy = auth.UserName
	subscription.Status = SubscriptionStatusActive
	subscription.AppType = SubscriptionAppTypePOSKasir
	plan, err := SubscriptionPlanByCodeFromDB(db, "")
	if err != nil {
		plan = SubscriptionPlanByCode("")
	}
	subscription.PlanCode = plan.Code
	subscription.PlanName = plan.Name
	subscription.PlanLimits = plan.Limits
	return subscription
}

func (m Subscription) Migrate() error {
	return subscriptionMModelFromModel(m).Migrate()
}

func (m Subscription) TableName() string {
	return mmodels.SubscriptionTable
}

func (m Subscription) ValidateCreate() error {
	if _, err := m.resolveAccountID(); err != nil {
		return err
	}
	m.AppType = NormalizeSubscriptionAppType(m.AppType)
	if !IsAllowedSubscriptionAppType(m.AppType) {
		return errors.New("app type subscription tidak valid")
	}
	if m.NumberOfUsers <= 0 {
		return errors.New("number of users subscription wajib lebih dari 0")
	}
	if m.PricePerUser < 0 {
		return errors.New("price per user subscription tidak valid")
	}
	if m.Amount < 0 {
		return errors.New("amount subscription tidak valid")
	}
	if m.CurrentStartDate.IsZero() {
		return errors.New("current start date subscription wajib diisi")
	}
	if m.CurrentEndDate.IsZero() {
		return errors.New("current end date subscription wajib diisi")
	}
	if m.CurrentEndDate.Before(m.CurrentStartDate) {
		return errors.New("current end date tidak boleh lebih kecil dari current start date")
	}
	return nil
}

func (m Subscription) ValidateID() error {
	if m.ID <= 0 {
		return errors.New("id subscription tidak valid")
	}
	return nil
}

func (m Subscription) Create() (int64, error) {
	return subscriptionMModelFromModel(m).Create()
}

func (m Subscription) Update() error {
	return subscriptionMModelFromModel(m).Update()
}

func (m Subscription) Delete() error {
	return subscriptionMModelFromModel(m).Delete()
}

func (m Subscription) DeleteByID(id int64) error {
	m.ID = id
	return m.Delete()
}

func (m Subscription) FindByID() (*Subscription, error) {
	found, err := subscriptionMModelFromModel(m).FindByID()
	if err != nil {
		return nil, err
	}
	item := subscriptionFromMModel(*found, m)
	return &item, nil
}

func (m Subscription) FindByAccount() (*Subscription, error) {
	found, err := subscriptionMModelFromModel(m).FindByAccount()
	if err != nil {
		return nil, err
	}
	item := subscriptionFromMModel(*found, m)
	return &item, nil
}

func subscriptionMModelFromModel(value Subscription) *mmodels.SubscriptionMModel {
	model := mmodels.NewSubscriptionMModel(value.Context, value.DB)
	model.ID = value.ID
	model.AccountID = value.AccountID
	model.AuthAccountID = value.Auth.AccountID
	model.AuthUserName = value.Auth.UserName
	model.AppType = value.AppType
	model.PlanCode = value.PlanCode
	model.PlanName = value.PlanName
	model.PlanLimitsJSON = value.PlanLimitsJSON
	model.PlanLimits = value.PlanLimits
	model.NumberOfUsers = value.NumberOfUsers
	model.PricePerUser = value.PricePerUser
	model.Amount = value.Amount
	model.CurrentStartDate = value.CurrentStartDate
	model.CurrentEndDate = value.CurrentEndDate
	model.PendingNumberOfUsers = value.PendingNumberOfUsers
	model.PendingEffectiveDate = value.PendingEffectiveDate
	model.Status = value.Status
	model.CreatedAt = value.CreatedAt
	model.CreatedBy = value.CreatedBy
	if value.UpdatedAt != nil {
		model.UpdatedAt = sql.NullTime{Time: *value.UpdatedAt, Valid: true}
	}
	if value.UpdatedBy != nil {
		model.UpdatedBy = sql.NullString{String: *value.UpdatedBy, Valid: true}
	}
	if strings.TrimSpace(model.CreatedBy) == "" {
		model.CreatedBy = strings.TrimSpace(value.Auth.UserName)
	}
	if strings.TrimSpace(model.CreatedBy) == "" {
		model.CreatedBy = "system"
	}
	return model
}

func subscriptionFromMModel(value mmodels.SubscriptionMModel, source Subscription) Subscription {
	item := Subscription{
		BaseModel: BaseModel{Context: source.Context, DB: source.DB, Auth: source.Auth},
		ID:        value.ID, AccountID: value.AccountID, AccountName: value.AccountName, CompanyName: value.CompanyName,
		AppType: value.AppType, PlanCode: value.PlanCode, PlanName: value.PlanName,
		PlanLimitsJSON: value.PlanLimitsJSON, PlanLimits: value.PlanLimits,
		NumberOfUsers: value.NumberOfUsers, PricePerUser: value.PricePerUser, Amount: value.Amount,
		CurrentStartDate: value.CurrentStartDate, CurrentEndDate: value.CurrentEndDate,
		PendingNumberOfUsers: value.PendingNumberOfUsers, PendingEffectiveDate: value.PendingEffectiveDate, Status: value.Status,
	}
	item.CreatedAt = value.CreatedAt
	item.CreatedBy = value.CreatedBy
	if value.UpdatedAt.Valid {
		updatedAt := value.UpdatedAt.Time
		item.UpdatedAt = &updatedAt
	}
	if value.UpdatedBy.Valid {
		updatedBy := value.UpdatedBy.String
		item.UpdatedBy = &updatedBy
	}
	return item
}

func (m Subscription) resolveAccountID() (int64, error) {
	if m.Auth.AccountID > 0 {
		return m.Auth.AccountID, nil
	}
	if m.AccountID > 0 {
		return m.AccountID, nil
	}
	return 0, errors.New("id akun wajib diisi")
}

func NormalizeSubscriptionAppType(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return SubscriptionAppTypePOSKasir
	}

	switch strings.ToLower(normalized) {
	case strings.ToLower(SubscriptionAppTypePOSKasir):
		return SubscriptionAppTypePOSKasir
	case strings.ToLower(SubscriptionAppTypeHumanResource):
		return SubscriptionAppTypeHumanResource
	default:
		return normalized
	}
}

func IsAllowedSubscriptionAppType(value string) bool {
	switch NormalizeSubscriptionAppType(value) {
	case SubscriptionAppTypePOSKasir, SubscriptionAppTypeHumanResource:
		return true
	default:
		return false
	}
}

func RoundMoney(value float64) float64 {
	return float64(int64(value*100+0.5)) / 100
}

func StartOfDay(value time.Time) time.Time {
	return time.Date(value.UTC().Year(), value.UTC().Month(), value.UTC().Day(), 0, 0, 0, 0, time.UTC)
}

func InclusiveDayDiff(start, end time.Time) int {
	start = StartOfDay(start)
	end = StartOfDay(end)
	if end.Before(start) {
		return 0
	}
	return int(end.Sub(start).Hours()/24) + 1
}

func NextDay(value time.Time) time.Time {
	value = StartOfDay(value)
	return value.AddDate(0, 0, 1)
}

func DefaultSubscriptionWindow() (time.Time, time.Time) {
	start := StartOfDay(time.Now().UTC())
	end := start.AddDate(0, 0, 30)
	return start, end
}

func DaysToDate(value time.Time) string {
	return StartOfDay(value).Format("2006-01-02")
}

func FmtMoney(value float64) string {
	return fmt.Sprintf("%.2f", RoundMoney(value))
}
