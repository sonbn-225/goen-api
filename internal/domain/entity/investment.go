package entity

import (
	"time"

	"github.com/google/uuid"
)


// SecurityEventElection tracks a user's participation or entitlement to a corporate action.
type SecurityEventElection struct {
	AuditEntity
	UserID                       uuid.UUID                   `json:"user_id"`                              // ID of the user
	AccountID                    uuid.UUID                   `json:"account_id"`                           // ID of the account
	SecurityEventID              uuid.UUID                   `json:"security_event_id"`                    // ID of the corporate action
	SecurityID                   uuid.UUID                   `json:"security_id"`                          // ID of the security
	EntitlementDate              string                      `json:"entitlement_date"`                     // Date used for eligibility
	HoldingQuantityAtEntitlement string                      `json:"holding_quantity_at_entitlement_date"` // Shares held on entitlement date
	EntitledQuantity             string                      `json:"entitled_quantity"`                    // Calculated entitlement amount
	ElectedQuantity              string                      `json:"elected_quantity"`                     // Quantity user chose to claim
	Status                       SecurityEventElectionStatus `json:"status"`                               // Status of the claim (eligible/claimed/etc.)
	ConfirmedAt                  *time.Time                  `json:"confirmed_at,omitempty"`               // Timestamp of confirmation
	Note                         *string                     `json:"note,omitempty"`                       // User notes
}

// Trade represents a single buy or sell transaction of a security.
type Trade struct {
	AuditEntity
	AccountID        uuid.UUID  `json:"account_id"`                   // ID of the account
	SecurityID       uuid.UUID  `json:"security_id"`                  // ID of the security traded
	FeeTransactionID *uuid.UUID `json:"fee_transaction_id,omitempty"` // Linked fee expense transaction
	TaxTransactionID *uuid.UUID `json:"tax_transaction_id,omitempty"` // Linked tax expense transaction
	Side             TradeSide  `json:"side"`                         // Buy or Sell
	Quantity         string     `json:"quantity"`                     // Number of shares (decimal string)
	Price            string     `json:"price"`                        // Execution price per share (decimal string)
	Fees             string     `json:"fees"`                         // Total trading fees (decimal string)
	Taxes            string     `json:"taxes"`                        // Total trading taxes (decimal string)
	OccurredAt       time.Time  `json:"occurred_at"`                  // Timestamp of the trade
	Note             *string    `json:"note,omitempty"`               // Optional trade memo
}

// Holding represents the current quantity and value of a security held in a specific account.
type Holding struct {
	AuditEntity
	AccountID       uuid.UUID  `json:"account_id"`                 // ID of the account
	SecurityID      uuid.UUID  `json:"security_id"`                // ID of the security
	Quantity        string     `json:"quantity"`                   // Total shares held (decimal string)
	CostBasisTotal  *string    `json:"cost_basis_total,omitempty"` // Total cost paid for the holding
	AvgCost         *string    `json:"avg_cost,omitempty"`         // Average cost per share
	MarketPrice     *string    `json:"market_price,omitempty"`     // Last known market price
	MarketValue     *string    `json:"market_value,omitempty"`     // Total market value of the holding
	UnrealizedPnL   *string    `json:"unrealized_pnl,omitempty"`   // Total paper profit/loss
	AsOf            *time.Time `json:"as_of,omitempty"`            // Timestamp of the market valuation
}

// ShareLot represents a specific batch of shares acquired at the same time and price.
type ShareLot struct {
	AuditEntity
	AccountID       uuid.UUID      `json:"account_id"`             // ID of the account
	SecurityID      uuid.UUID      `json:"security_id"`            // ID of the security
	Quantity        string         `json:"quantity"`               // Remaining shares in this lot (decimal string)
	AcquisitionDate string         `json:"acquisition_date"`       // Date shares were acquired (YYYY-MM-DD)
	CostBasisPer    string         `json:"cost_basis_per_share"`   // Cost per share in this lot
	Provenance      string         `json:"provenance"`             // Origin of shares (regular_buy, dividend, etc.)
	Status          ShareLotStatus `json:"status"`                 // active or closed
	BuyTradeID      *uuid.UUID     `json:"buy_trade_id,omitempty"` // ID of the trade that created this lot
}

// RealizedTradeLog records the profit or loss from selling shares from a specific lot.
type RealizedTradeLog struct {
	AuditEntity
	AccountID       uuid.UUID `json:"account_id"`          // ID of the account
	SecurityID      uuid.UUID `json:"security_id"`         // ID of the security
	SellTradeID     uuid.UUID `json:"sell_trade_id"`       // ID of the selling trade
	SourceShareLot  uuid.UUID `json:"source_share_lot_id"` // ID of the original share lot
	Quantity        string    `json:"quantity"`            // Shares sold from this lot (decimal string)
	AcquisitionDate string    `json:"acquisition_date"`    // Date original shares were acquired
	CostBasisTotal  string    `json:"cost_basis_total"`    // Initial cost of the sold shares
	SellPrice       string    `json:"sell_price"`          // Executed sale price per share
	Proceeds        string    `json:"proceeds"`            // Total proceeds from this sale lot
	RealizedPnL     string    `json:"realized_pnl"`        // Actual profit or loss from this sale
	Provenance      string    `json:"provenance"`          // Origin of the original lot
}
