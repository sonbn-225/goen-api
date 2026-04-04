package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
	"github.com/sonbn-225/goen-api-v2/internal/core/money"
	"github.com/sonbn-225/goen-api-v2/internal/domains/account"
)

type AccountRepository struct {
	db *pgxpool.Pool
}

func NewAccountRepository(db *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, acc *account.Account) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "account", "operation", "create", "user_id", acc.UserID, "account_id", acc.ID)
	now := time.Now().UTC()
	tx, err := r.db.Begin(ctx)
	if err != nil {
		logger.Error("repo_account_create_failed", "error", err)
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO accounts (
			id,
			name,
			account_number,
			color,
			account_type,
			currency,
			parent_account_id,
			status,
			closed_at,
			created_at,
			updated_at,
			created_by,
			updated_by
		) VALUES ($1, $2, $3, $4, $5::account_type, $6, $7, $8::account_status, $9, $10, $11, $12, $13)
	`, acc.ID, acc.Name, acc.AccountNumber, acc.Color, acc.Type, acc.Currency, acc.ParentAccountID, acc.Status, acc.ClosedAt, now, now, acc.UserID, acc.UserID)
	if err != nil {
		logger.Error("repo_account_create_failed", "error", err)
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO user_accounts (
			id,
			account_id,
			user_id,
			permission,
			status,
			created_at,
			updated_at,
			created_by,
			updated_by
		) VALUES ($1, $2, $3, 'owner', 'active', $4, $5, $6, $7)
	`, uuid.NewString(), acc.ID, acc.UserID, now, now, acc.UserID, acc.UserID)
	if err != nil {
		logger.Error("repo_account_create_failed", "error", err)
		return err
	}

	if acc.Type == "broker" {
		_, err = tx.Exec(ctx, `
			INSERT INTO investment_accounts (
				id,
				account_id,
				created_at,
				updated_at
			) VALUES ($1, $2, $3, $4)
		`, uuid.NewString(), acc.ID, now, now)
		if err != nil {
			logger.Error("repo_account_create_failed", "error", err)
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Error("repo_account_create_failed", "error", err)
		return err
	}
	logger.Info("repo_account_create_succeeded")
	return nil
}

func (r *AccountRepository) ListByUser(ctx context.Context, userID string) ([]account.Account, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "account", "operation", "list_by_user", "user_id", userID)
	rows, err := r.db.Query(ctx, `
		SELECT
			a.id,
			a.name,
			a.account_number,
			a.color,
			a.account_type::text,
			a.currency,
			a.parent_account_id,
			a.status::text,
			a.closed_at,
			a.updated_at,
			COALESCE(
				SUM(
					CASE
						WHEN t.type = 'income' AND t.account_id = a.id THEN t.amount
						WHEN t.type = 'expense' AND t.account_id = a.id THEN -t.amount
						WHEN t.type = 'transfer' AND t.to_account_id = a.id THEN COALESCE(t.to_amount, t.amount)
						WHEN t.type = 'transfer' AND t.from_account_id = a.id THEN -COALESCE(t.from_amount, t.amount)
						ELSE 0
					END
				),
				0
			)::text AS balance,
			a.created_at
		FROM accounts a
		JOIN user_accounts ua ON ua.account_id = a.id
		LEFT JOIN transactions t
		  ON t.deleted_at IS NULL
		 AND t.status = 'posted'
		 AND (
		   t.account_id = a.id
		   OR t.from_account_id = a.id
		   OR t.to_account_id = a.id
		 )
		WHERE ua.user_id = $1
		  AND ua.status = 'active'
		  AND a.deleted_at IS NULL
		GROUP BY a.id, a.name, a.account_number, a.color, a.account_type, a.currency, a.parent_account_id, a.status, a.closed_at, a.updated_at, a.created_at
		ORDER BY a.created_at DESC, a.id DESC
	`, userID)
	if err != nil {
		logger.Error("repo_account_list_failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	items := make([]account.Account, 0)
	for rows.Next() {
		var item account.Account
		var balanceStr string
		item.UserID = userID
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.AccountNumber,
			&item.Color,
			&item.Type,
			&item.Currency,
			&item.ParentAccountID,
			&item.Status,
			&item.ClosedAt,
			&item.UpdatedAt,
			&balanceStr,
			&item.CreatedAt,
		); err != nil {
			logger.Error("repo_account_list_failed", "error", err)
			return nil, err
		}
		balance, err := money.NewFromString(balanceStr)
		if err != nil {
			logger.Error("repo_account_list_failed", "error", err)
			return nil, err
		}
		item.Balance = balance
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		logger.Error("repo_account_list_failed", "error", err)
		return nil, err
	}
	logger.Info("repo_account_list_succeeded", "count", len(items))

	return items, nil
}

