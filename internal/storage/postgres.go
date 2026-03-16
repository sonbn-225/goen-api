package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool        *pgxpool.Pool
	databaseURL string
}

func NewPostgres(databaseURL string) *Postgres {
	if databaseURL == "" {
		return nil
	}
	// NOTE: We connect lazily (on first Ping) so the process can start even if DB is temporarily unavailable.
	return &Postgres{pool: nil, databaseURL: databaseURL}
}

func (p *Postgres) Ping(ctx context.Context) error {
	if p == nil {
		return nil
	}
	pool, err := p.ensurePool(ctx)
	if err != nil {
		return err
	}
	return pool.Ping(ctx)
}

func (p *Postgres) Probe(ctx context.Context) (map[string]any, error) {
	if p == nil {
		return nil, nil
	}
	pool, err := p.ensurePool(ctx)
	if err != nil {
		return nil, err
	}

	var one int
	if err := pool.QueryRow(ctx, "select 1").Scan(&one); err != nil {
		return nil, err
	}

	var dbName string
	_ = pool.QueryRow(ctx, "select current_database()").Scan(&dbName)

	var serverVersion string
	_ = pool.QueryRow(ctx, "show server_version").Scan(&serverVersion)

	return map[string]any{
		"select_1":       one,
		"database":       dbName,
		"server_version": serverVersion,
	}, nil
}

func (p *Postgres) ensurePool(ctx context.Context) (*pgxpool.Pool, error) {
	if p.pool != nil {
		return p.pool, nil
	}
	pool, err := pgxpool.New(ctx, p.databaseURL)
	if err != nil {
		return nil, fmt.Errorf("postgres connect: %w", err)
	}
	p.pool = pool
	return p.pool, nil
}

func (p *Postgres) Pool(ctx context.Context) (*pgxpool.Pool, error) {
	if p == nil {
		return nil, nil
	}
	return p.ensurePool(ctx)
}

func (p *Postgres) Close() {
	if p == nil {
		return
	}
	if p.pool != nil {
		p.pool.Close()
		p.pool = nil
	}
}

