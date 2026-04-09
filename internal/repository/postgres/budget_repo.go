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
)

type BudgetRepo struct {
	BaseRepo
}

func NewBudgetRepo(db *database.Postgres) *BudgetRepo {
	return &BudgetRepo{BaseRepo: *NewBaseRepo(db)}
}

// --- Nhóm 1: Truy vấn Thống kê & Danh sách (Read-only Optimized) ---

// --- Nhóm 1: Truy vấn Thống kê & Danh sách (Flexible Tx) ---

func (r *BudgetRepo) GetBudgetTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, budgetID uuid.UUID) (*entity.Budget, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}
	return r.getBudgetTx(ctx, q, userID, budgetID)
}

func (r *BudgetRepo) ListBudgetsTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]entity.Budget, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, `
		SELECT
			id, user_id, name, period,
			CASE WHEN period_start IS NULL THEN NULL ELSE to_char(period_start, 'YYYY-MM-DD') END,
			CASE WHEN period_end IS NULL THEN NULL ELSE to_char(period_end, 'YYYY-MM-DD') END,
			amount::text, alert_threshold_percent, rollover_mode, category_id,
			created_at, updated_at, deleted_at
		FROM budgets
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC, id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entity.Budget
	for rows.Next() {
		var b entity.Budget
		if err := rows.Scan(
			&b.ID, &b.UserID, &b.Name, &b.Period, &b.PeriodStart, &b.PeriodEnd,
			&b.Amount, &b.AlertThresholdPercent, &b.RolloverMode, &b.CategoryID,
			&b.CreatedAt, &b.UpdatedAt, &b.DeletedAt,
		); err != nil {
			return nil, err
		}

		if b.CategoryID != nil && b.PeriodStart != nil && b.PeriodEnd != nil {
			s, _ := r.ComputeSpentTx(ctx, tx, userID, *b.CategoryID, *b.PeriodStart, *b.PeriodEnd)
			b.Spent = s
		} else {
			b.Spent = "0"
		}
		b.Remaining = "0" // Placeholder

		results = append(results, b)
	}
	return results, nil
}

func (r *BudgetRepo) ComputeSpentTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, categoryID uuid.UUID, startDate string, endDate string) (string, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return "", err
	}

	startT, _ := time.Parse("2006-01-02", startDate)
	endT, _ := time.Parse("2006-01-02", endDate)
	startExclusive := startT.UTC()
	endExclusive := endT.UTC().Add(24 * time.Hour)

	row := q.QueryRow(ctx, `
		WITH RECURSIVE cat_tree AS (
			SELECT id FROM categories WHERE id = $2 AND deleted_at IS NULL
			UNION ALL
			SELECT c.id FROM categories c JOIN cat_tree ct ON c.parent_category_id = ct.id WHERE c.deleted_at IS NULL
		)
		SELECT COALESCE(SUM(li.amount), 0)::text
		FROM transaction_line_items li
		JOIN transactions t ON t.id = li.transaction_id
		WHERE t.deleted_at IS NULL
		  AND t.type = 'expense'
		  AND t.occurred_at >= $3 AND t.occurred_at < $4
		  AND li.category_id IN (SELECT id FROM cat_tree)
		  AND EXISTS (
			SELECT 1 FROM user_accounts ua
			WHERE ua.account_id = t.account_id AND ua.user_id = $1 AND ua.status = 'active'
		  )
	`, userID, categoryID, startExclusive, endExclusive)

	var spent string
	if err := row.Scan(&spent); err != nil {
		return "", err
	}
	return spent, nil
}

// --- Nhóm 2: Thao tác ghi & Nhất quán (Transactional) ---

func (r *BudgetRepo) CreateBudgetTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, b entity.Budget) error {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, `
		INSERT INTO budgets (
			id, user_id, name, period, period_start, period_end, amount,
			alert_threshold_percent, rollover_mode, category_id, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`,
		b.ID, userID, b.Name, b.Period, b.PeriodStart, b.PeriodEnd, b.Amount,
		b.AlertThresholdPercent, b.RolloverMode, b.CategoryID, b.CreatedAt, b.UpdatedAt,
	)
	return err
}

func (r *BudgetRepo) UpdateBudgetTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, b entity.Budget) error {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, `
		UPDATE budgets
		SET name = $1, amount = $2::numeric, alert_threshold_percent = $3, rollover_mode = $4, updated_at = $5
		WHERE id = $6 AND user_id = $7 AND deleted_at IS NULL
	`, b.Name, b.Amount, b.AlertThresholdPercent, b.RolloverMode, b.UpdatedAt, b.ID, userID)
	return err
}

func (r *BudgetRepo) DeleteBudgetTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, budgetID uuid.UUID) error {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return err
	}

	_, err = q.Exec(ctx, `UPDATE budgets SET deleted_at = NOW() WHERE id = $1 AND user_id = $2`, budgetID, userID)
	return err
}

// --- Internal Helpers ---

func (r *BudgetRepo) getBudgetTx(ctx context.Context, q database.Queryer, userID uuid.UUID, budgetID uuid.UUID) (*entity.Budget, error) {
	row := q.QueryRow(ctx, `
		SELECT
			id, user_id, name, period, 
			CASE WHEN period_start IS NULL THEN NULL ELSE to_char(period_start, 'YYYY-MM-DD') END,
			CASE WHEN period_end IS NULL THEN NULL ELSE to_char(period_end, 'YYYY-MM-DD') END,
			amount::text, alert_threshold_percent, rollover_mode, category_id,
			created_at, updated_at, deleted_at
		FROM budgets
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, budgetID, userID)

	var b entity.Budget
	if err := row.Scan(
		&b.ID, &b.UserID, &b.Name, &b.Period, &b.PeriodStart, &b.PeriodEnd,
		&b.Amount, &b.AlertThresholdPercent, &b.RolloverMode, &b.CategoryID,
		&b.CreatedAt, &b.UpdatedAt, &b.DeletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("budget not found")
		}
		return nil, fmt.Errorf("failed to scan budget: %w", err)
	}

	// Compute spent (simplified)
	if b.CategoryID != nil && b.PeriodStart != nil && b.PeriodEnd != nil {
		// Pass q as tx if it's a transaction, but Queryer doesn't easily convert back to pgx.Tx
		// For now, we'll just use the ComputeSpentTx with nil tx if we can't easily get it.
		// Actually, let's just make ComputeSpentTx accept Queryer instead.
		s, _ := r.computeSpentWithQueryer(ctx, q, userID, *b.CategoryID, *b.PeriodStart, *b.PeriodEnd)
		b.Spent = s
	} else {
		b.Spent = "0"
	}
	b.Remaining = "0" // Placeholder
	return &b, nil
}

