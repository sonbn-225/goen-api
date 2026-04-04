package app

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/sonbn-225/goen-api-v2/docs"
	"github.com/sonbn-225/goen-api-v2/internal/core/config"
	"github.com/sonbn-225/goen-api-v2/internal/core/httpx"
	"github.com/sonbn-225/goen-api-v2/internal/core/response"
	"github.com/sonbn-225/goen-api-v2/internal/core/security"
	"github.com/sonbn-225/goen-api-v2/internal/domains/account"
	"github.com/sonbn-225/goen-api-v2/internal/domains/auth"
	"github.com/sonbn-225/goen-api-v2/internal/domains/budget"
	"github.com/sonbn-225/goen-api-v2/internal/domains/category"
	"github.com/sonbn-225/goen-api-v2/internal/domains/contact"
	"github.com/sonbn-225/goen-api-v2/internal/domains/debt"
	"github.com/sonbn-225/goen-api-v2/internal/domains/investment"
	"github.com/sonbn-225/goen-api-v2/internal/domains/media"
	"github.com/sonbn-225/goen-api-v2/internal/domains/profile"
	"github.com/sonbn-225/goen-api-v2/internal/domains/report"
	rotatingsavings "github.com/sonbn-225/goen-api-v2/internal/domains/rotating_savings"
	"github.com/sonbn-225/goen-api-v2/internal/domains/savings"
	"github.com/sonbn-225/goen-api-v2/internal/domains/setting"
	"github.com/sonbn-225/goen-api-v2/internal/domains/tag"
	"github.com/sonbn-225/goen-api-v2/internal/domains/transaction"
	"github.com/sonbn-225/goen-api-v2/internal/infra/objectstorage"
	"github.com/sonbn-225/goen-api-v2/internal/infra/postgres"
	repository "github.com/sonbn-225/goen-api-v2/internal/repository"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type App struct {
	Handler http.Handler
	cleanup func(context.Context)
}

func (a *App) Close(ctx context.Context) {
	if a == nil || a.cleanup == nil {
		return
	}
	a.cleanup(ctx)
}

func New(cfg *config.Config) *App {
	db, err := postgres.NewPool(cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}

	userRepo := repository.NewUserRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	budgetRepo := repository.NewBudgetRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	contactRepo := repository.NewContactRepository(db)
	debtRepo := repository.NewDebtRepository(db)
	investmentRepo := repository.NewInvestmentRepository(db)
	reportRepo := repository.NewReportRepository(db)
	rotatingSavingsRepo := repository.NewRotatingSavingsRepository(db)
	savingsRepo := repository.NewSavingsRepository(db)
	tagRepo := repository.NewTagRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	hasher := security.NewPasswordHasher()
	avatarStorage := objectstorage.NewSeaweedClient(objectstorage.Config{
		Endpoint:      cfg.S3Endpoint,
		AccessKey:     cfg.S3AccessKey,
		SecretKey:     cfg.S3SecretKey,
		Bucket:        cfg.S3Bucket,
		UseSSL:        cfg.S3UseSSL,
		PublicBaseURL: cfg.S3PublicBaseURL,
	})

	authMod := auth.NewModule(auth.ModuleDeps{
		UserRepo:         userRepo,
		Hasher:           hasher,
		Issuer:           security.NewTokenIssuer(cfg.JWTSecret, time.Duration(cfg.JWTAccessTTLMinutes)*time.Minute),
		AccessTTLMinutes: cfg.JWTAccessTTLMinutes,
	})
	mediaMod := media.NewModule(media.ModuleDeps{Storage: newMediaStorageAdapter(avatarStorage)})
	profileSvc := profile.NewService(userRepo, hasher, avatarStorage)
	settingSvc := setting.NewService(userRepo)
	profileMod := profile.NewModule(profile.ModuleDeps{Service: profileSvc})
	settingMod := setting.NewModule(setting.ModuleDeps{Service: settingSvc})

	accountMod := account.NewModule(account.ModuleDeps{Repo: accountRepo})
	budgetMod := budget.NewModule(budget.ModuleDeps{Repo: budgetRepo, CategoryRepo: categoryRepo})
	categoryMod := category.NewModule(category.ModuleDeps{Repo: categoryRepo})
	contactMod := contact.NewModule(contact.ModuleDeps{Repo: contactRepo})
	tagMod := tag.NewModule(tag.ModuleDeps{Repo: tagRepo})
	txMod := transaction.NewModule(transaction.ModuleDeps{Repo: txRepo})
	savingsMod := savings.NewModule(savings.ModuleDeps{Repo: savingsRepo, TxService: txMod.Service})
	debtMod := debt.NewModule(debt.ModuleDeps{Repo: debtRepo, TxService: txMod.Service, ContactService: contactMod.Service})
	investmentMod := investment.NewModule(investment.ModuleDeps{Repo: investmentRepo})
	rotatingSavingsMod := rotatingsavings.NewModule(rotatingsavings.ModuleDeps{Repo: rotatingSavingsRepo, TxService: txMod.Service})
	reportMod := report.NewModule(report.ModuleDeps{Repo: reportRepo})

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(httpx.RequestLogger())
	r.Use(middleware.Heartbeat("/healthz"))

	r.Get("/swagger", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/swagger/", http.StatusMovedPermanently)
	})
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
			response.WriteData(w, http.StatusOK, map[string]string{"message": "ok"})
		})

		authMod.RegisterPublicRoutes(r)
		r.Group(func(r chi.Router) {
			r.Use(httpx.AuthMiddleware(cfg.JWTSecret))
			authMod.RegisterProtectedRoutes(r)
			mediaMod.RegisterRoutes(r)
			profileMod.RegisterRoutes(r)
			settingMod.RegisterRoutes(r)
			accountMod.RegisterRoutes(r)
			budgetMod.RegisterRoutes(r)
			categoryMod.RegisterRoutes(r)
			contactMod.RegisterRoutes(r)
			debtMod.RegisterRoutes(r)
			investmentMod.RegisterRoutes(r)
			rotatingSavingsMod.RegisterRoutes(r)
			savingsMod.RegisterRoutes(r)
			reportMod.RegisterRoutes(r)
			tagMod.RegisterRoutes(r)
			txMod.RegisterRoutes(r)
		})
	})

	return &App{
		Handler: r,
		cleanup: func(_ context.Context) {
			db.Close()
		},
	}
}

type mediaStorageAdapter struct {
	client *objectstorage.SeaweedClient
}

func newMediaStorageAdapter(client *objectstorage.SeaweedClient) media.Storage {
	if client == nil {
		return nil
	}
	return mediaStorageAdapter{client: client}
}

func (a mediaStorageAdapter) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, media.ObjectInfo, error) {
	obj, info, err := a.client.GetObject(ctx, bucket, key)
	if err != nil {
		return nil, media.ObjectInfo{}, err
	}
	return obj, media.ObjectInfo{ContentType: info.ContentType}, nil
}
