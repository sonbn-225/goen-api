package postgres

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

type TransactionRepo struct {
	db *database.Postgres
}

func NewTransactionRepo(db *database.Postgres) *TransactionRepo {
	return &TransactionRepo{db: db}
}


func (r *TransactionRepo) CreateTransactionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, txEntity entity.Transaction, lineItems []entity.TransactionLineItem, tagIDs []uuid.UUID) error {
	var q database.Queryer = tx
	if tx == nil {
		pool, err := r.db.Pool(ctx)
		if err != nil {
			return err
		}
		q = pool
	}
	return CreateTransactionTx(ctx, q, userID, txEntity, lineItems, tagIDs)
}

func (r *TransactionRepo) GetTransaction(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*entity.Transaction, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}
	return r.getTransaction(ctx, pool, userID, id)
}

func (r *TransactionRepo) GetTransactionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, id uuid.UUID) (*entity.Transaction, error) {
	q, err := r.db.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}
	return r.getTransaction(ctx, q, userID, id)
}

func (r *TransactionRepo) getTransaction(ctx context.Context, q database.Queryer, userID uuid.UUID, id uuid.UUID) (*entity.Transaction, error) {
	row := q.QueryRow(ctx, `
		SELECT
			t.id, t.external_ref, t.type, t.occurred_at,
			to_char(t.occurred_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS occurred_date,
			t.amount::text, t.from_amount::text, t.to_amount::text,
			(SELECT li.note FROM transaction_line_items li WHERE li.transaction_id = t.id ORDER BY li.id LIMIT 1) AS description,
			t.account_id, a.name AS account_name, t.from_account_id, t.to_account_id, t.exchange_rate::text,
			a.currency AS account_currency, fa.currency AS from_currency, ta.currency AS to_currency,
			t.status, t.created_at, t.updated_at, t.deleted_at,
			COALESCE((SELECT array_agg(tt.tag_id ORDER BY tt.tag_id) FROM transaction_tags tt WHERE tt.transaction_id = t.id), '{}'::uuid[]) AS tag_ids
		FROM transactions t
		LEFT JOIN accounts a ON a.id = t.account_id
		LEFT JOIN accounts fa ON fa.id = t.from_account_id
		LEFT JOIN accounts ta ON ta.id = t.to_account_id
		WHERE t.id = $1 AND t.deleted_at IS NULL
		  AND (
			(t.type IN ('expense','income') AND EXISTS (SELECT 1 FROM user_accounts ua WHERE ua.user_id = $2 AND ua.account_id = t.account_id AND ua.status = 'active'))
			OR
			(t.type = 'transfer' AND EXISTS (SELECT 1 FROM user_accounts ua WHERE ua.user_id = $2 AND ua.account_id = t.from_account_id AND ua.status = 'active')
			                 AND EXISTS (SELECT 1 FROM user_accounts ua WHERE ua.user_id = $2 AND ua.account_id = t.to_account_id AND ua.status = 'active'))
		  )
	`, id, userID)

	var t entity.Transaction
	var catNames, catColors, tagNames, tagColors []string
	err := row.Scan(
		&t.ID, &t.ExternalRef, &t.Type, &t.OccurredAt, &t.OccurredDate,
		&t.Amount, &t.FromAmount, &t.ToAmount, &t.Description, &t.AccountID, &t.AccountName, &t.FromAccountID, &t.ToAccountID,
		&t.ExchangeRate, &t.AccountCurrency, &t.FromCurrency, &t.ToCurrency, &t.Status,
		&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt, &t.TagIDs,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}

	// Enrichment
	_ = q.QueryRow(ctx, `
		SELECT 
			COALESCE(array_agg(DISTINCT c.key), '{}'::text[]),
			COALESCE(array_agg(DISTINCT c.color), '{}'::text[]),
			COALESCE(array_agg(DISTINCT tg.name_vi), '{}'::text[]),
			COALESCE(array_agg(DISTINCT tg.color), '{}'::text[])
		FROM transaction_line_items li
		LEFT JOIN categories c ON c.id = li.category_id
		LEFT JOIN transaction_tags tt ON tt.transaction_id = li.transaction_id
		LEFT JOIN tags tg ON tg.id = tt.tag_id
		WHERE li.transaction_id = $1
	`, t.ID).Scan(&catNames, &catColors, &tagNames, &tagColors)

	t.CategoryNames = catNames
	t.CategoryColors = catColors
	t.TagNames = tagNames
	t.TagColors = tagColors

	// Line Items
	rows, err := q.Query(ctx, `
		SELECT li.id, li.category_id, li.amount::text, li.note,
		       COALESCE((SELECT array_agg(tlit.tag_id ORDER BY tlit.tag_id) FROM transaction_line_item_tags tlit WHERE tlit.line_item_id = li.id), '{}'::uuid[])
		FROM transaction_line_items li
		WHERE li.transaction_id = $1 ORDER BY li.id ASC
	`, t.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var li entity.TransactionLineItem
		if err := rows.Scan(&li.ID, &li.CategoryID, &li.Amount, &li.Note, &li.TagIDs); err != nil {
			return nil, err
		}
		t.LineItems = append(t.LineItems, li)
	}

	return &t, nil
}

