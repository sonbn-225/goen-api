package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// Migrate runs database migrations using goose
func Migrate(ctx context.Context, db *Postgres, migrationDir string) error {
	slog.Info("running database migrations", "dir", migrationDir)

	pool, err := db.Pool(ctx)
	if err != nil {
		return fmt.Errorf("failed to get db pool: %w", err)
	}

	// Convert pgxpool to standard sql.DB for goose
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(sqlDB, migrationDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("database migrations completed successfully")
	return nil
}
