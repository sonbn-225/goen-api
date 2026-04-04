package savings

// HTTP contract models used by handlers and API docs.

type CreateSavingsInstrumentRequest struct {
	Name             *string `json:"name,omitempty"`
	SavingsAccountID *string `json:"savings_account_id,omitempty"`
	ParentAccountID  *string `json:"parent_account_id,omitempty"`
	Principal        string  `json:"principal"`
	InterestRate     *string `json:"interest_rate,omitempty"`
	TermMonths       *int    `json:"term_months,omitempty"`
	StartDate        *string `json:"start_date,omitempty"`
	MaturityDate     *string `json:"maturity_date,omitempty"`
	AutoRenew        *bool   `json:"auto_renew,omitempty"`
	AccruedInterest  *string `json:"accrued_interest,omitempty"`
	Status           *string `json:"status,omitempty"`
}

type PatchSavingsInstrumentRequest struct {
	Principal       *string `json:"principal,omitempty"`
	InterestRate    *string `json:"interest_rate,omitempty"`
	TermMonths      *int    `json:"term_months,omitempty"`
	StartDate       *string `json:"start_date,omitempty"`
	MaturityDate    *string `json:"maturity_date,omitempty"`
	AutoRenew       *bool   `json:"auto_renew,omitempty"`
	AccruedInterest *string `json:"accrued_interest,omitempty"`
	Status          *string `json:"status,omitempty"`
}
