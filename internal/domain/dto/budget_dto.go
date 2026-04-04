package dto

import "github.com/sonbn-225/goen-api/internal/domain/entity"

type CreateBudgetRequest struct {
	Name                  *string `json:"name,omitempty"`
	Period                string  `json:"period" binding:"required"`
	PeriodStart           *string `json:"period_start,omitempty"`
	PeriodEnd             *string `json:"period_end,omitempty"`
	Amount                string  `json:"amount" binding:"required"`
	AlertThresholdPercent *int    `json:"alert_threshold_percent,omitempty"`
	RolloverMode          *string `json:"rollover_mode,omitempty"`
	CategoryID            *string `json:"category_id,omitempty"`
}

type UpdateBudgetRequest struct {
	Name                  *string `json:"name,omitempty"`
	Amount                *string `json:"amount,omitempty"`
	AlertThresholdPercent *int    `json:"alert_threshold_percent,omitempty"`
	RolloverMode          *string `json:"rollover_mode,omitempty"`
}

type BudgetWithStatsResponse struct {
	entity.Budget
	Spent       string `json:"spent"`
	Remaining   string `json:"remaining"`
	PercentUsed int    `json:"percent_used"`
}
