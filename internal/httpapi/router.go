package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sonbn-225/goen-api/internal/auth"
	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/handlers"
	"github.com/sonbn-225/goen-api/internal/services"
	"github.com/sonbn-225/goen-api/internal/storage"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func NewRouter(cfg *config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(CORSMiddleware(cfg))
	r.Use(middleware.Heartbeat("/healthz"))

	db := storage.NewPostgres(cfg.DatabaseURL)
	redis := storage.NewRedis(cfg.RedisURL)
	userRepo := storage.NewUserRepo(db)
	accountRepo := storage.NewAccountRepo(db)
	txRepo := storage.NewTransactionRepo(db)
	categoryRepo := storage.NewCategoryRepo(db)
	tagRepo := storage.NewTagRepo(db)
	budgetRepo := storage.NewBudgetRepo(db)
	savingsRepo := storage.NewSavingsRepo(db)
	rotatingSavingsRepo := storage.NewRotatingSavingsRepo(db)
	authService := services.NewAuthService(userRepo, cfg)
	accountService := services.NewAccountService(accountRepo)
	transactionService := services.NewTransactionService(txRepo)
	categoryService := services.NewCategoryService(categoryRepo)
	tagService := services.NewTagService(tagRepo)
	budgetService := services.NewBudgetService(budgetRepo, categoryRepo)
	savingsService := services.NewSavingsService(accountRepo, savingsRepo)
	rotatingSavingsService := services.NewRotatingSavingsService(accountRepo, transactionService, rotatingSavingsRepo)

	deps := handlers.Deps{
		Cfg:         cfg,
		DB:          db,
		Redis:       redis,
		AuthService: authService,
		AccountService: accountService,
		TransactionService: transactionService,
		CategoryService: categoryService,
		TagService: tagService,
		BudgetService: budgetService,
		SavingsService: savingsService,
		RotatingSavingsService: rotatingSavingsService,
	}

	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/signup", handlers.Signup(deps))
			r.Post("/signin", handlers.Signin(deps))
			r.With(auth.Middleware(cfg)).Get("/me", handlers.Me(deps))
		})

		accountsAuth := auth.Middleware(cfg)
		r.With(accountsAuth).Get("/accounts", handlers.ListAccounts(deps))
		r.With(accountsAuth).Get("/accounts/", handlers.ListAccounts(deps))
		r.With(accountsAuth).Post("/accounts", handlers.CreateAccount(deps))
		r.With(accountsAuth).Post("/accounts/", handlers.CreateAccount(deps))
		r.With(accountsAuth).Get("/accounts/{accountId}", handlers.GetAccount(deps))

		txAuth := auth.Middleware(cfg)
		r.With(txAuth).Get("/transactions", handlers.ListTransactions(deps))
		r.With(txAuth).Get("/transactions/", handlers.ListTransactions(deps))
		r.With(txAuth).Post("/transactions", handlers.CreateTransaction(deps))
		r.With(txAuth).Post("/transactions/", handlers.CreateTransaction(deps))
		r.With(txAuth).Get("/transactions/{transactionId}", handlers.GetTransaction(deps))
		r.With(txAuth).Patch("/transactions/{transactionId}", handlers.PatchTransaction(deps))
		r.With(txAuth).Delete("/transactions/{transactionId}", handlers.DeleteTransaction(deps))

		catAuth := auth.Middleware(cfg)
		r.With(catAuth).Get("/categories", handlers.ListCategories(deps))
		r.With(catAuth).Get("/categories/", handlers.ListCategories(deps))
		r.With(catAuth).Post("/categories", handlers.CreateCategory(deps))
		r.With(catAuth).Post("/categories/", handlers.CreateCategory(deps))
		r.With(catAuth).Get("/categories/{categoryId}", handlers.GetCategory(deps))

		tagAuth := auth.Middleware(cfg)
		r.With(tagAuth).Get("/tags", handlers.ListTags(deps))
		r.With(tagAuth).Get("/tags/", handlers.ListTags(deps))
		r.With(tagAuth).Post("/tags", handlers.CreateTag(deps))
		r.With(tagAuth).Post("/tags/", handlers.CreateTag(deps))
		r.With(tagAuth).Get("/tags/{tagId}", handlers.GetTag(deps))

		budgetAuth := auth.Middleware(cfg)
		r.With(budgetAuth).Get("/budgets", handlers.ListBudgets(deps))
		r.With(budgetAuth).Get("/budgets/", handlers.ListBudgets(deps))
		r.With(budgetAuth).Post("/budgets", handlers.CreateBudget(deps))
		r.With(budgetAuth).Post("/budgets/", handlers.CreateBudget(deps))
		r.With(budgetAuth).Get("/budgets/{budgetId}", handlers.GetBudget(deps))

		savingsAuth := auth.Middleware(cfg)
		r.With(savingsAuth).Get("/savings/instruments", handlers.ListSavingsInstruments(deps))
		r.With(savingsAuth).Get("/savings/instruments/", handlers.ListSavingsInstruments(deps))
		r.With(savingsAuth).Post("/savings/instruments", handlers.CreateSavingsInstrument(deps))
		r.With(savingsAuth).Post("/savings/instruments/", handlers.CreateSavingsInstrument(deps))
		r.With(savingsAuth).Get("/savings/instruments/{instrumentId}", handlers.GetSavingsInstrument(deps))

		rotAuth := auth.Middleware(cfg)
		r.With(rotAuth).Get("/rotating-savings/groups", handlers.ListRotatingSavingsGroups(deps))
		r.With(rotAuth).Get("/rotating-savings/groups/", handlers.ListRotatingSavingsGroups(deps))
		r.With(rotAuth).Post("/rotating-savings/groups", handlers.CreateRotatingSavingsGroup(deps))
		r.With(rotAuth).Post("/rotating-savings/groups/", handlers.CreateRotatingSavingsGroup(deps))
		r.With(rotAuth).Get("/rotating-savings/groups/{groupId}", handlers.GetRotatingSavingsGroup(deps))
		r.With(rotAuth).Get("/rotating-savings/groups/{groupId}/contributions", handlers.ListRotatingSavingsContributions(deps))
		r.With(rotAuth).Post("/rotating-savings/groups/{groupId}/contributions", handlers.CreateRotatingSavingsContribution(deps))

		r.Get("/healthz", handlers.Healthz(deps))
		r.Get("/readyz", handlers.Readyz(deps))
		r.Get("/ping", handlers.Ping(deps))
		r.Get("/connectivity", handlers.Connectivity(deps))
	})

	return r
}
