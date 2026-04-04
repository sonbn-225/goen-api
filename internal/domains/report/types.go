package report

import "context"

type AccountBalance struct {
	AccountID string `json:"account_id"`
	Currency  string `json:"currency"`
	Balance   string `json:"balance"`
}

type CashflowStat struct {
	Month   string `json:"month"`
	Income  string `json:"income"`
	Expense string `json:"expense"`
}

type CategoryExpenseStat struct {
	CategoryID string `json:"category_id"`
	Amount     string `json:"amount"`
}

type DashboardReport struct {
	TotalBalances    []AccountBalance      `json:"total_balances"`
	Cashflow6Months  []CashflowStat        `json:"cashflow_6_months"`
	TopExpensesMonth []CategoryExpenseStat `json:"top_expenses_month"`
}

type Repository interface {
	ListAccountBalances(ctx context.Context, userID string) ([]AccountBalance, error)
	GetCashflow(ctx context.Context, userID string, months int) ([]CashflowStat, error)
	GetTopExpenses(ctx context.Context, userID string, year int, month int, limit int) ([]CategoryExpenseStat, error)
}

type Service interface {
	GetDashboardReport(ctx context.Context, userID string) (*DashboardReport, error)
}

type ModuleDeps struct {
	Repo    Repository
	Service Service
}

type Module struct {
	Service Service
	Handler *Handler
}
