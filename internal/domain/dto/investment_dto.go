package dto

import (
	"time"

	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type PatchInvestmentAccountRequest struct {
	FeeSettings any `json:"fee_settings,omitempty"`
	TaxSettings any `json:"tax_settings,omitempty"`
}

type InvestmentAccountResponse struct {
	ID          string `json:"id"`
	AccountID   string `json:"account_id"`
	Currency    string `json:"currency"`
	FeeSettings any    `json:"fee_settings,omitempty"`
	TaxSettings any    `json:"tax_settings,omitempty"`
}

func NewInvestmentAccountResponse(a entity.InvestmentAccount) InvestmentAccountResponse {
	return InvestmentAccountResponse{
		ID:          a.ID,
		AccountID:   a.AccountID,
		Currency:    a.Currency,
		FeeSettings: a.FeeSettings,
		TaxSettings: a.TaxSettings,
	}
}

func NewInvestmentAccountResponses(items []entity.InvestmentAccount) []InvestmentAccountResponse {
	out := make([]InvestmentAccountResponse, len(items))
	for i, it := range items {
		out[i] = NewInvestmentAccountResponse(it)
	}
	return out
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

type TradeResponse struct {
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
}

func NewTradeResponse(t entity.Trade) TradeResponse {
	return TradeResponse{
		ID:               t.ID,
		ClientID:         t.ClientID,
		BrokerAccountID:  t.BrokerAccountID,
		SecurityID:       t.SecurityID,
		FeeTransactionID: t.FeeTransactionID,
		TaxTransactionID: t.TaxTransactionID,
		Side:             t.Side,
		Quantity:         t.Quantity,
		Price:            t.Price,
		Fees:             t.Fees,
		Taxes:            t.Taxes,
		OccurredAt:       t.OccurredAt,
		Note:             t.Note,
	}
}

func NewTradeResponses(items []entity.Trade) []TradeResponse {
	out := make([]TradeResponse, len(items))
	for i, it := range items {
		out[i] = NewTradeResponse(it)
	}
	return out
}

type SecurityResponse struct {
	ID         string  `json:"id"`
	Symbol     string  `json:"symbol"`
	Name       *string `json:"name,omitempty"`
	AssetClass *string `json:"asset_class,omitempty"`
	Currency   *string `json:"currency,omitempty"`
}

func NewSecurityResponse(s entity.Security) SecurityResponse {
	return SecurityResponse{
		ID:         s.ID,
		Symbol:     s.Symbol,
		Name:       s.Name,
		AssetClass: s.AssetClass,
		Currency:   s.Currency,
	}
}

func NewSecurityResponses(items []entity.Security) []SecurityResponse {
	out := make([]SecurityResponse, len(items))
	for i, it := range items {
		out[i] = NewSecurityResponse(it)
	}
	return out
}

type HoldingResponse struct {
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
}

func NewHoldingResponse(h entity.Holding) HoldingResponse {
	return HoldingResponse{
		ID:              h.ID,
		BrokerAccountID: h.BrokerAccountID,
		SecurityID:      h.SecurityID,
		Quantity:        h.Quantity,
		CostBasisTotal:  h.CostBasisTotal,
		AvgCost:         h.AvgCost,
		MarketPrice:     h.MarketPrice,
		MarketValue:     h.MarketValue,
		UnrealizedPnL:   h.UnrealizedPnL,
		AsOf:            h.AsOf,
	}
}

func NewHoldingResponses(items []entity.Holding) []HoldingResponse {
	out := make([]HoldingResponse, len(items))
	for i, it := range items {
		out[i] = NewHoldingResponse(it)
	}
	return out
}

type SecurityPriceDailyResponse struct {
	ID         string  `json:"id"`
	SecurityID string  `json:"security_id"`
	PriceDate  string  `json:"price_date"`
	Open       *string `json:"open,omitempty"`
	High       *string `json:"high,omitempty"`
	Low        *string `json:"low,omitempty"`
	Close      string  `json:"close"`
	Volume     *string `json:"volume,omitempty"`
}

func NewSecurityPriceDailyResponse(p entity.SecurityPriceDaily) SecurityPriceDailyResponse {
	return SecurityPriceDailyResponse{
		ID:         p.ID,
		SecurityID: p.SecurityID,
		PriceDate:  p.PriceDate,
		Open:       p.Open,
		High:       p.High,
		Low:        p.Low,
		Close:      p.Close,
		Volume:     p.Volume,
	}
}

func NewSecurityPriceDailyResponses(items []entity.SecurityPriceDaily) []SecurityPriceDailyResponse {
	out := make([]SecurityPriceDailyResponse, len(items))
	for i, it := range items {
		out[i] = NewSecurityPriceDailyResponse(it)
	}
	return out
}

type SecurityEventResponse struct {
	ID                 string  `json:"id"`
	SecurityID         string  `json:"security_id"`
	EventType          string  `json:"event_type"`
	ExDate             *string `json:"ex_date,omitempty"`
	RecordDate         *string `json:"record_date,omitempty"`
	PayDate            *string `json:"pay_date,omitempty"`
	EffectiveDate      *string `json:"effective_date,omitempty"`
	CashAmountPerShare *string `json:"cash_amount_per_share,omitempty"`
	RatioNumerator     *string `json:"ratio_numerator,omitempty"`
	RatioDenominator   *string `json:"ratio_denominator,omitempty"`
	SubscriptionPrice  *string `json:"subscription_price,omitempty"`
	Currency           *string `json:"currency,omitempty"`
	Note               *string `json:"note,omitempty"`
}

func NewSecurityEventResponse(e entity.SecurityEvent) SecurityEventResponse {
	return SecurityEventResponse{
		ID:                 e.ID,
		SecurityID:         e.SecurityID,
		EventType:          e.EventType,
		ExDate:             e.ExDate,
		RecordDate:         e.RecordDate,
		PayDate:            e.PayDate,
		EffectiveDate:      e.EffectiveDate,
		CashAmountPerShare: e.CashAmountPerShare,
		RatioNumerator:     e.RatioNumerator,
		RatioDenominator:   e.RatioDenominator,
		SubscriptionPrice:  e.SubscriptionPrice,
		Currency:           e.Currency,
		Note:               e.Note,
	}
}

func NewSecurityEventResponses(items []entity.SecurityEvent) []SecurityEventResponse {
	out := make([]SecurityEventResponse, len(items))
	for i, it := range items {
		out[i] = NewSecurityEventResponse(it)
	}
	return out
}

type EligibleAction struct {
	Event            SecurityEventResponse `json:"event"`
	HoldingQuantity  string                `json:"holding_quantity"`
	EntitledQuantity string                `json:"entitled_quantity"`
	Status           string                `json:"status"` // eligible, claimed, dismissed
	ElectionID       *string               `json:"election_id,omitempty"`
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
	TotalNet   string                  `json:"total_net"`
	TotalGross string                  `json:"total_gross"`
}
