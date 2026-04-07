package dto

import (
	"math/big"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

type CreateBudgetRequest struct {
	Name                  *string `json:"name,omitempty"`
	Period                entity.BudgetPeriod `json:"period" binding:"required"`
	PeriodStart           *string `json:"period_start,omitempty"`
	PeriodEnd             *string `json:"period_end,omitempty"`
	Amount                string  `json:"amount" binding:"required"`
	AlertThresholdPercent *int    `json:"alert_threshold_percent,omitempty"`
	RolloverMode          *entity.BudgetRolloverMode `json:"rollover_mode,omitempty"`
	CategoryID            *string `json:"category_id,omitempty"`
}

type UpdateBudgetRequest struct {
	Name                  *string `json:"name,omitempty"`
	Amount                *string `json:"amount,omitempty"`
	AlertThresholdPercent *int    `json:"alert_threshold_percent,omitempty"`
	RolloverMode          *entity.BudgetRolloverMode `json:"rollover_mode,omitempty"`
}

type BudgetWithStatsResponse struct {
	ID                    uuid.UUID  `json:"id"`
	UserID                uuid.UUID  `json:"user_id"`
	Name                  *string    `json:"name,omitempty"`
	Period                entity.BudgetPeriod `json:"period"`
	PeriodStart           *string    `json:"period_start,omitempty"`
	PeriodEnd             *string    `json:"period_end,omitempty"`
	Amount                string     `json:"amount"`
	AlertThresholdPercent *int       `json:"alert_threshold_percent,omitempty"`
	RolloverMode          *entity.BudgetRolloverMode `json:"rollover_mode,omitempty"`
	CategoryID            *uuid.UUID `json:"category_id,omitempty"`
	Spent                 string     `json:"spent"`
	Remaining             string     `json:"remaining"`
	PercentUsed           int        `json:"percent_used"`
}

func NewBudgetWithStatsResponse(b entity.Budget) BudgetWithStatsResponse {
	percentUsed := 0
	if utils.IsValidDecimal(b.Amount) && utils.IsValidDecimal(b.Spent) {
		amt, _ := new(big.Rat).SetString(b.Amount)
		spent, _ := new(big.Rat).SetString(b.Spent)
		if amt != nil && amt.Sign() > 0 {
			res := new(big.Rat).Quo(spent, amt)
			res.Mul(res, big.NewRat(100, 1))
			f, _ := res.Float64()
			percentUsed = int(f)
		}
	}

	return BudgetWithStatsResponse{
		ID:                    b.ID,
		UserID:                b.UserID,
		Name:                  b.Name,
		Period:                b.Period,
		PeriodStart:           b.PeriodStart,
		PeriodEnd:             b.PeriodEnd,
		Amount:                b.Amount,
		AlertThresholdPercent: b.AlertThresholdPercent,
		RolloverMode:          b.RolloverMode,
		CategoryID:            b.CategoryID,
		Spent:                 b.Spent,
		Remaining:             b.Remaining,
		PercentUsed:           percentUsed,
	}
}

func NewBudgetWithStatsResponses(items []entity.Budget) []BudgetWithStatsResponse {
	out := make([]BudgetWithStatsResponse, len(items))
	for i, it := range items {
		out[i] = NewBudgetWithStatsResponse(it)
	}
	return out
}

