package storage

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type TransactionRepo struct {
	db *Postgres
}

func NewTransactionRepo(db *Postgres) *TransactionRepo {
	return &TransactionRepo{db: db}
}

func (r *TransactionRepo) CreateTransaction(ctx context.Context, userID string, tx domain.Transaction, lineItems []domain.TransactionLineItem, tagIDs []string) error {
	if r.db == nil {
		return errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	return withTx(ctx, pool, func(dbtx pgx.Tx) error {
		auditAt := time.Now().UTC()
		auditDiff := map[string]any{
			"type":        tx.Type,
			"amount":      tx.Amount,
			"currency":    tx.Currency,
			"occurred_at": tx.OccurredAt.UTC().Format(time.RFC3339Nano),
		}
		// Permission checks
		switch tx.Type {
		case "expense", "income":
			if tx.AccountID == nil || strings.TrimSpace(*tx.AccountID) == "" {
				return errors.New("account_id is required")
			}
			if err := requireAccountPermission(ctx, dbtx, userID, *tx.AccountID, true); err != nil {
				return err
			}
		case "transfer":
			if tx.FromAccountID == nil || strings.TrimSpace(*tx.FromAccountID) == "" {
				return errors.New("from_account_id is required")
			}
			if tx.ToAccountID == nil || strings.TrimSpace(*tx.ToAccountID) == "" {
				return errors.New("to_account_id is required")
			}
			if err := requireAccountPermission(ctx, dbtx, userID, *tx.FromAccountID, true); err != nil {
				return err
			}
			if err := requireAccountPermission(ctx, dbtx, userID, *tx.ToAccountID, true); err != nil {
				return err
			}
		default:
			return errors.New("type is invalid")
		}

		_, err := dbtx.Exec(ctx, `
			INSERT INTO transactions (
				id, client_id, external_ref, type, occurred_at, amount, currency, description,
				account_id, from_account_id, to_account_id, exchange_rate,
				counterparty, notes, status, created_at, updated_at, created_by, updated_by, deleted_at
			) VALUES ($1,$2,$3,$4,$5,$6::numeric,$7,$8,$9,$10,$11,$12::numeric,$13,$14,$15,$16,$17,$18,$19,$20)
		`,
			tx.ID,
			tx.ClientID,
			tx.ExternalRef,
			tx.Type,
			tx.OccurredAt,
			tx.Amount,
			tx.Currency,
			tx.Description,
			tx.AccountID,
			tx.FromAccountID,
			tx.ToAccountID,
			tx.ExchangeRate,
			tx.Counterparty,
			tx.Notes,
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
				return errors.New("tag_ids contains invalid tag")
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
	})
}

func (r *TransactionRepo) GetTransaction(ctx context.Context, userID string, transactionID string) (*domain.Transaction, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
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
			t.currency,
			t.description,
			t.account_id,
			t.from_account_id,
			t.to_account_id,
			CASE WHEN t.exchange_rate IS NULL THEN NULL ELSE t.exchange_rate::text END,
			t.counterparty,
			t.notes,
			t.status,
			t.created_at,
			t.updated_at,
			t.created_by,
			t.updated_by,
			t.deleted_at,
			COALESCE((SELECT array_agg(tt.tag_id ORDER BY tt.tag_id) FROM transaction_tags tt WHERE tt.transaction_id = t.id), '{}'::text[]) AS tag_ids
		FROM transactions t
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
		&t.Currency,
		&t.Description,
		&t.AccountID,
		&t.FromAccountID,
		&t.ToAccountID,
		&t.ExchangeRate,
		&t.Counterparty,
		&t.Notes,
		&t.Status,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.CreatedBy,
		&t.UpdatedBy,
		&t.DeletedAt,
		&t.TagIDs,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTransactionNotFound
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
		return nil, nil, errors.New("database not ready")
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
		return nil, nil, errors.New("invalid cursor")
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
			t.currency,
			t.description,
			t.account_id,
			t.from_account_id,
			t.to_account_id,
			CASE WHEN t.exchange_rate IS NULL THEN NULL ELSE t.exchange_rate::text END,
			t.counterparty,
			t.notes,
			t.status,
			t.created_at,
			t.updated_at,
			t.created_by,
			t.updated_by,
			t.deleted_at,
			COALESCE((SELECT array_agg(tt.tag_id ORDER BY tt.tag_id) FROM transaction_tags tt WHERE tt.transaction_id = t.id), '{}'::text[]) AS tag_ids
		FROM transactions t
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
			&t.Currency,
			&t.Description,
			&t.AccountID,
			&t.FromAccountID,
			&t.ToAccountID,
			&t.ExchangeRate,
			&t.Counterparty,
			&t.Notes,
			&t.Status,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.CreatedBy,
			&t.UpdatedBy,
			&t.DeletedAt,
			&t.TagIDs,
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
		return nil, errors.New("database not ready")
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
				return domain.ErrTransactionForbidden
			}
			if err := requireAccountPermission(ctx, dbtx, userID, *cur.AccountID, true); err != nil {
				return err
			}
		case "transfer":
			if cur.FromAccountID == nil || cur.ToAccountID == nil {
				return domain.ErrTransactionForbidden
			}
			if err := requireAccountPermission(ctx, dbtx, userID, *cur.FromAccountID, true); err != nil {
				return err
			}
			if err := requireAccountPermission(ctx, dbtx, userID, *cur.ToAccountID, true); err != nil {
				return err
			}
		}

		desc := cur.Description
		notes := cur.Notes
		cp := cur.Counterparty
		if patch.Description != nil {
			desc = patch.Description
		}
		if patch.Notes != nil {
			notes = patch.Notes
		}
		if patch.Counterparty != nil {
			cp = patch.Counterparty
		}

		_, err = dbtx.Exec(ctx, `
			UPDATE transactions
			SET description = $1,
			    notes = $2,
			    counterparty = $3,
			    updated_at = $4,
			    updated_by = $5
			WHERE id = $6 AND deleted_at IS NULL
		`, desc, notes, cp, now, userID, transactionID)
		if err != nil {
			return err
		}

		fetched, err := r.GetTransaction(ctx, userID, transactionID)
		if err != nil {
			return err
		}

		// Audit (UC-007)
		auditDiff := map[string]any{
			"description":  patch.Description,
			"notes":        patch.Notes,
			"counterparty": patch.Counterparty,
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
		return nil, errors.New("patch failed")
	}
	return updated, nil
}

