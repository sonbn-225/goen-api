package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

type DebtRepo struct {
	BaseRepo
}

func NewDebtRepo(db *database.Postgres) *DebtRepo {
	return &DebtRepo{BaseRepo: *NewBaseRepo(db)}
}


func (r *DebtRepo) CreateDebtTx(ctx context.Context, tx pgx.Tx, debt entity.Debt) error {
	tag, err := tx.Exec(ctx, `
		INSERT INTO debts (
			id, user_id, account_id, direction, name, contact_id, principal,
			start_date, due_date, interest_rate, interest_rule,
			outstanding_principal, accrued_interest, status, closed_at,
			originating_transaction_id, created_at, updated_at
		)
		SELECT
			$1,$2,$3,$4,$5,$6,$7::numeric,
			$8::date,$9::date,$10::numeric,$11,
			$12::numeric,$13::numeric,$14,$15,
			$16,$17,$18
		WHERE EXISTS (
			SELECT 1 FROM user_accounts ua
			WHERE ua.user_id = $2 AND ua.account_id = $3 AND ua.status = 'active'
		)
	`,
		debt.ID, debt.UserID, debt.AccountID, debt.Direction, debt.Name, debt.ContactID,
		debt.Principal, debt.StartDate, debt.DueDate, debt.InterestRate, debt.InterestRule,
		debt.OutstandingPrincipal, debt.AccruedInterest, debt.Status, debt.ClosedAt,
		debt.OriginatingTransactionID, debt.CreatedAt, debt.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("forbidden: account access required")
	}
	return nil
}

func (r *DebtRepo) GetDebt(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*entity.Debt, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT
			d.id, d.user_id, d.account_id, d.direction, d.name, d.contact_id,
			COALESCE(u.display_name, c.name) AS contact_name,
			COALESCE(u.avatar_url, c.avatar_url) AS contact_avatar_url,
			d.principal::text, a.currency,
			to_char(d.start_date, 'YYYY-MM-DD'), to_char(d.due_date, 'YYYY-MM-DD'),
			d.interest_rate::text, d.interest_rule,
			d.outstanding_principal::text, d.accrued_interest::text,
			d.status, d.closed_at, d.originating_transaction_id, d.created_at, d.updated_at, d.deleted_at
			
		FROM debts d
		LEFT JOIN accounts a ON a.id = d.account_id
		LEFT JOIN contacts c ON d.contact_id = c.id
		LEFT JOIN users u ON c.linked_user_id = u.id
		WHERE d.id = $1 AND d.user_id = $2 AND d.deleted_at IS NULL
	`, id, userID)

	var d entity.Debt
	err = row.Scan(
		&d.ID, &d.UserID, &d.AccountID, &d.Direction, &d.Name, &d.ContactID,
		&d.ContactName, &d.ContactAvatarURL, &d.Principal, &d.Currency,
		&d.StartDate, &d.DueDate, &d.InterestRate, &d.InterestRule,
		&d.OutstandingPrincipal, &d.AccruedInterest, &d.Status, &d.ClosedAt,
		&d.OriginatingTransactionID, &d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("debt not found")
		}
		return nil, err
	}
	return &d, nil
}

func (r *DebtRepo) ListDebts(ctx context.Context, userID uuid.UUID) ([]entity.Debt, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT
			d.id, d.user_id, d.account_id, d.direction, d.name, d.contact_id,
			COALESCE(u.display_name, c.name) AS contact_name,
			COALESCE(u.avatar_url, c.avatar_url) AS contact_avatar_url,
			d.principal::text, a.currency,
			to_char(d.start_date, 'YYYY-MM-DD'), to_char(d.due_date, 'YYYY-MM-DD'),
			d.interest_rate::text, d.interest_rule,
			d.outstanding_principal::text, d.accrued_interest::text,
			d.status, d.closed_at, d.originating_transaction_id, d.created_at, d.updated_at, d.deleted_at
			
		FROM debts d
		LEFT JOIN accounts a ON a.id = d.account_id
		LEFT JOIN contacts c ON d.contact_id = c.id
		LEFT JOIN users u ON c.linked_user_id = u.id
		WHERE d.user_id = $1 AND d.deleted_at IS NULL
		ORDER BY d.due_date ASC, d.id ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []entity.Debt
	for rows.Next() {
		var d entity.Debt
		err := rows.Scan(
			&d.ID, &d.UserID, &d.AccountID, &d.Direction, &d.Name, &d.ContactID,
			&d.ContactName, &d.ContactAvatarURL, &d.Principal, &d.Currency,
			&d.StartDate, &d.DueDate, &d.InterestRate, &d.InterestRule,
			&d.OutstandingPrincipal, &d.AccruedInterest, &d.Status, &d.ClosedAt,
			&d.OriginatingTransactionID, &d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, nil
}


func (r *DebtRepo) UpdateDebtTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, d entity.Debt) error {
	_, err := tx.Exec(ctx, `
		UPDATE debts
		SET name = $1, due_date = $2::date, status = $3, interest_rate = $4::numeric, 
		    principal = $5::numeric, outstanding_principal = $6::numeric, accrued_interest = $7::numeric,
		    updated_at = $8, closed_at = $9
		WHERE id = $10 AND user_id = $11 AND deleted_at IS NULL
	`, d.Name, d.DueDate, d.Status, d.InterestRate, 
		d.Principal, d.OutstandingPrincipal, d.AccruedInterest, 
		d.UpdatedAt, d.ClosedAt, d.ID, userID)
	return err
}

func (r *DebtRepo) DeleteDebt(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	return r.SoftDelete(ctx, "debts", id, &userID)
}

func (r *DebtRepo) DeleteDebtsByOriginatingTransactionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, transactionID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		UPDATE debts 
		SET deleted_at = $1 
		WHERE user_id = $2 AND originating_transaction_id = $3 AND deleted_at IS NULL
	`, utils.Now(), userID, transactionID)
	return err
}


