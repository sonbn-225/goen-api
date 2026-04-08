package dto

import (
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// CreateBudgetRequest is used when creating a new budget.
// Used in: BudgetHandler, BudgetService, BudgetInterface
// CreateBudgetRequest is used when creating a new budget.
// Used in: BudgetHandler, BudgetService, BudgetInterface
type CreateBudgetRequest struct {
	Name                  *string                    `json:"name,omitempty"`                  // Optional descriptive name for the budget
	Period                entity.BudgetPeriod        `json:"period" binding:"required"`       // Recurrence period (monthly, weekly, etc.)
	PeriodStart           *string                    `json:"period_start,omitempty"`          // Start date of the first period (YYYY-MM-DD)
	PeriodEnd             *string                    `json:"period_end,omitempty"`            // Optional fixed end date for the budget
	Amount                string                     `json:"amount" binding:"required"`       // Total amount allowed for the period (decimal string)
	AlertThresholdPercent *int                       `json:"alert_threshold_percent,omitempty"` // Percentage of spending that triggers an alert
	RolloverMode          *entity.BudgetRolloverMode `json:"rollover_mode,omitempty"`          // How to handle unused funds in the next period
	CategoryID            *string                    `json:"category_id,omitempty"`          // ID of the category this budget tracks
}

// UpdateBudgetRequest is used when updating an existing budget.
// Used in: BudgetHandler, BudgetService, BudgetInterface
type UpdateBudgetRequest struct {
	Name                  *string                    `json:"name,omitempty"`                  // New descriptive name
	Amount                *string                    `json:"amount,omitempty"`                // New budget amount (decimal string)
	AlertThresholdPercent *int                       `json:"alert_threshold_percent,omitempty"` // New alert threshold percentage
	RolloverMode          *entity.BudgetRolloverMode `json:"rollover_mode,omitempty"`          // New rollover behavior
}

// BudgetWithStatsResponse represents a budget with current spending statistics.
// Used in: BudgetHandler, BudgetService, BudgetInterface
type BudgetWithStatsResponse struct {
	ID                    uuid.UUID                  `json:"id"`                             // Unique budget identifier
	UserID                uuid.UUID                  `json:"user_id"`                        // ID of the user who owns the budget
	Name                  *string                    `json:"name,omitempty"`                  // Budget name
	Period                entity.BudgetPeriod        `json:"period"`                         // Budget period type
	PeriodStart           *string                    `json:"period_start,omitempty"`          // Current period start date
	PeriodEnd             *string                    `json:"period_end,omitempty"`            // Current period end date
	Amount                string                     `json:"amount"`                         // Total budget amount
	AlertThresholdPercent *int                       `json:"alert_threshold_percent,omitempty"` // Alert threshold percentage
	RolloverMode          *entity.BudgetRolloverMode `json:"rollover_mode,omitempty"`          // Rollover configuration
	CategoryID            *uuid.UUID                 `json:"category_id,omitempty"`          // ID of the tracked category
	Spent                 string                     `json:"spent"`                          // Total amount spent in the current period
	Remaining             string                     `json:"remaining"`                      // Amount left in the budget
	PercentUsed           int                        `json:"percent_used"`                   // Percentage of budget already spent
}