func (r *TransactionRepo) DeleteTransaction(ctx context.Context, userID string, transactionID string) error {
	if r.db == nil {
		return errors.New("database not ready")
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
				return domain.ErrTransactionForbidden
			}
			if err := requireAccountPermission(ctx, dbtx, userID, *cur.AccountID, true); err != nil {
				return err
			}
		case "transfer":
			if cur.FromAccountID == nil || cur.ToAccountID == nil {
				return domain.ErrTransactionForbidden
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
			return domain.ErrTransactionNotFound
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

func requireAccountPermission(ctx context.Context, dbtx pgx.Tx, userID, accountID string, requireWrite bool) error {
	var one int
	permClause := ""
	if requireWrite {
		permClause = " AND ua.permission IN ('owner','editor')"
	}
	err := dbtx.QueryRow(ctx, `
		SELECT 1
		FROM user_accounts ua
		WHERE ua.user_id = $1 AND ua.account_id = $2 AND ua.status = 'active'`+permClause+`
	`, userID, accountID).Scan(&one)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrTransactionForbidden
		}
		return err
	}
	return nil
}

func encodeCursor(occurredAt time.Time, id string) string {
	raw := fmt.Sprintf("%s|%s", occurredAt.UTC().Format(time.RFC3339Nano), id)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(cursor *string) (*time.Time, *string, error) {
	if cursor == nil {
		return nil, nil, nil
	}
	c := strings.TrimSpace(*cursor)
	if c == "" {
		return nil, nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(c)
	if err != nil {
		return nil, nil, err
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return nil, nil, errors.New("invalid cursor")
	}
	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return nil, nil, err
	}
	id := parts[1]
	return &ts, &id, nil
}
