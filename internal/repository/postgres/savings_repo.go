package postgres

import (
	"context"
	"errors"

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

func (r *SavingsRepo) CreateSavings(ctx context.Context, userID string, s entity.Savings) error {
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

func (r *SavingsRepo) GetSavings(ctx context.Context, userID string, id string) (*entity.Savings, error) {
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

func (r *SavingsRepo) ListSavings(ctx context.Context, userID string) ([]entity.Savings, error) {
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

func (r *SavingsRepo) UpdateSavings(ctx context.Context, userID string, s entity.Savings) error {
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

func (r *SavingsRepo) DeleteSavings(ctx context.Context, userID, id string) error {
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

// Rotating Savings
func (r *SavingsRepo) CreateRotatingGroup(ctx context.Context, g entity.RotatingSavingsGroup) error {
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

func (r *SavingsRepo) GetRotatingGroup(ctx context.Context, userID, groupID string) (*entity.RotatingSavingsGroup, error) {
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

func (r *SavingsRepo) ListRotatingGroups(ctx context.Context, userID string) ([]entity.RotatingSavingsGroup, error) {
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

func (r *SavingsRepo) UpdateRotatingGroup(ctx context.Context, g entity.RotatingSavingsGroup) error {
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

func (r *SavingsRepo) DeleteRotatingGroup(ctx context.Context, userID, groupID string) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `DELETE FROM rotating_savings_groups WHERE id = $1 AND user_id = $2`, groupID, userID)
	return err
}

func (r *SavingsRepo) CreateContribution(ctx context.Context, c entity.RotatingSavingsContribution) error {
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

func (r *SavingsRepo) GetContributions(ctx context.Context, groupID string) ([]entity.RotatingSavingsContribution, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
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

func (r *SavingsRepo) DeleteContribution(ctx context.Context, contributionID string) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `DELETE FROM rotating_savings_contributions WHERE id = $1`, contributionID)
	return err
}

// Audit Logs
func (r *SavingsRepo) AddAuditLog(ctx context.Context, log entity.RotatingSavingsAuditLog) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO rotating_savings_audit_logs (id, user_id, group_id, action, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, log.ID, log.UserID, log.GroupID, log.Action, log.Details, log.CreatedAt)
	return err
}

func (r *SavingsRepo) GetAuditLogs(ctx context.Context, groupID string) ([]entity.RotatingSavingsAuditLog, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
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
