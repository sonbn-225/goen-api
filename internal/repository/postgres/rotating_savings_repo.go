package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
)

type RotatingSavingsRepo struct {
	BaseRepo
}

func NewRotatingSavingsRepo(db *database.Postgres) *RotatingSavingsRepo {
	return &RotatingSavingsRepo{BaseRepo: *NewBaseRepo(db)}
}

// --- Nhóm 1: Quản lý Nhóm (Flexible Tx) ---

func (r *RotatingSavingsRepo) GetRotatingGroupTx(ctx context.Context, tx pgx.Tx, userID, groupID uuid.UUID) (*entity.RotatingSavingsGroup, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	var g entity.RotatingSavingsGroup
	err = q.QueryRow(ctx, `
		SELECT
			rg.id, rg.user_id, rg.account_id, rg.name, a.currency, rg.member_count, rg.user_slots,
			rg.contribution_amount, rg.payout_cycle_no, rg.fixed_interest_amount,
			rg.cycle_frequency, to_char(rg.start_date, 'YYYY-MM-DD'), rg.status, rg.created_at, rg.updated_at
		FROM rotating_savings_groups rg
		JOIN accounts a ON a.id = rg.account_id
		WHERE rg.id = $1 AND rg.user_id = $2
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

func (r *RotatingSavingsRepo) ListRotatingGroupsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]entity.RotatingSavingsGroup, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, `
		SELECT
			rg.id, rg.user_id, rg.account_id, rg.name, a.currency, rg.member_count, rg.user_slots,
			rg.contribution_amount, rg.payout_cycle_no, rg.fixed_interest_amount,
			rg.cycle_frequency, to_char(rg.start_date, 'YYYY-MM-DD'), rg.status, rg.created_at, rg.updated_at
		FROM rotating_savings_groups rg
		JOIN accounts a ON a.id = rg.account_id
		WHERE rg.user_id = $1
		ORDER BY rg.created_at DESC
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

// --- Nhóm 2: Thao tác Nhóm (Transactional) ---

func (r *RotatingSavingsRepo) CreateRotatingGroupTx(ctx context.Context, tx pgx.Tx, g entity.RotatingSavingsGroup) error {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, `
		INSERT INTO rotating_savings_groups (
			id, user_id, account_id, name, member_count, user_slots,
			contribution_amount, payout_cycle_no, fixed_interest_amount,
			cycle_frequency, start_date, status, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`,
		g.ID, g.UserID, g.AccountID, g.Name, g.MemberCount, g.UserSlots,
		g.ContributionAmount, g.PayoutCycleNo, g.FixedInterestAmount,
		g.CycleFrequency, g.StartDate, g.Status, g.CreatedAt, g.UpdatedAt,
	)
	return err
}

func (r *RotatingSavingsRepo) UpdateRotatingGroupTx(ctx context.Context, tx pgx.Tx, g entity.RotatingSavingsGroup) error {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, `
		UPDATE rotating_savings_groups
		SET
			account_id = $3,
			name = $4,
			member_count = $5,
			user_slots = $6,
			contribution_amount = $7,
			payout_cycle_no = $8,
			fixed_interest_amount = $9,
			cycle_frequency = $10,
			start_date = $11,
			status = $12,
			updated_at = $13
		WHERE id = $1 AND user_id = $2
	`,
		g.ID, g.UserID, g.AccountID, g.Name, g.MemberCount, g.UserSlots,
		g.ContributionAmount, g.PayoutCycleNo, g.FixedInterestAmount,
		g.CycleFrequency, g.StartDate, g.Status, g.UpdatedAt,
	)
	return err
}

func (r *RotatingSavingsRepo) DeleteRotatingGroupTx(ctx context.Context, tx pgx.Tx, userID, groupID uuid.UUID) error {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, `DELETE FROM rotating_savings_groups WHERE id = $1 AND user_id = $2`, groupID, userID)
	return err
}

// --- Nhóm 3: Quản lý Đóng góp (Flexible Tx) ---

func (r *RotatingSavingsRepo) ListContributionsTx(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) ([]entity.RotatingSavingsContribution, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, `
		SELECT
			id, group_id, transaction_id, kind, cycle_no, to_char(due_date, 'YYYY-MM-DD'),
			amount, slots_taken, collected_fee_per_slot, occurred_at, note, created_at
		FROM rotating_savings_contributions
		WHERE group_id = $1
		ORDER BY cycle_no ASC, occurred_at ASC
	`, groupID)
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

// --- Nhóm 4: Thao tác Đóng góp (Transactional) ---

func (r *RotatingSavingsRepo) CreateContributionTx(ctx context.Context, tx pgx.Tx, c entity.RotatingSavingsContribution) error {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, `
		INSERT INTO rotating_savings_contributions (
			id, group_id, transaction_id, kind, cycle_no, due_date,
			amount, slots_taken, collected_fee_per_slot, occurred_at, note, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`,
		c.ID, c.GroupID, c.TransactionID, c.Kind, c.CycleNo, c.DueDate,
		c.Amount, c.SlotsTaken, c.CollectedFeePerSlot, c.OccurredAt, c.Note, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

func (r *RotatingSavingsRepo) DeleteContributionTx(ctx context.Context, tx pgx.Tx, contributionID uuid.UUID) error {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, `DELETE FROM rotating_savings_contributions WHERE id = $1`, contributionID)
	return err
}

func (r *RotatingSavingsRepo) DeleteContributionByTransactionTx(ctx context.Context, tx pgx.Tx, transactionID uuid.UUID) error {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, `
		UPDATE rotating_savings_contributions
		SET deleted_at = NOW()
		WHERE transaction_id = $1 AND deleted_at IS NULL
	`, transactionID)
	return err
}

// --- Nhóm 5: Nhật ký & Kiểm toán (Flexible Tx) ---

func (r *RotatingSavingsRepo) ListAuditLogsTx(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) ([]entity.RotatingSavingsAuditLog, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, `
		SELECT id, user_id, group_id, action, details, created_at
		FROM rotating_savings_audit_logs
		WHERE group_id = $1
		ORDER BY created_at DESC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.RotatingSavingsAuditLog
	for rows.Next() {
		var l entity.RotatingSavingsAuditLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.GroupID, &l.Action, &l.Details, &l.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, nil
}

// --- Nhóm 6: Nhật ký & Kiểm toán (Transactional) ---

func (r *RotatingSavingsRepo) AddAuditLogTx(ctx context.Context, tx pgx.Tx, log entity.RotatingSavingsAuditLog) error {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, `
		INSERT INTO rotating_savings_audit_logs (id, user_id, group_id, action, details, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, log.ID, log.UserID, log.GroupID, log.Action, log.Details, log.CreatedAt, log.UpdatedAt)
	return err
}
