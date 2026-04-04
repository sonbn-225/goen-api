package entity

import (
	"time"
)

type Budget struct {
	ID                    string    `json:"id"`
	UserID                string    `json:"user_id"`
	Name                  *string   `json:"name,omitempty"`
	Period                string    `json:"period"` // month, week, custom
	PeriodStart           *string   `json:"period_start,omitempty"`
	PeriodEnd             *string   `json:"period_end,omitempty"`
	Amount                string    `json:"amount"`
	AlertThresholdPercent *int      `json:"alert_threshold_percent,omitempty"`
	RolloverMode          *string   `json:"rollover_mode,omitempty"`
	CategoryID            *string   `json:"category_id,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}
