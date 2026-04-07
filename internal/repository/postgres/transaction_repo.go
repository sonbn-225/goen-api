package postgres

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

type TransactionRepo struct {
	db *database.Postgres
}

func NewTransactionRepo(db *database.Postgres) *TransactionRepo {
	return &TransactionRepo{db: db}
}

func (r *TransactionRepo) CreateTransaction(ctx context.Context, userID uuid.UUID, tx entity.Transaction, lineItems []entity.TransactionLineItem, tagIDs []uuid.UUID, participants []entity.GroupExpenseParticipant) error {
	return r.db.WithTx(ctx, func(txConn pgx.Tx) error {
		return createTransactionTx(ctx, txConn, userID, tx, lineItems, tagIDs, participants)
	})
}

func (r *TransactionRepo) GetTransaction(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*entity.Transaction, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT
			t.id, t.external_ref, t.type, t.occurred_at,
			to_char(t.occurred_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS occurred_date,
			t.amount::text, t.from_amount::text, t.to_amount::text,
			(SELECT li.note FROM transaction_line_items li WHERE li.transaction_id = t.id ORDER BY li.id LIMIT 1) AS description,
			t.account_id, t.from_account_id, t.to_account_id, t.exchange_rate::text,
			a.currency AS account_currency, fa.currency AS from_currency, ta.currency AS to_currency,
			t.status, t.created_at, t.updated_at, t.deleted_at,
			COALESCE((SELECT array_agg(tt.tag_id ORDER BY tt.tag_id) FROM transaction_tags tt WHERE tt.transaction_id = t.id), '{}'::text[]) AS tag_ids
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
	err = row.Scan(
		&t.ID, &t.ExternalRef, &t.Type, &t.OccurredAt, &t.OccurredDate,
		&t.Amount, &t.FromAmount, &t.ToAmount, &t.Description, &t.AccountID, &t.FromAccountID, &t.ToAccountID,
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
	_ = pool.QueryRow(ctx, `
		SELECT 
			COALESCE(array_agg(DISTINCT c.name), '{}'::text[]),
			COALESCE(array_agg(DISTINCT c.color), '{}'::text[]),
			COALESCE(array_agg(DISTINCT tg.name), '{}'::text[]),
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
	rows, err := pool.Query(ctx, `
		SELECT li.id, li.category_id, li.amount::text, li.note,
		       COALESCE((SELECT array_agg(tlit.tag_id ORDER BY tlit.tag_id) FROM transaction_line_item_tags tlit WHERE tlit.line_item_id = li.id), '{}'::text[])
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
			t.account_id, t.from_account_id, t.to_account_id, t.exchange_rate::text,
			a.currency AS account_currency, fa.currency AS from_currency, ta.currency AS to_currency,
			t.status, t.created_at, t.updated_at, t.deleted_at,
			COALESCE((SELECT array_agg(tt.tag_id ORDER BY tt.tag_id) FROM transaction_tags tt WHERE tt.transaction_id = t.id), '{}'::text[]) AS tag_ids,
			COALESCE((SELECT array_agg(DISTINCT tli.category_id ORDER BY tli.category_id) FROM transaction_line_items tli WHERE tli.transaction_id = t.id AND tli.category_id IS NOT NULL), '{}'::text[]) AS category_ids
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
			&t.Amount, &t.FromAmount, &t.ToAmount, &t.Description, &t.AccountID, &t.FromAccountID, &t.ToAccountID,
			&t.ExchangeRate, &t.AccountCurrency, &t.FromCurrency, &t.ToCurrency, &t.Status,
			&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt, &t.TagIDs, &t.CategoryIDs,
		)
		if err != nil {
			return nil, nil, 0, err
		}

		// Quick enrichment join per item (or could use a larger join above, but array_agg per item is safer for complex many-to-many)
		_ = pool.QueryRow(ctx, `
			SELECT 
				COALESCE(array_agg(DISTINCT c.name), '{}'::text[]),
				COALESCE(array_agg(DISTINCT c.color), '{}'::text[]),
				COALESCE(array_agg(DISTINCT tg.name), '{}'::text[]),
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

func (r *TransactionRepo) PatchTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID, patch entity.TransactionPatch) (*entity.Transaction, error) {
	now := time.Now().UTC()
	err := r.db.WithTx(ctx, func(txConn pgx.Tx) error {
		var err error
		if _, err = r.GetTransaction(ctx, userID, transactionID); err != nil {
			return err
		}

		// Basic SQL update
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

		if len(set) > 2 {
			args = append(args, transactionID)
			query := fmt.Sprintf("UPDATE transactions SET %s WHERE id = $%d", strings.Join(set, ", "), len(args))
			if _, err := txConn.Exec(ctx, query, args...); err != nil {
				return err
			}
		}

		// Handle LineItems: replace all
		if patch.LineItems != nil {
			_, err = txConn.Exec(ctx, "DELETE FROM transaction_line_items WHERE transaction_id = $1", transactionID)
			if err != nil {
				return err
			}
			for _, li := range *patch.LineItems {
				liID := li.ID
				if liID == uuid.Nil {
					liID = uuid.New()
				}
				_, err = txConn.Exec(ctx, "INSERT INTO transaction_line_items (id, transaction_id, category_id, amount, note) VALUES ($1,$2,$3,$4,$5)",
					liID, transactionID, li.CategoryID, li.Amount, li.Note)
				if err != nil {
					return err
				}
				if len(li.TagIDs) > 0 {
					for _, tid := range li.TagIDs {
						_, err = txConn.Exec(ctx, "INSERT INTO transaction_line_item_tags (line_item_id, tag_id, created_at, updated_at) VALUES ($1, $2, $3, $4)", liID, tid, now, now)
						if err != nil {
							return err
						}
					}
				}
			}
		}

		// Handle Tags
		if patch.TagIDs != nil {
			_, err = txConn.Exec(ctx, "DELETE FROM transaction_tags WHERE transaction_id = $1", transactionID)
			if err != nil {
				return err
			}
			for _, tid := range patch.TagIDs {
				_, err = txConn.Exec(ctx, "INSERT INTO transaction_tags (transaction_id, tag_id, created_at, updated_at) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING", transactionID, tid, now, now)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return r.GetTransaction(ctx, userID, transactionID)
}

func (r *TransactionRepo) BatchPatchTransactions(ctx context.Context, userID uuid.UUID, transactionIDs []uuid.UUID, patches map[uuid.UUID]entity.TransactionPatch, mode string) ([]uuid.UUID, []uuid.UUID, error) {
	// Simple implementation: iterate and patch.
	// In "atomic" mode, use one big transaction. In "partial", individual transactions.

	if mode == "atomic" {
		var updated []uuid.UUID
		err := r.db.WithTx(ctx, func(txConn pgx.Tx) error {
			for _, id := range transactionIDs {
				p, ok := patches[id]
				if !ok {
					continue
				}
				_, err := r.PatchTransaction(ctx, userID, id, p)
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
		_, err := r.PatchTransaction(ctx, userID, id, p)
		if err != nil {
			failed = append(failed, id)
		} else {
			updated = append(updated, id)
		}
	}
	return updated, failed, nil
}

func (r *TransactionRepo) DeleteTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithTx(ctx, func(txConn pgx.Tx) error {
		// Verify owner
		_, err := r.GetTransaction(ctx, userID, transactionID)
		if err != nil {
			return err
		}

		_, err = txConn.Exec(ctx, "UPDATE transactions SET deleted_at = $1, updated_at = $1 WHERE id = $2", now, transactionID)
		return err
	})
}

// Helper: requireAccountPermission
func (r *TransactionRepo) requireAccountPermission(ctx context.Context, tx pgx.Tx, userID, accountID uuid.UUID) error {
	return requireAccountPermission(ctx, tx, userID, accountID)
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

	nano, err := database.ParseInt64(parts[0])
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

// Imported Transactions
func (r *TransactionRepo) CreateImportedTransactions(ctx context.Context, userID uuid.UUID, items []entity.ImportedTransactionCreate) ([]entity.ImportedTransaction, error) {
	now := time.Now().UTC()
	created := make([]entity.ImportedTransaction, 0, len(items))

	err := r.db.WithTx(ctx, func(txConn pgx.Tx) error {
		for _, item := range items {
			id := uuid.New()

			// Payload to JSON
			payloadBytes, err := json.Marshal(item.RawPayload)
			if err != nil {
				return err
			}

			if item.MappedAccountID != nil && *item.MappedAccountID != uuid.Nil {
				if err := requireAccountPermission(ctx, txConn, userID, *item.MappedAccountID); err != nil {
					return err
				}
			}

			if item.MappedCategoryID != nil && *item.MappedCategoryID != uuid.Nil {
				var exists bool
				err := txConn.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM categories WHERE id = $1 AND deleted_at IS NULL)`, *item.MappedCategoryID).Scan(&exists)
				if err != nil {
					return err
				}
				if !exists {
					return errors.New("invalid category id")
				}
			}

			_, err = txConn.Exec(ctx, `
				INSERT INTO imported_transactions (
					id, user_id, source, transaction_date, amount, description, transaction_type,
					imported_account_name, imported_category_name, mapped_account_id, mapped_category_id,
					raw_payload, created_at, updated_at
				) VALUES ($1, $2, $3, $4::date, $5::numeric, $6, $7, $8, $9, $10, $11, $12::jsonb, $13, $14)
			`,
				id, userID, item.Source, item.TransactionDate, item.Amount, item.Description, item.TransactionType,
				item.ImportedAccountName, item.ImportedCategoryName,
				normalizeOptionalImportUUID(item.MappedAccountID), normalizeOptionalImportUUID(item.MappedCategoryID),
				payloadBytes, now, now,
			)
			if err != nil {
				return err
			}

			created = append(created, entity.ImportedTransaction{
				AuditEntity: entity.AuditEntity{
					BaseEntity: entity.BaseEntity{
						ID: id,
					},
					CreatedAt: now,
					UpdatedAt: now,
				},
				UserID:               userID,
				Source:               item.Source,
				TransactionDate:      item.TransactionDate,
				Amount:               item.Amount,
				Description:          item.Description,
				TransactionType:      item.TransactionType,
				ImportedAccountName:  item.ImportedAccountName,
				ImportedCategoryName: item.ImportedCategoryName,
				MappedAccountID:      normalizeOptionalImportUUID(item.MappedAccountID),
				MappedCategoryID:     normalizeOptionalImportUUID(item.MappedCategoryID),
				RawPayload:           item.RawPayload,
			})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return created, nil
}

func (r *TransactionRepo) ListImportedTransactions(ctx context.Context, userID uuid.UUID) ([]entity.ImportedTransaction, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT id, source, transaction_date, amount::text, description, transaction_type,
		       imported_account_name, imported_category_name, mapped_account_id, mapped_category_id,
		       raw_payload, created_at, updated_at
		FROM imported_transactions
		WHERE user_id = $1
		ORDER BY transaction_date DESC, created_at DESC, id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []entity.ImportedTransaction
	for rows.Next() {
		var it entity.ImportedTransaction
		var tDate time.Time
		var raw []byte
		err := rows.Scan(
			&it.ID, &it.Source, &tDate, &it.Amount, &it.Description, &it.TransactionType,
			&it.ImportedAccountName, &it.ImportedCategoryName, &it.MappedAccountID, &it.MappedCategoryID,
			&raw, &it.CreatedAt, &it.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		it.TransactionDate = tDate.Format("2006-01-02")
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &it.RawPayload)
		}
		items = append(items, it)
	}
	return items, nil
}

