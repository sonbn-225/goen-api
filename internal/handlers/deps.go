package handlers

import (
	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/services"
)

// Deps is the dependency bundle injected into HTTP handlers.
// Keep this as interfaces to support testing and decouple handlers from concrete implementations.
// (Many services are already interfaces; DB/Redis are optional infra dependencies.)
type Deps struct {
	Cfg *config.Config

	DiagnosticsService services.DiagnosticsService
	MarketDataService  services.MarketDataService

	AuthService            services.AuthService
	AccountService         services.AccountService
	AuditService           services.AuditService
	TransactionService     services.TransactionService
	CategoryService        services.CategoryService
	TagService             services.TagService
	BudgetService          services.BudgetService
	SavingsService         services.SavingsService
	RotatingSavingsService services.RotatingSavingsService
	DebtService            services.DebtService
	InvestmentService      services.InvestmentService
}
