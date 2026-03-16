package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type ReportRepo struct {
	db *Postgres
}

func NewReportRepo(db *Postgres) *ReportRepo {
	return &ReportRepo{db: db}
}

func (r *ReportRepo) GetCashflow(ctx context.Context, userID string, months int) ([]domain.CashflowStat, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	startDate := time.Now().UTC().AddDate(0, -months, 0)
	
	rows, err := pool.Query(ctx, `
		SELECT
			to_char(t.occurred_at AT TIME ZONE 'UTC', 'YYYY-MM') AS month,
			t.type,
			SUM(t.amount)::text AS total
		FROM transactions t
		WHERE t.deleted_at IS NULL 
		  AND t.type IN ('income', 'expense')
		  AND t.occurred_at >= $2
		  AND EXISTS (
		      SELECT 1 FROM user_accounts ua
		      WHERE ua.user_id = $1 AND ua.account_id = t.account_id AND ua.status = 'active'
		  )
		GROUP BY month, t.type
		ORDER BY month ASC
	`, userID, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Aggregate into mapping by month while preserving order
	var order []string
	dataMap := make(map[string]*domain.CashflowStat)
	for rows.Next() {
		var month string
		var tType string
		var total string
		if err := rows.Scan(&month, &tType, &total); err != nil {
			return nil, err
		}
		
		stat, ok := dataMap[month]
		if !ok {
			stat = &domain.CashflowStat{Month: month, Income: "0", Expense: "0"}
			dataMap[month] = stat
			order = append(order, month)
		}
		if tType == "income" {
			stat.Income = total
		} else if tType == "expense" {
			stat.Expense = total
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var result []domain.CashflowStat
	for _, m := range order {
		result = append(result, *dataMap[m])
	}
	
	return result, nil
}

func (r *ReportRepo) GetTopExpenses(ctx context.Context, userID string, year int, month int, limit int) ([]domain.CategoryExpenseStat, error) {
	if r.db == nil {
		return nil, apperrors.ErrDatabaseNotReady
	}
	pool, err := r.db.Pool(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := pool.Query(ctx, `
		SELECT
			COALESCE(tli.category_id, '') AS category_id,
			SUM(tli.amount)::text AS total
		FROM transaction_line_items tli
		JOIN transactions t ON t.id = tli.transaction_id
		WHERE t.deleted_at IS NULL
		  AND t.type = 'expense'
		  AND EXTRACT(YEAR FROM t.occurred_at AT TIME ZONE 'UTC') = $2
		  AND EXTRACT(MONTH FROM t.occurred_at AT TIME ZONE 'UTC') = $3
		  AND tli.category_id IS NOT NULL
		  AND EXISTS (
		      SELECT 1 FROM user_accounts ua
		      WHERE ua.user_id = $1 AND ua.account_id = t.account_id AND ua.status = 'active'
		  )
		GROUP BY tli.category_id
		ORDER BY SUM(tli.amount) DESC
		LIMIT $4
	`, userID, year, month, limit)

	if err != nil {
		if err == pgx.ErrNoRows {
			return []domain.CategoryExpenseStat{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	var result []domain.CategoryExpenseStat
	for rows.Next() {
		var stat domain.CategoryExpenseStat
		if err := rows.Scan(&stat.CategoryID, &stat.Amount); err != nil {
			return nil, err
		}
		result = append(result, stat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