func (r *TransactionRepo) PatchImportedTransaction(ctx context.Context, userID uuid.UUID, importID uuid.UUID, patch entity.ImportedTransactionPatch) (*entity.ImportedTransaction, error) {
	now := time.Now().UTC()
	var it entity.ImportedTransaction

	err := r.db.WithTx(ctx, func(txConn pgx.Tx) error {
		if patch.MappedAccountID != nil && *patch.MappedAccountID != uuid.Nil {
			if err := requireAccountPermission(ctx, txConn, userID, *patch.MappedAccountID); err != nil {
				return err
			}
		}
		if patch.MappedCategoryID != nil && *patch.MappedCategoryID != uuid.Nil {
			var exists bool
			err := txConn.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM categories WHERE id = $1 AND deleted_at IS NULL)`, *patch.MappedCategoryID).Scan(&exists)
			if err != nil {
				return err
			}
			if !exists {
				return errors.New("invalid category id")
			}
		}

		var tDate time.Time
		var raw []byte
		err := txConn.QueryRow(ctx, `
			UPDATE imported_transactions
			SET mapped_account_id = COALESCE($1, mapped_account_id),
			    mapped_category_id = COALESCE($2, mapped_category_id),
			    updated_at = $3
			WHERE id = $4 AND user_id = $5
			RETURNING id, source, transaction_date, amount::text, description, transaction_type,
			          imported_account_name, imported_category_name, mapped_account_id, mapped_category_id,
			          raw_payload, created_at, updated_at
		`, normalizeOptionalImportUUID(patch.MappedAccountID), normalizeOptionalImportUUID(patch.MappedCategoryID), now, importID, userID).Scan(
			&it.ID, &it.Source, &tDate, &it.Amount, &it.Description, &it.TransactionType,
			&it.ImportedAccountName, &it.ImportedCategoryName, &it.MappedAccountID, &it.MappedCategoryID,
			&raw, &it.CreatedAt, &it.UpdatedAt,
		)
		if err != nil {
			return err
		}
		it.TransactionDate = tDate.Format("2006-01-02")
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &it.RawPayload)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &it, nil
}

func (r *TransactionRepo) DeleteImportedTransaction(ctx context.Context, userID uuid.UUID, importID uuid.UUID) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}
	ct, err := pool.Exec(ctx, `DELETE FROM imported_transactions WHERE id = $1 AND user_id = $2`, importID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("imported transaction not found")
	}
	return nil
}

func (r *TransactionRepo) GetImportedTransaction(ctx context.Context, userID uuid.UUID, importID uuid.UUID) (*entity.ImportedTransaction, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT id, source, transaction_date, amount::text, description, transaction_type,
		       imported_account_name, imported_category_name, mapped_account_id, mapped_category_id,
		       raw_payload, created_at, updated_at
		FROM imported_transactions
		WHERE id = $1 AND user_id = $2
	`, importID, userID)

	var it entity.ImportedTransaction
	var tDate time.Time
	var raw []byte
	if err := row.Scan(
		&it.ID, &it.Source, &tDate, &it.Amount, &it.Description, &it.TransactionType,
		&it.ImportedAccountName, &it.ImportedCategoryName, &it.MappedAccountID, &it.MappedCategoryID,
		&raw, &it.CreatedAt, &it.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("imported transaction not found")
		}
		return nil, err
	}
	it.TransactionDate = tDate.Format("2006-01-02")
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &it.RawPayload)
	}
	return &it, nil
}

