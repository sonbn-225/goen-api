package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
	"github.com/sonbn-225/goen-api-v2/internal/core/money"
	"github.com/sonbn-225/goen-api-v2/internal/domains/transaction"
)

type TransactionRepository struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, txEntity *transaction.Transaction, opts transaction.CreateOptions) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "transaction", "operation", "create", "user_id", txEntity.UserID, "transaction_id", txEntity.ID, "type", txEntity.Type)
	now := time.Now().UTC()
	createdAt := txEntity.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	var rowsAffected int64
	var err error

	dbTx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		logger.Error("repo_transaction_create_failed", "error", err)
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = dbTx.Rollback(ctx)
		}
	}()

	if txEntity.Type == "transfer" {
		commandTag, execErr := dbTx.Exec(ctx, `
			INSERT INTO transactions (
				id,
				type,
				occurred_at,
				amount,
				from_account_id,
				to_account_id,
				status,
				created_at,
				updated_at,
				created_by,
				updated_by
			)
			SELECT $1, $2::transaction_type, $3, $4, $5, $6, 'pending', $7, $8, $9, $10
			WHERE EXISTS (
				SELECT 1 FROM user_accounts ua
				WHERE ua.user_id = $9
				  AND ua.account_id = $5
				  AND ua.status = 'active'
			)
			AND EXISTS (
				SELECT 1 FROM user_accounts ua
				WHERE ua.user_id = $9
				  AND ua.account_id = $6
				  AND ua.status = 'active'
			)
		`,
			txEntity.ID,
			txEntity.Type,
			now,
			txEntity.Amount.String(),
			txEntity.FromAccountID,
			txEntity.ToAccountID,
			createdAt,
			now,
			txEntity.UserID,
			txEntity.UserID,
		)
		err = execErr
		if execErr == nil {
			rowsAffected = commandTag.RowsAffected()
		}
	} else {
		commandTag, execErr := dbTx.Exec(ctx, `
			INSERT INTO transactions (
				id,
				type,
				occurred_at,
				amount,
				account_id,
				status,
				created_at,
				updated_at,
				created_by,
				updated_by
			)
			SELECT $1, $2::transaction_type, $3, $4, $5, 'pending', $6, $7, $8, $9
			WHERE EXISTS (
				SELECT 1
				FROM user_accounts ua
				WHERE ua.user_id = $8
				  AND ua.account_id = $5
				  AND ua.status = 'active'
			)
		`,
			txEntity.ID,
			txEntity.Type,
			now,
			txEntity.Amount.String(),
			txEntity.AccountID,
			createdAt,
			now,
			txEntity.UserID,
			txEntity.UserID,
		)
		err = execErr
		if execErr == nil {
			rowsAffected = commandTag.RowsAffected()
		}
	}
	if err != nil {
		logger.Error("repo_transaction_create_failed", "error", err)
		return err
	}

	if rowsAffected == 0 {
		logger.Warn("repo_transaction_create_failed", "reason", "account does not exist or is not owned by user")
		return apperrors.New(apperrors.KindForbidden, "account does not exist or is not owned by user")
	}

	for _, lineItem := range opts.LineItems {
		if lineItem.CategoryID == nil || strings.TrimSpace(*lineItem.CategoryID) == "" {
			return apperrors.New(apperrors.KindValidation, "line_items.category_id is required")
		}

		var categoryExists bool
		if err := dbTx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM categories
				WHERE id = $1
				  AND deleted_at IS NULL
				  AND is_active = true
			)
		`, *lineItem.CategoryID).Scan(&categoryExists); err != nil {
			logger.Error("repo_transaction_create_failed", "error", err)
			return err
		}
		if !categoryExists {
			return apperrors.New(apperrors.KindValidation, "line_items.category_id is invalid")
		}

		lineItemID := uuid.NewString()
		if _, err := dbTx.Exec(ctx, `
			INSERT INTO transaction_line_items (id, transaction_id, category_id, amount, note)
			VALUES ($1, $2, $3, $4, $5)
		`,
			lineItemID,
			txEntity.ID,
			lineItem.CategoryID,
			lineItem.Amount.String(),
			lineItem.Note,
		); err != nil {
			logger.Error("repo_transaction_create_failed", "error", err)
			return err
		}

		if len(lineItem.TagIDs) == 0 {
			continue
		}

		var validTagCount int
		if err := dbTx.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM tags
			WHERE user_id = $1
			  AND id = ANY($2::text[])
		`, txEntity.UserID, lineItem.TagIDs).Scan(&validTagCount); err != nil {
			logger.Error("repo_transaction_create_failed", "error", err)
			return err
		}
		if validTagCount != len(lineItem.TagIDs) {
			return apperrors.New(apperrors.KindValidation, "line_items.tag_ids are invalid")
		}

		if _, err := dbTx.Exec(ctx, `
			INSERT INTO transaction_line_item_tags (line_item_id, tag_id, created_at)
			SELECT $1, unnest($2::text[]), $3
		`, lineItemID, lineItem.TagIDs, now); err != nil {
			logger.Error("repo_transaction_create_failed", "error", err)
			return err
		}
	}

	for _, participant := range opts.GroupParticipants {
		if _, err := dbTx.Exec(ctx, `
			INSERT INTO group_expense_participants (
				id,
				user_id,
				transaction_id,
				participant_name,
				original_amount,
				share_amount,
				is_settled,
				created_at,
				updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, false, $7, $8)
		`,
			uuid.NewString(),
			txEntity.UserID,
			txEntity.ID,
			participant.ParticipantName,
			participant.OriginalAmount.String(),
			participant.ShareAmount.String(),
			now,
			now,
		); err != nil {
			logger.Error("repo_transaction_create_failed", "error", err)
			return err
		}
	}

	if err := dbTx.Commit(ctx); err != nil {
		logger.Error("repo_transaction_create_failed", "error", err)
		return err
	}
	committed = true
	logger.Info("repo_transaction_create_succeeded")

	return nil
}

