package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type DebtRepo struct {
	db *Postgres
}

func NewDebtRepo(db *Postgres) *DebtRepo {
	return &DebtRepo{db: db}
}

func (r *DebtRepo) CreateDebt(ctx context.Context, debt domain.Debt) error {
	if r.db == nil {
		return errors.New("database not ready")
	}
	if debt.AccountID == nil || *debt.AccountID == "" {
		return domain.ErrAccountInvalidInput
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	tag, err := pool.Exec(ctx, `
		INSERT INTO debts (
			id, client_id, user_id, account_id, direction, name, principal,
			start_date, due_date, interest_rate, interest_rule,
			outstanding_principal, accrued_interest, status, closed_at,
			created_at, updated_at
		)
		SELECT
			$1,$2,$3,$4,$5,$6,$7::numeric,
			$8::date,$9::date,$10::numeric,$11,
			$12::numeric,$13::numeric,$14,$15,
			$16,$17
		WHERE EXISTS (
			SELECT 1
			FROM user_accounts ua
			WHERE ua.user_id = $3 AND ua.account_id = $4 AND ua.status = 'active'
		)
	`,
		debt.ID,
		debt.ClientID,
		debt.UserID,
		debt.AccountID,
		debt.Direction,
		debt.Name,
		debt.Principal,
		debt.StartDate,
		debt.DueDate,
		debt.InterestRate,
		debt.InterestRule,
		debt.OutstandingPrincipal,
		debt.AccruedInterest,
		debt.Status,
		debt.ClosedAt,
		debt.CreatedAt,
		debt.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		// Either account doesn't exist, or user can't access it.
		var exists bool
		if err := pool.QueryRow(ctx, `SELECT TRUE FROM accounts WHERE id = $1`, *debt.AccountID).Scan(&exists); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrAccountNotFound
			}
			return err
		}
		return domain.ErrAccountForbidden
	}
	return nil
}

