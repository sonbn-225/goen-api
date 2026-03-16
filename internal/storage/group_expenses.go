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

type GroupExpenseRepo struct {
	db *Postgres
}

func NewGroupExpenseRepo(db *Postgres) *GroupExpenseRepo {
	return &GroupExpenseRepo{db: db}
}

func (r *GroupExpenseRepo) CreateGroupExpense(ctx context.Context, userID string, tx domain.Transaction, lineItems []domain.TransactionLineItem, tagIDs []string, participants []domain.GroupExpenseParticipant) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	return withTx(ctx, pool, func(dbtx pgx.Tx) error {
		if err := createTransactionTx(ctx, dbtx, userID, tx, lineItems, tagIDs); err != nil {
			return err
		}

		for _, p := range participants {
			name := strings.TrimSpace(p.ParticipantName)
			if name == "" {
				return apperrors.Validation("participant_name is required", map[string]any{"field": "participant_name"})
			}

			_, err := dbtx.Exec(ctx, `
        INSERT INTO group_expense_participants (
          id, user_id, transaction_id, participant_name,
          original_amount, share_amount,
          is_settled, settlement_transaction_id,
          created_at, updated_at
        ) VALUES ($1,$2,$3,$4,$5::numeric,$6::numeric,$7,$8,$9,$10)
      `,
				p.ID,
				userID,
				tx.ID,
				name,
				p.OriginalAmount,
				p.ShareAmount,
				p.IsSettled,
				p.SettlementTransactionID,
				p.CreatedAt,
				p.UpdatedAt,
			)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *GroupExpenseRepo) ListParticipantsByTransaction(ctx context.Context, userID, transactionID string) ([]domain.GroupExpenseParticipant, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
    SELECT
      id,
      user_id,
      transaction_id,
      participant_name,
      original_amount::text,
      share_amount::text,
      is_settled,
      settlement_transaction_id,
      created_at,
      updated_at
    FROM group_expense_participants
    WHERE user_id = $1 AND transaction_id = $2
    ORDER BY created_at ASC, id ASC
  `, userID, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.GroupExpenseParticipant{}
	for rows.Next() {
		var it domain.GroupExpenseParticipant
		if err := rows.Scan(
			&it.ID,
			&it.UserID,
			&it.TransactionID,
			&it.ParticipantName,
			&it.OriginalAmount,
			&it.ShareAmount,
			&it.IsSettled,
			&it.SettlementTransactionID,
			&it.CreatedAt,
			&it.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *GroupExpenseRepo) SettleParticipant(ctx context.Context, userID, participantID string, settlementTx domain.Transaction, settlementLineItems []domain.TransactionLineItem, settlementTagIDs []string) (string, error) {
	if r.db == nil {
		return "", apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return "", err
	}

	var createdID string
	err = withTx(ctx, pool, func(dbtx pgx.Tx) error {
		// Lock participant row.
		var (
			isSettled bool
			txID      string
			shareAmt  string
			name      string
		)
		err := dbtx.QueryRow(ctx, `
      SELECT is_settled, transaction_id, share_amount::text, participant_name
      FROM group_expense_participants
      WHERE id = $1 AND user_id = $2
      FOR UPDATE
    `, participantID, userID).Scan(&isSettled, &txID, &shareAmt, &name)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperrors.ErrGroupExpenseParticipantNotFound
			}
			return err
		}
		if isSettled {
			return apperrors.ErrGroupExpenseParticipantAlreadySettled
		}

		shareAmt = strings.TrimSpace(shareAmt)
		if shareAmt == "" {
			return apperrors.Validation("participant share_amount is invalid", nil)
		}

		// Fill settlement tx amounts and metadata.
		settlementTx.Amount = shareAmt
		if settlementTx.Description == nil {
			d := fmt.Sprintf("Reimbursement from %s", strings.TrimSpace(name))
			settlementTx.Description = &d
		}
		if settlementTx.Description != nil {
			d := fmt.Sprintf("%s | Settlement for group expense transaction %s", strings.TrimSpace(*settlementTx.Description), txID)
			settlementTx.Description = &d
		}

		catID := "cat_def_income_reimbursement"
		if len(settlementLineItems) == 0 {
			settlementLineItems = []domain.TransactionLineItem{{ID: uuid.NewString(), CategoryID: &catID, Amount: shareAmt}}
		} else {
			settlementLineItems[0].CategoryID = &catID
			settlementLineItems[0].Amount = shareAmt
			for i := 1; i < len(settlementLineItems); i++ {
				settlementLineItems[i].Amount = "0.00"
			}
		}

		// Create settlement income transaction.
		if err := createTransactionTx(ctx, dbtx, userID, settlementTx, settlementLineItems, settlementTagIDs); err != nil {
			return err
		}
		createdID = settlementTx.ID

		// Mark participant as settled.
		_, err = dbtx.Exec(ctx, `
      UPDATE group_expense_participants
      SET is_settled = true,
          settlement_transaction_id = $1,
          updated_at = $2
      WHERE id = $3 AND user_id = $4
    `, settlementTx.ID, time.Now().UTC(), participantID, userID)
		if err != nil {
			return err
		}

		// Audit is covered by createTransactionTx.
		return nil
	})
	if err != nil {
		return "", err
	}
	return createdID, nil
}

func (r *GroupExpenseRepo) ListUniqueParticipantNames(ctx context.Context, userID string, limit int) ([]string, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	rows, err := pool.Query(ctx, `
    SELECT DISTINCT participant_name
    FROM group_expense_participants
    WHERE user_id = $1
    ORDER BY participant_name ASC
    LIMIT $2
  `, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		name = strings.TrimSpace(name)
		if name != "" {
			out = append(out, name)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

