package marketdata

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/sonbn-225/goen-api/internal/storage"
)

// PostgresRepo implements Repository using PostgreSQL.
type PostgresRepo struct {
	db *storage.Postgres
}

func NewPostgresRepo(db *storage.Postgres) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) LoadSecurityIDsBySymbols(ctx context.Context, symbols []string) (map[string]string, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}
	if pool == nil {
		return nil, nil
	}

	rows, err := pool.Query(ctx, `
		SELECT symbol, id
		FROM securities
		WHERE symbol = ANY($1)
	`, symbols)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var sym, id string
		if err := rows.Scan(&sym, &id); err != nil {
			return nil, err
		}
		out[strings.ToUpper(strings.TrimSpace(sym))] = id
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *PostgresRepo) LoadSyncState(ctx context.Context, syncKey string) (*SyncState, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}
	if pool == nil {
		return nil, nil
	}

	var (
		minIntervalSeconds int
		lastStartedAt      *time.Time
		lastSuccessAt      *time.Time
		lastFailureAt      *time.Time
		lastStatus         string
		lastError          *string
	)

	err = pool.QueryRow(ctx, `
		SELECT min_interval_seconds, last_started_at, last_success_at, last_failure_at, last_status, last_error
		FROM market_data_sync_states
		WHERE sync_key = $1
	`, syncKey).Scan(&minIntervalSeconds, &lastStartedAt, &lastSuccessAt, &lastFailureAt, &lastStatus, &lastError)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	st := &SyncState{
		SyncKey:            syncKey,
		MinIntervalSeconds: minIntervalSeconds,
		LastStartedAt:      lastStartedAt,
		LastSuccessAt:      lastSuccessAt,
		LastFailureAt:      lastFailureAt,
		LastStatus:         lastStatus,
		LastError:          lastError,
	}

	if lastSuccessAt != nil {
		next := lastSuccessAt.Add(time.Duration(minIntervalSeconds) * time.Second)
		st.NextDueAt = &next
		cd := int(time.Until(next).Seconds())
		if cd < 0 {
			cd = 0
		}
		st.CooldownSeconds = cd
	}

	return st, nil
}
