package entity

type CashflowStat struct {
	Month   string `json:"month"` // e.g. "2023-10"
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
