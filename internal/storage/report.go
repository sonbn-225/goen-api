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

	startDate := time.Now().UTC().AddDate(0, -months, 0).Format("2006-01-01")

	rows, err := pool.Query(ctx, `
		WITH RECURSIVE months AS (
			SELECT date_trunc('month', $2::date) as m
			UNION ALL
			SELECT m + interval '1 month'
			FROM months
			WHERE m < date_trunc('month', CURRENT_DATE)
		),
		monthly_data AS (
			SELECT
				date_trunc('month', t.occurred_at) as month,
				SUM(CASE WHEN t.type = 'income' THEN t.amount ELSE 0 END) as income,
				SUM(CASE WHEN t.type = 'expense' THEN t.amount ELSE 0 END) as expense
			FROM transactions t
			WHERE t.deleted_at IS NULL
			  AND t.occurred_at >= $2
			  AND EXISTS (
					SELECT 1 FROM user_accounts ua
					WHERE ua.user_id = $1 AND ua.account_id = t.account_id AND ua.status = 'active'
			  )
			GROUP BY 1
		)
		SELECT
			to_char(m.m, 'YYYY-MM') as month,
			COALESCE(d.income, 0)::text as income,
			COALESCE(d.expense, 0)::text as expense
		FROM months m
		LEFT JOIN monthly_data d ON d.month = m.m
		ORDER BY m.m DESC
	`, userID, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.CashflowStat, 0)
	for rows.Next() {
		var s domain.CashflowStat
		if err := rows.Scan(&s.Month, &s.Income, &s.Expense); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
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
			li.category_id,
			SUM(li.amount)::text as amount
		FROM transaction_line_items li
		JOIN transactions t ON t.id = li.transaction_id
		WHERE t.deleted_at IS NULL
		  AND t.type = 'expense'
		  AND extract(year from t.occurred_at) = $2
		  AND extract(month from t.occurred_at) = $3
		  AND EXISTS (
				SELECT 1 FROM user_accounts ua
				WHERE ua.user_id = $1 AND ua.account_id = t.account_id AND ua.status = 'active'
		  )
		GROUP BY li.category_id
		ORDER BY SUM(li.amount) DESC
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


