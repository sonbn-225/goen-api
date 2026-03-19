package storage

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func RunMigrations(databaseURL, migrationDir string) error {
	if databaseURL == "" {
		return nil
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(db, migrationDir)
}
