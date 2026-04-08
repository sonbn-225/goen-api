package dto

// AccountSettingsRequest and AccountSettingsResponse define the structure for account-specific configs.
type AccountSettingsRequest struct {
	Color      *string                     `json:"color,omitempty"`      // Account UI Color
	Investment *InvestmentSettingsRequest `json:"investment,omitempty"` // Brokerage specific configuration
	Savings    *SavingsSettingsRequest    `json:"savings,omitempty"`    // Savings goal specific configuration
}

type AccountSettingsResponse struct {
	Color      *string                     `json:"color,omitempty"`      // Account UI Color
	Investment *InvestmentSettingsResponse `json:"investment,omitempty"` // Brokerage specific configuration
	Savings    *SavingsSettingsResponse    `json:"savings,omitempty"`    // Savings goal specific configuration
}

type InvestmentSettingsRequest struct {
	FeeSettings map[string]any `json:"fee_settings,omitempty"` // Configuration for trade fee calculations
	TaxSettings map[string]any `json:"tax_settings,omitempty"` // Configuration for trade tax calculations
}

type InvestmentSettingsResponse struct {
	FeeSettings map[string]any `json:"fee_settings,omitempty"` // Configuration for trade fee calculations
	TaxSettings map[string]any `json:"tax_settings,omitempty"` // Configuration for trade tax calculations
}

type SavingsSettingsRequest struct {
	TargetAmount string `json:"target_amount,omitempty"` // Financial goal amount (decimal string)
	TargetDate   string `json:"target_date,omitempty"`   // Target date for achieving the goal (YYYY-MM-DD)
}

type SavingsSettingsResponse struct {
	TargetAmount string `json:"target_amount,omitempty"` // Financial goal amount (decimal string)
	TargetDate   string `json:"target_date,omitempty"`   // Target date for achieving the goal (YYYY-MM-DD)
}
