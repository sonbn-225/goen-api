package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	v1 "github.com/sonbn-225/goen-api/internal/handler/http/v1"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
	"github.com/sonbn-225/goen-api/internal/pkg/storage"
	"github.com/sonbn-225/goen-api/internal/repository/postgres"
	"github.com/sonbn-225/goen-api/internal/service"
)

type App struct {
	Config *config.Config
	DB     *database.Postgres
	Redis  *database.Redis
	S3     *storage.S3Client
	Router *chi.Mux
}

func New(cfg *config.Config) *App {
	db := database.NewPostgres(cfg.DatabaseURL)
	
	var rds *database.Redis
	if cfg.RedisURL != "" {
		var err error
		rds, err = database.NewRedis(cfg.RedisURL)
		if err != nil {
			slog.Error("failed to connect to redis", "error", err)
		}
	}

	s3 := storage.NewS3Client(storage.S3Config{
		Endpoint:      cfg.S3Endpoint,
		AccessKey:     cfg.S3AccessKey,
		SecretKey:     cfg.S3SecretKey,
		Bucket:        cfg.S3Bucket,
		UseSSL:        cfg.S3UseSSL,
		PublicBaseURL: cfg.PublicBaseURL,
	})

	// Repositories
	userRepo := postgres.NewUserRepo(db)
	categoryRepo := postgres.NewCategoryRepo(db)
	tagRepo := postgres.NewTagRepo(db)
	accountRepo := postgres.NewAccountRepo(db)
	transactionRepo := postgres.NewTransactionRepo(db)
	contactRepo := postgres.NewContactRepo(db)
	debtRepo := postgres.NewDebtRepo(db)
	budgetRepo := postgres.NewBudgetRepo(db)
	reportRepo := postgres.NewReportRepo(db)
	groupExpenseRepo := postgres.NewGroupExpenseRepo(db)
	investmentRepo := postgres.NewInvestmentRepo(db)
	marketDataRepo := postgres.NewMarketDataRepo(db)
	savingsRepo := postgres.NewSavingsRepo(db)
	rotatingSavingsRepo := postgres.NewRotatingSavingsRepo(db)

	// Services
	authSvc := service.NewAuthService(userRepo, s3, cfg)
	categorySvc := service.NewCategoryService(categoryRepo)
	tagSvc := service.NewTagService(tagRepo)
	accountSvc := service.NewAccountService(accountRepo, userRepo)
	transactionSvc := service.NewTransactionService(transactionRepo, tagSvc)
	contactSvc := service.NewContactService(contactRepo)
	debtSvc := service.NewDebtService(debtRepo, contactSvc)
	budgetSvc := service.NewBudgetService(budgetRepo, categoryRepo)
	reportSvc := service.NewReportService(reportRepo, accountRepo)
	groupExpenseSvc := service.NewGroupExpenseService(transactionSvc, debtSvc, groupExpenseRepo)
	investmentSvc := service.NewInvestmentService(investmentRepo, accountSvc, transactionSvc)
	marketDataSvc := service.NewMarketDataService(cfg, marketDataRepo, rds, investmentSvc)
	savingsSvc := service.NewSavingsService(savingsRepo)
	rotatingSavingsSvc := service.NewRotatingSavingsService(rotatingSavingsRepo, accountSvc, transactionSvc)
	publicSvc := service.NewPublicService(userRepo, accountRepo, groupExpenseRepo)
	diagnosticsSvc := service.NewDiagnosticsService(db)

	// Cross-inject
	transactionSvc.SetDebtService(debtSvc)

	// Handlers
	authHandler := v1.NewAuthHandler(authSvc, s3)
	categoryHandler := v1.NewCategoryHandler(categorySvc)
	tagHandler := v1.NewTagHandler(tagSvc)
	accountHandler := v1.NewAccountHandler(accountSvc)
	transactionHandler := v1.NewTransactionHandler(transactionSvc)
	contactHandler := v1.NewContactHandler(contactSvc)
	debtHandler := v1.NewDebtHandler(debtSvc)
	budgetHandler := v1.NewBudgetHandler(budgetSvc)
	reportHandler := v1.NewReportHandler(reportSvc)
	groupExpenseHandler := v1.NewGroupExpenseHandler(groupExpenseSvc)
	investmentHandler := v1.NewInvestmentHandler(investmentSvc)
	marketDataHandler := v1.NewMarketDataHandler(marketDataSvc)
	savingsHandler := v1.NewSavingsHandler(savingsSvc, rotatingSavingsSvc)
	publicHandler := v1.NewPublicHandler(publicSvc)
	diagnosticsHandler := v1.NewDiagnosticsHandler(diagnosticsSvc)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		})

		authHandler.RegisterRoutes(r, cfg)
		categoryHandler.RegisterRoutes(r, cfg)
		tagHandler.RegisterRoutes(r, cfg)
		accountHandler.RegisterRoutes(r, cfg)
		transactionHandler.RegisterRoutes(r, cfg)
		contactHandler.RegisterRoutes(r, cfg)
		debtHandler.RegisterRoutes(r, cfg)
		budgetHandler.RegisterRoutes(r, cfg)
		reportHandler.RegisterRoutes(r, cfg)
		groupExpenseHandler.RegisterRoutes(r, cfg)
		investmentHandler.RegisterRoutes(r, cfg)
		marketDataHandler.RegisterRoutes(r, cfg)
		savingsHandler.RegisterRoutes(r, cfg)
		publicHandler.RegisterRoutes(r, cfg)
		diagnosticsHandler.RegisterRoutes(r, cfg)
	})

	return &App{
		Config: cfg,
		DB:     db,
		Redis:  rds,
		S3:     s3,
		Router: r,
	}
}

func (a *App) Close(ctx context.Context) {
	if a.DB != nil {
		a.DB.Close()
	}
	if a.Redis != nil {
		a.Redis.Close()
	}
}
