package models

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"

	"backend/shared/mmodels"
	"github.com/jmoiron/sqlx"
)

type SubscriptionPlanLimits = mmodels.SubscriptionPlanLimits

type SubscriptionPlan struct {
	ID          int64                  `json:"id,omitempty"`
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	Price       float64                `json:"price"`
	Currency    string                 `json:"currency"`
	Description string                 `json:"description"`
	Limits      SubscriptionPlanLimits `json:"limits"`
	Active      bool                   `json:"active"`
	SortOrder   int                    `json:"sortOrder"`
}

func SubscriptionPlansFromEnv() []SubscriptionPlan {
	plans := []SubscriptionPlan{
		defaultSubscriptionPlan("STARTER", "Starter", 399000, SubscriptionPlanLimits{Users: 1, Locations: 1, Products: 1000, Categories: 100, Brands: 100, Customers: 2000, Suppliers: 200, PaymentMethods: 10, PriceLists: 5, Promotions: 20, StorageGB: 2}),
		defaultSubscriptionPlan("BUSINESS", "Business", 699000, SubscriptionPlanLimits{Users: 3, Locations: 2, Products: 5000, Categories: 500, Brands: 500, Customers: 10000, Suppliers: 1000, PaymentMethods: 30, PriceLists: 20, Promotions: 100, StorageGB: 10}),
		defaultSubscriptionPlan("PRO", "Pro", 1199000, SubscriptionPlanLimits{Users: 10, Locations: 5, Products: 20000, Categories: 2000, Brands: 2000, Customers: 100000, Suppliers: 100000, PaymentMethods: 100, PriceLists: 100000, Promotions: 100000, StorageGB: 50}),
	}
	for index := range plans {
		plans[index].Active = true
		plans[index].SortOrder = (index + 1) * 10
		applySubscriptionPlanEnv(&plans[index])
	}
	return plans
}

func SubscriptionPlansFromDB(db *sqlx.DB, activeOnly bool) ([]SubscriptionPlan, error) {
	if db == nil {
		return nil, errors.New("subscription plan database is nil")
	}
	model := mmodels.NewSubscriptionPlanMModel(context.Background(), db)
	rows, err := model.FindList(activeOnly)
	if err != nil {
		return nil, err
	}
	plans := make([]SubscriptionPlan, 0, len(rows))
	for _, row := range rows {
		plans = append(plans, subscriptionPlanFromMModel(row))
	}
	return plans, nil
}

func SubscriptionPlanByCodeFromDB(db *sqlx.DB, code string) (SubscriptionPlan, error) {
	if db == nil {
		return SubscriptionPlan{}, errors.New("subscription plan database is nil")
	}
	model := mmodels.NewSubscriptionPlanMModel(context.Background(), db)
	row, err := model.FindByCode(code)
	if err != nil {
		return SubscriptionPlan{}, err
	}
	return subscriptionPlanFromMModel(*row), nil
}

func subscriptionPlanFromMModel(value mmodels.SubscriptionPlanMModel) SubscriptionPlan {
	return SubscriptionPlan{
		ID: value.ID, Code: value.Code, Name: value.Name, Price: value.Price,
		Currency: value.Currency, Description: value.Description, Limits: value.Limits,
		Active: value.Active, SortOrder: value.SortOrder,
	}
}

func SubscriptionPlanByCode(code string) SubscriptionPlan {
	normalized := NormalizeSubscriptionPlanCode(code)
	for _, plan := range SubscriptionPlansFromEnv() {
		if plan.Code == normalized {
			return plan
		}
	}
	return SubscriptionPlansFromEnv()[0]
}

func NormalizeSubscriptionPlanCode(code string) string {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	if normalized == "" {
		return "STARTER"
	}
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")
	return normalized
}

func defaultSubscriptionPlan(code, name string, price float64, limits SubscriptionPlanLimits) SubscriptionPlan {
	return SubscriptionPlan{Code: code, Name: name, Price: price, Currency: getenvSubscriptionPlan("SUBSCRIPTION_CURRENCY", "IDR"), Limits: limits}
}

func applySubscriptionPlanEnv(plan *SubscriptionPlan) {
	prefix := "SUBSCRIPTION_PLAN_" + plan.Code + "_"
	plan.Name = getenvSubscriptionPlan(prefix+"NAME", plan.Name)
	plan.Description = getenvSubscriptionPlan(prefix+"DESCRIPTION", plan.Description)
	plan.Currency = getenvSubscriptionPlan(prefix+"CURRENCY", getenvSubscriptionPlan("SUBSCRIPTION_CURRENCY", plan.Currency))
	plan.Price = getenvSubscriptionPlanFloat(prefix+"PRICE", plan.Price)
	plan.Limits.Users = getenvSubscriptionPlanInt(prefix+"MAX_USERS", plan.Limits.Users)
	plan.Limits.Locations = getenvSubscriptionPlanInt(prefix+"MAX_LOCATIONS", plan.Limits.Locations)
	plan.Limits.Products = getenvSubscriptionPlanInt(prefix+"MAX_PRODUCTS", plan.Limits.Products)
	plan.Limits.Categories = getenvSubscriptionPlanInt(prefix+"MAX_CATEGORIES", plan.Limits.Categories)
	plan.Limits.Brands = getenvSubscriptionPlanInt(prefix+"MAX_BRANDS", plan.Limits.Brands)
	plan.Limits.Customers = getenvSubscriptionPlanInt(prefix+"MAX_CUSTOMERS", plan.Limits.Customers)
	plan.Limits.Suppliers = getenvSubscriptionPlanInt(prefix+"MAX_SUPPLIERS", plan.Limits.Suppliers)
	plan.Limits.PaymentMethods = getenvSubscriptionPlanInt(prefix+"MAX_PAYMENT_METHODS", plan.Limits.PaymentMethods)
	plan.Limits.PriceLists = getenvSubscriptionPlanInt(prefix+"MAX_PRICE_LISTS", plan.Limits.PriceLists)
	plan.Limits.Promotions = getenvSubscriptionPlanInt(prefix+"MAX_PROMOTIONS", plan.Limits.Promotions)
	plan.Limits.StorageGB = getenvSubscriptionPlanInt(prefix+"MAX_STORAGE_GB", plan.Limits.StorageGB)
}

func getenvSubscriptionPlan(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getenvSubscriptionPlanInt(key string, fallback int) int {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed >= 0 {
			return parsed
		}
	}
	return fallback
}

func getenvSubscriptionPlanFloat(key string, fallback float64) float64 {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil && parsed >= 0 {
			return RoundMoney(parsed)
		}
	}
	return fallback
}
