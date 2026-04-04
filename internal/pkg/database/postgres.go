package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool        *pgxpool.Pool
	databaseURL string
	mu          sync.Mutex
}

func NewPostgres(databaseURL string) *Postgres {
	return &Postgres{
		databaseURL: databaseURL,
	}
}

func (p *Postgres) Pool(ctx context.Context) (*pgxpool.Pool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pool != nil {
		return p.pool, nil
	}

	pool, err := pgxpool.New(ctx, p.databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}

	p.pool = pool
	return p.pool, nil
}

func (p *Postgres) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.pool != nil {
		p.pool.Close()
		p.pool = nil
	}
}

func (p *Postgres) Ping(ctx context.Context) error {
	pool, err := p.Pool(ctx)
	if err != nil {
		return err
	}
	return pool.Ping(ctx)
}

func (p *Postgres) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	pool, err := p.Pool(ctx)
	if err != nil {
		return err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func ParseInt64(s string) (int64, error) {
	var i int64
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}
