package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

// CreateTransactionTx is a shared helper to insert a transaction and its associated data
// within an existing database transaction (pgx.Tx).
func CreateTransactionTx(ctx context.Context, txConn pgx.Tx, userID uuid.UUID, tx entity.Transaction, lineItems []entity.TransactionLineItem, tagIDs []uuid.UUID) error {
	// 1. Permission Check
	switch tx.Type {
	case "expense", "income":
		if tx.AccountID == nil {
			return errors.New("account ID is required")
		}
		if err := requireAccountPermission(ctx, txConn, userID, *tx.AccountID); err != nil {
			return err
		}
	case "transfer":
		if tx.FromAccountID == nil || tx.ToAccountID == nil {
			return errors.New("from/to account IDs are required for transfer")
		}
		if err := requireAccountPermission(ctx, txConn, userID, *tx.FromAccountID); err != nil {
			return err
		}
		if err := requireAccountPermission(ctx, txConn, userID, *tx.ToAccountID); err != nil {
			return err
		}
	}

	// 2. Insert Transaction
	_, err := txConn.Exec(ctx, `
		INSERT INTO transactions (
			id, external_ref, type, occurred_at, amount,
			from_amount, to_amount, account_id, from_account_id, to_account_id,
			exchange_rate, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`,
		tx.ID, tx.ExternalRef, tx.Type, tx.OccurredAt, tx.Amount,
		tx.FromAmount, tx.ToAmount, tx.AccountID, tx.FromAccountID, tx.ToAccountID,
		tx.ExchangeRate, tx.Status, tx.CreatedAt, tx.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// 3. Insert Line Items
	for _, li := range lineItems {
		liID := li.ID
		if liID == uuid.Nil {
			liID = uuid.New()
		}
		_, err = txConn.Exec(ctx, `
			INSERT INTO transaction_line_items (id, transaction_id, category_id, amount, note)
			VALUES ($1, $2, $3, $4, $5)
		`, liID, tx.ID, li.CategoryID, li.Amount, li.Note)
		if err != nil {
			return err
		}

		// Line Item Tags
		if len(li.TagIDs) > 0 {
			for _, tid := range li.TagIDs {
				_, err = txConn.Exec(ctx, `
					INSERT INTO transaction_line_item_tags (line_item_id, tag_id, created_at, updated_at)
					VALUES ($1, $2, $3, $4)
				`, liID, tid, tx.CreatedAt, tx.UpdatedAt)
				if err != nil {
					return err
				}
			}
		}
	}

	// 4. Transaction Tags
	if len(tagIDs) > 0 {
		for _, tid := range tagIDs {
			_, err = txConn.Exec(ctx, `
				INSERT INTO transaction_tags (transaction_id, tag_id, created_at, updated_at)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT DO NOTHING
			`, tx.ID, tid, tx.CreatedAt, tx.UpdatedAt)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// requireAccountPermission is a shared helper to check if a user has permission to write to an account.
func requireAccountPermission(ctx context.Context, tx pgx.Tx, userID, accountID uuid.UUID) error {
	var perm string
	err := tx.QueryRow(ctx, "SELECT permission FROM user_accounts WHERE user_id = $1 AND account_id = $2 AND status = 'active'", userID, accountID).Scan(&perm)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("forbidden: account access required")
		}
		return err
	}
	if perm != "owner" && perm != "editor" {
		return errors.New("forbidden: insufficient permission on account")
	}
	return nil
}

// DeleteTransactionTx is a shared helper to soft-delete a transaction and its associated data
// within an existing database transaction (pgx.Tx).
func DeleteTransactionTx(ctx context.Context, tx pgx.Tx, userID, transactionID uuid.UUID) error {
	now := utils.Now()
	// 1. Verify owner (simple check for existence and user_accounts link)
	var id uuid.UUID
	err := tx.QueryRow(ctx, `
		SELECT t.id 
		FROM transactions t
		LEFT JOIN accounts a ON a.id = t.account_id
		LEFT JOIN user_accounts ua ON ua.account_id = a.id
		WHERE t.id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND t.deleted_at IS NULL
	`, transactionID, userID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("transaction not found or access denied")
		}
		return err
	}

	// 2. Soft Delete
	_, err = tx.Exec(ctx, "UPDATE transactions SET deleted_at = $1, updated_at = $1 WHERE id = $2", now, transactionID)
	return err
}
