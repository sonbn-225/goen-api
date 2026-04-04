package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

type RotatingSavingsRepo struct {
	db *database.Postgres
}

func NewRotatingSavingsRepo(db *database.Postgres) *RotatingSavingsRepo {
	return &RotatingSavingsRepo{db: db}
}

func (r *RotatingSavingsRepo) CreateGroup(ctx context.Context, g entity.RotatingSavingsGroup) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO rotating_savings_groups (
			id, user_id, account_id, name, currency, member_count, user_slots,
			contribution_amount, payout_cycle_no, fixed_interest_amount,
			cycle_frequency, start_date, status, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
	`,
		g.ID, g.UserID, g.AccountID, g.Name, g.Currency, g.MemberCount, g.UserSlots,
		g.ContributionAmount, g.PayoutCycleNo, g.FixedInterestAmount,
		g.CycleFrequency, g.StartDate, g.Status, g.CreatedAt, g.UpdatedAt,
	)
	return err
}

func (r *RotatingSavingsRepo) GetGroup(ctx context.Context, userID, groupID string) (*entity.RotatingSavingsGroup, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	var g entity.RotatingSavingsGroup
	err = pool.QueryRow(ctx, `
		SELECT
			id, user_id, account_id, name, currency, member_count, user_slots,
			contribution_amount, payout_cycle_no, fixed_interest_amount,
			cycle_frequency, to_char(start_date, 'YYYY-MM-DD'), status, created_at, updated_at
		FROM rotating_savings_groups
		WHERE id = $1 AND user_id = $2
	`, groupID, userID).Scan(
		&g.ID, &g.UserID, &g.AccountID, &g.Name, &g.Currency, &g.MemberCount, &g.UserSlots,
		&g.ContributionAmount, &g.PayoutCycleNo, &g.FixedInterestAmount,
		&g.CycleFrequency, &g.StartDate, &g.Status, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("group not found")
		}
		return nil, err
	}
	return &g, nil
}

func (r *RotatingSavingsRepo) UpdateGroup(ctx context.Context, g entity.RotatingSavingsGroup) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		UPDATE rotating_savings_groups
		SET
			account_id = $3, name = $4, currency = $5, member_count = $6, user_slots = $7,
			contribution_amount = $8, payout_cycle_no = $9, fixed_interest_amount = $10,
			cycle_frequency = $11, start_date = $12, status = $13, updated_at = $14
		WHERE id = $1 AND user_id = $2
	`,
		g.ID, g.UserID, g.AccountID, g.Name, g.Currency, g.MemberCount, g.UserSlots,
		g.ContributionAmount, g.PayoutCycleNo, g.FixedInterestAmount,
		g.CycleFrequency, g.StartDate, g.Status, g.UpdatedAt,
	)
	return err
}

func (r *RotatingSavingsRepo) DeleteGroup(ctx context.Context, userID, groupID string) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `DELETE FROM rotating_savings_groups WHERE id = $1 AND user_id = $2`, groupID, userID)
	return err
}

