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
	Principal    string  `json:"principal,omitempty"`     // Initial deposit amount (decimal string)
	InterestRate *string `json:"interest_rate,omitempty"` // Annual interest rate (decimal string)
	TermMonths   *int    `json:"term_months,omitempty"`   // Duration in months
	StartDate    *string `json:"start_date,omitempty"`    // Start date (YYYY-MM-DD)
	MaturityDate *string `json:"maturity_date,omitempty"` // Expected maturity date (YYYY-MM-DD)
	AutoRenew    bool    `json:"auto_renew,omitempty"`    // Whether to auto-renew on maturity
}
