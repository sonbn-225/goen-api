package dto

import "github.com/sonbn-225/goen-api/internal/domain/entity"

type CashflowStatResponse struct {
	Month   string `json:"month"`
	Income  string `json:"income"`
	Expense string `json:"expense"`
}

type CategoryExpenseStatResponse struct {
	CategoryID string `json:"category_id"`
	Amount     string `json:"amount"`
}

type DashboardReportResponse struct {
	TotalBalances    []AccountBalanceResponse      `json:"total_balances"`
	Cashflow6Months  []CashflowStatResponse        `json:"cashflow_6_months"`
	TopExpensesMonth []CategoryExpenseStatResponse `json:"top_expenses_month"`
}

func NewCashflowStatResponse(s entity.CashflowStat) CashflowStatResponse {
	return CashflowStatResponse{
		Month:   s.Month,
		Income:  s.Income,
		Expense: s.Expense,
	}
}

func NewCashflowStatResponses(items []entity.CashflowStat) []CashflowStatResponse {
	out := make([]CashflowStatResponse, len(items))
	for i, it := range items {
		out[i] = NewCashflowStatResponse(it)
	}
	return out
}

func NewCategoryExpenseStatResponse(s entity.CategoryExpenseStat) CategoryExpenseStatResponse {
	return CategoryExpenseStatResponse{
		CategoryID: s.CategoryID,
		Amount:     s.Amount,
	}
}

func NewCategoryExpenseStatResponses(items []entity.CategoryExpenseStat) []CategoryExpenseStatResponse {
	out := make([]CategoryExpenseStatResponse, len(items))
	for i, it := range items {
		out[i] = NewCategoryExpenseStatResponse(it)
	}
	return out
}
