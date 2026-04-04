package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

type GroupExpenseRepo struct {
	db *database.Postgres
}

func NewGroupExpenseRepo(db *database.Postgres) *GroupExpenseRepo {
	return &GroupExpenseRepo{db: db}
}

func (r *GroupExpenseRepo) CreateGroupExpense(ctx context.Context, userID string, tx entity.Transaction, lineItems []entity.TransactionLineItem, tagIDs []string, participants []entity.GroupExpenseParticipant) error {
	return r.db.WithTx(ctx, func(txConn pgx.Tx) error {
		return createTransactionTx(ctx, txConn, userID, tx, lineItems, tagIDs, participants)
	})
}

func (r *GroupExpenseRepo) ListParticipantsByTransaction(ctx context.Context, userID, transactionID string) ([]entity.GroupExpenseParticipant, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT 
			id, user_id, transaction_id, participant_name, 
			original_amount::text, share_amount::text, is_settled, 
			settlement_transaction_id, created_at, updated_at
		FROM group_expense_participants
		WHERE user_id = $1 AND transaction_id = $2
		ORDER BY created_at ASC, id ASC
	`, userID, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.GroupExpenseParticipant
	for rows.Next() {
		var p entity.GroupExpenseParticipant
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.TransactionID, &p.ParticipantName,
			&p.OriginalAmount, &p.ShareAmount, &p.IsSettled,
			&p.SettlementTransactionID, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, p)
	}
	return results, nil
}

func (r *GroupExpenseRepo) SettleParticipant(ctx context.Context, userID, participantID string, settlementTx entity.Transaction, settlementLineItems []entity.TransactionLineItem, settlementTagIDs []string) (string, error) {
	var createdID string
	err := r.db.WithTx(ctx, func(txConn pgx.Tx) error {
		var (
			isSettled bool
			txID      string
			shareAmt  string
			name      string
		)
		err := txConn.QueryRow(ctx, `
			SELECT is_settled, transaction_id, share_amount::text, participant_name
			FROM group_expense_participants
			WHERE id = $1 AND user_id = $2
			FOR UPDATE
		`, participantID, userID).Scan(&isSettled, &txID, &shareAmt, &name)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errors.New("group expense participant not found")
			}
			return err
		}
		if isSettled {
			return errors.New("participant already settled")
		}

		// Update settlement transaction with the actual share amount
		settlementTx.Amount = shareAmt
		if settlementTx.Description == nil {
			d := fmt.Sprintf("Reimbursement from %s", strings.TrimSpace(name))
			settlementTx.Description = &d
		}
		
		// Fill line item amounts
		for i := range settlementLineItems {
			if i == 0 {
				settlementLineItems[i].Amount = shareAmt
			} else {
				settlementLineItems[i].Amount = "0.00"
			}
		}

		// Create settlement income transaction
		if err := createTransactionTx(ctx, txConn, userID, settlementTx, settlementLineItems, settlementTagIDs, nil); err != nil {
			return err
		}
		createdID = settlementTx.ID

		// Mark participant as settled
		_, err = txConn.Exec(ctx, `
			UPDATE group_expense_participants
			SET is_settled = true, settlement_transaction_id = $1, updated_at = $2
			WHERE id = $3 AND user_id = $4
		`, settlementTx.ID, time.Now().UTC(), participantID, userID)
		
		return err
	})
	
	if err != nil {
		return "", err
	}
	return createdID, nil
}

func (r *GroupExpenseRepo) ListUniqueParticipantNames(ctx context.Context, userID string, limit int) ([]string, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 50
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

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

func (r *GroupExpenseRepo) ListUnsettledParticipantsByName(ctx context.Context, userID string, name string) ([]entity.GroupExpenseParticipant, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT 
			id, user_id, transaction_id, participant_name, 
			original_amount::text, share_amount::text, is_settled, 
			settlement_transaction_id, created_at, updated_at
		FROM group_expense_participants
		WHERE user_id = $1 AND LOWER(participant_name) = LOWER($2) AND is_settled = false
		ORDER BY created_at DESC
	`, userID, strings.TrimSpace(name))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.GroupExpenseParticipant
	for rows.Next() {
		var p entity.GroupExpenseParticipant
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.TransactionID, &p.ParticipantName,
			&p.OriginalAmount, &p.ShareAmount, &p.IsSettled,
			&p.SettlementTransactionID, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, p)
	}
	return results, nil
}
