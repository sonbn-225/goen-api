package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type AccountRepo struct {
	db *Postgres
}

func NewAccountRepo(db *Postgres) *AccountRepo {
	return &AccountRepo{db: db}
}

func (r *AccountRepo) CreateAccountWithOwner(ctx context.Context, account domain.Account, ownerUserID string) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	// Create account + owner link in one transaction.
	return withTx(ctx, pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO accounts (
				id, client_id, name, account_number, color, account_type, currency, parent_account_id, status, closed_at,
				created_at, updated_at, created_by, updated_by, deleted_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		`,
			account.ID,
			account.ClientID,
			account.Name,
			account.AccountNumber,
			account.Color,
			account.AccountType,
			account.Currency,
			account.ParentAccountID,
			account.Status,
			account.ClosedAt,
			account.CreatedAt,
			account.UpdatedAt,
			account.CreatedBy,
			account.UpdatedBy,
			account.DeletedAt,
		)
		if err != nil {
			return err
		}

		uaID := uuid.NewString()
		_, err = tx.Exec(ctx, `
			INSERT INTO user_accounts (
				id, account_id, user_id, permission, status, revoked_at, created_at, updated_at, created_by, updated_by
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		`,
			uaID,
			account.ID,
			ownerUserID,
			"owner",
			"active",
			nil,
			account.CreatedAt,
			account.UpdatedAt,
			ownerUserID,
			ownerUserID,
		)
		if err != nil {
			return err
		}

		// Broker accounts must always have an investment_accounts extension row.
		if account.AccountType == "broker" {
			_, err := tx.Exec(ctx, `
				INSERT INTO investment_accounts (
					id, account_id, created_at, updated_at
				) VALUES ($1,$2,$3,$4)
			`,
				uuid.NewString(),
				account.ID,
				account.CreatedAt,
				account.UpdatedAt,
			)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *AccountRepo) ListAccountsForUser(ctx context.Context, userID string) ([]domain.Account, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT a.id, a.client_id, a.name, a.account_number, a.color, a.account_type, a.currency, a.parent_account_id, a.status, a.closed_at,
		       a.created_at, a.updated_at, a.created_by, a.updated_by, a.deleted_at
		FROM accounts a
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE ua.user_id = $1 AND ua.status = 'active' AND a.deleted_at IS NULL
		ORDER BY a.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.Account{}
	for rows.Next() {
		var a domain.Account
		err := rows.Scan(
			&a.ID,
			&a.ClientID,
			&a.Name,
			&a.AccountNumber,
			&a.Color,
			&a.AccountType,
			&a.Currency,
			&a.ParentAccountID,
			&a.Status,
			&a.ClosedAt,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.CreatedBy,
			&a.UpdatedBy,
			&a.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AccountRepo) GetAccountForUser(ctx context.Context, userID string, accountID string) (*domain.Account, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT a.id, a.client_id, a.name, a.account_number, a.color, a.account_type, a.currency, a.parent_account_id, a.status, a.closed_at,
		       a.created_at, a.updated_at, a.created_by, a.updated_by, a.deleted_at
		FROM accounts a
		JOIN user_accounts ua ON ua.account_id = a.id
		WHERE a.id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
	`, accountID, userID)

	var a domain.Account
	err = row.Scan(
		&a.ID,
		&a.ClientID,
		&a.Name,
		&a.AccountNumber,
		&a.Color,
		&a.AccountType,
		&a.Currency,
		&a.ParentAccountID,
		&a.Status,
		&a.ClosedAt,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.CreatedBy,
		&a.UpdatedBy,
		&a.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrAccountNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *AccountRepo) PatchAccount(ctx context.Context, actorUserID string, accountID string, patch domain.AccountPatch) (*domain.Account, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	if strings.TrimSpace(accountID) == "" {
		return nil, apperrors.ErrAccountInvalidInput
	}

	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	var out *domain.Account

	err = withTx(ctx, pool, func(tx pgx.Tx) error {
		if err := requireAccountOwnerForAccount(ctx, tx, actorUserID, accountID); err != nil {
			return err
		}

		cur, err := r.GetAccountForUser(ctx, actorUserID, accountID)
		if err != nil {
			return err
		}

		name := cur.Name
		if patch.Name != nil {
			name = strings.TrimSpace(*patch.Name)
			if name == "" {
				return apperrors.ErrAccountInvalidInput
			}
		}

		status := cur.Status
		closedAt := cur.ClosedAt
		color := cur.Color
		if patch.Status != nil {
			s := strings.TrimSpace(*patch.Status)
			if s != "active" && s != "closed" {
				return apperrors.ErrAccountInvalidInput
			}
			status = s
			if status == "closed" {
				if closedAt == nil {
					closedAt = &now
				}
			} else {
				closedAt = nil
			}
		}
		if patch.Color != nil {
			c := strings.TrimSpace(*patch.Color)
			if c == "" {
				color = nil
			} else {
				lc := strings.ToLower(c)
				color = &lc
			}
		}

		ct, err := tx.Exec(ctx, `
			UPDATE accounts
			SET name = $1,
			    color = $2,
			    status = $3,
			    closed_at = $4,
			    updated_at = $5,
			    updated_by = $6
			WHERE id = $7 AND deleted_at IS NULL
		`, name, color, status, closedAt, now, actorUserID, accountID)
		if err != nil {
			return err
		}
		if ct.RowsAffected() == 0 {
			return apperrors.ErrAccountNotFound
		}

		updated, err := r.GetAccountForUser(ctx, actorUserID, accountID)
		if err != nil {
			return err
		}
		out = updated
		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (r *AccountRepo) DeleteAccount(ctx context.Context, actorUserID string, accountID string) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	if strings.TrimSpace(accountID) == "" {
		return apperrors.ErrAccountInvalidInput
	}

	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	return withTx(ctx, pool, func(tx pgx.Tx) error {
		if err := requireAccountOwnerForAccount(ctx, tx, actorUserID, accountID); err != nil {
			return err
		}

		ct, err := tx.Exec(ctx, `
			UPDATE accounts
			SET deleted_at = $1,
			    updated_at = $1,
			    updated_by = $2
			WHERE id = $3 AND deleted_at IS NULL
		`, now, actorUserID, accountID)
		if err != nil {
			return err
		}
		if ct.RowsAffected() == 0 {
			return apperrors.ErrAccountNotFound
		}
		return nil
	})
}

func (r *AccountRepo) ListAccountBalancesForUser(ctx context.Context, userID string) ([]domain.AccountBalance, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT a.id AS account_id,
		       a.currency,
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
		  ON t.deleted_at IS NULL
		 AND t.status = 'posted'
		 AND (
		   t.account_id = a.id
		   OR t.from_account_id = a.id
		   OR t.to_account_id = a.id
		 )
		WHERE ua.user_id = $1 AND ua.status = 'active' AND a.deleted_at IS NULL
		GROUP BY a.id, a.currency
		ORDER BY a.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.AccountBalance, 0)
	for rows.Next() {
		var b domain.AccountBalance
		if err := rows.Scan(&b.AccountID, &b.Currency, &b.Balance); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *AccountRepo) ListAccountShares(ctx context.Context, actorUserID string, accountID string) ([]domain.AccountShare, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	// require owner
	err = withTx(ctx, pool, func(tx pgx.Tx) error {
		return requireAccountOwner(ctx, tx, actorUserID, accountID)
	})
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

	items := make([]domain.AccountShare, 0)
	for rows.Next() {
		var it domain.AccountShare
		if err := rows.Scan(
			&it.ID,
			&it.AccountID,
			&it.UserID,
			&it.Permission,
			&it.Status,
			&it.RevokedAt,
			&it.CreatedAt,
			&it.UpdatedAt,
			&it.CreatedBy,
			&it.UpdatedBy,
			&it.UserEmail,
			&it.UserPhone,
			&it.UserDisplayName,
		); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *AccountRepo) UpsertAccountShare(ctx context.Context, actorUserID string, accountID string, targetUserID string, permission string) (*domain.AccountShare, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	permission = strings.TrimSpace(permission)
	if permission != "viewer" && permission != "editor" {
		return nil, apperrors.ErrAccountShareInvalidInput
	}
	if strings.TrimSpace(accountID) == "" || strings.TrimSpace(targetUserID) == "" {
		return nil, apperrors.ErrAccountShareInvalidInput
	}
	if targetUserID == actorUserID {
		return nil, apperrors.ErrAccountShareInvalidInput
	}

	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	var out *domain.AccountShare

	err = withTx(ctx, pool, func(tx pgx.Tx) error {
		if err := requireAccountOwner(ctx, tx, actorUserID, accountID); err != nil {
			return err
		}

		// prevent modifying owner
		var existingPerm string
		err := tx.QueryRow(ctx, `
			SELECT permission
			FROM user_accounts
			WHERE account_id = $1 AND user_id = $2
		`, accountID, targetUserID).Scan(&existingPerm)
		if err == nil {
			if existingPerm == "owner" {
				return apperrors.ErrAccountShareInvalidInput
			}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		uaID := uuid.NewString()
		row := tx.QueryRow(ctx, `
			INSERT INTO user_accounts (
				id, account_id, user_id, permission, status, revoked_at,
				created_at, updated_at, created_by, updated_by
			) VALUES ($1,$2,$3,$4,'active',NULL,$5,$5,$6,$6)
			ON CONFLICT (account_id, user_id) DO UPDATE
			SET permission = EXCLUDED.permission,
			    status = 'active',
			    revoked_at = NULL,
			    updated_at = $5,
			    updated_by = $6
			RETURNING id, account_id, user_id, permission, status, revoked_at, created_at, updated_at, created_by, updated_by
		`, uaID, accountID, targetUserID, permission, now, actorUserID)

		var it domain.AccountShare
		if err := row.Scan(
			&it.ID,
			&it.AccountID,
			&it.UserID,
			&it.Permission,
			&it.Status,
			&it.RevokedAt,
			&it.CreatedAt,
			&it.UpdatedAt,
			&it.CreatedBy,
			&it.UpdatedBy,
		); err != nil {
			return err
		}

		// hydrate user fields
		_ = tx.QueryRow(ctx, `SELECT email, phone, display_name FROM users WHERE id = $1`, targetUserID).Scan(&it.UserEmail, &it.UserPhone, &it.UserDisplayName)
		out = &it

		_ = insertAuditEvent(ctx, tx, accountID, actorUserID, "user_account.upsert", "user_account", it.ID, now, map[string]any{
			"target_user_id": targetUserID,
			"permission":     permission,
			"status":         "active",
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (r *AccountRepo) RevokeAccountShare(ctx context.Context, actorUserID string, accountID string, targetUserID string) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	if strings.TrimSpace(accountID) == "" || strings.TrimSpace(targetUserID) == "" {
		return apperrors.ErrAccountShareInvalidInput
	}
	if targetUserID == actorUserID {
		return apperrors.ErrAccountShareInvalidInput
	}

	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	return withTx(ctx, pool, func(tx pgx.Tx) error {
		if err := requireAccountOwner(ctx, tx, actorUserID, accountID); err != nil {
			return err
		}

		var uaID string
		var perm string
		err := tx.QueryRow(ctx, `
			SELECT id, permission
			FROM user_accounts
			WHERE account_id = $1 AND user_id = $2
		`, accountID, targetUserID).Scan(&uaID, &perm)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if perm == "owner" {
			return apperrors.ErrAccountShareInvalidInput
		}

		_, err = tx.Exec(ctx, `
			UPDATE user_accounts
			SET status = 'revoked',
			    revoked_at = $1,
			    updated_at = $1,
			    updated_by = $2
			WHERE id = $3
		`, now, actorUserID, uaID)
		if err != nil {
			return err
		}

		_ = insertAuditEvent(ctx, tx, accountID, actorUserID, "user_account.revoke", "user_account", uaID, now, map[string]any{
			"target_user_id": targetUserID,
			"status":         "revoked",
		})

		return nil
	})
}

func requireAccountOwner(ctx context.Context, tx pgx.Tx, actorUserID string, accountID string) error {
	var one int
	err := tx.QueryRow(ctx, `
		SELECT 1
		FROM user_accounts ua
		WHERE ua.user_id = $1 AND ua.account_id = $2 AND ua.status = 'active' AND ua.permission = 'owner'
	`, actorUserID, accountID).Scan(&one)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrAccountShareForbidden
		}
		return err
	}
	return nil
}

