package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type RotatingSavingsRepo struct {
	db *Postgres
}

func NewRotatingSavingsRepo(db *Postgres) *RotatingSavingsRepo {
	return &RotatingSavingsRepo{db: db}
}

func (r *RotatingSavingsRepo) CreateGroup(ctx context.Context, g domain.RotatingSavingsGroup) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO rotating_savings_groups (
			id, user_id, account_id, name, member_count, user_slots,
			contribution_amount, payout_cycle_no,
			fixed_interest_amount, cycle_frequency, start_date, status,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`,
		g.ID,
		g.UserID,
		g.AccountID,
		g.Name,
		g.MemberCount,
		g.UserSlots,
		g.ContributionAmount,
		g.PayoutCycleNo,
		g.FixedInterestAmount,
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
			g.id, g.user_id, g.account_id, g.name, a.currency, g.member_count, g.user_slots,
			g.contribution_amount,
			g.payout_cycle_no,
			g.fixed_interest_amount,
			g.cycle_frequency::text,
			to_char(g.start_date, 'YYYY-MM-DD'),
			g.status::text,
			g.created_at, g.updated_at
		FROM rotating_savings_groups g
		JOIN accounts a ON a.id = g.account_id
		WHERE g.id = $1 AND g.user_id = $2
	`, groupID, userID)

	var g domain.RotatingSavingsGroup
	var interestNull sql.NullFloat64
	var payoutNull sql.NullInt32

	if err := row.Scan(
		&g.ID,
		&g.UserID,
		&g.AccountID,
		&g.Name,
		&g.Currency,
		&g.MemberCount,
		&g.UserSlots,
		&g.ContributionAmount,
		&payoutNull,
		&interestNull,
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

	if payoutNull.Valid {
		v := int(payoutNull.Int32)
		g.PayoutCycleNo = &v
	}
	if interestNull.Valid {
		g.FixedInterestAmount = &interestNull.Float64
	}

	return &g, nil
}

func (r *RotatingSavingsRepo) UpdateGroup(ctx context.Context, g domain.RotatingSavingsGroup) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		UPDATE rotating_savings_groups SET
			account_id = $1,
			name = $2,
			member_count = $3,
			user_slots = $4,
			contribution_amount = $5,
			payout_cycle_no = $6,
			fixed_interest_amount = $7,
			cycle_frequency = $8,
			start_date = $9,
			status = $10,
			updated_at = $11
		WHERE id = $12
	`,
		g.AccountID,
		g.Name,
		g.MemberCount,
		g.UserSlots,
		g.ContributionAmount,
		g.PayoutCycleNo,
		g.FixedInterestAmount,
		g.CycleFrequency,
		g.StartDate,
		g.Status,
		g.UpdatedAt,
		g.ID,
	)
	return err
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
			g.id, g.user_id, g.account_id, g.name, a.currency, g.member_count, g.user_slots,
			g.contribution_amount,
			g.payout_cycle_no,
			g.fixed_interest_amount,
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
		var interestNull sql.NullFloat64
		var payoutNull sql.NullInt32

		if err := rows.Scan(
			&g.ID,
			&g.UserID,
			&g.AccountID,
			&g.Name,
			&g.Currency,
			&g.MemberCount,
			&g.UserSlots,
			&g.ContributionAmount,
			&payoutNull,
			&interestNull,
			&g.CycleFrequency,
			&g.StartDate,
			&g.Status,
			&g.CreatedAt,
			&g.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if payoutNull.Valid {
			v := int(payoutNull.Int32)
			g.PayoutCycleNo = &v
		}
		if interestNull.Valid {
			g.FixedInterestAmount = &interestNull.Float64
		}

		out = append(out, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *RotatingSavingsRepo) CreateContribution(ctx context.Context, c domain.RotatingSavingsContribution) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO rotating_savings_contributions (
			id, group_id, transaction_id, kind, cycle_no, due_date, amount, slots_taken, collected_fee_per_slot, occurred_at, note, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`,
		c.ID,
		c.GroupID,
		c.TransactionID,
		c.Kind,
		c.CycleNo,
		c.DueDate,
		c.Amount,
		c.SlotsTaken,
		c.CollectedFeePerSlot,
		c.OccurredAt,
		c.Note,
		c.CreatedAt,
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
			c.amount,
			c.slots_taken,
			c.collected_fee_per_slot,
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
			&c.SlotsTaken,
			&c.CollectedFeePerSlot,
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
func (r *RotatingSavingsRepo) GetContribution(ctx context.Context, userID string, id string) (*domain.RotatingSavingsContribution, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
		SELECT
			c.id, c.group_id, c.transaction_id, c.kind::text, c.cycle_no,
			to_char(c.due_date, 'YYYY-MM-DD'), c.amount, c.slots_taken,
			c.collected_fee_per_slot, c.occurred_at, c.note, c.created_at
		FROM rotating_savings_contributions c
		JOIN rotating_savings_groups g ON g.id = c.group_id
		WHERE c.id = $1 AND g.user_id = $2
	`, id, userID)

	var c domain.RotatingSavingsContribution
	var cycleNull sql.NullInt32
	var dueNull sql.NullString
	var noteNull sql.NullString

	if err := row.Scan(
		&c.ID,
		&c.GroupID,
		&c.TransactionID,
		&c.Kind,
		&cycleNull,
		&dueNull,
		&c.Amount,
		&c.SlotsTaken,
		&c.CollectedFeePerSlot,
		&c.OccurredAt,
		&noteNull,
		&c.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrRotatingSavingsContributionNotFound
		}
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

	return &c, nil
}

func (r *RotatingSavingsRepo) DeleteContribution(ctx context.Context, userID string, id string) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM rotating_savings_contributions
		WHERE id = $1 AND group_id IN (
			SELECT id FROM rotating_savings_groups WHERE user_id = $2
		)
	`, id, userID)
	return err
}
func (r *RotatingSavingsRepo) DeleteGroup(ctx context.Context, userID string, groupID string) error {
	if r.db == nil {
		return apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM rotating_savings_groups
		WHERE id = $1 AND user_id = $2
	`, groupID, userID)
	return err
}
