package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

type MarketDataRepo struct {
	db *database.Postgres
}

func NewMarketDataRepo(db *database.Postgres) *MarketDataRepo {
	return &MarketDataRepo{db: db}
}

func (r *MarketDataRepo) LoadSecurityIDsBySymbols(ctx context.Context, symbols []string) (map[string]string, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
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
	return out, nil
}

func (r *MarketDataRepo) LoadSyncState(ctx context.Context, syncKey string) (*entity.SyncState, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	var st entity.SyncState
	st.SyncKey = syncKey

	err = pool.QueryRow(ctx, `
		SELECT min_interval_seconds, last_started_at, last_success_at, last_failure_at, last_status, last_error
		FROM market_data_sync_states
		WHERE sync_key = $1
	`, syncKey).Scan(
		&st.MinIntervalSeconds, &st.LastStartedAt, &st.LastSuccessAt, &st.LastFailureAt, &st.LastStatus, &st.LastError,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if st.LastSuccessAt != nil {
		next := st.LastSuccessAt.Add(time.Duration(st.MinIntervalSeconds) * time.Second)
		st.NextDueAt = &next
		cd := int(time.Until(next).Seconds())
		if cd < 0 {
			cd = 0
		}
		st.CooldownSeconds = cd
	}

	return &st, nil
}
