// Package app provides the main application entry point.
package app

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/httpapi"
	"github.com/sonbn-225/goen-api/internal/modules/account"
	authMod "github.com/sonbn-225/goen-api/internal/modules/auth"
	"github.com/sonbn-225/goen-api/internal/modules/budget"
	"github.com/sonbn-225/goen-api/internal/modules/category"
	"github.com/sonbn-225/goen-api/internal/modules/debt"
	"github.com/sonbn-225/goen-api/internal/modules/diagnostics"
	"github.com/sonbn-225/goen-api/internal/modules/investment"
	"github.com/sonbn-225/goen-api/internal/modules/marketdata"
	rotatingsavings "github.com/sonbn-225/goen-api/internal/modules/rotating_savings"
	"github.com/sonbn-225/goen-api/internal/modules/savings"
	"github.com/sonbn-225/goen-api/internal/modules/tag"
	"github.com/sonbn-225/goen-api/internal/modules/transaction"
	"github.com/sonbn-225/goen-api/internal/storage"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// App represents the application.
type App struct {
	Handler http.Handler
	cleanup func(context.Context)
}

// Close gracefully shuts down the application.
func (a *App) Close(ctx context.Context) {
	if a == nil || a.cleanup == nil {
		return
	}
	a.cleanup(ctx)
}

// New creates a new App using the modular architecture.
func New(cfg *config.Config) *App {
	db := storage.NewPostgres(cfg.DatabaseURL)
	redis := storage.NewRedis(cfg.RedisURL)

	// Repositories
	accountRepo := storage.NewAccountRepo(db)
	auditRepo := storage.NewAuditRepo(db)
	txRepo := storage.NewTransactionRepo(db)
	budgetRepo := storage.NewBudgetRepo(db)
	categoryRepo := storage.NewCategoryRepo(db)
	tagRepo := storage.NewTagRepo(db)
	userRepo := storage.NewUserRepo(db)
	savingsRepo := storage.NewSavingsRepo(db)
	rotatingSavingsRepo := storage.NewRotatingSavingsRepo(db)
	debtRepo := storage.NewDebtRepo(db)
	investmentRepo := storage.NewInvestmentRepo(db)

	// Independent modules (no cross-module dependencies)
	diagMod := diagnostics.NewModule(diagnostics.ModuleDeps{
		Cfg:   cfg,
		DB:    db,
		Redis: redis,
	})

	authModule := authMod.NewModule(authMod.ModuleDeps{
		UserRepo: userRepo,
		Config:   cfg,
	})

	categoryMod := category.NewModule(category.ModuleDeps{
		Repo: categoryRepo,
	})

	tagMod := tag.NewModule(tag.ModuleDeps{
		Repo: tagRepo,
	})

	// Transaction module
	txMod := transaction.NewModule(transaction.ModuleDeps{
		Repo: txRepo,
	})

	// Budget module (depends on category repo)
	budgetMod := budget.NewModule(budget.ModuleDeps{
		BudgetRepo:   budgetRepo,
		CategoryRepo: categoryRepo,
	})

	// Account module (depends on audit)
	auditSvc := &auditServiceAdapter{repo: auditRepo}
	accountMod := account.NewModule(account.ModuleDeps{
		AccountRepo:  accountRepo,
		UserRepo:     userRepo,
		AuditService: auditSvc,
	})

	// Savings module (depends on account and transaction services)
	savingsMod := savings.NewModule(savings.ModuleDeps{
		Repo:       savingsRepo,
		AccountSvc: accountMod.Service,
		TxSvc:      txMod.Service,
	})

	// Debt module (depends on transaction service)
	debtMod := debt.NewModule(debt.ModuleDeps{
		Repo:  debtRepo,
		TxSvc: txMod.Service,
	})

	// Rotating Savings module (depends on account repo and transaction service)
	rotTxAdapter := &rotTxServiceAdapter{svc: txMod.Service}
	rotMod := rotatingsavings.NewModule(rotatingsavings.ModuleDeps{
		Repo:        rotatingSavingsRepo,
		AccountRepo: accountRepo,
		TxSvc:       rotTxAdapter,
	})

	// Investment module (depends on account and transaction services)
	investAccountAdapter := &investmentAccountServiceAdapter{svc: accountMod.Service}
	investMod := investment.NewModule(investment.ModuleDeps{
		Repo:               investmentRepo,
		Redis:              redis,
		Config:             cfg,
		AccountService:     investAccountAdapter,
		TransactionService: txMod.Service,
	})

	// Market Data module (depends on investment service)
	marketDataMod := marketdata.NewModule(marketdata.ModuleDeps{
		Cfg:       cfg,
		Redis:     redis,
		Repo:      marketdata.NewPostgresRepo(db),
		InvestSvc: investMod.Service,
	})

	// Create router
	h := newModularRouter(cfg, &modules{
		diagnostics:     diagMod,
		auth:            authModule,
		account:         accountMod,
		transaction:     txMod,
		category:        categoryMod,
		tag:             tagMod,
		budget:          budgetMod,
		savings:         savingsMod,
		rotatingSavings: rotMod,
		debt:            debtMod,
		investment:      investMod,
		marketData:      marketDataMod,
	})

	return &App{
		Handler: h,
		cleanup: func(_ context.Context) {
			if db != nil {
				db.Close()
			}
			if redis != nil {
				redis.Close()
			}
		},
	}
}

// modules holds all application modules internally.
type modules struct {
	diagnostics     *diagnostics.Module
	auth            *authMod.Module
	account         *account.Module
	transaction     *transaction.Module
	category        *category.Module
	tag             *tag.Module
	budget          *budget.Module
	savings         *savings.Module
	rotatingSavings *rotatingsavings.Module
	debt            *debt.Module
	investment      *investment.Module
	marketData      *marketdata.Module
}

func newModularRouter(cfg *config.Config, mods *modules) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(httpapi.OptionalAuthMiddleware(cfg))
	r.Use(httpapi.RequestLogger())
	r.Use(httpapi.CORSMiddleware(cfg))
	r.Use(middleware.Heartbeat("/healthz"))

	r.Get("/swagger", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/swagger/", http.StatusMovedPermanently)
	})
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Route("/api/v1", func(r chi.Router) {
		// Diagnostics (no auth required)
		mods.diagnostics.Handler.RegisterRoutes(r)

		// Auth - uses its own cfg-based middleware
		mods.auth.Handler.RegisterRoutes(r, cfg)

		// Others use auth middleware
		authMiddleware := httpapi.AuthMiddleware(cfg)

		// Account
		mods.account.Handler.RegisterRoutes(r, authMiddleware)

		// Transaction
		mods.transaction.Handler.RegisterRoutes(r, authMiddleware)

		// Category
		mods.category.Handler.RegisterRoutes(r, authMiddleware)

		// Tag
		mods.tag.Handler.RegisterRoutes(r, authMiddleware)

		// Budget
		mods.budget.Handler.RegisterRoutes(r, authMiddleware)

		// Savings
		mods.savings.Handler.RegisterRoutes(r, authMiddleware)

		// Rotating Savings
		mods.rotatingSavings.Handler.RegisterRoutes(r, authMiddleware)

		// Debt
		mods.debt.Handler.RegisterRoutes(r, authMiddleware)

		// Investment
		mods.investment.Handler.RegisterRoutes(r, authMiddleware)

		// Market Data
		mods.marketData.Handler.RegisterRoutes(r, authMiddleware)
	})

	return r
}
