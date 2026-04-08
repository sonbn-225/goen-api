package entity

import (
	"github.com/google/uuid"
)

// CashflowStat represents income and expense totals for a specific month.
type CashflowStat struct {
	Month   string `json:"month"`   // The month of the statistic (YYYY-MM)
	Income  string `json:"income"`  // Total income for the month (decimal string)
	Expense string `json:"expense"` // Total expenses for the month (decimal string)
}

// CategoryExpenseStat represents the total amount spent in a specific category.
type CategoryExpenseStat struct {
	CategoryID uuid.UUID `json:"category_id"` // ID of the category
	Amount     string    `json:"amount"`      // Total spent amount (decimal string)
}

// DashboardReport provides a summary of account balances and recent cashflow for the user dashboard.
type DashboardReport struct {
	TotalBalances    []AccountBalance      `json:"total_balances"`     // Current balances of all user accounts
	Cashflow6Months  []CashflowStat        `json:"cashflow_6_months"`  // Monthly cashflow for the last 6 months
	TopExpensesMonth []CategoryExpenseStat `json:"top_expenses_month"` // Top expense categories for the current month
}

