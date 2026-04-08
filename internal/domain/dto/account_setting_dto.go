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
	Principal    string  `json:"principal,omitempty"`     // Initial deposit amount (decimal string)
	InterestRate *string `json:"interest_rate,omitempty"` // Annual interest rate (decimal string)
	TermMonths   *int    `json:"term_months,omitempty"`   // Duration in months
	StartDate    *string `json:"start_date,omitempty"`    // Start date (YYYY-MM-DD)
	MaturityDate *string `json:"maturity_date,omitempty"` // Expected maturity date (YYYY-MM-DD)
	AutoRenew    *bool   `json:"auto_renew,omitempty"`    // Whether to auto-renew on maturity
}

type SavingsSettingsResponse struct {
	Principal    string  `json:"principal,omitempty"`     // Initial deposit amount (decimal string)
	InterestRate *string `json:"interest_rate,omitempty"` // Annual interest rate (decimal string)
	TermMonths   *int    `json:"term_months,omitempty"`   // Duration in months
	StartDate    *string `json:"start_date,omitempty"`    // Start date (YYYY-MM-DD)
	MaturityDate *string `json:"maturity_date,omitempty"` // Expected maturity date (YYYY-MM-DD)
	AutoRenew    bool    `json:"auto_renew"`              // Whether to auto-renew on maturity
}
