package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// CreateTradeRequest is the payload for recording a new security trade.
// Used in: InvestmentHandler, InvestmentService, InvestmentInterface
type CreateTradeRequest struct {
	SecurityID       uuid.UUID  `json:"security_id"`                // ID of the security being traded
	FeeTransactionID *uuid.UUID `json:"fee_transaction_id,omitempty"` // Optional ID of an existing fee transaction
	TaxTransactionID *uuid.UUID `json:"tax_transaction_id,omitempty"` // Optional ID of an existing tax transaction
	Provenance       *string    `json:"provenance,omitempty"`        // Origin of the trade (e.g., "manual")
	Side             entity.TradeSide  `json:"side"`                 // Buy or Sell
	Quantity         string     `json:"quantity"`                   // Number of shares traded (decimal string)
	Price            string     `json:"price"`                      // Price per share (decimal string)
	Fees             *string    `json:"fees,omitempty"`             // Total trading fees (decimal string)
	Taxes            *string    `json:"taxes,omitempty"`            // Total trading taxes (decimal string)
	OccurredAt       *string    `json:"occurred_at,omitempty"`       // Full occurrence timestamp
	OccurredDate     *string    `json:"occurred_date,omitempty"`     // Date of the trade (YYYY-MM-DD)
	OccurredTime     *string    `json:"occurred_time,omitempty"`     // Time of the trade (HH:MM:SS)
	Note             *string    `json:"note,omitempty"`              // Optional trade memo

	PrincipalCategoryID  *uuid.UUID `json:"principal_category_id,omitempty"`  // Category ID for the principal transaction
	PrincipalDescription *string    `json:"principal_description,omitempty"` // Description for the principal transaction
	FeeCategoryID        *uuid.UUID `json:"fee_category_id,omitempty"`        // Category ID for the fee transaction
	FeeDescription       *string    `json:"fee_description,omitempty"`       // Description for the fee transaction
	TaxCategoryID        *uuid.UUID `json:"tax_category_id,omitempty"`        // Category ID for the tax transaction
	TaxDescription       *string    `json:"tax_description,omitempty"`       // Description for the tax transaction
}

// TradeResponse represents a single security trade transaction.
// Used in: InvestmentHandler, InvestmentService, InvestmentInterface
type TradeResponse struct {
	ID               uuid.UUID  `json:"id"`                             // Unique trade identifier
	AccountID        uuid.UUID  `json:"account_id"`                     // ID of the account
	SecurityID       uuid.UUID  `json:"security_id"`                    // ID of the traded security
	FeeTransactionID *uuid.UUID `json:"fee_transaction_id,omitempty"`     // ID of the linked fee transaction
	TaxTransactionID *uuid.UUID `json:"tax_transaction_id,omitempty"`     // ID of the linked tax transaction
	Side             entity.TradeSide  `json:"side"`                     // Buy or Sell
	Quantity         string     `json:"quantity"`                       // Shares traded
	Price            string     `json:"price"`                          // Execution price
	Fees             string     `json:"fees"`                           // Total fees paid
	Taxes            string     `json:"taxes"`                          // Total taxes paid
	OccurredAt       time.Time  `json:"occurred_at"`                    // Time of the trade
	Note             *string    `json:"note,omitempty"`                 // Trade memo
}

// HoldingResponse represents a security position in an investment account.
// Used in: InvestmentHandler, InvestmentService, InvestmentInterface
type HoldingResponse struct {
	ID              uuid.UUID  `json:"id"`                             // Unique identifier for the holding record
	AccountID       uuid.UUID  `json:"account_id"`                     // ID of the account
	SecurityID      uuid.UUID  `json:"security_id"`                    // ID of the security held
	Quantity        string     `json:"quantity"`                       // Total shares currently held
	CostBasisTotal  *string    `json:"cost_basis_total,omitempty"`      // Total cost of all shares in position
	AvgCost         *string    `json:"avg_cost,omitempty"`              // Average cost per share
	MarketPrice     *string    `json:"market_price,omitempty"`          // Current market price
	MarketValue     *string    `json:"market_value,omitempty"`          // Total market value of position
	UnrealizedPnL   *string    `json:"unrealized_pnl,omitempty"`        // Paper profit/loss
	AsOf            *time.Time `json:"as_of,omitempty"`                 // Timestamp of the market valuation
}

// EligibleAction represents a corporate action that a user is entitled to.
// Used in: InvestmentHandler, InvestmentService, InvestmentInterface
type EligibleAction struct {
	Event            SecurityEventResponse `json:"event"`            // The corporate action details
	HoldingQuantity  string                `json:"holding_quantity"` // Shares held on the eligibility date
	EntitledQuantity string                `json:"entitled_quantity"` // Calculated entitlement (decimal string)
	Status           entity.SecurityEventElectionStatus `json:"status"` // Current claim status (eligible/claimed)
	ElectionID       *uuid.UUID            `json:"election_id,omitempty"` // ID of the existing claim/election record
}

// ClaimCorporateActionRequest is the payload for claiming a corporate action.
// Used in: InvestmentHandler, InvestmentService, InvestmentInterface
type ClaimCorporateActionRequest struct {
	ElectedQuantity *string `json:"elected_quantity,omitempty"` // Amount of entitlement user chooses to claim
	Note            *string `json:"note,omitempty"`             // Optional user notes for the claim
}

// BackfillTradePrincipalResponse represents the result of the trade principal backfill operation.
// Used in: InvestmentHandler, InvestmentService, InvestmentInterface
type BackfillTradePrincipalResponse struct {
	TradesTotal          int `json:"trades_total"`           // Total trades scanned
	TransactionsCreated  int `json:"transactions_created"`   // New principal transactions created
	TransactionsExisting int `json:"transactions_existing"`  // Trades that already had principal transactions
	SkippedZeroNotional  int `json:"skipped_zero_notional"`  // Trades skipped because notional amount was zero
	SkippedStockDividend int `json:"skipped_stock_dividend"`  // Trades skipped because they were stock dividends
}

// RealizedPNLReportItem represents realized profit and loss for a specific security.
// Used in: InvestmentHandler, InvestmentService, InvestmentInterface
type RealizedPNLReportItem struct {
	SecurityID        uuid.UUID `json:"security_id"`          // ID of the security
	Symbol            string    `json:"symbol"`               // Security ticker
	GrossRealizedGain string    `json:"gross_realized_gain"` // Total gain before fees and taxes
	TradeGain         string    `json:"trade_gain"`           // Gain from price appreciation
	DividendGain      string    `json:"dividend_gain"`        // Gain from dividends
	Proceeds          string    `json:"proceeds"`             // Total sale proceeds
	CostBasis         string    `json:"cost_basis"`           // Original cost of sold shares
	Fees              string    `json:"fees"`                 // Total fees associated with the sales
	Taxes             string    `json:"taxes"`                // Total taxes associated with the sales
	NetPNL            string    `json:"net_pnl"`              // Final profit/loss after fees and taxes
}

// RealizedPNLReport represents the overall realized profit and loss report.
// Used in: InvestmentHandler, InvestmentService, InvestmentInterface
type RealizedPNLReport struct {
	Items      []RealizedPNLReportItem `json:"items"`       // PnL breakdown per security
	TotalNet   string                  `json:"total_net"`    // Overall net profit/loss
	TotalGross string                  `json:"total_gross"`  // Overall gross profit/loss
}
