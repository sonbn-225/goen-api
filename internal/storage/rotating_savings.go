package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/apperrors"
)

type RotatingSavingsRepo struct {
	db *Postgres
}

func NewRotatingSavingsRepo(db *Postgres) *RotatingSavingsRepo {
	return &RotatingSavingsRepo{db: db}
}

func (r *RotatingSavingsRepo) CreateGroup(ctx context.Context, userID string, g domain.RotatingSavingsGroup) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO rotating_savings_groups (
			id, user_id, self_label, account_id, name, member_count,
			contribution_amount, early_payout_fee_rate, cycle_frequency, start_date, status,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`,
		g.ID,
		g.UserID,
		g.SelfLabel,
		g.AccountID,
		g.Name,
		g.MemberCount,
		g.ContributionAmount,
		g.EarlyPayoutFeeRate,
		g.CycleFrequency,
		g.StartDate,
		g.Status,
		g.CreatedAt,
		g.UpdatedAt,
	)
	return err
}

func (r *RotatingSavingsRepo) GetGroup(ctx context.Context, userID string, groupID string) (*domain.RotatingSavingsGroup, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT
			g.id, g.user_id, g.self_label, g.account_id, g.name, a.currency, g.member_count,
			g.contribution_amount::text,
			CASE WHEN g.early_payout_fee_rate IS NULL THEN NULL ELSE g.early_payout_fee_rate::text END,
			g.cycle_frequency::text,
			to_char(g.start_date, 'YYYY-MM-DD'),
			g.status::text,
			g.created_at, g.updated_at
		FROM rotating_savings_groups g
		JOIN accounts a ON a.id = g.account_id
		WHERE g.id = $1 AND g.user_id = $2
	`, groupID, userID)

	var g domain.RotatingSavingsGroup
	var selfNull sql.NullString
	var earlyNull sql.NullString

	if err := row.Scan(
		&g.ID,
		&g.UserID,
		&selfNull,
		&g.AccountID,
		&g.Name,
		&g.Currency,
		&g.MemberCount,
		&g.ContributionAmount,
		&earlyNull,
		&g.CycleFrequency,
		&g.StartDate,
		&g.Status,
		&g.CreatedAt,
		&g.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrRotatingSavingsGroupNotFound
		}
		return nil, err
	}

	if selfNull.Valid {
		g.SelfLabel = &selfNull.String
	}
	if earlyNull.Valid {
		g.EarlyPayoutFeeRate = &earlyNull.String
	}

	return &g, nil
}

func (r *RotatingSavingsRepo) ListGroups(ctx context.Context, userID string) ([]domain.RotatingSavingsGroup, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT
			g.id, g.user_id, g.self_label, g.account_id, g.name, a.currency, g.member_count,
			g.contribution_amount::text,
			CASE WHEN g.early_payout_fee_rate IS NULL THEN NULL ELSE g.early_payout_fee_rate::text END,
			g.cycle_frequency::text,
			to_char(g.start_date, 'YYYY-MM-DD'),
			g.status::text,
			g.created_at, g.updated_at
		FROM rotating_savings_groups g
		JOIN accounts a ON a.id = g.account_id
		WHERE g.user_id = $1
		ORDER BY g.start_date DESC, g.id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.RotatingSavingsGroup, 0)
	for rows.Next() {
		var g domain.RotatingSavingsGroup
		var selfNull sql.NullString
		var earlyNull sql.NullString

		if err := rows.Scan(
			&g.ID,
			&g.UserID,
			&selfNull,
			&g.AccountID,
			&g.Name,
			&g.Currency,
			&g.MemberCount,
			&g.ContributionAmount,
			&earlyNull,
			&g.CycleFrequency,
			&g.StartDate,
			&g.Status,
			&g.CreatedAt,
			&g.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if selfNull.Valid {
			g.SelfLabel = &selfNull.String
		}
		if earlyNull.Valid {
			g.EarlyPayoutFeeRate = &earlyNull.String
		}

		out = append(out, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *RotatingSavingsRepo) CreateContribution(ctx context.Context, userID string, c domain.RotatingSavingsContribution) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	// Ensure the group belongs to the user.
	_, err = pool.Exec(ctx, `
		INSERT INTO rotating_savings_contributions (
			id, group_id, transaction_id, kind, cycle_no, due_date, amount, occurred_at, note, created_at
		)
		SELECT $1,$2,$3,$4,$5,$6,$7,$8,$9,$10
		WHERE EXISTS (
			SELECT 1 FROM rotating_savings_groups g WHERE g.id = $2 AND g.user_id = $11
		)
	`,
		c.ID,
		c.GroupID,
		c.TransactionID,
		c.Kind,
		c.CycleNo,
		c.DueDate,
		c.Amount,
		c.OccurredAt,
		c.Note,
		c.CreatedAt,
		userID,
	)
	return err
}

func (r *RotatingSavingsRepo) ListContributions(ctx context.Context, userID string, groupID string) ([]domain.RotatingSavingsContribution, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT
			c.id,
			c.group_id,
			c.transaction_id,
			c.kind::text,
			c.cycle_no,
			CASE WHEN c.due_date IS NULL THEN NULL ELSE to_char(c.due_date, 'YYYY-MM-DD') END,
			c.amount::text,
			c.occurred_at,
			c.note,
			c.created_at
		FROM rotating_savings_contributions c
		JOIN rotating_savings_groups g ON g.id = c.group_id
		WHERE c.group_id = $1 AND g.user_id = $2
		ORDER BY c.occurred_at DESC, c.id DESC
	`, groupID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.RotatingSavingsContribution, 0)
	for rows.Next() {
		var c domain.RotatingSavingsContribution
		var cycleNull sql.NullInt32
		var dueNull sql.NullString
		var noteNull sql.NullString

		if err := rows.Scan(
			&c.ID,
			&c.GroupID,
			&c.TransactionID,
			&c.Kind,
			&cycleNull,
			&dueNull,
			&c.Amount,
			&c.OccurredAt,
			&noteNull,
			&c.CreatedAt,
		); err != nil {
			return nil, err
		}
		if cycleNull.Valid {
			v := int(cycleNull.Int32)
			c.CycleNo = &v
		}
		if dueNull.Valid {
			c.DueDate = &dueNull.String
		}
		if noteNull.Valid {
			c.Note = &noteNull.String
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
