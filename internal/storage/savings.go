package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/apperrors"
)

type SavingsRepo struct {
	db *Postgres
}

func NewSavingsRepo(db *Postgres) *SavingsRepo {
	return &SavingsRepo{db: db}
}

func (r *SavingsRepo) CreateSavingsInstrument(ctx context.Context, userID string, s domain.SavingsInstrument) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
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
		s.ID,
		s.SavingsAccountID,
		s.ParentAccountID,
		s.Principal,
		s.InterestRate,
		s.TermMonths,
		s.StartDate,
		s.MaturityDate,
		s.AutoRenew,
		s.AccruedInterest,
		s.Status,
		s.ClosedAt,
		s.CreatedAt,
		s.UpdatedAt,
	)
	return err
}

func (r *SavingsRepo) GetSavingsInstrument(ctx context.Context, userID string, savingsInstrumentID string) (*domain.SavingsInstrument, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT
			si.id,
			si.savings_account_id,
			si.parent_account_id,
			si.principal::text,
			CASE WHEN si.interest_rate IS NULL THEN NULL ELSE si.interest_rate::text END,
			si.term_months,
			CASE WHEN si.start_date IS NULL THEN NULL ELSE to_char(si.start_date, 'YYYY-MM-DD') END,
			CASE WHEN si.maturity_date IS NULL THEN NULL ELSE to_char(si.maturity_date, 'YYYY-MM-DD') END,
			si.auto_renew,
			si.accrued_interest::text,
			si.status::text,
			si.closed_at,
			si.created_at,
			si.updated_at
		FROM savings_instruments si
		JOIN accounts a ON a.id = si.savings_account_id
		JOIN user_accounts ua ON ua.account_id = si.savings_account_id
		WHERE si.id = $1 AND ua.user_id = $2 AND ua.status = 'active' AND a.deleted_at IS NULL
	`, savingsInstrumentID, userID)

	var s domain.SavingsInstrument
	var interestNull sql.NullString
	var termNull sql.NullInt32
	var startNull sql.NullString
	var maturityNull sql.NullString

	if err := row.Scan(
		&s.ID,
		&s.SavingsAccountID,
		&s.ParentAccountID,
		&s.Principal,
		&interestNull,
		&termNull,
		&startNull,
		&maturityNull,
		&s.AutoRenew,
		&s.AccruedInterest,
		&s.Status,
		&s.ClosedAt,
		&s.CreatedAt,
		&s.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrSavingsInstrumentNotFound
		}
		return nil, err
	}

	if interestNull.Valid {
		s.InterestRate = &interestNull.String
	}
	if termNull.Valid {
		v := int(termNull.Int32)
		s.TermMonths = &v
	}
	if startNull.Valid {
		s.StartDate = &startNull.String
	}
	if maturityNull.Valid {
		s.MaturityDate = &maturityNull.String
	}

	return &s, nil
}

func (r *SavingsRepo) ListSavingsInstruments(ctx context.Context, userID string) ([]domain.SavingsInstrument, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT
			si.id,
			si.savings_account_id,
			si.parent_account_id,
			si.principal::text,
			CASE WHEN si.interest_rate IS NULL THEN NULL ELSE si.interest_rate::text END,
			si.term_months,
			CASE WHEN si.start_date IS NULL THEN NULL ELSE to_char(si.start_date, 'YYYY-MM-DD') END,
			CASE WHEN si.maturity_date IS NULL THEN NULL ELSE to_char(si.maturity_date, 'YYYY-MM-DD') END,
			si.auto_renew,
			si.accrued_interest::text,
			si.status::text,
			si.closed_at,
			si.created_at,
			si.updated_at
		FROM savings_instruments si
		JOIN accounts a ON a.id = si.savings_account_id
		JOIN user_accounts ua ON ua.account_id = si.savings_account_id
		WHERE ua.user_id = $1 AND ua.status = 'active' AND a.deleted_at IS NULL
		ORDER BY si.created_at DESC, si.id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.SavingsInstrument, 0)
	for rows.Next() {
		var s domain.SavingsInstrument
		var interestNull sql.NullString
		var termNull sql.NullInt32
		var startNull sql.NullString
		var maturityNull sql.NullString

		if err := rows.Scan(
			&s.ID,
			&s.SavingsAccountID,
			&s.ParentAccountID,
			&s.Principal,
			&interestNull,
			&termNull,
			&startNull,
			&maturityNull,
			&s.AutoRenew,
			&s.AccruedInterest,
			&s.Status,
			&s.ClosedAt,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if interestNull.Valid {
			s.InterestRate = &interestNull.String
		}
		if termNull.Valid {
			v := int(termNull.Int32)
			s.TermMonths = &v
		}
		if startNull.Valid {
			s.StartDate = &startNull.String
		}
		if maturityNull.Valid {
			s.MaturityDate = &maturityNull.String
		}

		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *SavingsRepo) UpdateSavingsInstrument(ctx context.Context, userID string, s domain.SavingsInstrument) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	cmd, err := pool.Exec(ctx, `
		UPDATE savings_instruments si
		SET
			principal = $3,
			interest_rate = $4,
			term_months = $5,
			start_date = $6,
			maturity_date = $7,
			auto_renew = $8,
			accrued_interest = $9,
			status = $10,
			closed_at = $11,
			updated_at = $12
		FROM user_accounts ua
		WHERE ua.account_id = si.savings_account_id AND ua.user_id = $2 AND ua.status = 'active'
			AND si.id = $1
	`,
		s.ID,
		userID,
		s.Principal,
		s.InterestRate,
		s.TermMonths,
		s.StartDate,
		s.MaturityDate,
		s.AutoRenew,
		s.AccruedInterest,
		s.Status,
		s.ClosedAt,
		s.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return apperrors.ErrSavingsInstrumentNotFound
	}
	return nil
}

func (r *SavingsRepo) DeleteSavingsInstrument(ctx context.Context, userID string, savingsInstrumentID string) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	cmd, err := pool.Exec(ctx, `
		DELETE FROM savings_instruments si
		USING user_accounts ua
		WHERE ua.account_id = si.savings_account_id AND ua.user_id = $2 AND ua.status = 'active'
			AND si.id = $1
	`, savingsInstrumentID, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return apperrors.ErrSavingsInstrumentNotFound
	}
	return nil
}
