package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

type ReportRepo struct {
	BaseRepo
}

func NewReportRepo(db *database.Postgres) *ReportRepo {
	return &ReportRepo{BaseRepo: *NewBaseRepo(db)}
}

func (r *ReportRepo) GetCashflowTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, months int) ([]entity.CashflowStat, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	// Calculate start date (first day of N months ago)
	startDate := utils.Now().AddDate(0, -months+1, 0)
	startDate = time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)

	rows, err := q.Query(ctx, `
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

	out := make([]entity.CashflowStat, 0)
	for rows.Next() {
		var s entity.CashflowStat
		if err := rows.Scan(&s.Month, &s.Income, &s.Expense); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *ReportRepo) GetTopExpensesTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, year int, month int, limit int) ([]entity.CategoryExpenseStat, error) {
	q, err := r.Queryer(ctx, tx)
	if err != nil {
		return nil, err
	}

	rows, err := q.Query(ctx, `
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
			return []entity.CategoryExpenseStat{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	var result []entity.CategoryExpenseStat
	for rows.Next() {
		var stat entity.CategoryExpenseStat
		if err := rows.Scan(&stat.CategoryID, &stat.Amount); err != nil {
			return nil, err
		}
		result = append(result, stat)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if result == nil {
		result = make([]entity.CategoryExpenseStat, 0)
	}

	return result, nil
}
