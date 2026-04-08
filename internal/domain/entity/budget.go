package entity

import (
	"github.com/google/uuid"
)

// Budget represents a spending limit for a specific category and period.
type Budget struct {
	AuditEntity
	UserID                uuid.UUID           `json:"user_id"`                          // ID of the user who owns the budget
	Name                  *string             `json:"name,omitempty"`                  // Optional name for the budget
	Period                BudgetPeriod        `json:"period"`                          // Budget cycle (monthly, weekly, custom)
	PeriodStart           *string             `json:"period_start,omitempty"`          // Start date of the budget period (YYYY-MM-DD)
	PeriodEnd             *string             `json:"period_end,omitempty"`            // End date of the budget period (YYYY-MM-DD)
	Amount                string              `json:"amount"`                          // Total budget amount (decimal string)
	AlertThresholdPercent *int                `json:"alert_threshold_percent,omitempty"` // Percentage of spending that triggers an alert
	RolloverMode          *BudgetRolloverMode `json:"rollover_mode,omitempty"`          // How to handle remaining funds at period end
	CategoryID            *uuid.UUID          `json:"category_id,omitempty"`          // ID of the category this budget applies to

	// Enriched fields
	Spent     string `json:"spent"`     // Total amount already spent in this period (decimal string)
	Remaining string `json:"remaining"` // Remaining budget amount (decimal string)
}

