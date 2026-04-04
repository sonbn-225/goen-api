package entity

import "time"

type InvestmentAccount struct {
	ID           string    `json:"id"`
	AccountID    string    `json:"account_id"`
	Currency     string    `json:"currency"`
	FeeSettings  any       `json:"fee_settings,omitempty"`
	TaxSettings  any       `json:"tax_settings,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Security struct {
	ID         string    `json:"id"`
	Symbol     string    `json:"symbol"`
	Name       *string   `json:"name,omitempty"`
	AssetClass *string   `json:"asset_class,omitempty"`
	Currency   *string   `json:"currency,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type SecurityPriceDaily struct {
	ID         string    `json:"id"`
	SecurityID string    `json:"security_id"`
	PriceDate  string    `json:"price_date"`
	Open       *string   `json:"open,omitempty"`
	High       *string   `json:"high,omitempty"`
	Low        *string   `json:"low,omitempty"`
	Close      string    `json:"close"`
	Volume     *string   `json:"volume,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type SecurityEvent struct {
	ID                 string    `json:"id"`
	SecurityID         string    `json:"security_id"`
	EventType          string    `json:"event_type"` // cash_dividend, stock_dividend, split, merger, etc.
	ExDate             *string   `json:"ex_date,omitempty"`
	RecordDate         *string   `json:"record_date,omitempty"`
	PayDate            *string   `json:"pay_date,omitempty"`
	EffectiveDate      *string   `json:"effective_date,omitempty"`
	CashAmountPerShare *string   `json:"cash_amount_per_share,omitempty"`
	RatioNumerator     *string   `json:"ratio_numerator,omitempty"`
	RatioDenominator   *string   `json:"ratio_denominator,omitempty"`
	SubscriptionPrice  *string   `json:"subscription_price,omitempty"`
	Currency           *string   `json:"currency,omitempty"`
	Note               *string   `json:"note,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type SecurityEventElection struct {
	ID                           string     `json:"id"`
	UserID                       string     `json:"user_id"`
	BrokerAccountID              string     `json:"broker_account_id"`
	SecurityEventID              string     `json:"security_event_id"`
	SecurityID                   string     `json:"security_id"`
	EntitlementDate              string     `json:"entitlement_date"`
	HoldingQuantityAtEntitlement string     `json:"holding_quantity_at_entitlement_date"`
	EntitledQuantity             string     `json:"entitled_quantity"`
	ElectedQuantity              string     `json:"elected_quantity"`
	Status                       string     `json:"status"` // eligible, claimed, dismissed
	ConfirmedAt                  *time.Time `json:"confirmed_at,omitempty"`
	Note                         *string    `json:"note,omitempty"`
	CreatedAt                    time.Time  `json:"created_at"`
	UpdatedAt                    time.Time  `json:"updated_at"`
}

type Trade struct {
	ID               string    `json:"id"`
	ClientID         *string   `json:"client_id,omitempty"`
	BrokerAccountID  string    `json:"broker_account_id"`
	SecurityID       string    `json:"security_id"`
	FeeTransactionID *string   `json:"fee_transaction_id,omitempty"`
	TaxTransactionID *string   `json:"tax_transaction_id,omitempty"`
	Side             string    `json:"side"` // buy, sell
	Quantity         string    `json:"quantity"`
	Price            string    `json:"price"`
	Fees             string    `json:"fees"`
	Taxes            string    `json:"taxes"`
	OccurredAt       time.Time `json:"occurred_at"`
	Note             *string   `json:"note,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Holding struct {
	ID              string     `json:"id"`
	BrokerAccountID string     `json:"broker_account_id"`
	SecurityID      string     `json:"security_id"`
	Quantity        string     `json:"quantity"`
	CostBasisTotal  *string    `json:"cost_basis_total,omitempty"`
	AvgCost         *string    `json:"avg_cost,omitempty"`
	MarketPrice     *string    `json:"market_price,omitempty"`
	MarketValue     *string    `json:"market_value,omitempty"`
	UnrealizedPnL   *string    `json:"unrealized_pnl,omitempty"`
	AsOf            *time.Time `json:"as_of,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type ShareLot struct {
	ID              string    `json:"id"`
	BrokerAccountID string    `json:"broker_account_id"`
	SecurityID      string    `json:"security_id"`
	Quantity        string    `json:"quantity"`
	AcquisitionDate string    `json:"acquisition_date"`
	CostBasisPer    string    `json:"cost_basis_per_share"`
	Provenance      string    `json:"provenance"` // regular_buy, stock_dividend, rights_offering
	Status          string    `json:"status"`     // active, closed
	BuyTradeID      *string   `json:"buy_trade_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type RealizedTradeLog struct {
	ID              string    `json:"id"`
	BrokerAccountID string    `json:"broker_account_id"`
	SecurityID      string    `json:"security_id"`
	SellTradeID     string    `json:"sell_trade_id"`
	SourceShareLot  string    `json:"source_share_lot_id"`
	Quantity        string    `json:"quantity"`
	AcquisitionDate string    `json:"acquisition_date"`
	CostBasisTotal  string    `json:"cost_basis_total"`
	SellPrice       string    `json:"sell_price"`
	Proceeds        string    `json:"proceeds"`
	RealizedPnL     string    `json:"realized_pnl"`
	Provenance      string    `json:"provenance"`
	CreatedAt       time.Time `json:"created_at"`
}
