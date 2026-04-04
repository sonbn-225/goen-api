package dto

type CreateDebtRequest struct {
	ClientID     *string `json:"client_id,omitempty"`
	AccountID    string  `json:"account_id" binding:"required"`
	Direction    string  `json:"direction" binding:"required"` // lent, borrowed
	Name         *string `json:"name,omitempty"`
	ContactID    *string `json:"contact_id,omitempty"`
	ContactName  *string `json:"contact_name,omitempty"`
	Principal    string  `json:"principal" binding:"required"`
	StartDate    string  `json:"start_date" binding:"required"`
	DueDate      string  `json:"due_date" binding:"required"`
	InterestRate *string `json:"interest_rate,omitempty"`
	InterestRule *string `json:"interest_rule,omitempty"`
}

type UpdateDebtRequest struct {
	Name         *string `json:"name,omitempty"`
	DueDate      *string `json:"due_date,omitempty"`
	Status       *string `json:"status,omitempty"`
	InterestRate *string `json:"interest_rate,omitempty"`
}

type DebtPaymentRequest struct {
	TransactionID string  `json:"transaction_id" binding:"required"`
	PrincipalPaid *string `json:"principal_paid,omitempty"`
	InterestPaid  *string `json:"interest_paid,omitempty"`
	AmountPaid    *string `json:"amount_paid,omitempty"` // Total paid, can be split by service
}
