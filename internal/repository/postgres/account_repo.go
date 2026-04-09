package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

type AccountRepo struct {
	BaseRepo
}

func NewAccountRepo(db *database.Postgres) *AccountRepo {
	return &AccountRepo{BaseRepo: *NewBaseRepo(db)}
}

func (r *AccountRepo) CreateAccountWithOwnerTx(ctx context.Context, tx pgx.Tx, account entity.Account, ownerUserID uuid.UUID) error {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	settingsJSON, _ := json.Marshal(account.Settings)

	_, err = q.Exec(ctx, `
		INSERT INTO accounts (
			id, name, account_number, account_type, currency, parent_account_id, status, settings, closed_at,
			created_at, updated_at, deleted_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`,
		account.ID, account.Name, account.AccountNumber,
		account.AccountType, account.Currency, account.ParentAccountID, account.Status,
		settingsJSON, account.ClosedAt, account.CreatedAt, account.UpdatedAt, account.DeletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert account: %w", err)
	}

	_, err = q.Exec(ctx, `
		INSERT INTO user_accounts (
			id, account_id, user_id, permission, status, revoked_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`,
		uuid.New(), account.ID, ownerUserID, "owner", "active", nil,
		account.CreatedAt, account.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to link user to account: %w", err)
	}

	return nil
}

func (r *AccountRepo) ListAccountsForUserTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]entity.Account, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, fmt.Sprintf(`
		SELECT %s, %s
		FROM accounts a
		JOIN user_accounts ua ON ua.account_id = a.id
		LEFT JOIN transactions t
		  ON t.deleted_at IS NULL AND t.status = 'posted'
		 AND (t.account_id = a.id OR t.from_account_id = a.id OR t.to_account_id = a.id)
		WHERE ua.user_id = $1 AND ua.status = 'active' AND a.deleted_at IS NULL
		GROUP BY a.id
		ORDER BY a.created_at DESC
	`, AccountColumnsSQL, AccountBalanceSQL), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.Account
	for rows.Next() {
		a, err := ScanAccount(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	return out, rows.Err()
}

func (r *AccountRepo) GetAccountForUserTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountID uuid.UUID) (*entity.Account, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	row := q.QueryRow(ctx, fmt.Sprintf(`
		SELECT %s, %s
		FROM accounts a
		JOIN user_accounts ua ON ua.account_id = a.id
		LEFT JOIN transactions t
		  ON t.deleted_at IS NULL AND t.status = 'posted'
		 AND (t.account_id = a.id OR t.from_account_id = a.id OR t.to_account_id = a.id)
		WHERE a.id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
		GROUP BY a.id
	`, AccountColumnsSQL, AccountBalanceSQL), accountID, userID)

	a, err := ScanAccount(row)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, errors.New("account not found")
	}
	return a, nil
}

func (r *AccountRepo) PatchAccountTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, accountID uuid.UUID, patch entity.AccountPatch) (*entity.Account, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	if err := r.requireAccountOwner(ctx, q, actorUserID, accountID); err != nil {
		return nil, err
	}

	cur, err := r.getAccountQueryer(ctx, q, actorUserID, accountID)
	if err != nil {
		return nil, err
	}

	now := utils.Now()
	name := cur.Name
	if patch.Name != nil {
		name = strings.TrimSpace(*patch.Name)
	}

	status := cur.Status
	closedAt := cur.ClosedAt
	if patch.Status != nil {
		status = *patch.Status
		if status == entity.AccountStatusArchived {
			if closedAt == nil {
				closedAt = &now
			}
		} else {
			closedAt = nil
		}
	}

	if patch.Settings != nil {
		if patch.Settings.Color != nil {
			cur.Settings.Color = patch.Settings.Color
		}
		if patch.Settings.Investment != nil {
			cur.Settings.Investment = patch.Settings.Investment
		}
		if patch.Settings.Savings != nil {
			cur.Settings.Savings = patch.Settings.Savings
		}
	}
	settingsJSON, _ := json.Marshal(cur.Settings)

	_, err = q.Exec(ctx, `
		UPDATE accounts
		SET name = $1, status = $2, closed_at = $3, settings = $4, updated_at = $5
		WHERE id = $6 AND deleted_at IS NULL
	`, name, status, closedAt, settingsJSON, now, accountID)
	if err != nil {
		return nil, err
	}

	return r.getAccountQueryer(ctx, q, actorUserID, accountID)
}

