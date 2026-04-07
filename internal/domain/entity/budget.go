package entity

import (
	"github.com/google/uuid"
)

type Budget struct {
	AuditEntity
	UserID                uuid.UUID  `json:"user_id"`
	Name                  *string    `json:"name,omitempty"`
	Period                string     `json:"period"` // month, week, custom
	PeriodStart           *string    `json:"period_start,omitempty"`
	PeriodEnd             *string    `json:"period_end,omitempty"`
	Amount                string     `json:"amount"`
	AlertThresholdPercent *int       `json:"alert_threshold_percent,omitempty"`
	RolloverMode          *string    `json:"rollover_mode,omitempty"`
	CategoryID            *uuid.UUID `json:"category_id,omitempty"`

	// Enriched fields
	Spent     string `json:"spent"`
	Remaining string `json:"remaining"`
}

