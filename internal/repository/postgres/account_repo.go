package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

type AccountRepo struct {
	db *database.Postgres
}

func NewAccountRepo(db *database.Postgres) *AccountRepo {
	return &AccountRepo{db: db}
}

func (r *AccountRepo) CreateAccountWithOwner(ctx context.Context, account entity.Account, ownerUserID string) error {
	return r.db.WithTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO accounts (
				id, client_id, name, account_number, color, account_type, currency, parent_account_id, status, closed_at,
				created_at, updated_at, created_by, updated_by, deleted_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		`,
			account.ID, account.ClientID, account.Name, account.AccountNumber, account.Color,
			account.AccountType, account.Currency, account.ParentAccountID, account.Status,
			account.ClosedAt, account.CreatedAt, account.UpdatedAt, account.CreatedBy,
			account.UpdatedBy, account.DeletedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert account: %w", err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO user_accounts (
				id, account_id, user_id, permission, status, revoked_at, created_at, updated_at, created_by, updated_by
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		`,
			uuid.NewString(), account.ID, ownerUserID, "owner", "active", nil,
			account.CreatedAt, account.UpdatedAt, ownerUserID, ownerUserID,
		)
		if err != nil {
			return fmt.Errorf("failed to link user to account: %w", err)
		}

		if account.AccountType == "broker" {
			_, err := tx.Exec(ctx, `
				INSERT INTO investment_accounts (id, account_id, created_at, updated_at)
				VALUES ($1,$2,$3,$4)
			`, uuid.NewString(), account.ID, account.CreatedAt, account.UpdatedAt)
			if err != nil {
				return fmt.Errorf("failed to create investment account extension: %w", err)
			}
		}

		return nil
	})
}