func (r *TransactionRepository) Update(ctx context.Context, userID, transactionID string, input transaction.UpdateInput) (*transaction.Transaction, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "transaction", "operation", "update", "user_id", userID, "transaction_id", transactionID)

	dbTx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		logger.Error("repo_transaction_update_failed", "error", err)
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = dbTx.Rollback(ctx)
		}
	}()

	var txType string
	err = dbTx.QueryRow(ctx, `
		SELECT t.type::text
		FROM transactions t
		WHERE t.id = $2
		  AND t.deleted_at IS NULL
		  AND `+accessibleTransactionCondition("$1")+`
		FOR UPDATE
	`, userID, transactionID).Scan(&txType)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		logger.Error("repo_transaction_update_failed", "error", err)
		return nil, err
	}

	now := time.Now().UTC()

	if input.LineItems != nil {
		if txType == "transfer" {
			if len(*input.LineItems) > 0 {
				return nil, apperrors.New(apperrors.KindValidation, "line_items must be empty for transfer")
			}
		} else {
			if len(*input.LineItems) == 0 {
				return nil, apperrors.New(apperrors.KindValidation, "line_items is required and must include at least one category")
			}

			if _, err := dbTx.Exec(ctx, `DELETE FROM transaction_line_items WHERE transaction_id = $1`, transactionID); err != nil {
				logger.Error("repo_transaction_update_failed", "error", err)
				return nil, err
			}

			totalAmount := money.Zero().Decimal
			for _, lineItem := range *input.LineItems {
				if lineItem.CategoryID == nil || strings.TrimSpace(*lineItem.CategoryID) == "" {
					return nil, apperrors.New(apperrors.KindValidation, "line_items.category_id is required")
				}

				var categoryExists bool
				if err := dbTx.QueryRow(ctx, `
					SELECT EXISTS (
						SELECT 1
						FROM categories
						WHERE id = $1
						  AND deleted_at IS NULL
						  AND is_active = true
					)
				`, *lineItem.CategoryID).Scan(&categoryExists); err != nil {
					logger.Error("repo_transaction_update_failed", "error", err)
					return nil, err
				}
				if !categoryExists {
					return nil, apperrors.New(apperrors.KindValidation, "line_items.category_id is invalid")
				}

				lineItemID := uuid.NewString()
				if _, err := dbTx.Exec(ctx, `
					INSERT INTO transaction_line_items (id, transaction_id, category_id, amount, note)
					VALUES ($1, $2, $3, $4, $5)
				`,
					lineItemID,
					transactionID,
					lineItem.CategoryID,
					lineItem.Amount.String(),
					lineItem.Note,
				); err != nil {
					logger.Error("repo_transaction_update_failed", "error", err)
					return nil, err
				}

				if len(lineItem.TagIDs) > 0 {
					var validTagCount int
					if err := dbTx.QueryRow(ctx, `
						SELECT COUNT(*)
						FROM tags
						WHERE user_id = $1
						  AND id = ANY($2::text[])
					`, userID, lineItem.TagIDs).Scan(&validTagCount); err != nil {
						logger.Error("repo_transaction_update_failed", "error", err)
						return nil, err
					}
					if validTagCount != len(lineItem.TagIDs) {
						return nil, apperrors.New(apperrors.KindValidation, "line_items.tag_ids are invalid")
					}

					if _, err := dbTx.Exec(ctx, `
						INSERT INTO transaction_line_item_tags (line_item_id, tag_id, created_at)
						SELECT $1, unnest($2::text[]), $3
					`, lineItemID, lineItem.TagIDs, now); err != nil {
						logger.Error("repo_transaction_update_failed", "error", err)
						return nil, err
					}
				}

				totalAmount = totalAmount.Add(lineItem.Amount.Decimal)
			}

			if _, err := dbTx.Exec(ctx, `
				UPDATE transactions
				SET amount = $1,
				    updated_at = $2,
				    updated_by = $3
				WHERE id = $4
			`, totalAmount.String(), now, userID, transactionID); err != nil {
				logger.Error("repo_transaction_update_failed", "error", err)
				return nil, err
			}
		}
	}

	if input.GroupParticipants != nil {
		if txType != "expense" {
			return nil, apperrors.New(apperrors.KindValidation, "group_participants are only supported for expense transactions")
		}

		if _, err := dbTx.Exec(ctx, `
			DELETE FROM group_expense_participants
			WHERE transaction_id = $1
			  AND is_settled = false
		`, transactionID); err != nil {
			logger.Error("repo_transaction_update_failed", "error", err)
			return nil, err
		}

		for _, participant := range *input.GroupParticipants {
			name := strings.TrimSpace(participant.ParticipantName)
			if name == "" {
				return nil, apperrors.New(apperrors.KindValidation, "participant_name is required")
			}
			if !participant.OriginalAmount.GreaterThan(money.Zero().Decimal) {
				return nil, apperrors.New(apperrors.KindValidation, "participant original_amount must be greater than zero")
			}
			if !participant.ShareAmount.GreaterThan(money.Zero().Decimal) {
				return nil, apperrors.New(apperrors.KindValidation, "participant share_amount must be greater than zero")
			}

			if _, err := dbTx.Exec(ctx, `
				INSERT INTO group_expense_participants (
					id,
					user_id,
					transaction_id,
					participant_name,
					original_amount,
					share_amount,
					is_settled,
					created_at,
					updated_at
				)
				VALUES ($1, $2, $3, $4, $5, $6, false, $7, $8)
			`,
				uuid.NewString(),
				userID,
				transactionID,
				name,
				participant.OriginalAmount.String(),
				participant.ShareAmount.String(),
				now,
				now,
			); err != nil {
				logger.Error("repo_transaction_update_failed", "error", err)
				return nil, err
			}
		}

		if _, err := dbTx.Exec(ctx, `
			UPDATE transactions
			SET updated_at = $1,
			    updated_by = $2
			WHERE id = $3
		`, now, userID, transactionID); err != nil {
			logger.Error("repo_transaction_update_failed", "error", err)
			return nil, err
		}
	}

	if input.Note != nil {
		note := strings.TrimSpace(*input.Note)
		if input.LineItems == nil || len(*input.LineItems) == 0 {
			var firstLineID string
			err := dbTx.QueryRow(ctx, `
				SELECT id
				FROM transaction_line_items
				WHERE transaction_id = $1
				ORDER BY id
				LIMIT 1
			`, transactionID).Scan(&firstLineID)
			if err != nil {
				if !isNoRows(err) {
					logger.Error("repo_transaction_update_failed", "error", err)
					return nil, err
				}
			} else {
				if _, err := dbTx.Exec(ctx, `UPDATE transaction_line_items SET note = $1 WHERE id = $2`, note, firstLineID); err != nil {
					logger.Error("repo_transaction_update_failed", "error", err)
					return nil, err
				}
			}
		}

		if _, err := dbTx.Exec(ctx, `
			UPDATE transactions
			SET updated_at = $1,
			    updated_by = $2
			WHERE id = $3
		`, now, userID, transactionID); err != nil {
			logger.Error("repo_transaction_update_failed", "error", err)
			return nil, err
		}
	}

	if err := dbTx.Commit(ctx); err != nil {
		logger.Error("repo_transaction_update_failed", "error", err)
		return nil, err
	}
	committed = true

	updated, err := r.GetByID(ctx, userID, transactionID)
	if err != nil {
		logger.Error("repo_transaction_update_failed", "error", err)
		return nil, err
	}

	logger.Info("repo_transaction_update_succeeded")
	return updated, nil
}

