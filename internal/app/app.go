package app

import (
	"context"
	"net/http"

	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/handlers"
	"github.com/sonbn-225/goen-api/internal/httpapi"
	"github.com/sonbn-225/goen-api/internal/services"
	"github.com/sonbn-225/goen-api/internal/storage"
)

type App struct {
	Handler http.Handler
	cleanup func(context.Context)
}

func New(cfg *config.Config) *App {
	db := storage.NewPostgres(cfg.DatabaseURL)
	redis := storage.NewRedis(cfg.RedisURL)

	userRepo := storage.NewUserRepo(db)
	accountRepo := storage.NewAccountRepo(db)
	auditRepo := storage.NewAuditRepo(db)
	txRepo := storage.NewTransactionRepo(db)
	categoryRepo := storage.NewCategoryRepo(db)
	tagRepo := storage.NewTagRepo(db)
	budgetRepo := storage.NewBudgetRepo(db)
	savingsRepo := storage.NewSavingsRepo(db)
	rotatingSavingsRepo := storage.NewRotatingSavingsRepo(db)
	debtRepo := storage.NewDebtRepo(db)
	investmentRepo := storage.NewInvestmentRepo(db)

	authService := services.NewAuthService(userRepo, cfg)
	accountService := services.NewAccountService(accountRepo, userRepo)
	auditService := services.NewAuditService(auditRepo)
	transactionService := services.NewTransactionService(txRepo)
	categoryService := services.NewCategoryService(categoryRepo)
	tagService := services.NewTagService(tagRepo)
	budgetService := services.NewBudgetService(budgetRepo, categoryRepo)
	savingsService := services.NewSavingsService(accountService, transactionService, savingsRepo)
	rotatingSavingsService := services.NewRotatingSavingsService(accountRepo, transactionService, rotatingSavingsRepo)
	debtService := services.NewDebtService(transactionService, debtRepo)
	investmentService := services.NewInvestmentService(accountService, transactionService, investmentRepo)
	diagnosticsService := services.NewDiagnosticsService(db, redis)
	marketDataService := services.NewMarketDataService(cfg, db, redis, investmentService)

	deps := handlers.Deps{
		Cfg:                    cfg,
		DiagnosticsService:     diagnosticsService,
		MarketDataService:      marketDataService,
		AuthService:            authService,
		AccountService:         accountService,
		AuditService:           auditService,
		TransactionService:     transactionService,
		CategoryService:        categoryService,
		TagService:             tagService,
		BudgetService:          budgetService,
		SavingsService:         savingsService,
		RotatingSavingsService: rotatingSavingsService,
		DebtService:            debtService,
		InvestmentService:      investmentService,
	}

	h := httpapi.NewRouter(cfg, deps)

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

func (a *App) Close(ctx context.Context) {
	if a == nil || a.cleanup == nil {
		return
	}
	a.cleanup(ctx)
}
