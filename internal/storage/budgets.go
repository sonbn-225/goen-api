package storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type BudgetRepo struct {
	db *Postgres
}

func NewBudgetRepo(db *Postgres) *BudgetRepo {
	return &BudgetRepo{db: db}
}

func (r *BudgetRepo) CreateBudget(ctx context.Context, userID string, b domain.Budget) error {
	if r.db == nil {
		return errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return err
	}

	if strings.TrimSpace(userID) == "" {
		return errors.New("userID is required")
	}
	if strings.TrimSpace(b.UserID) == "" {
		b.UserID = userID
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO budgets (
			id, user_id, name, period, period_start, period_end, amount,
			alert_threshold_percent, rollover_mode, category_id, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`,
		b.ID,
		b.UserID,
		b.Name,
		b.Period,
		b.PeriodStart,
		b.PeriodEnd,
		b.Amount,
		b.AlertThresholdPercent,
		b.RolloverMode,
		b.CategoryID,
		b.CreatedAt,
		b.UpdatedAt,
	)
	return err
}

func (r *BudgetRepo) GetBudget(ctx context.Context, userID string, budgetID string) (*domain.Budget, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	row := pool.QueryRow(ctx, `
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

	var b domain.Budget
	var nameNull sql.NullString
	var startNull sql.NullString
	var endNull sql.NullString
	var alertNull sql.NullInt32
	var rolloverNull sql.NullString
	var categoryNull sql.NullString

	if err := row.Scan(
		&b.ID,
		&b.UserID,
		&nameNull,
		&b.Period,
		&startNull,
		&endNull,
		&b.Amount,
		&alertNull,
		&rolloverNull,
		&categoryNull,
		&b.CreatedAt,
		&b.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBudgetNotFound
		}
		return nil, err
	}

	if nameNull.Valid {
		b.Name = &nameNull.String
	}
	if startNull.Valid {
		b.PeriodStart = &startNull.String
	}
	if endNull.Valid {
		b.PeriodEnd = &endNull.String
	}
	if alertNull.Valid {
		v := int(alertNull.Int32)
		b.AlertThresholdPercent = &v
	}
	if rolloverNull.Valid {
		b.RolloverMode = &rolloverNull.String
	}
	if categoryNull.Valid {
		b.CategoryID = &categoryNull.String
	}

	return &b, nil
}

func (r *BudgetRepo) ListBudgets(ctx context.Context, userID string) ([]domain.Budget, error) {
	if r.db == nil {
		return nil, errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
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
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Budget, 0)
	for rows.Next() {
		var b domain.Budget
		var nameNull sql.NullString
		var startNull sql.NullString
		var endNull sql.NullString
		var alertNull sql.NullInt32
		var rolloverNull sql.NullString
		var categoryNull sql.NullString

		if err := rows.Scan(
			&b.ID,
			&b.UserID,
			&nameNull,
			&b.Period,
			&startNull,
			&endNull,
			&b.Amount,
			&alertNull,
			&rolloverNull,
			&categoryNull,
			&b.CreatedAt,
			&b.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if nameNull.Valid {
			b.Name = &nameNull.String
		}
		if startNull.Valid {
			b.PeriodStart = &startNull.String
		}
		if endNull.Valid {
			b.PeriodEnd = &endNull.String
		}
		if alertNull.Valid {
			v := int(alertNull.Int32)
			b.AlertThresholdPercent = &v
		}
		if rolloverNull.Valid {
			b.RolloverMode = &rolloverNull.String
		}
		if categoryNull.Valid {
			b.CategoryID = &categoryNull.String
		}

		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *BudgetRepo) ComputeSpent(ctx context.Context, userID string, categoryID string, startDate string, endDate string) (string, error) {
	if r.db == nil {
		return "", errors.New("database not ready")
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return "", err
	}

	startT, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return "", err
	}
	endT, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return "", err
	}
	endExclusive := time.Date(endT.Year(), endT.Month(), endT.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour)
	startExclusive := time.Date(startT.Year(), startT.Month(), startT.Day(), 0, 0, 0, 0, time.UTC)

	row := pool.QueryRow(ctx, `
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
			SELECT 1 FROM user_accounts ua
			WHERE ua.user_id = $1 AND ua.account_id = t.account_id AND ua.status = 'active'
		  )
		  AND t.occurred_at >= $3
		  AND t.occurred_at < $4
		  AND li.category_id IN (SELECT id FROM cat_tree)
	`, userID, categoryID, startExclusive, endExclusive)

	var spent string
	if err := row.Scan(&spent); err != nil {
		return "", err
	}
	return spent, nil
}
