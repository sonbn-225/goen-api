package debt

// HTTP contract models used by handlers and API docs.

type CreateDebtRequest struct {
	ClientID     *string `json:"client_id,omitempty"`
	AccountID    string  `json:"account_id"`
	Direction    string  `json:"direction"`
	Name         *string `json:"name,omitempty"`
	ContactID    *string `json:"contact_id,omitempty"`
	Principal    string  `json:"principal"`
	StartDate    string  `json:"start_date"`
	DueDate      string  `json:"due_date"`
	InterestRate *string `json:"interest_rate,omitempty"`
	InterestRule *string `json:"interest_rule,omitempty"`
	Status       *string `json:"status,omitempty"`
}

type CreateDebtPaymentRequest struct {
	TransactionID string  `json:"transaction_id"`
	PrincipalPaid *string `json:"principal_paid,omitempty"`
	InterestPaid  *string `json:"interest_paid,omitempty"`
}

type CreateDebtInstallmentRequest struct {
	InstallmentNo int     `json:"installment_no"`
	DueDate       string  `json:"due_date"`
	AmountDue     string  `json:"amount_due"`
	AmountPaid    *string `json:"amount_paid,omitempty"`
	Status        *string `json:"status,omitempty"`
}