func (r *TransactionRepo) ListTransactions(ctx context.Context, userID uuid.UUID, filter entity.TransactionListFilter) ([]entity.Transaction, *string, int, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	args := []any{userID}
	where := ""

	if filter.From != nil {
		args = append(args, *filter.From)
		where += fmt.Sprintf(" AND t.occurred_at >= $%d", len(args))
	}
	if filter.To != nil {
		args = append(args, *filter.To)
		where += fmt.Sprintf(" AND t.occurred_at <= $%d", len(args))
	}
	if filter.AccountID != nil {
		args = append(args, *filter.AccountID)
		where += fmt.Sprintf(" AND (t.account_id = $%d OR t.from_account_id = $%d OR t.to_account_id = $%d)", len(args), len(args), len(args))
	}
	if filter.CategoryID != nil {
		args = append(args, *filter.CategoryID)
		where += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM transaction_line_items tli WHERE tli.transaction_id = t.id AND tli.category_id = $%d)", len(args))
	}
	if filter.Type != nil {
		args = append(args, *filter.Type)
		where += fmt.Sprintf(" AND t.type = $%d", len(args))
	}
	if filter.Search != nil {
		args = append(args, "%"+*filter.Search+"%")
		where += fmt.Sprintf(" AND (t.external_ref ILIKE $%d OR EXISTS (SELECT 1 FROM transaction_line_items li WHERE li.transaction_id = t.id AND li.note ILIKE $%d))", len(args), len(args))
	}

	countSQL := fmt.Sprintf(`
		SELECT COUNT(*) FROM transactions t
		WHERE t.deleted_at IS NULL %s
		  AND (
			(t.type IN ('expense','income') AND EXISTS (SELECT 1 FROM user_accounts ua WHERE ua.user_id = $1 AND ua.account_id = t.account_id AND ua.status = 'active'))
			OR
			(t.type = 'transfer' AND EXISTS (SELECT 1 FROM user_accounts ua WHERE ua.user_id = $1 AND ua.account_id = t.from_account_id AND ua.status = 'active')
			                 AND EXISTS (SELECT 1 FROM user_accounts ua WHERE ua.user_id = $1 AND ua.account_id = t.to_account_id AND ua.status = 'active'))
		  )
	`, where)

	var total int
	if err := pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, nil, 0, err
	}

	pagination := ""
	if filter.Cursor != nil {
		tAt, id, err := decodeCursor(*filter.Cursor)
		if err == nil {
			args = append(args, tAt, id)
			where += fmt.Sprintf(" AND (t.occurred_at, t.id) < ($%d, $%d)", len(args)-1, len(args))
		}
	} else if filter.Page > 0 {
		offset := (filter.Page - 1) * limit
		pagination = fmt.Sprintf(" OFFSET %d", offset)
	}

	querySQL := fmt.Sprintf(`
		SELECT
			t.id, t.external_ref, t.type, t.occurred_at,
			to_char(t.occurred_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS occurred_date,
			t.amount::text, t.from_amount::text, t.to_amount::text,
			(SELECT li.note FROM transaction_line_items li WHERE li.transaction_id = t.id ORDER BY li.id LIMIT 1) AS description,
			t.account_id, a.name AS account_name, t.from_account_id, t.to_account_id, t.exchange_rate::text,
			a.currency AS account_currency, fa.currency AS from_currency, ta.currency AS to_currency,
			t.status, t.created_at, t.updated_at, t.deleted_at,
			COALESCE((SELECT array_agg(tt.tag_id ORDER BY tt.tag_id) FROM transaction_tags tt WHERE tt.transaction_id = t.id), '{}'::uuid[]) AS tag_ids,
			COALESCE((SELECT array_agg(DISTINCT tli.category_id ORDER BY tli.category_id) FROM transaction_line_items tli WHERE tli.transaction_id = t.id AND tli.category_id IS NOT NULL), '{}'::uuid[]) AS category_ids
		FROM transactions t
		LEFT JOIN accounts a ON a.id = t.account_id
		LEFT JOIN accounts fa ON fa.id = t.from_account_id
		LEFT JOIN accounts ta ON ta.id = t.to_account_id
		WHERE t.deleted_at IS NULL %s
		  AND (
			(t.type IN ('expense','income') AND EXISTS (SELECT 1 FROM user_accounts ua WHERE ua.user_id = $1 AND ua.account_id = t.account_id AND ua.status = 'active'))
			OR
			(t.type = 'transfer' AND EXISTS (SELECT 1 FROM user_accounts ua WHERE ua.user_id = $1 AND ua.account_id = t.from_account_id AND ua.status = 'active')
			                 AND EXISTS (SELECT 1 FROM user_accounts ua WHERE ua.user_id = $1 AND ua.account_id = t.to_account_id AND ua.status = 'active'))
		  )
		ORDER BY t.occurred_at DESC, t.id DESC
		LIMIT %d %s
	`, where, limit+1, pagination)

	rows, err := pool.Query(ctx, querySQL, args...)
	if err != nil {
		return nil, nil, 0, err
	}
	defer rows.Close()

	var results []entity.Transaction
	for rows.Next() {
		var t entity.Transaction
		err := rows.Scan(
			&t.ID, &t.ExternalRef, &t.Type, &t.OccurredAt, &t.OccurredDate,
			&t.Amount, &t.FromAmount, &t.ToAmount, &t.Description, &t.AccountID, &t.AccountName, &t.FromAccountID, &t.ToAccountID,
			&t.ExchangeRate, &t.AccountCurrency, &t.FromCurrency, &t.ToCurrency, &t.Status,
			&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt, &t.TagIDs, &t.CategoryIDs,
		)
		if err != nil {
			return nil, nil, 0, err
		}

		// Quick enrichment join per item (or could use a larger join above, but array_agg per item is safer for complex many-to-many)
		_ = pool.QueryRow(ctx, `
			SELECT 
				COALESCE(array_agg(DISTINCT c.key), '{}'::text[]),
				COALESCE(array_agg(DISTINCT c.color), '{}'::text[]),
				COALESCE(array_agg(DISTINCT tg.name_vi), '{}'::text[]),
				COALESCE(array_agg(DISTINCT tg.color), '{}'::text[])
			FROM transaction_line_items li
			LEFT JOIN categories c ON c.id = li.category_id
			LEFT JOIN transaction_tags tt ON tt.transaction_id = li.transaction_id
			LEFT JOIN tags tg ON tg.id = tt.tag_id
			WHERE li.transaction_id = $1
		`, t.ID).Scan(&t.CategoryNames, &t.CategoryColors, &t.TagNames, &t.TagColors)

		results = append(results, t)
	}

	var nextCursor *string
	if len(results) > limit {
		last := results[limit-1]
		c := encodeCursor(last.OccurredAt, last.ID)
		nextCursor = &c
		results = results[:limit]
	}

	return results, nextCursor, total, nil
}