func (r *DebtRepo) GetDebt(ctx context.Context, userID string, debtID string) (*domain.Debt, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT
			d.id,
			d.client_id,
			d.user_id,
			d.account_id,
			d.direction,
			d.name,
			d.principal::text,
			a.currency,
			to_char(d.start_date, 'YYYY-MM-DD'),
			to_char(d.due_date, 'YYYY-MM-DD'),
			CASE WHEN d.interest_rate IS NULL THEN NULL ELSE d.interest_rate::text END,
			d.interest_rule,
			d.outstanding_principal::text,
			d.accrued_interest::text,
			d.status,
			d.closed_at,
			d.created_at,
			d.updated_at
		FROM debts d
		LEFT JOIN accounts a ON a.id = d.account_id
		WHERE d.id = $1 AND d.user_id = $2
	`, debtID, userID)

	var d domain.Debt
	if err := row.Scan(
		&d.ID,
		&d.ClientID,
		&d.UserID,
		&d.AccountID,
		&d.Direction,
		&d.Name,
		&d.Principal,
		&d.Currency,
		&d.StartDate,
		&d.DueDate,
		&d.InterestRate,
		&d.InterestRule,
		&d.OutstandingPrincipal,
		&d.AccruedInterest,
		&d.Status,
		&d.ClosedAt,
		&d.CreatedAt,
		&d.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDebtNotFound
		}
		return nil, err
	}
	return &d, nil
}

func (r *DebtRepo) ListDebts(ctx context.Context, userID string) ([]domain.Debt, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT
			d.id,
			d.client_id,
			d.user_id,
			d.account_id,
			d.direction,
			d.name,
			d.principal::text,
			a.currency,
			to_char(d.start_date, 'YYYY-MM-DD'),
			to_char(d.due_date, 'YYYY-MM-DD'),
			CASE WHEN d.interest_rate IS NULL THEN NULL ELSE d.interest_rate::text END,
			d.interest_rule,
			d.outstanding_principal::text,
			d.accrued_interest::text,
			d.status,
			d.closed_at,
			d.created_at,
			d.updated_at
		FROM debts d
		LEFT JOIN accounts a ON a.id = d.account_id
		WHERE d.user_id = $1
		ORDER BY d.due_date ASC, d.id ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Debt, 0)
	for rows.Next() {
		var d domain.Debt
		if err := rows.Scan(
			&d.ID,
			&d.ClientID,
			&d.UserID,
			&d.AccountID,
			&d.Direction,
			&d.Name,
			&d.Principal,
			&d.Currency,
			&d.StartDate,
			&d.DueDate,
			&d.InterestRate,
			&d.InterestRule,
			&d.OutstandingPrincipal,
			&d.AccruedInterest,
			&d.Status,
			&d.ClosedAt,
			&d.CreatedAt,
			&d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *DebtRepo) CreatePaymentLink(ctx context.Context, userID string, link domain.DebtPaymentLink, newOutstandingPrincipal string, newAccruedInterest string, newStatus string, closedAt *time.Time) error {
	if r.db == nil {
		return errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	return withTx(ctx, pool, func(dbtx pgx.Tx) error {
		// Ensure debt belongs to user
		var exists bool
		if err := dbtx.QueryRow(ctx, `SELECT TRUE FROM debts WHERE id = $1 AND user_id = $2`, link.DebtID, userID).Scan(&exists); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrDebtNotFound
			}
			return err
		}

		_, err := dbtx.Exec(ctx, `
			INSERT INTO debt_payment_links (id, debt_id, transaction_id, principal_paid, interest_paid, created_at)
			VALUES ($1,$2,$3,$4::numeric,$5::numeric,$6)
		`, link.ID, link.DebtID, link.TransactionID, link.PrincipalPaid, link.InterestPaid, link.CreatedAt)
		if err != nil {
			return err
		}

		_, err = dbtx.Exec(ctx, `
			UPDATE debts
			SET outstanding_principal = $1::numeric,
				accrued_interest = $2::numeric,
				status = $3,
				closed_at = $4,
				updated_at = $5
			WHERE id = $6 AND user_id = $7
		`, newOutstandingPrincipal, newAccruedInterest, newStatus, closedAt, link.CreatedAt, link.DebtID, userID)
		return err
	})
}

func (r *DebtRepo) ListPaymentLinks(ctx context.Context, userID string, debtID string) ([]domain.DebtPaymentLink, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	// Ensure debt belongs to user
	if _, err := r.GetDebt(ctx, userID, debtID); err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT id, debt_id, transaction_id,
			CASE WHEN principal_paid IS NULL THEN NULL ELSE principal_paid::text END,
			CASE WHEN interest_paid IS NULL THEN NULL ELSE interest_paid::text END,
			created_at
		FROM debt_payment_links
		WHERE debt_id = $1
		ORDER BY created_at DESC, id DESC
	`, debtID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.DebtPaymentLink, 0)
	for rows.Next() {
		var it domain.DebtPaymentLink
		if err := rows.Scan(&it.ID, &it.DebtID, &it.TransactionID, &it.PrincipalPaid, &it.InterestPaid, &it.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *DebtRepo) ListPaymentLinksByTransaction(ctx context.Context, userID string, transactionID string) ([]domain.DebtPaymentLink, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT l.id, l.debt_id, l.transaction_id,
			CASE WHEN l.principal_paid IS NULL THEN NULL ELSE l.principal_paid::text END,
			CASE WHEN l.interest_paid IS NULL THEN NULL ELSE l.interest_paid::text END,
			l.created_at
		FROM debt_payment_links l
		JOIN debts d ON d.id = l.debt_id
		WHERE d.user_id = $1 AND l.transaction_id = $2
		ORDER BY l.created_at DESC, l.id DESC
	`, userID, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.DebtPaymentLink, 0)
	for rows.Next() {
		var it domain.DebtPaymentLink
		if err := rows.Scan(&it.ID, &it.DebtID, &it.TransactionID, &it.PrincipalPaid, &it.InterestPaid, &it.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *DebtRepo) CreateInstallment(ctx context.Context, userID string, inst domain.DebtInstallment) error {
	if r.db == nil {
		return errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	// Ensure debt belongs to user
	if _, err := r.GetDebt(ctx, userID, inst.DebtID); err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO debt_installments (id, debt_id, installment_no, due_date, amount_due, amount_paid, status)
		VALUES ($1,$2,$3,$4::date,$5::numeric,$6::numeric,$7)
	`, inst.ID, inst.DebtID, inst.InstallmentNo, inst.DueDate, inst.AmountDue, inst.AmountPaid, inst.Status)
	return err
}

func (r *DebtRepo) ListInstallments(ctx context.Context, userID string, debtID string) ([]domain.DebtInstallment, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	// Ensure debt belongs to user
	if _, err := r.GetDebt(ctx, userID, debtID); err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT id, debt_id, installment_no, to_char(due_date, 'YYYY-MM-DD'), amount_due::text, amount_paid::text, status
		FROM debt_installments
		WHERE debt_id = $1
		ORDER BY installment_no ASC
	`, debtID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.DebtInstallment, 0)
	for rows.Next() {
		var it domain.DebtInstallment
		if err := rows.Scan(&it.ID, &it.DebtID, &it.InstallmentNo, &it.DueDate, &it.AmountDue, &it.AmountPaid, &it.Status); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
