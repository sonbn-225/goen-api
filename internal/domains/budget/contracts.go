package budget

// HTTP contract models used by handlers and API docs.

type CreateBudgetRequest struct {
	Name                  *string `json:"name,omitempty"`
	Period                string  `json:"period"`
	PeriodStart           *string `json:"period_start,omitempty"`
	PeriodEnd             *string `json:"period_end,omitempty"`
	Amount                string  `json:"amount"`
	AlertThresholdPercent *int    `json:"alert_threshold_percent,omitempty"`
	RolloverMode          *string `json:"rollover_mode,omitempty"`
	CategoryID            *string `json:"category_id,omitempty"`
}
