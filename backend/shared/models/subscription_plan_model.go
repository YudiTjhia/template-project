package models

import (
	"os"
	"strconv"
	"strings"
)

type SubscriptionPlanLimits struct {
	Users                     int `json:"users"`
	Locations                 int `json:"locations"`
	Products                  int `json:"products"`
	Categories                int `json:"categories"`
	Brands                    int `json:"brands"`
	Customers                 int `json:"customers"`
	Suppliers                 int `json:"suppliers"`
	SalesTransactionsPerMonth int `json:"salesTransactionsPerMonth"`
	SalesReturnsPerMonth      int `json:"salesReturnsPerMonth"`
	PurchaseStockInPerMonth   int `json:"purchaseStockInPerMonth"`
	StockTransfersPerMonth    int `json:"stockTransfersPerMonth"`
	StockOpnamesPerMonth      int `json:"stockOpnamesPerMonth"`
	PaymentMethods            int `json:"paymentMethods"`
	PriceLists                int `json:"priceLists"`
	Promotions                int `json:"promotions"`
	StorageGB                 int `json:"storageGb"`
}

type SubscriptionPlan struct {
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	Price       float64                `json:"price"`
	Currency    string                 `json:"currency"`
	Description string                 `json:"description"`
	Limits      SubscriptionPlanLimits `json:"limits"`
}

func SubscriptionPlansFromEnv() []SubscriptionPlan {
	plans := []SubscriptionPlan{
		defaultSubscriptionPlan("STARTER", "Starter", 399000, SubscriptionPlanLimits{Users: 1, Locations: 1, Products: 1000, Categories: 100, Brands: 100, Customers: 2000, Suppliers: 200, SalesTransactionsPerMonth: 5000, SalesReturnsPerMonth: 500, PurchaseStockInPerMonth: 1000, StockTransfersPerMonth: 500, StockOpnamesPerMonth: 20, PaymentMethods: 10, PriceLists: 5, Promotions: 20, StorageGB: 2}),
		defaultSubscriptionPlan("BUSINESS", "Business", 699000, SubscriptionPlanLimits{Users: 3, Locations: 2, Products: 5000, Categories: 500, Brands: 500, Customers: 10000, Suppliers: 1000, SalesTransactionsPerMonth: 15000, SalesReturnsPerMonth: 1500, PurchaseStockInPerMonth: 3000, StockTransfersPerMonth: 1500, StockOpnamesPerMonth: 60, PaymentMethods: 30, PriceLists: 20, Promotions: 100, StorageGB: 10}),
		defaultSubscriptionPlan("PRO", "Pro", 1199000, SubscriptionPlanLimits{Users: 10, Locations: 5, Products: 20000, Categories: 2000, Brands: 2000, Customers: 100000, Suppliers: 100000, SalesTransactionsPerMonth: 50000, SalesReturnsPerMonth: 5000, PurchaseStockInPerMonth: 10000, StockTransfersPerMonth: 5000, StockOpnamesPerMonth: 200, PaymentMethods: 100, PriceLists: 100000, Promotions: 100000, StorageGB: 50}),
	}
	for index := range plans {
		applySubscriptionPlanEnv(&plans[index])
	}
	return plans
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
	plan.Limits.SalesTransactionsPerMonth = getenvSubscriptionPlanInt(prefix+"MAX_SALES_TRANSACTIONS_PER_MONTH", plan.Limits.SalesTransactionsPerMonth)
	plan.Limits.SalesReturnsPerMonth = getenvSubscriptionPlanInt(prefix+"MAX_SALES_RETURNS_PER_MONTH", plan.Limits.SalesReturnsPerMonth)
	plan.Limits.PurchaseStockInPerMonth = getenvSubscriptionPlanInt(prefix+"MAX_PURCHASE_STOCK_IN_PER_MONTH", plan.Limits.PurchaseStockInPerMonth)
	plan.Limits.StockTransfersPerMonth = getenvSubscriptionPlanInt(prefix+"MAX_STOCK_TRANSFERS_PER_MONTH", plan.Limits.StockTransfersPerMonth)
	plan.Limits.StockOpnamesPerMonth = getenvSubscriptionPlanInt(prefix+"MAX_STOCK_OPNAMES_PER_MONTH", plan.Limits.StockOpnamesPerMonth)
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
