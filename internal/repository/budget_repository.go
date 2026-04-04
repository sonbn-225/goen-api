package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
	"github.com/sonbn-225/goen-api-v2/internal/domains/budget"
)

type BudgetRepository struct {
	db *pgxpool.Pool
}

func NewBudgetRepository(db *pgxpool.Pool) *BudgetRepository {
	return &BudgetRepository{db: db}
}

func (r *BudgetRepository) Create(ctx context.Context, userID string, input budget.Budget) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "budget", "operation", "create", "user_id", userID, "budget_id", input.ID)
	_, err := r.db.Exec(ctx, `
		INSERT INTO budgets (
			id, user_id, name, period, period_start, period_end, amount,
			alert_threshold_percent, rollover_mode, category_id, created_at, updated_at
		) VALUES ($1,$2,$3,$4::budget_period,$5,$6,$7,$8,$9::budget_rollover_mode,$10,$11,$12)
	`,
		input.ID,
		input.UserID,
		input.Name,
		input.Period,
		input.PeriodStart,
		input.PeriodEnd,
		input.Amount,
		input.AlertThresholdPercent,
		input.RolloverMode,
		input.CategoryID,
		input.CreatedAt,
		input.UpdatedAt,
	)
	if err != nil {
		logger.Error("repo_budget_create_failed", "error", err)
		return err
	}
	logger.Info("repo_budget_create_succeeded")
	return nil
}

func (r *BudgetRepository) GetByID(ctx context.Context, userID, budgetID string) (*budget.Budget, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "budget", "operation", "get_by_id", "user_id", userID, "budget_id", budgetID)
	row := r.db.QueryRow(ctx, `
		SELECT
			id,
			user_id,
			name,
			period::text,
			CASE WHEN period_start IS NULL THEN NULL ELSE to_char(period_start, 'YYYY-MM-DD') END,
			CASE WHEN period_end IS NULL THEN NULL ELSE to_char(period_end, 'YYYY-MM-DD') END,
			amount::text,
			alert_threshold_percent,
			CASE WHEN rollover_mode IS NULL THEN NULL ELSE rollover_mode::text END,
			category_id,
			created_at,
			updated_at
		FROM budgets
		WHERE id = $1 AND user_id = $2
	`, budgetID, userID)

	item, err := scanBudget(row)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		logger.Error("repo_budget_get_failed", "error", err)
		return nil, err
	}
	logger.Info("repo_budget_get_succeeded")
	return item, nil
}

func (r *BudgetRepository) ListByUser(ctx context.Context, userID string) ([]budget.Budget, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "budget", "operation", "list_by_user", "user_id", userID)
	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			user_id,
			name,
			period::text,
			CASE WHEN period_start IS NULL THEN NULL ELSE to_char(period_start, 'YYYY-MM-DD') END,
			CASE WHEN period_end IS NULL THEN NULL ELSE to_char(period_end, 'YYYY-MM-DD') END,
			amount::text,
			alert_threshold_percent,
			CASE WHEN rollover_mode IS NULL THEN NULL ELSE rollover_mode::text END,
			category_id,
			created_at,
			updated_at
		FROM budgets
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
	`, userID)
	if err != nil {
		logger.Error("repo_budget_list_failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	items := make([]budget.Budget, 0)
	for rows.Next() {
		item, err := scanBudget(rows)
		if err != nil {
			logger.Error("repo_budget_list_failed", "error", err)
			return nil, err
		}
		items = append(items, *item)
	}

	if err := rows.Err(); err != nil {
		logger.Error("repo_budget_list_failed", "error", err)
		return nil, err
	}

	logger.Info("repo_budget_list_succeeded", "count", len(items))
	return items, nil
}

func (r *BudgetRepository) ComputeSpent(ctx context.Context, userID, categoryID, startDate, endDate string) (string, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "budget", "operation", "compute_spent", "user_id", userID, "category_id", categoryID)

	startT, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		logger.Error("repo_budget_compute_spent_failed", "error", err)
		return "", err
	}
	endT, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		logger.Error("repo_budget_compute_spent_failed", "error", err)
		return "", err
	}

	startInclusive := time.Date(startT.Year(), startT.Month(), startT.Day(), 0, 0, 0, 0, time.UTC)
	endExclusive := time.Date(endT.Year(), endT.Month(), endT.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour)

	row := r.db.QueryRow(ctx, `
		WITH RECURSIVE cat_tree AS (
			SELECT id
			FROM categories
			WHERE id = $2 AND deleted_at IS NULL
			UNION ALL
			SELECT c.id
			FROM categories c
			JOIN cat_tree ct ON c.parent_category_id = ct.id
			WHERE c.deleted_at IS NULL
		)
		SELECT COALESCE(SUM(li.amount), 0)::text AS spent
		FROM transaction_line_items li
		JOIN transactions t ON t.id = li.transaction_id
		WHERE t.deleted_at IS NULL
		  AND t.type = 'expense'
		  AND t.account_id IS NOT NULL
		  AND EXISTS (
			SELECT 1
			FROM user_accounts ua
			WHERE ua.user_id = $1
			  AND ua.account_id = t.account_id
			  AND ua.status = 'active'
		  )
		  AND t.occurred_at >= $3
		  AND t.occurred_at < $4
		  AND li.category_id IN (SELECT id FROM cat_tree)
	`, userID, categoryID, startInclusive, endExclusive)

	var spent string
	if err := row.Scan(&spent); err != nil {
		logger.Error("repo_budget_compute_spent_failed", "error", err)
		return "", err
	}

	logger.Info("repo_budget_compute_spent_succeeded", "spent", spent)
	return spent, nil
}

type budgetScanner interface {
	Scan(dest ...any) error
}

func scanBudget(scanner budgetScanner) (*budget.Budget, error) {
	var item budget.Budget
	var name sql.NullString
	var periodStart sql.NullString
	var periodEnd sql.NullString
	var alert sql.NullInt32
	var rollover sql.NullString
	var categoryID sql.NullString

	err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&name,
		&item.Period,
		&periodStart,
		&periodEnd,
		&item.Amount,
		&alert,
		&rollover,
		&categoryID,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if name.Valid {
		item.Name = &name.String
	}
	if periodStart.Valid {
		item.PeriodStart = &periodStart.String
	}
	if periodEnd.Valid {
		item.PeriodEnd = &periodEnd.String
	}
	if alert.Valid {
		v := int(alert.Int32)
		item.AlertThresholdPercent = &v
	}
	if rollover.Valid {
		item.RolloverMode = &rollover.String
	}
	if categoryID.Valid {
		item.CategoryID = &categoryID.String
	}

	return &item, nil
}
