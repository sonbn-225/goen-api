package service

import (
	"context"
	"time"

	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

var startTime = time.Now()

type DiagnosticsService struct {
	db *database.Postgres
}

func NewDiagnosticsService(db *database.Postgres) *DiagnosticsService {
	return &DiagnosticsService{db: db}
}

func (s *DiagnosticsService) GetDiagnostics(ctx context.Context) (*entity.Diagnostics, error) {
	dbStatus := "unknown"
	var dbStats map[string]any

	pool, err := s.db.Pool(ctx)
	if err == nil {
		dbStatus = "connected"
		stats := pool.Stat()
		dbStats = map[string]any{
			"total_conns":            stats.TotalConns(),
			"idle_conns":             stats.IdleConns(),
			"acquired_conns":         stats.AcquiredConns(),
			"max_conns":              stats.MaxConns(),
			"new_conns_count":        stats.NewConnsCount(),
			"max_idle_destroy_count": stats.MaxIdleDestroyCount(),
			"max_lifetime_destroy":   stats.MaxLifetimeDestroyCount(),
		}
	} else {
		dbStatus = "disconnected: " + err.Error()
	}

	return &entity.Diagnostics{
		Status:    "running",
		DBStatus:  dbStatus,
		DBStats:   dbStats,
		Version:   "1.0.0", // Hardcoded or from build tags
		UptimeSec: time.Since(startTime).Seconds(),
	}, nil
}
