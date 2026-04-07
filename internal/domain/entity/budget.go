package entity

import (
	"github.com/google/uuid"
)

type BudgetPeriod string

const (
	BudgetPeriodMonth BudgetPeriod = "month"
	BudgetPeriodWeek  BudgetPeriod = "week"
	BudgetPeriodCustom BudgetPeriod = "custom"
)

type BudgetRolloverMode string

const (
	BudgetRolloverModeNone        BudgetRolloverMode = "none"
	BudgetRolloverModeAddPositive BudgetRolloverMode = "add_positive"
	BudgetRolloverModeAddAll      BudgetRolloverMode = "add_all"
)

type Budget struct {
	AuditEntity
	UserID                uuid.UUID           `json:"user_id"`
	Name                  *string             `json:"name,omitempty"`
	Period                BudgetPeriod        `json:"period"`
	PeriodStart           *string             `json:"period_start,omitempty"`
	PeriodEnd             *string             `json:"period_end,omitempty"`
	Amount                string              `json:"amount"`
	AlertThresholdPercent *int                `json:"alert_threshold_percent,omitempty"`
	RolloverMode          *BudgetRolloverMode `json:"rollover_mode,omitempty"`
	CategoryID            *uuid.UUID          `json:"category_id,omitempty"`

	// Enriched fields
	Spent     string `json:"spent"`
	Remaining string `json:"remaining"`
}