func (r *RotatingSavingsRepo) ListGroups(ctx context.Context, userID string) ([]entity.RotatingSavingsGroup, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT
			id, user_id, account_id, name, currency, member_count, user_slots,
			contribution_amount, payout_cycle_no, fixed_interest_amount,
			cycle_frequency, to_char(start_date, 'YYYY-MM-DD'), status, created_at, updated_at
		FROM rotating_savings_groups
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.RotatingSavingsGroup
	for rows.Next() {
		var g entity.RotatingSavingsGroup
		if err := rows.Scan(
			&g.ID, &g.UserID, &g.AccountID, &g.Name, &g.Currency, &g.MemberCount, &g.UserSlots,
			&g.ContributionAmount, &g.PayoutCycleNo, &g.FixedInterestAmount,
			&g.CycleFrequency, &g.StartDate, &g.Status, &g.CreatedAt, &g.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, nil
}

func (r *RotatingSavingsRepo) CreateContribution(ctx context.Context, c entity.RotatingSavingsContribution) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO rotating_savings_contributions (
			id, group_id, transaction_id, kind, cycle_no, due_date,
			amount, slots_taken, collected_fee_per_slot, occurred_at, note, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`,
		c.ID, c.GroupID, c.TransactionID, c.Kind, c.CycleNo, c.DueDate,
		c.Amount, c.SlotsTaken, c.CollectedFeePerSlot, c.OccurredAt, c.Note, c.CreatedAt,
	)
	return err
}

func (r *RotatingSavingsRepo) GetContribution(ctx context.Context, userID, id string) (*entity.RotatingSavingsContribution, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	var c entity.RotatingSavingsContribution
	err = pool.QueryRow(ctx, `
		SELECT
			c.id, c.group_id, c.transaction_id, c.kind, c.cycle_no, to_char(c.due_date, 'YYYY-MM-DD'),
			c.amount, c.slots_taken, c.collected_fee_per_slot, c.occurred_at, c.note, c.created_at
		FROM rotating_savings_contributions c
		JOIN rotating_savings_groups g ON g.id = c.group_id
		WHERE c.id = $1 AND g.user_id = $2
	`, id, userID).Scan(
		&c.ID, &c.GroupID, &c.TransactionID, &c.Kind, &c.CycleNo, &c.DueDate,
		&c.Amount, &c.SlotsTaken, &c.CollectedFeePerSlot, &c.OccurredAt, &c.Note, &c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("contribution not found")
		}
		return nil, err
	}
	return &c, nil
}

func (r *RotatingSavingsRepo) ListContributions(ctx context.Context, userID, groupID string) ([]entity.RotatingSavingsContribution, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT
			c.id, c.group_id, c.transaction_id, c.kind, c.cycle_no, to_char(c.due_date, 'YYYY-MM-DD'),
			c.amount, c.slots_taken, c.collected_fee_per_slot, c.occurred_at, c.note, c.created_at
		FROM rotating_savings_contributions c
		JOIN rotating_savings_groups g ON g.id = c.group_id
		WHERE c.group_id = $1 AND g.user_id = $2
		ORDER BY c.cycle_no ASC, c.occurred_at ASC
	`, groupID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.RotatingSavingsContribution
	for rows.Next() {
		var c entity.RotatingSavingsContribution
		if err := rows.Scan(
			&c.ID, &c.GroupID, &c.TransactionID, &c.Kind, &c.CycleNo, &c.DueDate,
			&c.Amount, &c.SlotsTaken, &c.CollectedFeePerSlot, &c.OccurredAt, &c.Note, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (r *RotatingSavingsRepo) DeleteContribution(ctx context.Context, userID, id string) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		DELETE FROM rotating_savings_contributions c
		USING rotating_savings_groups g
		WHERE g.id = c.group_id AND g.user_id = $2 AND c.id = $1
	`, id, userID)
	return err
}

func (r *RotatingSavingsRepo) CreateAuditLog(ctx context.Context, log entity.RotatingSavingsAuditLog) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO rotating_savings_audit_logs (
			id, user_id, group_id, action, details, created_at
		) VALUES ($1,$2,$3,$4,$5,$6)
	`, log.ID, log.UserID, log.GroupID, log.Action, log.Details, log.CreatedAt)
	return err
}

func (r *RotatingSavingsRepo) ListAuditLogs(ctx context.Context, userID, groupID string) ([]entity.RotatingSavingsAuditLog, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT id, user_id, group_id, action, details, created_at
		FROM rotating_savings_audit_logs
		WHERE user_id = $1 AND group_id = $2
		ORDER BY created_at DESC
	`, userID, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.RotatingSavingsAuditLog
	for rows.Next() {
		var log entity.RotatingSavingsAuditLog
		if err := rows.Scan(&log.ID, &log.UserID, &log.GroupID, &log.Action, &log.Details, &log.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, log)
	}
	return out, nil
}
