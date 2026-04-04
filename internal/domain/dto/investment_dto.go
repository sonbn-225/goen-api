package dto

import "github.com/sonbn-225/goen-api/internal/domain/entity"

type PatchInvestmentAccountRequest struct {
	FeeSettings any `json:"fee_settings,omitempty"`
	TaxSettings any `json:"tax_settings,omitempty"`
}

type CreateTradeRequest struct {
	ClientID         *string `json:"client_id,omitempty"`
	SecurityID       string  `json:"security_id"`
	FeeTransactionID *string `json:"fee_transaction_id,omitempty"`
	TaxTransactionID *string `json:"tax_transaction_id,omitempty"`
	Provenance       *string `json:"provenance,omitempty"`
	Side             string  `json:"side"` // buy, sell
	Quantity         string  `json:"quantity"`
	Price            string  `json:"price"`
	Fees             *string `json:"fees,omitempty"`
	Taxes            *string `json:"taxes,omitempty"`
	OccurredAt       *string `json:"occurred_at,omitempty"`
	OccurredDate     *string `json:"occurred_date,omitempty"`
	OccurredTime     *string `json:"occurred_time,omitempty"`
	Note             *string `json:"note,omitempty"`

	PrincipalCategoryID  *string `json:"principal_category_id,omitempty"`
	PrincipalDescription *string `json:"principal_description,omitempty"`
	FeeCategoryID        *string `json:"fee_category_id,omitempty"`
	FeeDescription       *string `json:"fee_description,omitempty"`
	TaxCategoryID        *string `json:"tax_category_id,omitempty"`
	TaxDescription       *string `json:"tax_description,omitempty"`
}

type EligibleAction struct {
	Event            entity.SecurityEvent `json:"event"`
	HoldingQuantity  string               `json:"holding_quantity"`
	EntitledQuantity string               `json:"entitled_quantity"`
	Status           string               `json:"status"` // eligible, claimed, dismissed
	ElectionID       *string              `json:"election_id,omitempty"`
}

type ClaimCorporateActionRequest struct {
	ElectedQuantity *string `json:"elected_quantity,omitempty"`
	Note            *string `json:"note,omitempty"`
}

type BackfillTradePrincipalResponse struct {
	TradesTotal          int `json:"trades_total"`
	TransactionsCreated  int `json:"transactions_created"`
	TransactionsExisting int `json:"transactions_existing"`
	SkippedZeroNotional  int `json:"skipped_zero_notional"`
	SkippedStockDividend int `json:"skipped_stock_dividend"`
}

type RealizedPNLReportItem struct {
	SecurityID        string `json:"security_id"`
	Symbol            string `json:"symbol"`
	GrossRealizedGain string `json:"gross_realized_gain"`
	TradeGain         string `json:"trade_gain"`
	DividendGain      string `json:"dividend_gain"`
	Proceeds          string `json:"proceeds"`
	CostBasis         string `json:"cost_basis"`
	Fees              string `json:"fees"`
	Taxes             string `json:"taxes"`
	NetPNL            string `json:"net_pnl"`
}

type RealizedPNLReport struct {
	Items      []RealizedPNLReportItem `json:"items"`
	TotalNet   string                 `json:"total_net"`
	TotalGross string                 `json:"total_gross"`
}
