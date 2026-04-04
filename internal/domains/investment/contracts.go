package investment

// HTTP contract models used by handlers and API docs.

type PatchInvestmentAccountRequest struct {
	FeeSettings any `json:"fee_settings,omitempty"`
	TaxSettings any `json:"tax_settings,omitempty"`
}

type CreateTradeRequest struct {
	ClientID         *string `json:"client_id,omitempty"`
	SecurityID       string  `json:"security_id"`
	FeeTransactionID *string `json:"fee_transaction_id,omitempty"`
	TaxTransactionID *string `json:"tax_transaction_id,omitempty"`
	Side             string  `json:"side"`
	Quantity         string  `json:"quantity"`
	Price            string  `json:"price"`
	Fees             *string `json:"fees,omitempty"`
	Taxes            *string `json:"taxes,omitempty"`
	OccurredAt       *string `json:"occurred_at,omitempty"`
	Note             *string `json:"note,omitempty"`
}
