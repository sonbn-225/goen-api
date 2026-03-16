package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type TransactionRepo struct {
	db *Postgres
}

// createTransactionTx inserts a transaction and its children using an existing dbtx.
// It mirrors CreateTransaction but does not begin/commit.
func createTransactionTx(ctx context.Context, dbtx pgx.Tx, userID string, tx domain.Transaction, lineItems []domain.TransactionLineItem, tagIDs []string) error {
	auditAt := time.Now().UTC()
	auditDiff := map[string]any{
		"type":        tx.Type,
		"amount":      tx.Amount,
		"from_amount": tx.FromAmount,
		"to_amount":   tx.ToAmount,
		"occurred_at": tx.OccurredAt.UTC().Format(time.RFC3339Nano),
	}
	// Permission checks
	switch tx.Type {
	case "expense", "income":
		if tx.AccountID == nil || strings.TrimSpace(*tx.AccountID) == "" {
			return apperrors.ErrAccountIDRequired
		}
		if err := requireAccountPermission(ctx, dbtx, userID, *tx.AccountID, true); err != nil {
			return err
		}
		if err := requireAccountActive(ctx, dbtx, *tx.AccountID); err != nil {
			return err
		}
	case "transfer":
		if tx.FromAccountID == nil || strings.TrimSpace(*tx.FromAccountID) == "" {
			return apperrors.ErrFromAccountIDRequired
		}
		if tx.ToAccountID == nil || strings.TrimSpace(*tx.ToAccountID) == "" {
			return apperrors.ErrToAccountIDRequired
		}
		if err := requireAccountPermission(ctx, dbtx, userID, *tx.FromAccountID, true); err != nil {
			return err
		}
		if err := requireAccountPermission(ctx, dbtx, userID, *tx.ToAccountID, true); err != nil {
			return err
		}
		if err := requireAccountActive(ctx, dbtx, *tx.FromAccountID); err != nil {
			return err
		}
		if err := requireAccountActive(ctx, dbtx, *tx.ToAccountID); err != nil {
			return err
		}

		fromCur, err := getAccountCurrency(ctx, dbtx, *tx.FromAccountID)
		if err != nil {
			return err
		}
		toCur, err := getAccountCurrency(ctx, dbtx, *tx.ToAccountID)
		if err != nil {
			return err
		}
		fx := !strings.EqualFold(strings.TrimSpace(fromCur), strings.TrimSpace(toCur))

		// Default amounts for same-currency transfers.
		if tx.FromAmount == nil {
			tx.FromAmount = &tx.Amount
		}
		if tx.ToAmount == nil {
			if fx {
				return apperrors.ErrFXAmountsRequired
			}
			tx.ToAmount = &tx.Amount
		}
		if fx {
			if tx.FromAmount == nil || tx.ToAmount == nil {
				return apperrors.ErrFXAmountsRequired
			}
		}

		// Auto-compute exchange_rate if omitted and amounts are provided.
		if tx.ExchangeRate == nil && tx.FromAmount != nil && tx.ToAmount != nil {
			rate, err := computeExchangeRate(*tx.FromAmount, *tx.ToAmount)
			if err != nil {
				return err
			}
			if rate != nil {
				tx.ExchangeRate = rate
			}
		}
	default:
		return apperrors.ErrTransactionInvalidType
	}

	_, err := dbtx.Exec(ctx, `
		INSERT INTO transactions (
			id, client_id, external_ref, type, occurred_at, amount, description,
			from_amount, to_amount,
			account_id, from_account_id, to_account_id, exchange_rate,
			status, created_at, updated_at, created_by, updated_by, deleted_at
		) VALUES ($1,$2,$3,$4,$5,$6::numeric,$7,$8::numeric,$9::numeric,$10,$11,$12,$13::numeric,$14,$15,$16,$17,$18,$19)
	`,
		tx.ID,
		tx.ClientID,
		tx.ExternalRef,
		tx.Type,
		tx.OccurredAt,
		tx.Amount,
		tx.Description,
		tx.FromAmount,
		tx.ToAmount,
		tx.AccountID,
		tx.FromAccountID,
		tx.ToAccountID,
		tx.ExchangeRate,
		tx.Status,
		tx.CreatedAt,
		tx.UpdatedAt,
		tx.CreatedBy,
		tx.UpdatedBy,
		tx.DeletedAt,
	)
	if err != nil {
		return err
	}

	for _, li := range lineItems {
		if li.ID == "" {
			li.ID = uuid.NewString()
		}
		if li.CategoryID != nil && strings.TrimSpace(*li.CategoryID) != "" {
			var ok bool
			err := dbtx.QueryRow(ctx, `
				SELECT EXISTS (
					SELECT 1
					FROM categories
					WHERE id = $1
					  AND deleted_at IS NULL
					  AND is_active = true
				)
			`, strings.TrimSpace(*li.CategoryID)).Scan(&ok)
			if err != nil {
				return err
			}
			if !ok {
				return apperrors.ErrCategoryIDInvalid
			}
		}
		_, err := dbtx.Exec(ctx, `
			INSERT INTO transaction_line_items (id, transaction_id, category_id, amount, note)
			VALUES ($1,$2,$3,$4::numeric,$5)
		`, li.ID, tx.ID, li.CategoryID, li.Amount, li.Note)
		if err != nil {
			return err
		}
	}

	if len(tagIDs) > 0 {
		var okCount int
		err := dbtx.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM tags
			WHERE user_id = $1 AND id = ANY($2::text[])
		`, userID, tagIDs).Scan(&okCount)
		if err != nil {
			return err
		}
		if okCount != len(tagIDs) {
			return apperrors.ErrTagIDsInvalid
		}

		_, err = dbtx.Exec(ctx, `
			INSERT INTO transaction_tags (transaction_id, tag_id, created_at)
			SELECT $1, unnest($2::text[]), $3
			ON CONFLICT DO NOTHING
		`, tx.ID, tagIDs, tx.CreatedAt)
		if err != nil {
			return err
		}
	}

	// Audit (UC-007)
	switch tx.Type {
	case "expense", "income":
		_ = insertAuditEvent(ctx, dbtx, *tx.AccountID, userID, "transaction.create", "transaction", tx.ID, auditAt, auditDiff)
	case "transfer":
		_ = insertAuditEvent(ctx, dbtx, *tx.FromAccountID, userID, "transaction.create", "transaction", tx.ID, auditAt, auditDiff)
		_ = insertAuditEvent(ctx, dbtx, *tx.ToAccountID, userID, "transaction.create", "transaction", tx.ID, auditAt, auditDiff)
	}

	return nil
}

func NewTransactionRepo(db *Postgres) *TransactionRepo {
	return &TransactionRepo{db: db}
}

func (r *TransactionRepo) CreateTransaction(ctx context.Context, userID string, tx domain.Transaction, lineItems []domain.TransactionLineItem, tagIDs []string) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	return withTx(ctx, pool, func(dbtx pgx.Tx) error {
		return createTransactionTx(ctx, dbtx, userID, tx, lineItems, tagIDs)
	})
}

func (r *TransactionRepo) GetTransaction(ctx context.Context, userID string, transactionID string) (*domain.Transaction, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT
			t.id,
			t.client_id,
			t.external_ref,
			t.type,
			t.occurred_at,
			to_char(t.occurred_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS occurred_date,
			t.amount::text,
			CASE WHEN t.from_amount IS NULL THEN NULL ELSE t.from_amount::text END,
			CASE WHEN t.to_amount IS NULL THEN NULL ELSE t.to_amount::text END,
			t.description,
			t.account_id,
			t.from_account_id,
			t.to_account_id,
			CASE WHEN t.exchange_rate IS NULL THEN NULL ELSE t.exchange_rate::text END,
			a.currency AS account_currency,
			fa.currency AS from_currency,
			ta.currency AS to_currency,
			t.status,
			t.created_at,
			t.updated_at,
			t.created_by,
			t.updated_by,
			t.deleted_at,
			COALESCE((SELECT array_agg(tt.tag_id ORDER BY tt.tag_id) FROM transaction_tags tt WHERE tt.transaction_id = t.id), '{}'::text[]) AS tag_ids
		FROM transactions t
		LEFT JOIN accounts a ON a.id = t.account_id
		LEFT JOIN accounts fa ON fa.id = t.from_account_id
		LEFT JOIN accounts ta ON ta.id = t.to_account_id
		WHERE t.id = $1 AND t.deleted_at IS NULL
		  AND (
			(t.type IN ('expense','income') AND EXISTS (
				SELECT 1 FROM user_accounts ua
				WHERE ua.user_id = $2 AND ua.account_id = t.account_id AND ua.status = 'active'
			))
			OR
			(t.type = 'transfer' AND EXISTS (
				SELECT 1 FROM user_accounts ua
				WHERE ua.user_id = $2 AND ua.account_id = t.from_account_id AND ua.status = 'active'
			) AND EXISTS (
				SELECT 1 FROM user_accounts ua
				WHERE ua.user_id = $2 AND ua.account_id = t.to_account_id AND ua.status = 'active'
			))
		  )
	`, transactionID, userID)

	var t domain.Transaction
	if err := row.Scan(
		&t.ID,
		&t.ClientID,
		&t.ExternalRef,
		&t.Type,
		&t.OccurredAt,
		&t.OccurredDate,
		&t.Amount,
		&t.FromAmount,
		&t.ToAmount,
		&t.Description,
		&t.AccountID,
		&t.FromAccountID,
		&t.ToAccountID,
		&t.ExchangeRate,
		&t.AccountCurrency,
		&t.FromCurrency,
		&t.ToCurrency,
		&t.Status,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.CreatedBy,
		&t.UpdatedBy,
		&t.DeletedAt,
		&t.TagIDs,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrTransactionNotFound
		}
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT id, category_id, amount::text, note
		FROM transaction_line_items
		WHERE transaction_id = $1
		ORDER BY id ASC
	`, t.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.TransactionLineItem, 0)
	for rows.Next() {
		var li domain.TransactionLineItem
		if err := rows.Scan(&li.ID, &li.CategoryID, &li.Amount, &li.Note); err != nil {
			return nil, err
		}
		items = append(items, li)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	t.LineItems = items
	return &t, nil
}

func (r *TransactionRepo) ListTransactions(ctx context.Context, userID string, filter domain.TransactionListFilter) ([]domain.Transaction, *string, error) {
	if r.db == nil {
		return nil, nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, nil, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	cursorTime, cursorID, err := decodeCursor(filter.Cursor)
	if err != nil {
		return nil, nil, apperrors.ErrInvalidCursor
	}

	args := []any{userID}
	whereExtra := ""
	if filter.From != nil {
		args = append(args, *filter.From)
		whereExtra += fmt.Sprintf(" AND t.occurred_at >= $%d", len(args))
	}
	if filter.To != nil {
		args = append(args, *filter.To)
		whereExtra += fmt.Sprintf(" AND t.occurred_at <= $%d", len(args))
	}
	if filter.AccountID != nil && strings.TrimSpace(*filter.AccountID) != "" {
		args = append(args, *filter.AccountID)
		idx := len(args)
		whereExtra += fmt.Sprintf(" AND (t.account_id = $%d OR t.from_account_id = $%d OR t.to_account_id = $%d)", idx, idx, idx)
	}
	if filter.CategoryID != nil && strings.TrimSpace(*filter.CategoryID) != "" {
		args = append(args, *filter.CategoryID)
		whereExtra += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM transaction_line_items tli WHERE tli.transaction_id = t.id AND tli.category_id = $%d)", len(args))
	}
	if filter.Type != nil && strings.TrimSpace(*filter.Type) != "" {
		args = append(args, *filter.Type)
		whereExtra += fmt.Sprintf(" AND t.type = $%d", len(args))
	}
	if filter.Search != nil && strings.TrimSpace(*filter.Search) != "" {
		args = append(args, "%"+*filter.Search+"%")
		whereExtra += fmt.Sprintf(" AND t.description ILIKE $%d", len(args))
	}
	if filter.ExternalRefFamily != nil && strings.TrimSpace(*filter.ExternalRefFamily) != "" {
		args = append(args, strings.TrimSpace(*filter.ExternalRefFamily))
		whereExtra += fmt.Sprintf(" AND t.external_ref LIKE ($%d || ':%%')", len(args))
	}
	if cursorTime != nil && cursorID != nil {
		args = append(args, *cursorTime)
		args = append(args, *cursorID)
		whereExtra += fmt.Sprintf(" AND (t.occurred_at, t.id) < ($%d, $%d)", len(args)-1, len(args))
	}

	args = append(args, limit+1)
	limitArg := len(args)

	rows, err := pool.Query(ctx, fmt.Sprintf(`
		SELECT
			t.id,
			t.client_id,
			t.external_ref,
			t.type,
			t.occurred_at,
			to_char(t.occurred_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS occurred_date,
			t.amount::text,
			CASE WHEN t.from_amount IS NULL THEN NULL ELSE t.from_amount::text END,
			CASE WHEN t.to_amount IS NULL THEN NULL ELSE t.to_amount::text END,
			t.description,
			t.account_id,
			t.from_account_id,
			t.to_account_id,
			CASE WHEN t.exchange_rate IS NULL THEN NULL ELSE t.exchange_rate::text END,
			a.currency AS account_currency,
			fa.currency AS from_currency,
			ta.currency AS to_currency,
			t.status,
			t.created_at,
			t.updated_at,
			t.created_by,
			t.updated_by,
			t.deleted_at,
			COALESCE((SELECT array_agg(tt.tag_id ORDER BY tt.tag_id) FROM transaction_tags tt WHERE tt.transaction_id = t.id), '{}'::text[]) AS tag_ids,
			COALESCE((
				SELECT array_agg(DISTINCT tli.category_id ORDER BY tli.category_id)
				FROM transaction_line_items tli
				WHERE tli.transaction_id = t.id AND tli.category_id IS NOT NULL
			), '{}'::text[]) AS category_ids
		FROM transactions t
		LEFT JOIN accounts a ON a.id = t.account_id
		LEFT JOIN accounts fa ON fa.id = t.from_account_id
		LEFT JOIN accounts ta ON ta.id = t.to_account_id
		WHERE t.deleted_at IS NULL
		  AND (
			(t.type IN ('expense','income') AND EXISTS (
				SELECT 1 FROM user_accounts ua
				WHERE ua.user_id = $1 AND ua.account_id = t.account_id AND ua.status = 'active'
			))
			OR
			(t.type = 'transfer' AND EXISTS (
				SELECT 1 FROM user_accounts ua
				WHERE ua.user_id = $1 AND ua.account_id = t.from_account_id AND ua.status = 'active'
			) AND EXISTS (
				SELECT 1 FROM user_accounts ua
				WHERE ua.user_id = $1 AND ua.account_id = t.to_account_id AND ua.status = 'active'
			))
		  )
		  %s
		ORDER BY t.occurred_at DESC, t.id DESC
		LIMIT $%d
	`, whereExtra, limitArg), args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	out := make([]domain.Transaction, 0, limit)
	for rows.Next() {
		var t domain.Transaction
		if err := rows.Scan(
			&t.ID,
			&t.ClientID,
			&t.ExternalRef,
			&t.Type,
			&t.OccurredAt,
			&t.OccurredDate,
			&t.Amount,
			&t.FromAmount,
			&t.ToAmount,
			&t.Description,
			&t.AccountID,
			&t.FromAccountID,
			&t.ToAccountID,
			&t.ExchangeRate,
			&t.AccountCurrency,
			&t.FromCurrency,
			&t.ToCurrency,
			&t.Status,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.CreatedBy,
			&t.UpdatedBy,
			&t.DeletedAt,
			&t.TagIDs,
			&t.CategoryIDs,
		); err != nil {
			return nil, nil, err
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	var nextCursor *string
	if len(out) > limit {
		last := out[limit-1]
		out = out[:limit]
		c := encodeCursor(last.OccurredAt, last.ID)
		nextCursor = &c
	}

	return out, nextCursor, nil
}

func (r *TransactionRepo) PatchTransaction(ctx context.Context, userID string, transactionID string, patch domain.TransactionPatch) (*domain.Transaction, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	var updated *domain.Transaction
	err = withTx(ctx, pool, func(dbtx pgx.Tx) error {
		auditAt := time.Now().UTC()
		// Fetch current (for permission check + existing values)
		cur, err := r.GetTransaction(ctx, userID, transactionID)
		if err != nil {
			return err
		}

		// require write permission on linked accounts
		switch cur.Type {
		case "expense", "income":
			if cur.AccountID == nil {
				return apperrors.ErrTransactionForbidden
			}
			if err := requireAccountPermission(ctx, dbtx, userID, *cur.AccountID, true); err != nil {
				return err
			}
		case "transfer":
			if cur.FromAccountID == nil || cur.ToAccountID == nil {
				return apperrors.ErrTransactionForbidden
			}
			if err := requireAccountPermission(ctx, dbtx, userID, *cur.FromAccountID, true); err != nil {
				return err
			}
			if err := requireAccountPermission(ctx, dbtx, userID, *cur.ToAccountID, true); err != nil {
				return err
			}
		}

		desc := cur.Description
		if patch.Description != nil {
			desc = patch.Description
		}

		amount := cur.Amount
		if patch.Amount != nil {
			amount = *patch.Amount
		}

		occurredAt := cur.OccurredAt
		if patch.OccurredAt != nil {
			occurredAt = *patch.OccurredAt
		}

		_, err = dbtx.Exec(ctx, `
			UPDATE transactions
			SET description = $1,
			    amount = $2::numeric,
			    occurred_at = $3,
			    updated_at = $4,
			    updated_by = $5
			WHERE id = $6 AND deleted_at IS NULL
		`, desc, amount, occurredAt, now, userID, transactionID)
		if err != nil {
			return err
		}

		// Update FIRST line item if amount or single category changes
		if patch.Amount != nil || len(patch.CategoryIDs) == 1 {
			var liID string
			err := dbtx.QueryRow(ctx, `SELECT id FROM transaction_line_items WHERE transaction_id = $1 ORDER BY id LIMIT 1`, transactionID).Scan(&liID)
			if err == nil {
				setClause := ""
				args := []any{liID}
				if patch.Amount != nil {
					setClause += "amount = $2"
					args = append(args, *patch.Amount)
				}
				if len(patch.CategoryIDs) == 1 {
					if setClause != "" {
						setClause += ", "
					}
					setClause += fmt.Sprintf("category_id = $%d", len(args)+1)
					args = append(args, patch.CategoryIDs[0])
				}

				if setClause != "" {
					_, err = dbtx.Exec(ctx, "UPDATE transaction_line_items SET "+setClause+" WHERE id = $1", args...)
					if err != nil {
						return err
					}
				}
			}
		}

		// Handle Tags
		if patch.TagIDs != nil {
			_, err = dbtx.Exec(ctx, `DELETE FROM transaction_tags WHERE transaction_id = $1`, transactionID)
			if err != nil {
				return err
			}
			if len(patch.TagIDs) > 0 {
				_, err = dbtx.Exec(ctx, `
					INSERT INTO transaction_tags (transaction_id, tag_id, created_at)
					SELECT $1, unnest($2::text[]), $3
				`, transactionID, patch.TagIDs, now)
				if err != nil {
					return err
				}
			}
		}

		// Handle LineItems: replace all
		if patch.LineItems != nil {
			_, err = dbtx.Exec(ctx, `DELETE FROM transaction_line_items WHERE transaction_id = $1`, transactionID)
			if err != nil {
				return err
			}
			for _, li := range *patch.LineItems {
				if li.ID == "" {
					li.ID = uuid.NewString()
				}
				_, err = dbtx.Exec(ctx, `
					INSERT INTO transaction_line_items (id, transaction_id, category_id, amount, note)
					VALUES ($1, $2, $3, $4::numeric, $5)
				`, li.ID, transactionID, li.CategoryID, li.Amount, li.Note)
				if err != nil {
					return err
				}
			}
		}

		// Handle GroupParticipants: delete unsettled + insert new
		if patch.GroupParticipants != nil {
			_, err = dbtx.Exec(ctx, `
				DELETE FROM group_expense_participants
				WHERE transaction_id = $1 AND user_id = $2 AND is_settled = false
			`, transactionID, userID)
			if err != nil {
				return err
			}
			for _, p := range *patch.GroupParticipants {
				if p.ID == "" {
					p.ID = uuid.NewString()
				}
				_, err = dbtx.Exec(ctx, `
					INSERT INTO group_expense_participants (
						id, user_id, transaction_id, participant_name,
						original_amount, share_amount,
						is_settled, settlement_transaction_id,
						created_at, updated_at
					) VALUES ($1,$2,$3,$4,$5::numeric,$6::numeric,$7,$8,$9,$10)
				`,
					p.ID, p.UserID, transactionID, p.ParticipantName,
					p.OriginalAmount, p.ShareAmount,
					p.IsSettled, p.SettlementTransactionID,
					p.CreatedAt, p.UpdatedAt,
				)
				if err != nil {
					return err
				}
			}
		}

		fetched, err := r.GetTransaction(ctx, userID, transactionID)
		if err != nil {
			return err
		}

		// Audit (UC-007)
		auditDiff := map[string]any{
			"description": patch.Description,
		}
		switch cur.Type {
		case "expense", "income":
			if cur.AccountID != nil {
				_ = insertAuditEvent(ctx, dbtx, *cur.AccountID, userID, "transaction.update", "transaction", cur.ID, auditAt, auditDiff)
			}
		case "transfer":
			if cur.FromAccountID != nil {
				_ = insertAuditEvent(ctx, dbtx, *cur.FromAccountID, userID, "transaction.update", "transaction", cur.ID, auditAt, auditDiff)
			}
			if cur.ToAccountID != nil {
				_ = insertAuditEvent(ctx, dbtx, *cur.ToAccountID, userID, "transaction.update", "transaction", cur.ID, auditAt, auditDiff)
			}
		}

		updated = fetched
		return nil
	})
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, apperrors.ErrTransactionPatchFailed
	}
	return updated, nil
}

func (r *TransactionRepo) DeleteTransaction(ctx context.Context, userID string, transactionID string) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()

	return withTx(ctx, pool, func(dbtx pgx.Tx) error {
		auditAt := time.Now().UTC()
		cur, err := r.GetTransaction(ctx, userID, transactionID)
		if err != nil {
			return err
		}

		switch cur.Type {
		case "expense", "income":
			if cur.AccountID == nil {
				return apperrors.ErrTransactionForbidden
			}
			if err := requireAccountPermission(ctx, dbtx, userID, *cur.AccountID, true); err != nil {
				return err
			}
		case "transfer":
			if cur.FromAccountID == nil || cur.ToAccountID == nil {
				return apperrors.ErrTransactionForbidden
			}
			if err := requireAccountPermission(ctx, dbtx, userID, *cur.FromAccountID, true); err != nil {
				return err
			}
			if err := requireAccountPermission(ctx, dbtx, userID, *cur.ToAccountID, true); err != nil {
				return err
			}
		}

		ct, err := dbtx.Exec(ctx, `
			UPDATE transactions
			SET deleted_at = $1,
			    updated_at = $1,
			    updated_by = $2
			WHERE id = $3 AND deleted_at IS NULL
		`, now, userID, transactionID)
		if err != nil {
			return err
		}
		if ct.RowsAffected() == 0 {
			return apperrors.ErrTransactionNotFound
		}

		// Audit (UC-007)
		switch cur.Type {
		case "expense", "income":
			if cur.AccountID != nil {
				_ = insertAuditEvent(ctx, dbtx, *cur.AccountID, userID, "transaction.delete", "transaction", cur.ID, auditAt, nil)
			}
		case "transfer":
			if cur.FromAccountID != nil {
				_ = insertAuditEvent(ctx, dbtx, *cur.FromAccountID, userID, "transaction.delete", "transaction", cur.ID, auditAt, nil)
			}
			if cur.ToAccountID != nil {
				_ = insertAuditEvent(ctx, dbtx, *cur.ToAccountID, userID, "transaction.delete", "transaction", cur.ID, auditAt, nil)
			}
		}
		return nil
	})
}

