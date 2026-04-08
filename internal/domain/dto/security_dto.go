package dto

import (
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// SecurityResponse represents a financial security (stock, bond, etc.).
// Used in: SecurityHandler, SecurityService, SecurityInterface
type SecurityResponse struct {
	ID         uuid.UUID `json:"id"`                    // Unique security identifier
	Symbol     string    `json:"symbol"`                // Ticker symbol
	Name       *string   `json:"name,omitempty"`         // Full name of the security
	AssetClass *string   `json:"asset_class,omitempty"`  // Asset class (e.g., Equity)
	Currency   *string   `json:"currency,omitempty"`    // Trading currency
}

// SecurityPriceDailyResponse represents a security's price for a specific day.
// Used in: SecurityHandler, SecurityService, SecurityInterface
type SecurityPriceDailyResponse struct {
	ID         uuid.UUID `json:"id"`              // Unique price record identifier
	SecurityID uuid.UUID `json:"security_id"`    // ID of the security
	PriceDate  string    `json:"price_date"`     // Date of the price (YYYY-MM-DD)
	Open       *string   `json:"open,omitempty"`  // Opening price
	High       *string   `json:"high,omitempty"`  // Highest price of day
	Low        *string   `json:"low,omitempty"`   // Lowest price of day
	Close      string    `json:"close"`          // Closing price
	Volume     *string   `json:"volume,omitempty"` // Trading volume
}

// SecurityEventResponse represents a corporate action or security event (e.g., dividend, split).
// Used in: SecurityHandler, SecurityService, SecurityInterface
type SecurityEventResponse struct {
	ID                 uuid.UUID `json:"id"`                             // Unique event identifier
	SecurityID         uuid.UUID `json:"security_id"`                    // ID of the security
	EventType          entity.SecurityEventType `json:"event_type"`      // Type of action (Dividend, Split, etc.)
	ExDate             *string   `json:"ex_date,omitempty"`              // Ex-event date (YYYY-MM-DD)
	RecordDate         *string   `json:"record_date,omitempty"`          // Eligibility date (YYYY-MM-DD)
	PayDate            *string   `json:"pay_date,omitempty"`             // Payment date (YYYY-MM-DD)
	EffectiveDate      *string   `json:"effective_date,omitempty"`       // Effective date (YYYY-MM-DD)
	CashAmountPerShare *string   `json:"cash_amount_per_share,omitempty"` // Cash dividend per share
	RatioNumerator     *string   `json:"ratio_numerator,omitempty"`       // Split ratio numerator
	RatioDenominator   *string   `json:"ratio_denominator,omitempty"`     // Split ratio denominator
	SubscriptionPrice  *string   `json:"subscription_price,omitempty"`    // Rights offering price
	Currency           *string   `json:"currency,omitempty"`              // Currency of cash amount
	Note               *string   `json:"note,omitempty"`                  // Optional note
}