func (r *TransactionRepo) PatchTransactionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, transactionID uuid.UUID, patch entity.TransactionPatch) (*entity.Transaction, error) {
	q, err := r.db.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}
	return r.patchTransactionTx(ctx, q, userID, transactionID, patch)
}

func (r *TransactionRepo) patchTransactionTx(ctx context.Context, q database.Queryer, userID uuid.UUID, transactionID uuid.UUID, patch entity.TransactionPatch) (*entity.Transaction, error) {
	now := utils.Now()

	// Verify ownership/permission
	if _, err := r.getTransaction(ctx, q, userID, transactionID); err != nil {
		return nil, err
	}

	// 1. Basic SQL update for the transactions table
	set := []string{"updated_at = $1"}
	args := []any{now}

	if patch.Amount != nil {
		args = append(args, *patch.Amount)
		set = append(set, fmt.Sprintf("amount = $%d", len(args)))
	}
	if patch.OccurredAt != nil {
		args = append(args, *patch.OccurredAt)
		set = append(set, fmt.Sprintf("occurred_at = $%d", len(args)))
	}
	if patch.Status != nil {
		args = append(args, *patch.Status)
		set = append(set, fmt.Sprintf("status = $%d", len(args)))
	}

	if len(set) > 1 {
		args = append(args, transactionID)
		query := fmt.Sprintf("UPDATE transactions SET %s WHERE id = $%d", strings.Join(set, ", "), len(args))
		if _, err := q.Exec(ctx, query, args...); err != nil {
			return nil, err
		}
	}

	// 2. Handle Description and CategoryIDs (updating the first line item)
	if patch.Description != nil || patch.CategoryIDs != nil {
		// Identify the first line item
		var firstID uuid.UUID
		err := q.QueryRow(ctx, "SELECT id FROM transaction_line_items WHERE transaction_id = $1 ORDER BY id LIMIT 1", transactionID).Scan(&firstID)
		if err == nil {
			liSet := []string{}
			liArgs := []any{}

			if patch.Description != nil {
				liArgs = append(liArgs, *patch.Description)
				liSet = append(liSet, fmt.Sprintf("note = $%d", len(liArgs)))
			}
			if len(patch.CategoryIDs) > 0 {
				liArgs = append(liArgs, patch.CategoryIDs[0])
				liSet = append(liSet, fmt.Sprintf("category_id = $%d", len(liArgs)))
			}

			if len(liSet) > 0 {
				liArgs = append(liArgs, firstID)
				liQuery := fmt.Sprintf("UPDATE transaction_line_items SET %s WHERE id = $%d", strings.Join(liSet, ", "), len(liArgs))
				if _, err := q.Exec(ctx, liQuery, liArgs...); err != nil {
					return nil, err
				}
			}
		}
	}

	// 3. Handle LineItems: replace all
	if patch.LineItems != nil {
		_, err := q.Exec(ctx, "DELETE FROM transaction_line_items WHERE transaction_id = $1", transactionID)
		if err != nil {
			return nil, err
		}
		for _, li := range *patch.LineItems {
			liID := li.ID
			if liID == uuid.Nil {
				liID = uuid.New()
			}
			_, err = q.Exec(ctx, "INSERT INTO transaction_line_items (id, transaction_id, category_id, amount, note) VALUES ($1,$2,$3,$4,$5)",
				liID, transactionID, li.CategoryID, li.Amount, li.Note)
			if err != nil {
				return nil, err
			}
			if len(li.TagIDs) > 0 {
				for _, tid := range li.TagIDs {
					_, err = q.Exec(ctx, "INSERT INTO transaction_line_item_tags (line_item_id, tag_id, created_at, updated_at) VALUES ($1, $2, $3, $4)", liID, tid, now, now)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	// 4. Handle Tags
	if patch.TagIDs != nil {
		_, err := q.Exec(ctx, "DELETE FROM transaction_tags WHERE transaction_id = $1", transactionID)
		if err != nil {
			return nil, err
		}
		for _, tid := range patch.TagIDs {
			_, err = q.Exec(ctx, "INSERT INTO transaction_tags (transaction_id, tag_id, created_at, updated_at) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING", transactionID, tid, now, now)
			if err != nil {
				return nil, err
			}
		}
	}

	// Return enriched transaction
	return r.getTransaction(ctx, q, userID, transactionID)
}

func (r *TransactionRepo) BatchPatchTransactionsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, transactionIDs []uuid.UUID, patches map[uuid.UUID]entity.TransactionPatch, mode string) ([]uuid.UUID, []uuid.UUID, error) {
	// Simple implementation: iterate and patch.
	// In "atomic" mode, use one big transaction. In "partial", individual transactions.

	if mode == "atomic" {
		var updated []uuid.UUID
		if tx != nil {
			for _, id := range transactionIDs {
				p, ok := patches[id]
				if !ok {
					continue
				}
				_, err := r.patchTransactionTx(ctx, tx, userID, id, p)
				if err != nil {
					return nil, transactionIDs, err
				}
				updated = append(updated, id)
			}
			return updated, []uuid.UUID{}, nil
		}

		err := r.db.WithTx(ctx, func(txConn pgx.Tx) error {
			for _, id := range transactionIDs {
				p, ok := patches[id]
				if !ok {
					continue
				}
				_, err := r.patchTransactionTx(ctx, txConn, userID, id, p)
				if err != nil {
					return err
				}
				updated = append(updated, id)
			}
			return nil
		})
		if err != nil {
			return nil, transactionIDs, err
		}
		return updated, []uuid.UUID{}, nil
	}

	// Partial mode
	var updated []uuid.UUID
	var failed []uuid.UUID
	for _, id := range transactionIDs {
		p, ok := patches[id]
		if !ok {
			continue
		}
		_, err := r.PatchTransactionTx(ctx, nil, userID, id, p)
		if err != nil {
			failed = append(failed, id)
		} else {
			updated = append(updated, id)
		}
	}
	return updated, failed, nil
}


func (r *TransactionRepo) DeleteTransactionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, transactionID uuid.UUID) error {
	q, err := r.db.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	now := utils.Now()
	// Verify owner using the transaction
	_, err = r.getTransaction(ctx, q, userID, transactionID)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, "UPDATE transactions SET deleted_at = $1, updated_at = $1 WHERE id = $2", now, transactionID)
	return err
}

func (r *TransactionRepo) DeleteTransactionsByAccountTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, accountID uuid.UUID) error {
	q, err := r.db.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	now := utils.Now()
	_, err = q.Exec(ctx, `
		UPDATE transactions
		SET deleted_at = $1, updated_at = $1
		WHERE (from_account_id = $2 OR to_account_id = $2)
		  AND deleted_at IS NULL
	`, now, accountID)
	return err
}

