package dto
 
import (
	"github.com/google/uuid"
)
 
// CashflowStatResponse represents income and expense statistics for a specific month.
// Used in: ReportHandler, ReportService, ReportInterface
type CashflowStatResponse struct {
	Month   string `json:"month"`   // Month of the statistic (YYYY-MM)
	Income  string `json:"income"`  // Total income for the month (decimal string)
	Expense string `json:"expense"` // Total expenses for the month (decimal string)
}
 
// CategoryExpenseStatResponse represents the total expenses for a specific category.
// Used in: ReportHandler, ReportService, ReportInterface
type CategoryExpenseStatResponse struct {
	CategoryID uuid.UUID `json:"category_id"` // ID of the category
	Amount     string    `json:"amount"`      // Total expense amount for the category (decimal string)
}
 
// DashboardReportResponse represents the aggregated data for the user dashboard.
// Used in: ReportHandler, ReportService, ReportInterface
type DashboardReportResponse struct {
	TotalBalances    []AccountBalanceResponse      `json:"total_balances"`     // Current balances of all user accounts
	Cashflow6Months  []CashflowStatResponse        `json:"cashflow_6_months"`  // Monthly cashflow for the last 6 months
	TopExpensesMonth []CategoryExpenseStatResponse `json:"top_expenses_month"` // Top expense categories for the current month
}