func (r *AccountRepo) ListAccountsForUser(ctx context.Context, userID string) ([]entity.Account, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT a.id, a.client_id, a.name, a.account_number, a.color, a.account_type, a.currency, a.parent_account_id, a.status, a.closed_at,
		       a.created_at, a.updated_at, a.created_by, a.updated_by, a.deleted_at,
		       COALESCE(SUM(
		         CASE
		           WHEN t.type = 'income' AND t.account_id = a.id THEN t.amount
		           WHEN t.type = 'expense' AND t.account_id = a.id THEN -t.amount
		           WHEN t.type = 'transfer' AND t.to_account_id = a.id THEN COALESCE(t.to_amount, t.amount)
		           WHEN t.type = 'transfer' AND t.from_account_id = a.id THEN -COALESCE(t.from_amount, t.amount)
		           ELSE 0
		         END
		       ), 0)::text AS balance,
		       ia.id AS investment_account_id
		FROM accounts a
		JOIN user_accounts ua ON ua.account_id = a.id
		LEFT JOIN transactions t
		  ON t.deleted_at IS NULL AND t.status = 'posted'
		 AND (t.account_id = a.id OR t.from_account_id = a.id OR t.to_account_id = a.id)
		LEFT JOIN investment_accounts ia ON ia.account_id = a.id
		WHERE ua.user_id = $1 AND ua.status = 'active' AND a.deleted_at IS NULL
		GROUP BY a.id, ia.id
		ORDER BY a.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.Account
	for rows.Next() {
		var a entity.Account
		err := rows.Scan(
			&a.ID, &a.ClientID, &a.Name, &a.AccountNumber, &a.Color, &a.AccountType, &a.Currency,
			&a.ParentAccountID, &a.Status, &a.ClosedAt, &a.CreatedAt, &a.UpdatedAt, &a.CreatedBy,
			&a.UpdatedBy, &a.DeletedAt, &a.Balance, &a.InvestmentAccountID,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AccountRepo) GetAccountForUser(ctx context.Context, userID string, accountID string) (*entity.Account, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT a.id, a.client_id, a.name, a.account_number, a.color, a.account_type, a.currency, a.parent_account_id, a.status, a.closed_at,
		       a.created_at, a.updated_at, a.created_by, a.updated_by, a.deleted_at,
		       COALESCE(SUM(
		         CASE
		           WHEN t.type = 'income' AND t.account_id = a.id THEN t.amount
		           WHEN t.type = 'expense' AND t.account_id = a.id THEN -t.amount
		           WHEN t.type = 'transfer' AND t.to_account_id = a.id THEN COALESCE(t.to_amount, t.amount)
		           WHEN t.type = 'transfer' AND t.from_account_id = a.id THEN -COALESCE(t.from_amount, t.amount)
		           ELSE 0
		         END
		       ), 0)::text AS balance,
		       ia.id AS investment_account_id
		FROM accounts a
		JOIN user_accounts ua ON ua.account_id = a.id
		LEFT JOIN transactions t
		  ON t.deleted_at IS NULL AND t.status = 'posted'
		 AND (t.account_id = a.id OR t.from_account_id = a.id OR t.to_account_id = a.id)
		LEFT JOIN investment_accounts ia ON ia.account_id = a.id
		WHERE a.id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
		GROUP BY a.id, ia.id
	`, accountID, userID)

	var a entity.Account
	err = row.Scan(
		&a.ID, &a.ClientID, &a.Name, &a.AccountNumber, &a.Color, &a.AccountType, &a.Currency,
		&a.ParentAccountID, &a.Status, &a.ClosedAt, &a.CreatedAt, &a.UpdatedAt, &a.CreatedBy,
		&a.UpdatedBy, &a.DeletedAt, &a.Balance, &a.InvestmentAccountID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("account not found")
		}
		return nil, err
	}
	return &a, nil
}

func (r *AccountRepo) PatchAccount(ctx context.Context, actorUserID string, accountID string, patch entity.AccountPatch) (*entity.Account, error) {
	var out *entity.Account
	now := time.Now().UTC()

	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		if err := r.requireAccountOwner(ctx, tx, actorUserID, accountID); err != nil {
			return err
		}

		cur, err := r.getAccountInTx(ctx, tx, actorUserID, accountID)
		if err != nil {
			return err
		}

		name := cur.Name
		if patch.Name != nil {
			name = strings.TrimSpace(*patch.Name)
		}

		status := cur.Status
		closedAt := cur.ClosedAt
		if patch.Status != nil {
			status = *patch.Status
			if status == "closed" {
				if closedAt == nil {
					closedAt = &now
				}
			} else {
				closedAt = nil
			}
		}

		color := cur.Color
		if patch.Color != nil {
			color = patch.Color
		}

		_, err = tx.Exec(ctx, `
			UPDATE accounts
			SET name = $1, color = $2, status = $3, closed_at = $4, updated_at = $5, updated_by = $6
			WHERE id = $7 AND deleted_at IS NULL
		`, name, color, status, closedAt, now, actorUserID, accountID)
		if err != nil {
			return err
		}

		updated, err := r.getAccountInTx(ctx, tx, actorUserID, accountID)
		if err != nil {
			return err
		}
		out = updated
		return nil
	})

	return out, err
}

func (r *AccountRepo) DeleteAccount(ctx context.Context, actorUserID string, accountID string) error {
	now := time.Now().UTC()
	return r.db.WithTx(ctx, func(tx pgx.Tx) error {
		if err := r.requireAccountOwner(ctx, tx, actorUserID, accountID); err != nil {
			return err
		}

		_, err := tx.Exec(ctx, `
			UPDATE accounts
			SET deleted_at = $1, updated_at = $1, updated_by = $2
			WHERE id = $3 AND deleted_at IS NULL
		`, now, actorUserID, accountID)
		return err
	})
}

func (r *AccountRepo) HasRelatedTransferTransactionsForAccount(ctx context.Context, accountID string) (bool, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return false, err
	}

	var exists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM transactions
			WHERE deleted_at IS NULL AND type = 'transfer'
			  AND (from_account_id = $1 OR to_account_id = $1)
		)
	`, accountID).Scan(&exists)
	return exists, err
}

func (r *AccountRepo) ListAccountBalancesForUser(ctx context.Context, userID string) ([]entity.AccountBalance, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT a.id AS account_id, a.currency,
		       COALESCE(SUM(
		         CASE
		           WHEN t.type = 'income' AND t.account_id = a.id THEN t.amount
		           WHEN t.type = 'expense' AND t.account_id = a.id THEN -t.amount
		           WHEN t.type = 'transfer' AND t.to_account_id = a.id THEN COALESCE(t.to_amount, t.amount)
		           WHEN t.type = 'transfer' AND t.from_account_id = a.id THEN -COALESCE(t.from_amount, t.amount)
		           ELSE 0
		         END
		       ), 0)::text AS balance
		FROM accounts a
		JOIN user_accounts ua ON ua.account_id = a.id
		LEFT JOIN transactions t
		  ON t.deleted_at IS NULL AND t.status = 'posted'
		 AND (t.account_id = a.id OR t.from_account_id = a.id OR t.to_account_id = a.id)
		WHERE ua.user_id = $1 AND ua.status = 'active' AND a.deleted_at IS NULL
		GROUP BY a.id, a.currency
		ORDER BY a.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.AccountBalance
	for rows.Next() {
		var b entity.AccountBalance
		if err := rows.Scan(&b.AccountID, &b.Currency, &b.Balance); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *AccountRepo) ListAccountShares(ctx context.Context, actorUserID string, accountID string) ([]entity.AccountShare, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT ua.id, ua.account_id, ua.user_id, ua.permission, ua.status, ua.revoked_at,
		       ua.created_at, ua.updated_at, ua.created_by, ua.updated_by,
		       u.email, u.phone, u.display_name
		FROM user_accounts ua
		JOIN users u ON u.id = ua.user_id
		WHERE ua.account_id = $1
		ORDER BY (ua.permission = 'owner') DESC, ua.created_at ASC
	`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.AccountShare
	for rows.Next() {
		var it entity.AccountShare
		err := rows.Scan(
			&it.ID, &it.AccountID, &it.UserID, &it.Permission, &it.Status, &it.RevokedAt,
			&it.CreatedAt, &it.UpdatedAt, &it.CreatedBy, &it.UpdatedBy,
			&it.UserEmail, &it.UserPhone, &it.UserDisplayName,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *AccountRepo) UpsertAccountShare(ctx context.Context, actorUserID string, accountID string, targetUserID string, permission string) (*entity.AccountShare, error) {
	now := time.Now().UTC()
	var out *entity.AccountShare

	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		if err := r.requireAccountOwner(ctx, tx, actorUserID, accountID); err != nil {
			return err
		}

		uaID := uuid.NewString()
		row := tx.QueryRow(ctx, `
			INSERT INTO user_accounts (
				id, account_id, user_id, permission, status, revoked_at,
				created_at, updated_at, created_by, updated_by
			) VALUES ($1,$2,$3,$4,'active',NULL,$5,$5,$6,$6)
			ON CONFLICT (account_id, user_id) DO UPDATE
			SET permission = EXCLUDED.permission, status = 'active', revoked_at = NULL, updated_at = $5, updated_by = $6
			RETURNING id, account_id, user_id, permission, status, revoked_at, created_at, updated_at, created_by, updated_by
		`, uaID, accountID, targetUserID, permission, now, actorUserID)

		var it entity.AccountShare
		if err := row.Scan(
			&it.ID, &it.AccountID, &it.UserID, &it.Permission, &it.Status, &it.RevokedAt,
			&it.CreatedAt, &it.UpdatedAt, &it.CreatedBy, &it.UpdatedBy,
		); err != nil {
			return err
		}

		_ = tx.QueryRow(ctx, `SELECT email, phone, display_name FROM users WHERE id = $1`, targetUserID).Scan(&it.UserEmail, &it.UserPhone, &it.UserDisplayName)
		out = &it
		return nil
	})

	return out, err
}

func (r *AccountRepo) RevokeAccountShare(ctx context.Context, actorUserID string, accountID string, targetUserID string) error {
	now := time.Now().UTC()
	return r.db.WithTx(ctx, func(tx pgx.Tx) error {
		if err := r.requireAccountOwner(ctx, tx, actorUserID, accountID); err != nil {
			return err
		}

		_, err := tx.Exec(ctx, `
			UPDATE user_accounts
			SET status = 'revoked', revoked_at = $1, updated_at = $1, updated_by = $2
			WHERE account_id = $3 AND user_id = $4 AND permission != 'owner'
		`, now, actorUserID, accountID, targetUserID)
		return err
	})
}

func (r *AccountRepo) requireAccountOwner(ctx context.Context, tx pgx.Tx, userID string, accountID string) error {
	var one int
	err := tx.QueryRow(ctx, `
		SELECT 1 FROM user_accounts
		WHERE user_id = $1 AND account_id = $2 AND status = 'active' AND permission = 'owner'
	`, userID, accountID).Scan(&one)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("forbidden: account owner required")
		}
		return err
	}
	return nil
}

func (r *AccountRepo) getAccountInTx(ctx context.Context, tx pgx.Tx, userID, accountID string) (*entity.Account, error) {
	row := tx.QueryRow(ctx, `
		SELECT a.id, a.client_id, a.name, a.account_number, a.color, a.account_type, a.currency, a.parent_account_id, a.status, a.closed_at,
		       a.created_at, a.updated_at, a.created_by, a.updated_by, a.deleted_at,
		       COALESCE(SUM(
		         CASE
		           WHEN t.type = 'income' AND t.account_id = a.id THEN t.amount
		           WHEN t.type = 'expense' AND t.account_id = a.id THEN -t.amount
		           WHEN t.type = 'transfer' AND t.to_account_id = a.id THEN COALESCE(t.to_amount, t.amount)
		           WHEN t.type = 'transfer' AND t.from_account_id = a.id THEN -COALESCE(t.from_amount, t.amount)
		           ELSE 0
		         END
		       ), 0)::text AS balance,
		       ia.id AS investment_account_id
		FROM accounts a
		JOIN user_accounts ua ON ua.account_id = a.id
		LEFT JOIN transactions t
		  ON t.deleted_at IS NULL AND t.status = 'posted'
		 AND (t.account_id = a.id OR t.from_account_id = a.id OR t.to_account_id = a.id)
		LEFT JOIN investment_accounts ia ON ia.account_id = a.id
		WHERE a.id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
		GROUP BY a.id, ia.id
	`, accountID, userID)

	var a entity.Account
	err := row.Scan(
		&a.ID, &a.ClientID, &a.Name, &a.AccountNumber, &a.Color, &a.AccountType, &a.Currency,
		&a.ParentAccountID, &a.Status, &a.ClosedAt, &a.CreatedAt, &a.UpdatedAt, &a.CreatedBy,
		&a.UpdatedBy, &a.DeletedAt, &a.Balance, &a.InvestmentAccountID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("account not found")
		}
		return nil, err
	}
	return &a, nil
}
