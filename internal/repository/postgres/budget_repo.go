package postgres

import (
	"context"
	"errors"
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

func (r *BudgetRepo) CreateBudget(ctx context.Context, userID uuid.UUID, b entity.Budget) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
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

func (r *BudgetRepo) GetBudget(ctx context.Context, userID uuid.UUID, budgetID uuid.UUID) (*entity.Budget, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
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
	var spent string
	if err := row.Scan(
		&b.ID, &b.UserID, &b.Name, &b.Period, &b.PeriodStart, &b.PeriodEnd,
		&b.Amount, &b.AlertThresholdPercent, &b.RolloverMode, &b.CategoryID,
		&b.CreatedAt, &b.UpdatedAt, &b.DeletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("budget not found")
		}
		return nil, err
	}

	// Compute spent (simplified, should use recursive tree if category_id exists)
	if b.CategoryID != nil && b.PeriodStart != nil && b.PeriodEnd != nil {
		s, _ := r.ComputeSpent(ctx, userID, *b.CategoryID, *b.PeriodStart, *b.PeriodEnd)
		spent = s
	} else {
		spent = "0"
	}
	b.Spent = spent
	// Remaining calculation (simplified, assuming numeric amounts)
	b.Remaining = "0" // Placeholder, in real app would use decimal arithmetic
	return &b, nil
}

func (r *BudgetRepo) ListBudgets(ctx context.Context, userID uuid.UUID) ([]entity.Budget, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
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
			s, _ := r.ComputeSpent(ctx, userID, *b.CategoryID, *b.PeriodStart, *b.PeriodEnd)
			b.Spent = s
		} else {
			b.Spent = "0"
		}
		b.Remaining = "0" // Placeholder
		
		results = append(results, b)
	}
	return results, nil
}

func (r *BudgetRepo) UpdateBudget(ctx context.Context, userID uuid.UUID, b entity.Budget) error {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		UPDATE budgets
		SET name = $1, amount = $2::numeric, alert_threshold_percent = $3, rollover_mode = $4, updated_at = $5
		WHERE id = $6 AND user_id = $7 AND deleted_at IS NULL
	`, b.Name, b.Amount, b.AlertThresholdPercent, b.RolloverMode, b.UpdatedAt, b.ID, userID)
	return err
}

func (r *BudgetRepo) DeleteBudget(ctx context.Context, userID uuid.UUID, budgetID uuid.UUID) error {
	return r.SoftDelete(ctx, "budgets", budgetID, &userID)
}

func (r *BudgetRepo) ComputeSpent(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID, startDate string, endDate string) (string, error) {
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return "", err
	}

	startT, _ := time.Parse("2006-01-02", startDate)
	endT, _ := time.Parse("2006-01-02", endDate)
	startExclusive := startT.UTC()
	endExclusive := endT.UTC().Add(24 * time.Hour)

	row := pool.QueryRow(ctx, `
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