func (r *TransactionRepo) DeleteAllImportedTransactions(ctx context.Context, userID uuid.UUID) (int64, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return 0, err
	}
	ct, err := pool.Exec(ctx, `DELETE FROM imported_transactions WHERE user_id = $1`, userID)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

// Import Mapping Rules
func (r *TransactionRepo) UpsertImportMappingRules(ctx context.Context, userID uuid.UUID, rules []entity.ImportMappingRuleUpsert) ([]entity.ImportMappingRule, error) {
	now := time.Now().UTC()
	out := make([]entity.ImportMappingRule, 0, len(rules))

	err := r.db.WithTx(ctx, func(txConn pgx.Tx) error {
		for _, rule := range rules {
			kind := strings.ToLower(strings.TrimSpace(rule.Kind))
			sourceName := strings.TrimSpace(rule.SourceName)
			mappedID := rule.MappedID
			if sourceName == "" || mappedID == uuid.Nil {
				continue
			}

			switch kind {
			case "account":
				if err := requireAccountPermission(ctx, txConn, userID, mappedID); err != nil {
					return err
				}
			case "category":
				var exists bool
				err := txConn.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM categories WHERE id = $1 AND deleted_at IS NULL)`, mappedID).Scan(&exists)
				if err != nil {
					return err
				}
				if !exists {
					return errors.New("invalid category id")
				}
			default:
				return errors.New("invalid rule kind")
			}

			normalized := normalizeRuleSourceName(sourceName)
			var outRule entity.ImportMappingRule
			err := txConn.QueryRow(ctx, `
				INSERT INTO import_mapping_rules (
					id, user_id, kind, source_name, normalized_source_name, mapped_id, created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				ON CONFLICT (user_id, kind, normalized_source_name)
				DO UPDATE SET
					source_name = EXCLUDED.source_name,
					mapped_id = EXCLUDED.mapped_id,
					updated_at = EXCLUDED.updated_at
				RETURNING id, kind, source_name, mapped_id, created_at, updated_at
			`, uuid.New(), userID, kind, sourceName, normalized, mappedID, now, now).Scan(
				&outRule.ID, &outRule.Kind, &outRule.SourceName, &outRule.MappedID, &outRule.CreatedAt, &outRule.UpdatedAt,
			)
			if err != nil {
				return err
			}
			out = append(out, outRule)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *TransactionRepo) ListImportMappingRules(ctx context.Context, userID uuid.UUID) ([]entity.ImportMappingRule, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT id, kind, source_name, mapped_id, created_at, updated_at
		FROM import_mapping_rules
		WHERE user_id = $1
		ORDER BY kind ASC, source_name ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.ImportMappingRule
	for rows.Next() {
		var rule entity.ImportMappingRule
		if err := rows.Scan(&rule.ID, &rule.Kind, &rule.SourceName, &rule.MappedID, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, rule)
	}
	return out, nil
}

func (r *TransactionRepo) DeleteImportMappingRule(ctx context.Context, userID uuid.UUID, ruleID uuid.UUID) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}
	ct, err := pool.Exec(ctx, `DELETE FROM import_mapping_rules WHERE id = $1 AND user_id = $2`, ruleID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("rule not found")
	}
	return nil
}

// Helpers
func normalizeOptionalImportUUID(v *uuid.UUID) *uuid.UUID {
	if v == nil || *v == uuid.Nil {
		return nil
	}
	return v
}

func normalizeRuleSourceName(v string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(v)), " "))
}
