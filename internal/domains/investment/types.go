package investment

import (
	"context"
	"time"
)

type InvestmentAccount struct {
	ID          string    `json:"id"`
	AccountID   string    `json:"account_id"`
	Currency    string    `json:"currency"`
	FeeSettings any       `json:"fee_settings,omitempty"`
	TaxSettings any       `json:"tax_settings,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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

type UpdateInvestmentAccountSettingsInput struct {
	FeeSettings any `json:"fee_settings,omitempty"`
	TaxSettings any `json:"tax_settings,omitempty"`
}

type CreateTradeInput struct {
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

type Repository interface {
	GetInvestmentAccount(ctx context.Context, userID, investmentAccountID string) (*InvestmentAccount, error)
	ListInvestmentAccounts(ctx context.Context, userID string) ([]InvestmentAccount, error)
	UpdateInvestmentAccountSettings(ctx context.Context, userID, investmentAccountID string, feeSettings any, taxSettings any) (*InvestmentAccount, error)

	GetSecurity(ctx context.Context, securityID string) (*Security, error)
	ListSecurities(ctx context.Context) ([]Security, error)
	ListSecurityPrices(ctx context.Context, securityID string, from *string, to *string) ([]SecurityPriceDaily, error)
	ListSecurityEvents(ctx context.Context, securityID string, from *string, to *string) ([]SecurityEvent, error)

	CreateTrade(ctx context.Context, userID string, trade Trade) error
	ListTrades(ctx context.Context, userID, brokerAccountID string) ([]Trade, error)

	ListHoldings(ctx context.Context, userID, brokerAccountID string) ([]Holding, error)
	GetHolding(ctx context.Context, userID, brokerAccountID, securityID string) (*Holding, error)
	UpsertHolding(ctx context.Context, userID string, holding Holding) (*Holding, error)
}

type Service interface {
	ListInvestmentAccounts(ctx context.Context, userID string) ([]InvestmentAccount, error)
	GetInvestmentAccount(ctx context.Context, userID, investmentAccountID string) (*InvestmentAccount, error)
	UpdateInvestmentAccountSettings(ctx context.Context, userID, investmentAccountID string, input UpdateInvestmentAccountSettingsInput) (*InvestmentAccount, error)

	ListSecurities(ctx context.Context, userID string) ([]Security, error)
	GetSecurity(ctx context.Context, userID, securityID string) (*Security, error)
	ListSecurityPrices(ctx context.Context, userID, securityID string, from *string, to *string) ([]SecurityPriceDaily, error)
	ListSecurityEvents(ctx context.Context, userID, securityID string, from *string, to *string) ([]SecurityEvent, error)

	CreateTrade(ctx context.Context, userID, investmentAccountID string, input CreateTradeInput) (*Trade, error)
	ListTrades(ctx context.Context, userID, investmentAccountID string) ([]Trade, error)
	ListHoldings(ctx context.Context, userID, investmentAccountID string) ([]Holding, error)
}

type ModuleDeps struct {
	Repo    Repository
	Service Service
}

type Module struct {
	Service Service
	Handler *Handler
}
