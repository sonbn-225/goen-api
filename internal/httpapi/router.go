package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sonbn-225/goen-api/internal/auth"
	"github.com/sonbn-225/goen-api/internal/config"
	"github.com/sonbn-225/goen-api/internal/handlers"
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

	deps := handlers.Deps{
		Cfg:   cfg,
		DB:    storage.NewPostgres(cfg.DatabaseURL),
		Redis: storage.NewRedis(cfg.RedisURL),
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

		r.Get("/healthz", handlers.Healthz(deps))
		r.Get("/readyz", handlers.Readyz(deps))
		r.Get("/ping", handlers.Ping(deps))
		r.Get("/connectivity", handlers.Connectivity(deps))
	})

	return r
}