func (r *DebtRepo) CreatePaymentLinkTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, link entity.DebtPaymentLink, newPrincipal string, newOutstandingPrincipal string, newAccruedInterest string, newStatus entity.DebtStatus, closedAt *time.Time) error {
	// 1. Verify ownership
	var ok bool
	err := tx.QueryRow(ctx, "SELECT TRUE FROM debts WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL", link.DebtID, userID).Scan(&ok)
	if err != nil {
		return fmt.Errorf("debt ownership verification failed: %w", err)
	}

	// 2. Insert Link
	_, err = tx.Exec(ctx, `
		INSERT INTO debt_payment_links (id, debt_id, transaction_id, principal_paid, interest_paid, created_at, updated_at)
		VALUES ($1, $2, $3, $4::numeric, $5::numeric, $6, $7)
	`, link.ID, link.DebtID, link.TransactionID, link.PrincipalPaid, link.InterestPaid, link.CreatedAt, link.CreatedAt)
	if err != nil {
		return err
	}

	// 3. Update Debt
	_, err = tx.Exec(ctx, `
		UPDATE debts
		SET principal = $1::numeric, outstanding_principal = $2::numeric, accrued_interest = $3::numeric,
			status = $4, closed_at = $5, updated_at = $6
		WHERE id = $7 AND user_id = $8 AND deleted_at IS NULL
	`, newPrincipal, newOutstandingPrincipal, newAccruedInterest, newStatus, closedAt, link.CreatedAt, link.DebtID, userID)
	return err
}

func (r *DebtRepo) ListPaymentLinks(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) ([]entity.DebtPaymentLink, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT l.id, l.debt_id, l.transaction_id, l.principal_paid::text, l.interest_paid::text, l.created_at
		FROM debt_payment_links l
		JOIN debts d ON d.id = l.debt_id
		WHERE l.debt_id = $1 AND d.user_id = $2 AND d.deleted_at IS NULL
		ORDER BY l.created_at DESC
	`, debtID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.DebtPaymentLink
	for rows.Next() {
		var l entity.DebtPaymentLink
		err := rows.Scan(&l.ID, &l.DebtID, &l.TransactionID, &l.PrincipalPaid, &l.InterestPaid, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		results = append(results, l)
	}
	return results, nil
}

func (r *DebtRepo) ListPaymentLinksByTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) ([]entity.DebtPaymentLink, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT l.id, l.debt_id, l.transaction_id, l.principal_paid::text, l.interest_paid::text, l.created_at
		FROM debt_payment_links l
		JOIN debts d ON d.id = l.debt_id
		WHERE d.user_id = $1 AND l.transaction_id = $2 AND d.deleted_at IS NULL
	`, userID, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.DebtPaymentLink
	for rows.Next() {
		var l entity.DebtPaymentLink
		err := rows.Scan(&l.ID, &l.DebtID, &l.TransactionID, &l.PrincipalPaid, &l.InterestPaid, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		results = append(results, l)
	}
	return results, nil
}

func (r *DebtRepo) CreateInstallment(ctx context.Context, userID uuid.UUID, inst entity.DebtInstallment) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO debt_installments (id, debt_id, installment_no, due_date, amount_due, amount_paid, status)
		SELECT $1, $2, $3, $4::date, $5::numeric, $6::numeric, $7
		WHERE EXISTS (SELECT 1 FROM debts WHERE id = $2 AND user_id = $8 AND deleted_at IS NULL)
	`, inst.ID, inst.DebtID, inst.InstallmentNo, inst.DueDate, inst.AmountDue, inst.AmountPaid, inst.Status, userID)
	return err
}

func (r *DebtRepo) ListInstallments(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) ([]entity.DebtInstallment, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT i.id, i.debt_id, i.installment_no, to_char(i.due_date, 'YYYY-MM-DD'), i.amount_due::text, i.amount_paid::text, i.status
		FROM debt_installments i
		JOIN debts d ON d.id = i.debt_id
		WHERE i.debt_id = $1 AND d.user_id = $2 AND d.deleted_at IS NULL
		ORDER BY i.installment_no ASC
	`, debtID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.DebtInstallment
	for rows.Next() {
		var i entity.DebtInstallment
		err := rows.Scan(&i.ID, &i.DebtID, &i.InstallmentNo, &i.DueDate, &i.AmountDue, &i.AmountPaid, &i.Status)
		if err != nil {
			return nil, err
		}
		results = append(results, i)
	}
	return results, nil
}

func (r *DebtRepo) DeletePaymentLinksByTransactionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, transactionID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		DELETE FROM debt_payment_links 
		WHERE id IN (
			SELECT l.id FROM debt_payment_links l
			JOIN debts d ON d.id = l.debt_id
			WHERE d.user_id = $1 AND l.transaction_id = $2
		)
	`, userID, transactionID)
	return err
}
