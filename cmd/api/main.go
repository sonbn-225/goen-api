package main

// @title Goen API
// @version 1.0
// @description REST API for Goen Project
// @host api.goen.wisteria.local
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sonbn-225/goen-api/internal/app"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/logger"
)

// HealthCheck handles the basic health check endpoint
// @Summary System Health Check
// @Description Returns the health status of the API
// @Tags System
// @Produce plain
// @Success 200 {string} string "ok"
// @Router /health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.LogLevel)

	a, err := app.New(cfg)
	if err != nil {
		slog.Error("failed to initialize app", "error", err)
		os.Exit(1)
	}

	a.StartJobs(context.Background())

	srv := &http.Server{
		Addr:         cfg.HTTPAddr(),
		Handler:      a.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("starting http server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server stopped with error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "error", err)
	}
	a.Close(ctx)
}
