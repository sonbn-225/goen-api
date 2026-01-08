package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvestmentAccountNotFound     = errors.New("investment account not found")
	ErrSecurityNotFound              = errors.New("security not found")
	ErrTradeNotFound                 = errors.New("trade not found")
	ErrHoldingNotFound               = errors.New("holding not found")
	ErrSecurityEventNotFound         = errors.New("security event not found")
	ErrSecurityEventElectionNotFound = errors.New("security event election not found")
	ErrInvestmentForbidden           = errors.New("investment forbidden")
)

type InvestmentAccount struct {
	ID           string    `json:"id"`
	AccountID    string    `json:"account_id"`
	BrokerName   *string   `json:"broker_name,omitempty"`
	Currency     string    `json:"currency"`
	SyncEnabled  bool      `json:"sync_enabled"`
	SyncSettings any       `json:"sync_settings,omitempty"`
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
	EventType          string    `json:"event_type"`
	ExDate             *string   `json:"ex_date,omitempty"`
	RecordDate         *string   `json:"record_date,omitempty"`
	PayDate            *string   `json:"pay_date,omitempty"`
	EffectiveDate      *string   `json:"effective_date,omitempty"`
	CashAmountPerShare *string   `json:"cash_amount_per_share,omitempty"`
	RatioNumerator     *string   `json:"ratio_numerator,omitempty"`
	RatioDenominator   *string   `json:"ratio_denominator,omitempty"`
	SubscriptionPrice  *string   `json:"subscription_price,omitempty"`
	Currency           *string   `json:"currency,omitempty"`
	VnstockEventID     *string   `json:"vnstock_event_id,omitempty"`
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
	Status                       string     `json:"status"`
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
	Side             string    `json:"side"`
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
	SourceOfTruth   string     `json:"source_of_truth"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type InvestmentRepository interface {
	// Investment accounts
	CreateInvestmentAccount(ctx context.Context, userID string, ia InvestmentAccount) error
	GetInvestmentAccount(ctx context.Context, userID string, investmentAccountID string) (*InvestmentAccount, error)
	ListInvestmentAccounts(ctx context.Context, userID string) ([]InvestmentAccount, error)

	// Securities
	CreateSecurity(ctx context.Context, s Security) error
	GetSecurity(ctx context.Context, securityID string) (*Security, error)
	ListSecurities(ctx context.Context) ([]Security, error)

	// Read-only market data
	ListSecurityPrices(ctx context.Context, securityID string, from *string, to *string) ([]SecurityPriceDaily, error)
	ListSecurityEvents(ctx context.Context, securityID string, from *string, to *string) ([]SecurityEvent, error)
	GetSecurityEvent(ctx context.Context, securityEventID string) (*SecurityEvent, error)

	// Elections
	UpsertSecurityEventElection(ctx context.Context, userID string, e SecurityEventElection) (*SecurityEventElection, error)
	ListSecurityEventElections(ctx context.Context, userID string, brokerAccountID string, status *string) ([]SecurityEventElection, error)

	// Trades
	CreateTrade(ctx context.Context, userID string, t Trade) error
	ListTrades(ctx context.Context, userID string, brokerAccountID string) ([]Trade, error)

	// Holdings
	ListHoldings(ctx context.Context, userID string, brokerAccountID string) ([]Holding, error)
	GetHolding(ctx context.Context, userID string, brokerAccountID string, securityID string) (*Holding, error)
}
