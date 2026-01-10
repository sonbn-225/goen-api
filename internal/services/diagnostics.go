package services

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/storage"
)

type ConnectivityItem struct {
	OK      bool           `json:"ok"`
	Details map[string]any `json:"details,omitempty"`
	Error   string         `json:"error,omitempty"`
}

type ConnectivityResponse struct {
	Postgres ConnectivityItem `json:"postgres"`
	Redis    ConnectivityItem `json:"redis"`
}

type DiagnosticsService interface {
	Readiness(ctx context.Context) (checks map[string]string, ready bool)
	Connectivity(ctx context.Context) ConnectivityResponse
}

type diagnosticsService struct {
	db    *storage.Postgres
	redis *storage.Redis
}

func NewDiagnosticsService(db *storage.Postgres, redis *storage.Redis) DiagnosticsService {
	return &diagnosticsService{db: db, redis: redis}
}

func (s *diagnosticsService) Readiness(ctx context.Context) (map[string]string, bool) {
	checks := map[string]string{}
	ready := true

	if s.db != nil {
		if err := s.db.Ping(ctx); err != nil {
			checks["postgres"] = "error"
			ready = false
		} else {
			checks["postgres"] = "ok"
		}
	}

	if s.redis != nil {
		if err := s.redis.Ping(ctx); err != nil {
			checks["redis"] = "error"
			ready = false
		} else {
			checks["redis"] = "ok"
		}
	}

	return checks, ready
}

func (s *diagnosticsService) Connectivity(ctx context.Context) ConnectivityResponse {
	resp := ConnectivityResponse{}

	if s.db == nil {
		resp.Postgres = ConnectivityItem{OK: false, Error: "DATABASE_URL not set"}
	} else if details, err := s.db.Probe(ctx); err != nil {
		resp.Postgres = ConnectivityItem{OK: false, Error: err.Error()}
	} else {
		resp.Postgres = ConnectivityItem{OK: true, Details: details}
	}

	if s.redis == nil {
		resp.Redis = ConnectivityItem{OK: false, Error: "REDIS_URL not set or invalid"}
	} else if details, err := s.redis.Probe(ctx); err != nil {
		resp.Redis = ConnectivityItem{OK: false, Error: err.Error()}
	} else {
		resp.Redis = ConnectivityItem{OK: true, Details: details}
	}

	return resp
}
