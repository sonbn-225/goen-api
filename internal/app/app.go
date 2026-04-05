package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	_ "github.com/sonbn-225/goen-api/docs"
	v1 "github.com/sonbn-225/goen-api/internal/handler/http/v1"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
	"github.com/sonbn-225/goen-api/internal/pkg/storage"
	"github.com/sonbn-225/goen-api/internal/repository/postgres"
	"github.com/sonbn-225/goen-api/internal/service"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type App struct {
	Config *config.Config
	DB     *database.Postgres
	Redis  *database.Redis
	S3     *storage.S3Client
	Router *chi.Mux
}

func New(cfg *config.Config) (*App, error) {
	db := database.NewPostgres(cfg.DatabaseURL)
	if cfg.MigrateOnStart {
		if err := database.Migrate(context.Background(), db, cfg.MigrationDir); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	var rds *database.Redis
	if cfg.RedisURL != "" {
		var err error
		rds, err = database.NewRedis(cfg.RedisURL)
		if err != nil {
			slog.Error("failed to connect to redis", "error", err)
		}
	}

	s3, err := storage.NewS3Client(storage.S3Config{
		Endpoint:      cfg.S3Endpoint,
		AccessKey:     cfg.S3AccessKey,
		SecretKey:     cfg.S3SecretKey,
		Bucket:        cfg.S3Bucket,
		UseSSL:        cfg.S3UseSSL,
		PublicBaseURL: cfg.PublicBaseURL,
	})
	if err != nil {
		slog.Error("failed to initialize s3 storage", "error", err)
	}

	// Repositories
	userRepo := postgres.NewUserRepo(db)
	refreshRepo := postgres.NewRefreshTokenRepo(db)
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
	authSvc := service.NewAuthService(userRepo, refreshRepo, s3, cfg)
	categorySvc := service.NewCategoryService(categoryRepo, rds)
	tagSvc := service.NewTagService(tagRepo)
	accountSvc := service.NewAccountService(accountRepo, userRepo)
	transactionSvc := service.NewTransactionService(transactionRepo, tagSvc)
	contactSvc := service.NewContactService(contactRepo)
	debtSvc := service.NewDebtService(debtRepo, contactSvc)
	budgetSvc := service.NewBudgetService(budgetRepo, categoryRepo)
	reportSvc := service.NewReportService(reportRepo, accountRepo)
	groupExpenseSvc := service.NewGroupExpenseService(transactionSvc, debtSvc, groupExpenseRepo)
	investmentSvc := service.NewInvestmentService(investmentRepo, transactionSvc)
	marketDataSvc := service.NewMarketDataService(cfg, marketDataRepo, rds, investmentSvc)
	savingsSvc := service.NewSavingsService(savingsRepo)
	rotatingSavingsSvc := service.NewRotatingSavingsService(rotatingSavingsRepo, accountSvc, transactionSvc)
	publicSvc := service.NewPublicService(userRepo, accountRepo, groupExpenseRepo)
	diagnosticsSvc := service.NewDiagnosticsService(db)

	// Cross-inject
	transactionSvc.SetDebtService(debtSvc)

	// Handlers
	authHandler := v1.NewAuthHandler(authSvc)
	profileHandler := v1.NewProfileHandler(authSvc)
	settingsHandler := v1.NewSettingsHandler(authSvc)
	mediaHandler := v1.NewMediaHandler(s3)
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
	savingsHandler := v1.NewSavingsHandler(savingsSvc)
	rotatingSavingsHandler := v1.NewRotatingSavingsHandler(rotatingSavingsSvc)
	publicHandler := v1.NewPublicHandler(publicSvc)
	diagnosticsHandler := v1.NewDiagnosticsHandler(diagnosticsSvc)

	r := chi.NewRouter()

	// CORS
	corsOpts := cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}
	// Wildcard "*" is incompatible with AllowCredentials=true per CORS spec.
	// Use AllowOriginFunc to dynamically mirror the request Origin header instead.
	if len(cfg.CORSOrigins) == 1 && cfg.CORSOrigins[0] == "*" {
		corsOpts.AllowOriginFunc = func(origin string) bool {
			return true
		}
	} else {
		corsOpts.AllowedOrigins = cfg.CORSOrigins
		corsOpts.AllowOriginFunc = func(origin string) bool {
			for _, o := range cfg.CORSOrigins {
				if o == origin || o == "*" {
					return true
				}
			}
			return false
		}
	}
	r.Use(cors.New(corsOpts).Handler)

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), //The url pointing to API definition
	))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		})
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		})

		authHandler.RegisterRoutes(r, cfg)
		profileHandler.RegisterRoutes(r, cfg)
		settingsHandler.RegisterRoutes(r, cfg)
		mediaHandler.RegisterRoutes(r, cfg)
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
		rotatingSavingsHandler.RegisterRoutes(r, cfg)
		publicHandler.RegisterRoutes(r, cfg)
		diagnosticsHandler.RegisterRoutes(r, cfg)
	})

	return &App{
		Config: cfg,
		DB:     db,
		Redis:  rds,
		S3:     s3,
		Router: r,
	}, nil
}

func (a *App) Close(ctx context.Context) {
	if a.DB != nil {
		a.DB.Close()
	}
	if a.Redis != nil {
		a.Redis.Close()
	}
}
