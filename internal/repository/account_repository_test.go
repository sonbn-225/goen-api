package repository

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonbn-225/goen-api-v2/internal/infra/postgres"
)

func TestAccountRepository_ListByUser_TransferBidirectionalBalance(t *testing.T) {
	dbURL := strings.TrimSpace(os.Getenv("GOEN_TEST_DATABASE_URL"))
	if dbURL == "" {
		t.Skip("set GOEN_TEST_DATABASE_URL to run Postgres integration tests")
	}

	pool := mustOpenAccountTestPool(t, dbURL)
	t.Cleanup(func() { pool.Close() })

	ensureAccountBalanceSchema(t, pool)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	userID := "u_acc_it_" + suffix
	acc1 := "acc_it_source_" + suffix
	acc2 := "acc_it_target_" + suffix

	seedAccountBalanceData(t, pool, userID, acc1, acc2)
	t.Cleanup(func() { cleanupAccountBalanceData(t, pool, userID, acc1, acc2) })

	repo := NewAccountRepository(pool)
	items, err := repo.ListByUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(items))
	}

	byID := map[string]string{}
	for _, it := range items {
		byID[it.ID] = it.Balance.Decimal.String()
	}

	if got := byID[acc1]; got != "500" {
		t.Fatalf("expected source account balance 500, got %s", got)
	}
	if got := byID[acc2]; got != "290" {
		t.Fatalf("expected target account balance 290, got %s", got)
	}
}

func mustOpenAccountTestPool(t *testing.T, dbURL string) *pgxpool.Pool {
	t.Helper()
	pool, err := postgres.NewPool(dbURL)
	if err != nil {
		t.Fatalf("failed to open postgres pool: %v", err)
	}
	return pool
}

func ensureAccountBalanceSchema(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx, `
		DO $$ BEGIN
			CREATE TYPE account_type AS ENUM ('bank','wallet','cash','broker','card','savings');
		EXCEPTION WHEN duplicate_object THEN NULL;
		END $$;

		DO $$ BEGIN
			CREATE TYPE user_account_permission AS ENUM ('owner','viewer','editor');
		EXCEPTION WHEN duplicate_object THEN NULL;
		END $$;

		DO $$ BEGIN
			CREATE TYPE user_account_status AS ENUM ('active','revoked');
		EXCEPTION WHEN duplicate_object THEN NULL;
		END $$;

		DO $$ BEGIN
			CREATE TYPE transaction_type AS ENUM ('expense','income','transfer');
		EXCEPTION WHEN duplicate_object THEN NULL;
		END $$;

		DO $$ BEGIN
			CREATE TYPE transaction_status AS ENUM ('pending','posted','cancelled');
		EXCEPTION WHEN duplicate_object THEN NULL;
		END $$;

		CREATE TABLE IF NOT EXISTS users (
			id text PRIMARY KEY,
			username text NOT NULL UNIQUE,
			email text,
			phone text,
			display_name text,
			status text NOT NULL DEFAULT 'active',
			password_hash text NOT NULL DEFAULT 'x',
			settings jsonb NOT NULL DEFAULT '{}'::jsonb,
			created_at timestamptz NOT NULL,
			updated_at timestamptz NOT NULL,
			deleted_at timestamptz
		);

		CREATE TABLE IF NOT EXISTS accounts (
			id text PRIMARY KEY,
			name varchar NOT NULL,
			account_number text,
			color varchar,
			account_type account_type NOT NULL,
			currency varchar NOT NULL,
			parent_account_id text,
			status text NOT NULL DEFAULT 'active',
			closed_at timestamptz,
			created_at timestamptz NOT NULL,
			updated_at timestamptz NOT NULL,
			created_by text,
			updated_by text,
			deleted_at timestamptz
		);

		CREATE TABLE IF NOT EXISTS user_accounts (
			id text PRIMARY KEY,
			account_id text NOT NULL,
			user_id text NOT NULL,
			permission user_account_permission NOT NULL,
			status user_account_status NOT NULL DEFAULT 'active',
			created_at timestamptz NOT NULL,
			updated_at timestamptz NOT NULL,
			created_by text,
			updated_by text,
			CONSTRAINT uq_user_accounts_account_user UNIQUE (account_id, user_id)
		);

		CREATE TABLE IF NOT EXISTS transactions (
			id text PRIMARY KEY,
			type transaction_type NOT NULL,
			occurred_at timestamptz NOT NULL,
			amount numeric(18,2) NOT NULL,
			from_amount numeric(18,2),
			to_amount numeric(18,2),
			account_id text,
			from_account_id text,
			to_account_id text,
			status transaction_status NOT NULL DEFAULT 'pending',
			created_at timestamptz NOT NULL,
			updated_at timestamptz NOT NULL,
			created_by text,
			updated_by text,
			deleted_at timestamptz
		);
	`)
	if err != nil {
		t.Fatalf("failed to ensure schema: %v", err)
	}
}

func seedAccountBalanceData(t *testing.T, pool *pgxpool.Pool, userID, acc1, acc2 string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	now := time.Now().UTC()

	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, username, email, display_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, "user_"+userID, userID+"@example.com", "IT User", now, now)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO accounts (id, name, account_type, currency, created_at, updated_at, created_by, updated_by)
		VALUES
			($1, 'Wallet', 'cash', 'VND', $3, $3, $2, $2),
			($4, 'Savings', 'savings', 'VND', $3, $3, $2, $2)
	`, acc1, userID, now, acc2)
	if err != nil {
		t.Fatalf("failed to insert accounts: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO user_accounts (id, account_id, user_id, permission, status, created_at, updated_at, created_by, updated_by)
		VALUES
			($1, $2, $3, 'owner', 'active', $4, $4, $3, $3),
			($5, $6, $3, 'owner', 'active', $4, $4, $3, $3)
	`, "ua_"+acc1, acc1, userID, now, "ua_"+acc2, acc2)
	if err != nil {
		t.Fatalf("failed to insert user_accounts: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO transactions (
			id, type, occurred_at, amount, from_amount, to_amount, account_id, from_account_id, to_account_id,
			status, created_at, updated_at, created_by, updated_by
		)
		VALUES
			('tx_income_' || $1, 'income', $4, 1000, NULL, NULL, $2, NULL, NULL, 'posted', $4, $4, $1, $1),
			('tx_expense_' || $1, 'expense', $4, 200, NULL, NULL, $2, NULL, NULL, 'posted', $4, $4, $1, $1),
			('tx_transfer_' || $1, 'transfer', $4, 300, 300, 290, NULL, $2, $3, 'posted', $4, $4, $1, $1),
			('tx_pending_' || $1, 'income', $4, 9999, NULL, NULL, $2, NULL, NULL, 'pending', $4, $4, $1, $1)
	`, userID, acc1, acc2, now)
	if err != nil {
		t.Fatalf("failed to insert transactions: %v", err)
	}
}

func cleanupAccountBalanceData(t *testing.T, pool *pgxpool.Pool, userID, acc1, acc2 string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, _ = pool.Exec(ctx, `DELETE FROM transactions WHERE created_by = $1`, userID)
	_, _ = pool.Exec(ctx, `DELETE FROM user_accounts WHERE user_id = $1`, userID)
	_, _ = pool.Exec(ctx, `DELETE FROM accounts WHERE id IN ($1, $2)`, acc1, acc2)
	_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
}