func (r *TransactionRepo) ListTransactionsByIDs(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) ([]entity.Transaction, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT 
			t.id, t.external_ref, t.type::text, t.occurred_at, t.occurred_date, t.amount::text,
			t.from_amount::text, t.to_amount::text, t.description, t.account_id,
			t.from_account_id, t.to_account_id, t.exchange_rate::text, t.status::text,
			t.created_at, t.updated_at
		FROM transactions t
		WHERE t.user_id = $1 AND t.id = ANY($2) AND t.deleted_at IS NULL
	`, userID, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.Transaction
	for rows.Next() {
		var t entity.Transaction
		if err := rows.Scan(
			&t.ID, &t.ExternalRef, &t.Type, &t.OccurredAt, &t.OccurredDate, &t.Amount,
			&t.FromAmount, &t.ToAmount, &t.Description, &t.AccountID,
			&t.FromAccountID, &t.ToAccountID, &t.ExchangeRate, &t.Status,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

// Helper: requireAccountPermission
func (r *TransactionRepo) requireAccountPermission(ctx context.Context, tx pgx.Tx, userID, accountID uuid.UUID) error {
	q, err := r.db.Queryer(ctx, tx)
	if err != nil {
		return err
	}
	return requireAccountPermission(ctx, q, userID, accountID)
}

// Cursor Encoding/Decoding
func encodeCursor(t time.Time, id uuid.UUID) string {
	s := fmt.Sprintf("%d,%s", t.UnixNano(), id.String())
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func decodeCursor(c string) (*time.Time, *uuid.UUID, error) {
	b, err := base64.StdEncoding.DecodeString(c)
	if err != nil {
		return nil, nil, err
	}
	parts := strings.Split(string(b), ",")
	if len(parts) != 2 {
		return nil, nil, errors.New("invalid cursor")
	}

	nano, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, nil, err
	}
	t := time.Unix(0, nano).UTC()
	uid, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, nil, err
	}
	return &t, &uid, nil
}