func (r *AccountRepo) DeleteAccountTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, accountID uuid.UUID) error {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	if err := r.requireAccountOwner(ctx, q, actorUserID, accountID); err != nil {
		return err
	}

	now := utils.Now()
	_, err = q.Exec(ctx, `
		UPDATE accounts
		SET deleted_at = $1, updated_at = $1, status = $2
		WHERE id = $3 AND deleted_at IS NULL
	`, now, entity.AccountStatusDeleted, accountID)
	return err
}

func (r *AccountRepo) HasRelatedTransferTransactionsForAccountTx(ctx context.Context, tx pgx.Tx, accountID uuid.UUID) (bool, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return false, err
	}

	var exists bool
	err = q.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM transactions
			WHERE deleted_at IS NULL AND type = 'transfer'
			  AND (from_account_id = $1 OR to_account_id = $1)
		)
	`, accountID).Scan(&exists)
	return exists, err
}

func (r *AccountRepo) ListAccountBalancesForUserTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]entity.AccountBalance, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, fmt.Sprintf(`
		SELECT a.id AS account_id, a.currency, %s
		FROM accounts a
		JOIN user_accounts ua ON ua.account_id = a.id
		LEFT JOIN transactions t
		  ON t.deleted_at IS NULL AND t.status = 'posted'
		 AND (t.account_id = a.id OR t.from_account_id = a.id OR t.to_account_id = a.id)
		WHERE ua.user_id = $1 AND ua.status = 'active' AND a.deleted_at IS NULL
		GROUP BY a.id, a.currency
		ORDER BY a.created_at DESC
	`, AccountBalanceSQL), userID)
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

func (r *AccountRepo) ListAccountSharesTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, accountID uuid.UUID) ([]entity.AccountShare, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, `
		SELECT ua.id, ua.account_id, ua.user_id, ua.permission, ua.status, ua.revoked_at,
		       ua.created_at, ua.updated_at,
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
			&it.CreatedAt, &it.UpdatedAt,
			&it.UserEmail, &it.UserPhone, &it.UserDisplayName,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *AccountRepo) UpsertAccountShareTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, accountID uuid.UUID, targetUserID uuid.UUID, permission string) (*entity.AccountShare, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	if err := r.requireAccountOwner(ctx, q, actorUserID, accountID); err != nil {
		return nil, err
	}

	now := utils.Now()
	uaID := uuid.New()
	row := q.QueryRow(ctx, `
		INSERT INTO user_accounts (
			id, account_id, user_id, permission, status, revoked_at,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,'active',NULL,$5,$5)
		ON CONFLICT (account_id, user_id) DO UPDATE
		SET permission = EXCLUDED.permission, status = 'active', revoked_at = NULL, updated_at = $5
		RETURNING id, account_id, user_id, permission, status, revoked_at, created_at, updated_at
	`, uaID, accountID, targetUserID, permission, now)

	var it entity.AccountShare
	if err := row.Scan(
		&it.ID, &it.AccountID, &it.UserID, &it.Permission, &it.Status, &it.RevokedAt,
		&it.CreatedAt, &it.UpdatedAt,
	); err != nil {
		return nil, err
	}

	_ = q.QueryRow(ctx, `SELECT email, phone, display_name FROM users WHERE id = $1`, targetUserID).Scan(&it.UserEmail, &it.UserPhone, &it.UserDisplayName)
	return &it, nil
}

func (r *AccountRepo) RevokeAccountShareTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, accountID uuid.UUID, targetUserID uuid.UUID) error {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	if err := r.requireAccountOwner(ctx, q, actorUserID, accountID); err != nil {
		return err
	}

	now := utils.Now()
	_, err = q.Exec(ctx, `
		UPDATE user_accounts
		SET status = 'revoked', revoked_at = $1, updated_at = $1
		WHERE account_id = $2 AND user_id = $3 AND permission != 'owner'
	`, now, accountID, targetUserID)
	return err
}


func (r *AccountRepo) requireAccountOwner(ctx context.Context, q database.Queryer, userID uuid.UUID, accountID uuid.UUID) error {
	var one int
	err := q.QueryRow(ctx, `
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

func (r *AccountRepo) getAccountQueryer(ctx context.Context, q database.Queryer, userID uuid.UUID, accountID uuid.UUID) (*entity.Account, error) {
	row := q.QueryRow(ctx, fmt.Sprintf(`
		SELECT %s, %s
		FROM accounts a
		JOIN user_accounts ua ON ua.account_id = a.id
		LEFT JOIN transactions t
		  ON t.deleted_at IS NULL AND t.status = 'posted'
		 AND (t.account_id = a.id OR t.from_account_id = a.id OR t.to_account_id = a.id)
		WHERE a.id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
		GROUP BY a.id
	`, AccountColumnsSQL, AccountBalanceSQL), accountID, userID)

	return ScanAccount(row)
}
