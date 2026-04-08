package entity

// AccountSettings contains specialized configurations for different account types.
type AccountSettings struct {
	Color      *string             `json:"color,omitempty"`      // UI color representation in hex
	Investment *InvestmentSettings `json:"investment,omitempty"` // Settings specific to brokerage/investment accounts
	Savings    *SavingsSettings    `json:"savings,omitempty"`    // Settings specific to savings goal accounts
}

// InvestmentSettings defines brokerage-specific behavior like automated tax/fee calculation.
type InvestmentSettings struct {
	FeeSettings map[string]any `json:"fee_settings,omitempty"` // Configuration for trade fee calculations
	TaxSettings map[string]any `json:"tax_settings,omitempty"` // Configuration for trade tax calculations
}

// SavingsSettings defines parameters for tracking savings progress.
type SavingsSettings struct {
	TargetAmount string `json:"target_amount,omitempty"` // Financial goal amount (decimal string)
	TargetDate   string `json:"target_date,omitempty"`   // Target date for achieving the goal (YYYY-MM-DD)
}
