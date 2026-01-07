package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/google/uuid"
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
		return errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	// Create account + owner link in one transaction.
	return withTx(ctx, pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO accounts (
				id, client_id, name, account_type, currency, parent_account_id, status, closed_at,
				created_at, updated_at, created_by, updated_by, deleted_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		`,
			account.ID,
			account.ClientID,
			account.Name,
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
		return err
	})
}

func (r *AccountRepo) ListAccountsForUser(ctx context.Context, userID string) ([]domain.Account, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT a.id, a.client_id, a.name, a.account_type, a.currency, a.parent_account_id, a.status, a.closed_at,
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
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT a.id, a.client_id, a.name, a.account_type, a.currency, a.parent_account_id, a.status, a.closed_at,
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
			return nil, domain.ErrAccountNotFound
		}
		return nil, err
	}
	return &a, nil
}

func withTx(ctx context.Context, pool interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}, fn func(tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
