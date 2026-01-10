package domain

import (
	"context"
	"time"
)

type Budget struct {
	ID                    string    `json:"id"`
	UserID                string    `json:"user_id"`
	Name                  *string   `json:"name,omitempty"`
	Period                string    `json:"period"`
	PeriodStart           *string   `json:"period_start,omitempty"`
	PeriodEnd             *string   `json:"period_end,omitempty"`
	Amount                string    `json:"amount"`
	AlertThresholdPercent *int      `json:"alert_threshold_percent,omitempty"`
	RolloverMode          *string   `json:"rollover_mode,omitempty"`
	CategoryID            *string   `json:"category_id,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type BudgetRepository interface {
	CreateBudget(ctx context.Context, userID string, b Budget) error
	GetBudget(ctx context.Context, userID string, budgetID string) (*Budget, error)
	ListBudgets(ctx context.Context, userID string) ([]Budget, error)
	ComputeSpent(ctx context.Context, userID string, categoryID string, startDate string, endDate string) (string, error)
}