func (r *TransactionRepository) ListByUser(ctx context.Context, userID string, filter transaction.ListFilter) ([]transaction.Transaction, int, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "transaction", "operation", "list_by_user", "user_id", userID)
	whereClauses := []string{"t.deleted_at IS NULL", accessibleTransactionCondition("$1")}
	args := []any{userID}
	nextArg := 2

	if filter.AccountID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("(t.account_id = $%d OR t.from_account_id = $%d OR t.to_account_id = $%d)", nextArg, nextArg, nextArg))
		args = append(args, *filter.AccountID)
		nextArg++
	}
	if filter.Status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("t.status = $%d::transaction_status", nextArg))
		args = append(args, *filter.Status)
		nextArg++
	}
	if filter.Search != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("COALESCE(t.external_ref, '') ILIKE '%%' || $%d || '%%'", nextArg))
		args = append(args, *filter.Search)
		nextArg++
	}
	if filter.From != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("t.occurred_at >= $%d", nextArg))
		args = append(args, *filter.From)
		nextArg++
	}
	if filter.To != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("t.occurred_at <= $%d", nextArg))
		args = append(args, *filter.To)
		nextArg++
	}
	if filter.Type != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("t.type = $%d::transaction_type", nextArg))
		args = append(args, *filter.Type)
		nextArg++
	}
	if filter.ExternalRefFamily != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("split_part(COALESCE(t.external_ref, ''), ':', 1) = $%d", nextArg))
		args = append(args, *filter.ExternalRefFamily)
		nextArg++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := "SELECT COUNT(1) FROM transactions t WHERE " + whereSQL
	var totalCount int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		logger.Error("repo_transaction_list_failed", "error", err)
		return nil, 0, err
	}

	listQuery := fmt.Sprintf(`
		SELECT
			t.id,
			t.external_ref,
			t.account_id,
			t.from_account_id,
			t.to_account_id,
			t.type::text,
			t.status::text,
			t.amount::text,
			t.occurred_at,
			t.created_at
		FROM transactions t
		WHERE %s
		ORDER BY t.occurred_at DESC, t.id DESC
		LIMIT $%d
	`, whereSQL, nextArg)

	argsWithLimit := append(append([]any{}, args...), filter.Limit)
	rows, err := r.db.Query(ctx, listQuery, argsWithLimit...)
	if err != nil {
		logger.Error("repo_transaction_list_failed", "error", err)
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]transaction.Transaction, 0)
	for rows.Next() {
		var item transaction.Transaction
		var amountStr string
		var externalRef *string
		var accountID *string
		var fromAccountID *string
		var toAccountID *string
		item.UserID = userID
		if err := rows.Scan(
			&item.ID,
			&externalRef,
			&accountID,
			&fromAccountID,
			&toAccountID,
			&item.Type,
			&item.Status,
			&amountStr,
			&item.OccurredAt,
			&item.CreatedAt,
		); err != nil {
			logger.Error("repo_transaction_list_failed", "error", err)
			return nil, 0, err
		}
		item.ExternalRef = externalRef
		item.AccountID = accountID
		item.FromAccountID = fromAccountID
		item.ToAccountID = toAccountID
		amount, err := money.NewFromString(amountStr)
		if err != nil {
			logger.Error("repo_transaction_list_failed", "error", err)
			return nil, 0, err
		}
		item.Amount = amount
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		logger.Error("repo_transaction_list_failed", "error", err)
		return nil, 0, err
	}
	logger.Info("repo_transaction_list_succeeded", "count", len(items))

	return items, totalCount, nil
}

