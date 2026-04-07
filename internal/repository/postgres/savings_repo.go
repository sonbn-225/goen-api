package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

type SavingsRepo struct {
	db *database.Postgres
}

func NewSavingsRepo(db *database.Postgres) *SavingsRepo {
	return &SavingsRepo{db: db}
}

func (r *SavingsRepo) CreateSavings(ctx context.Context, userID uuid.UUID, s entity.Savings) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO savings_instruments (
			id, savings_account_id, parent_account_id, principal, interest_rate, term_months,
			start_date, maturity_date, auto_renew, accrued_interest, status, closed_at,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`,
		s.ID, s.SavingsAccountID, s.ParentAccountID, s.Principal, s.InterestRate, s.TermMonths,
		s.StartDate, s.MaturityDate, s.AutoRenew, s.AccruedInterest, s.Status, s.ClosedAt,
		s.CreatedAt, s.UpdatedAt,
	)
	return err
}

func (r *SavingsRepo) GetSavings(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*entity.Savings, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	var s entity.Savings
	err = pool.QueryRow(ctx, `
		SELECT
			si.id, si.savings_account_id, si.parent_account_id, si.principal::text,
			si.interest_rate::text, si.term_months, to_char(si.start_date, 'YYYY-MM-DD'),
			to_char(si.maturity_date, 'YYYY-MM-DD'), si.auto_renew, si.accrued_interest::text,
			si.status::text, si.closed_at, si.created_at, si.updated_at
		FROM savings_instruments si
		JOIN user_accounts ua ON ua.account_id = si.savings_account_id
		WHERE si.id = $1 AND ua.user_id = $2 AND ua.status = 'active'
	`, id, userID).Scan(
		&s.ID, &s.SavingsAccountID, &s.ParentAccountID, &s.Principal, &s.InterestRate, &s.TermMonths,
		&s.StartDate, &s.MaturityDate, &s.AutoRenew, &s.AccruedInterest, &s.Status, &s.ClosedAt,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("savings not found")
		}
		return nil, err
	}
	return &s, nil
}

func (r *SavingsRepo) ListSavings(ctx context.Context, userID uuid.UUID) ([]entity.Savings, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT
			si.id, si.savings_account_id, si.parent_account_id, si.principal::text,
			si.interest_rate::text, si.term_months, to_char(si.start_date, 'YYYY-MM-DD'),
			to_char(si.maturity_date, 'YYYY-MM-DD'), si.auto_renew, si.accrued_interest::text,
			si.status::text, si.closed_at, si.created_at, si.updated_at
		FROM savings_instruments si
		JOIN user_accounts ua ON ua.account_id = si.savings_account_id
		WHERE ua.user_id = $1 AND ua.status = 'active'
		ORDER BY si.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.Savings
	for rows.Next() {
		var s entity.Savings
		if err := rows.Scan(
			&s.ID, &s.SavingsAccountID, &s.ParentAccountID, &s.Principal, &s.InterestRate, &s.TermMonths,
			&s.StartDate, &s.MaturityDate, &s.AutoRenew, &s.AccruedInterest, &s.Status, &s.ClosedAt,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *SavingsRepo) UpdateSavings(ctx context.Context, userID uuid.UUID, s entity.Savings) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		UPDATE savings_instruments si
		SET
			principal = $3, interest_rate = $4, term_months = $5, start_date = $6,
			maturity_date = $7, auto_renew = $8, accrued_interest = $9, status = $10,
			closed_at = $11, updated_at = $12
		FROM user_accounts ua
		WHERE ua.account_id = si.savings_account_id AND ua.user_id = $2 AND si.id = $1
	`,
		s.ID, userID, s.Principal, s.InterestRate, s.TermMonths, s.StartDate,
		s.MaturityDate, s.AutoRenew, s.AccruedInterest, s.Status, s.ClosedAt, s.UpdatedAt,
	)
	return err
}

func (r *SavingsRepo) DeleteSavings(ctx context.Context, userID, id uuid.UUID) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM savings_instruments si
		USING user_accounts ua
		WHERE ua.account_id = si.savings_account_id AND ua.user_id = $2 AND si.id = $1
	`, id, userID)
	return err
}
