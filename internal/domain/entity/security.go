package entity

import (
	"github.com/google/uuid"
)

// Security represents a tradable financial asset (e.g. Stock, ETF).
type Security struct {
	AuditEntity
	Symbol     string  `json:"symbol"`                // Ticker symbol (e.g., "AAPL", "VNM")
	Name       *string `json:"name,omitempty"`        // Full name of the security
	AssetClass *string `json:"asset_class,omitempty"` // Asset class (e.g., "Equity", "ETF", "Crypto")
	Currency   *string `json:"currency,omitempty"`    // Trading currency of the security
}

// SecurityPriceDaily records the closing price of a security for a specific date.
type SecurityPriceDaily struct {
	AuditEntity
	SecurityID uuid.UUID `json:"security_id"`      // ID of the security
	PriceDate  string    `json:"price_date"`       // Date of the price record (YYYY-MM-DD)
	Open       *string   `json:"open,omitempty"`   // Opening price (decimal string)
	High       *string   `json:"high,omitempty"`   // Highest price during the session (decimal string)
	Low        *string   `json:"low,omitempty"`    // Lowest price during the session (decimal string)
	Close      string    `json:"close"`            // Closing price (decimal string)
	Volume     *string   `json:"volume,omitempty"` // Trading volume (decimal string)
}

// SecurityEvent represents a corporate action (Dividend, Split, etc.) for a security.
type SecurityEvent struct {
	AuditEntity
	SecurityID         uuid.UUID         `json:"security_id"`                     // ID of the security
	EventType          SecurityEventType `json:"event_type"`                      // Type of corporate action
	ExDate             *string           `json:"ex_date,omitempty"`               // Ex-dividend or ex-event date (YYYY-MM-DD)
	RecordDate         *string           `json:"record_date,omitempty"`           // Date to determine eligibility (YYYY-MM-DD)
	PayDate            *string           `json:"pay_date,omitempty"`              // Date of distribution/execution (YYYY-MM-DD)
	EffectiveDate      *string           `json:"effective_date,omitempty"`        // Date the event takes effect (for splits/mergers)
	CashAmountPerShare *string           `json:"cash_amount_per_share,omitempty"` // Cash dividend amount (decimal string)
	RatioNumerator     *string           `json:"ratio_numerator,omitempty"`       // Split/bonus ratio numerator
	RatioDenominator   *string           `json:"ratio_denominator,omitempty"`     // Split/bonus ratio denominator
	SubscriptionPrice  *string           `json:"subscription_price,omitempty"`    // Rights offering price (decimal string)
	Currency           *string           `json:"currency,omitempty"`              // Currency of the cash amount
	Note               *string           `json:"note,omitempty"`                  // Optional notes about the event
}
