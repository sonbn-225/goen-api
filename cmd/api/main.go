package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sonbn-225/goen-api-v2/internal/app"
	"github.com/sonbn-225/goen-api-v2/internal/core/config"
	"github.com/sonbn-225/goen-api-v2/internal/infra/postgres"
)

// @title goen-api-v2
// @version 0.1.0
// @description goen-api-v2 clean architecture backend API.
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(logger)

	if cfg.MigrateOnStart && cfg.DatabaseURL != "" {
		logger.Info("running database migrations", "dir", cfg.MigrationDir)
		if err := postgres.RunMigrations(cfg.DatabaseURL, cfg.MigrationDir); err != nil {
			logger.Error("failed to run database migrations", "err", err)
			os.Exit(1)
		}
		logger.Info("database migrations completed")
	}

	a := app.New(cfg)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr(),
		Handler:           a.Handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("http server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server stopped with error", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("shutting down")
	_ = srv.Shutdown(ctx)
	a.Close(ctx)
}