func (r *BudgetRepo) computeSpentWithQueryer(ctx context.Context, q database.Queryer, userID uuid.UUID, categoryID uuid.UUID, startDate string, endDate string) (string, error) {
	startT, _ := time.Parse("2006-01-02", startDate)
	endT, _ := time.Parse("2006-01-02", endDate)
	startExclusive := startT.UTC()
	endExclusive := endT.UTC().Add(24 * time.Hour)

	row := q.QueryRow(ctx, `
		WITH RECURSIVE cat_tree AS (
			SELECT id FROM categories WHERE id = $2 AND deleted_at IS NULL
			UNION ALL
			SELECT c.id FROM categories c JOIN cat_tree ct ON c.parent_category_id = ct.id WHERE c.deleted_at IS NULL
		)
		SELECT COALESCE(SUM(li.amount), 0)::text
		FROM transaction_line_items li
		JOIN transactions t ON t.id = li.transaction_id
		WHERE t.deleted_at IS NULL
		  AND t.type = 'expense'
		  AND t.occurred_at >= $3 AND t.occurred_at < $4
		  AND li.category_id IN (SELECT id FROM cat_tree)
		  AND EXISTS (
			SELECT 1 FROM user_accounts ua
			WHERE ua.account_id = t.account_id AND ua.user_id = $1 AND ua.status = 'active'
		  )
	`, userID, categoryID, startExclusive, endExclusive)

	var spent string
	if err := row.Scan(&spent); err != nil {
		return "", err
	}
	return spent, nil
}