func (r *TransactionRepository) GetByID(ctx context.Context, userID, transactionID string) (*transaction.Transaction, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "transaction", "operation", "get_by_id", "user_id", userID, "transaction_id", transactionID)

	query := `
		SELECT
			t.id,
			t.external_ref,
			t.account_id,
			t.from_account_id,
			t.to_account_id,
			t.type::text,
			t.status::text,
			t.amount::text,
			t.occurred_at,
			t.created_at
		FROM transactions t
		WHERE t.id = $2
		  AND t.deleted_at IS NULL
		  AND ` + accessibleTransactionCondition("$1")

	var (
		item          transaction.Transaction
		externalRef   *string
		accountID     *string
		fromAccountID *string
		toAccountID   *string
		amountStr     string
	)
	item.UserID = userID

	err := r.db.QueryRow(ctx, query, userID, transactionID).Scan(
		&item.ID,
		&externalRef,
		&accountID,
		&fromAccountID,
		&toAccountID,
		&item.Type,
		&item.Status,
		&amountStr,
		&item.OccurredAt,
		&item.CreatedAt,
	)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		logger.Error("repo_transaction_get_failed", "error", err)
		return nil, err
	}

	item.ExternalRef = externalRef
	item.AccountID = accountID
	item.FromAccountID = fromAccountID
	item.ToAccountID = toAccountID
	amount, err := money.NewFromString(amountStr)
	if err != nil {
		logger.Error("repo_transaction_get_failed", "error", err)
		return nil, err
	}
	item.Amount = amount

	lineItems, err := r.listTransactionLineItems(ctx, transactionID)
	if err != nil {
		logger.Error("repo_transaction_get_failed", "error", err)
		return nil, err
	}
	item.LineItems = lineItems

	logger.Info("repo_transaction_get_succeeded")
	return &item, nil
}

