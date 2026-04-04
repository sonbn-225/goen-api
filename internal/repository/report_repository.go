package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
	"github.com/sonbn-225/goen-api-v2/internal/domains/report"
)

type ReportRepository struct {
	db *pgxpool.Pool
}

var _ report.Repository = (*ReportRepository)(nil)

func NewReportRepository(db *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) ListAccountBalances(ctx context.Context, userID string) ([]report.AccountBalance, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "report", "operation", "list_account_balances", "user_id", userID)
	rows, err := r.db.Query(ctx, `
		SELECT
			a.id AS account_id,
			a.currency,
			COALESCE(
				SUM(
					CASE
						WHEN t.type = 'income' AND t.account_id = a.id THEN t.amount
						WHEN t.type = 'expense' AND t.account_id = a.id THEN -t.amount
						WHEN t.type = 'transfer' AND t.to_account_id = a.id THEN COALESCE(t.to_amount, t.amount)
						WHEN t.type = 'transfer' AND t.from_account_id = a.id THEN -COALESCE(t.from_amount, t.amount)
						ELSE 0
					END
				),
				0
			)::text AS balance
		FROM accounts a
		JOIN user_accounts ua ON ua.account_id = a.id
		LEFT JOIN transactions t
		  ON t.deleted_at IS NULL
		 AND t.status = 'posted'
		 AND (
		   t.account_id = a.id
		   OR t.from_account_id = a.id
		   OR t.to_account_id = a.id
		 )
		WHERE ua.user_id = $1
		  AND ua.status = 'active'
		  AND a.deleted_at IS NULL
		GROUP BY a.id, a.currency
		ORDER BY a.id ASC
	`, userID)
	if err != nil {
		logger.Error("repo_report_list_account_balances_failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	out := make([]report.AccountBalance, 0)
	for rows.Next() {
		var item report.AccountBalance
		if err := rows.Scan(&item.AccountID, &item.Currency, &item.Balance); err != nil {
			logger.Error("repo_report_list_account_balances_failed", "error", err)
			return nil, err
		}
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		logger.Error("repo_report_list_account_balances_failed", "error", err)
		return nil, err
	}

	logger.Info("repo_report_list_account_balances_succeeded", "count", len(out))
	return out, nil
}

func (r *ReportRepository) GetCashflow(ctx context.Context, userID string, months int) ([]report.CashflowStat, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "report", "operation", "get_cashflow", "user_id", userID, "months", months)
	startDate := time.Now().UTC().AddDate(0, -months, 0).Format("2006-01-01")

	rows, err := r.db.Query(ctx, `
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
			  AND t.status = 'posted'
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
		logger.Error("repo_report_get_cashflow_failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	out := make([]report.CashflowStat, 0)
	for rows.Next() {
		var item report.CashflowStat
		if err := rows.Scan(&item.Month, &item.Income, &item.Expense); err != nil {
			logger.Error("repo_report_get_cashflow_failed", "error", err)
			return nil, err
		}
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		logger.Error("repo_report_get_cashflow_failed", "error", err)
		return nil, err
	}

	logger.Info("repo_report_get_cashflow_succeeded", "count", len(out))
	return out, nil
}

func (r *ReportRepository) GetTopExpenses(ctx context.Context, userID string, year int, month int, limit int) ([]report.CategoryExpenseStat, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "repository", "domain", "report", "operation", "get_top_expenses", "user_id", userID, "year", year, "month", month, "limit", limit)
	rows, err := r.db.Query(ctx, `
		SELECT
			li.category_id,
			SUM(li.amount)::text as amount
		FROM transaction_line_items li
		JOIN transactions t ON t.id = li.transaction_id
		WHERE t.deleted_at IS NULL
		  AND t.status = 'posted'
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
		logger.Error("repo_report_get_top_expenses_failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	out := make([]report.CategoryExpenseStat, 0)
	for rows.Next() {
		var item report.CategoryExpenseStat
		if err := rows.Scan(&item.CategoryID, &item.Amount); err != nil {
			logger.Error("repo_report_get_top_expenses_failed", "error", err)
			return nil, err
		}
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		logger.Error("repo_report_get_top_expenses_failed", "error", err)
		return nil, err
	}

	logger.Info("repo_report_get_top_expenses_succeeded", "count", len(out))
	return out, nil
}