func (r *AccountRepository) GetByID(ctx context.Context, userID, accountID string) (*account.Account, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "account", "operation", "get_by_id", "user_id", userID, "account_id", accountID)
	row := r.db.QueryRow(ctx, `
		SELECT
			a.id,
			a.name,
			a.account_number,
			a.color,
			a.account_type::text,
			a.currency,
			a.parent_account_id,
			a.status::text,
			a.closed_at,
			a.updated_at,
			COALESCE(
				SUM(
					CASE
						WHEN t.type = 'income' AND t.account_id = a.id THEN t.amount
						WHEN t.type = 'expense' AND t.account_id = a.id THEN -t.amount
						WHEN t.type = 'transfer' AND t.to_account_id = a.id THEN COALESCE(t.to_amount, t.amount)
						WHEN t.type = 'transfer' AND t.from_account_id = a.id THEN -COALESCE(t.from_amount, t.amount)
						ELSE 0
					END
				),
				0
			)::text AS balance,
			a.created_at
		FROM accounts a
		JOIN user_accounts ua ON ua.account_id = a.id
		LEFT JOIN transactions t
		  ON t.deleted_at IS NULL
		 AND t.status = 'posted'
		 AND (
		   t.account_id = a.id
		   OR t.from_account_id = a.id
		   OR t.to_account_id = a.id
		 )
		WHERE ua.user_id = $1
		  AND ua.status = 'active'
		  AND a.deleted_at IS NULL
		  AND a.id = $2
		GROUP BY a.id, a.name, a.account_number, a.color, a.account_type, a.currency, a.parent_account_id, a.status, a.closed_at, a.updated_at, a.created_at
	`, userID, accountID)

	var item account.Account
	var balanceStr string
	item.UserID = userID
	if err := row.Scan(
		&item.ID,
		&item.Name,
		&item.AccountNumber,
		&item.Color,
		&item.Type,
		&item.Currency,
		&item.ParentAccountID,
		&item.Status,
		&item.ClosedAt,
		&item.UpdatedAt,
		&balanceStr,
		&item.CreatedAt,
	); err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		logger.Error("repo_account_get_failed", "error", err)
		return nil, err
	}

	balance, err := money.NewFromString(balanceStr)
	if err != nil {
		logger.Error("repo_account_get_failed", "error", err)
		return nil, err
	}
	item.Balance = balance
	logger.Info("repo_account_get_succeeded")

	return &item, nil
}

func (r *AccountRepository) GetDefaultCurrency(ctx context.Context, userID string) (string, error) {
	row := r.db.QueryRow(ctx, `
		SELECT COALESCE(settings->>'default_currency', '')
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, userID)

	var currency string
	if err := row.Scan(&currency); err != nil {
		if isNoRows(err) {
			return "", nil
		}
		return "", err
	}
	return currency, nil
}

func (r *AccountRepository) IsOwner(ctx context.Context, userID, accountID string) (bool, error) {
	row := r.db.QueryRow(ctx, `
		SELECT 1
		FROM user_accounts ua
		WHERE ua.user_id = $1
		  AND ua.account_id = $2
		  AND ua.status = 'active'
		  AND ua.permission = 'owner'
	`, userID, accountID)

	var one int
	if err := row.Scan(&one); err != nil {
		if isNoRows(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *AccountRepository) HasRelatedTransferTransactionsForAccount(ctx context.Context, accountID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM transactions t
			WHERE t.deleted_at IS NULL
			  AND t.type = 'transfer'
			  AND (t.from_account_id = $1 OR t.to_account_id = $1)
		)
	`, accountID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *AccountRepository) Delete(ctx context.Context, userID, accountID string) (bool, error) {
	now := time.Now().UTC()
	ct, err := r.db.Exec(ctx, `
		UPDATE accounts
		SET deleted_at = $1,
		    updated_at = $1,
		    updated_by = $2
		WHERE id = $3 AND deleted_at IS NULL
	`, now, userID, accountID)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() > 0, nil
}