func (r *TransactionRepository) listTransactionLineItems(ctx context.Context, transactionID string) ([]transaction.TransactionLineItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			li.id,
			li.category_id,
			li.amount::text,
			li.note,
			COALESCE((
				SELECT array_agg(tlit.tag_id ORDER BY tlit.tag_id)
				FROM transaction_line_item_tags tlit
				WHERE tlit.line_item_id = li.id
			), '{}'::text[]) AS tag_ids
		FROM transaction_line_items li
		WHERE li.transaction_id = $1
		ORDER BY li.id
	`, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]transaction.TransactionLineItem, 0)
	for rows.Next() {
		var item transaction.TransactionLineItem
		if err := rows.Scan(&item.ID, &item.CategoryID, &item.Amount, &item.Note, &item.TagIDs); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *TransactionRepository) BatchPatchStatus(ctx context.Context, userID string, transactionIDs []string, status string) ([]string, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "transaction", "operation", "batch_patch_status", "user_id", userID)

	rows, err := r.db.Query(ctx, `
		UPDATE transactions t
		SET status = $2::transaction_status,
			updated_at = $3,
			updated_by = $1
		WHERE t.deleted_at IS NULL
		  AND t.id = ANY($4::text[])
		  AND `+accessibleTransactionCondition("$1")+`
		RETURNING t.id
	`, userID, status, time.Now().UTC(), transactionIDs)
	if err != nil {
		logger.Error("repo_transaction_batch_patch_failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	updatedIDs := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			logger.Error("repo_transaction_batch_patch_failed", "error", err)
			return nil, err
		}
		updatedIDs = append(updatedIDs, id)
	}

	if err := rows.Err(); err != nil {
		logger.Error("repo_transaction_batch_patch_failed", "error", err)
		return nil, err
	}

	logger.Info("repo_transaction_batch_patch_succeeded", "updated_count", len(updatedIDs))
	return updatedIDs, nil
}

func (r *TransactionRepository) ListGroupParticipantsByTransaction(ctx context.Context, userID, transactionID string) ([]transaction.GroupExpenseParticipant, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "transaction", "operation", "list_group_participants", "user_id", userID, "transaction_id", transactionID)

	rows, err := r.db.Query(ctx, `
		SELECT
			gep.id,
			gep.user_id,
			gep.transaction_id,
			gep.participant_name,
			gep.original_amount::text,
			gep.share_amount::text,
			gep.is_settled,
			gep.settlement_transaction_id,
			gep.created_at,
			gep.updated_at
		FROM group_expense_participants gep
		JOIN transactions t ON t.id = gep.transaction_id
		WHERE gep.transaction_id = $2
		  AND t.deleted_at IS NULL
		  AND `+accessibleTransactionCondition("$1")+`
		ORDER BY gep.created_at DESC, gep.id DESC
	`, userID, transactionID)
	if err != nil {
		logger.Error("repo_transaction_list_group_participants_failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	items := make([]transaction.GroupExpenseParticipant, 0)
	for rows.Next() {
		var item transaction.GroupExpenseParticipant
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.TransactionID,
			&item.ParticipantName,
			&item.OriginalAmount,
			&item.ShareAmount,
			&item.IsSettled,
			&item.SettlementTransactionID,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			logger.Error("repo_transaction_list_group_participants_failed", "error", err)
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		logger.Error("repo_transaction_list_group_participants_failed", "error", err)
		return nil, err
	}

	logger.Info("repo_transaction_list_group_participants_succeeded", "count", len(items))
	return items, nil
}

func accessibleTransactionCondition(userPlaceholder string) string {
	return `(
		(
			t.type IN ('expense', 'income')
			AND EXISTS (
				SELECT 1
				FROM user_accounts ua
				WHERE ua.user_id = ` + userPlaceholder + `
				  AND ua.account_id = t.account_id
				  AND ua.status = 'active'
			)
		)
		OR (
			t.type = 'transfer'
			AND (
				EXISTS (
					SELECT 1
					FROM user_accounts ua
					WHERE ua.user_id = ` + userPlaceholder + `
					  AND ua.account_id = t.from_account_id
					  AND ua.status = 'active'
				)
				OR EXISTS (
					SELECT 1
					FROM user_accounts ua
					WHERE ua.user_id = ` + userPlaceholder + `
					  AND ua.account_id = t.to_account_id
					  AND ua.status = 'active'
				)
			)
		)
	)`
}
