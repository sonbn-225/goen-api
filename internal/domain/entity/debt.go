package entity

import (
	"time"
)

type Debt struct {
	ID                   string    `json:"id"`
	ClientID             *string   `json:"client_id,omitempty"`
	UserID               string    `json:"user_id"`
	AccountID            *string   `json:"account_id,omitempty"`
	Direction            string    `json:"direction"` // lent, borrowed
	Name                 *string   `json:"name,omitempty"`
	ContactID            *string   `json:"contact_id,omitempty"`
	ContactName          *string   `json:"contact_name,omitempty"`
	ContactAvatarURL     *string   `json:"contact_avatar_url,omitempty"`
	Principal            string    `json:"principal"`
	Currency             *string   `json:"currency,omitempty"`
	StartDate            string    `json:"start_date"`
	DueDate              string    `json:"due_date"`
	InterestRate         *string   `json:"interest_rate,omitempty"`
	InterestRule         *string   `json:"interest_rule,omitempty"`
	OutstandingPrincipal string    `json:"outstanding_principal"`
	AccruedInterest      string    `json:"accrued_interest"`
	Status               string    `json:"status"` // active, paid, cancelled
	ClosedAt             *time.Time `json:"closed_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type DebtPaymentLink struct {
	ID            string    `json:"id"`
	DebtID        string    `json:"debt_id"`
	TransactionID string    `json:"transaction_id"`
	PrincipalPaid *string   `json:"principal_paid,omitempty"`
	InterestPaid  *string   `json:"interest_paid,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type DebtInstallment struct {
	ID            string `json:"id"`
	DebtID        string `json:"debt_id"`
	InstallmentNo int    `json:"installment_no"`
	DueDate       string `json:"due_date"`
	AmountDue     string `json:"amount_due"`
	AmountPaid    string `json:"amount_paid"`
	Status        string `json:"status"`
}
