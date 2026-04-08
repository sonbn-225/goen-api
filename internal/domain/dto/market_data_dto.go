package dto

import (
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// RefreshOneResponse represents the result of a single security refresh request.
// Used in: MarketDataHandler, MarketDataService, MarketDataInterface
// RefreshOneResponse represents the result of a single security refresh request.
// Used in: MarketDataHandler, MarketDataService, MarketDataInterface
type RefreshOneResponse struct {
	Stream    string `json:"stream"`     // Name of the message stream/queue
	MessageID string `json:"message_id"` // Unique ID of the enqueued refresh message
}

// RefreshManyResponse represents the result of a batch security refresh request.
// Used in: MarketDataHandler, MarketDataService, MarketDataInterface
type RefreshManyResponse struct {
	Stream     string   `json:"stream"`               // Name of the message stream/queue
	Enqueued   int      `json:"enqueued"`             // Number of refresh messages successfully enqueued
	MessageIDs []string `json:"message_ids"`           // List of enqueued message IDs
	NotFound   []string `json:"not_found_symbols,omitempty"` // List of symbols that were not found in the system
}

// SecurityStatus represents the synchronization status of a specific security.
// Used in: MarketDataHandler, MarketDataService, MarketDataInterface
type SecurityStatus struct {
	SecurityID uuid.UUID         `json:"security_id"`      // ID of the security
	Prices     *entity.SyncState `json:"prices_daily"`     // Current sync state of daily price data
	Events     *entity.SyncState `json:"security_events"`   // Current sync state of corporate events
	RateLimit  *entity.RateLimit `json:"rate_limit,omitempty"` // Provider rate limit information (if applicable)
}

// GlobalStatus represents the synchronization status of the entire market data system.
// Used in: MarketDataHandler, MarketDataService, MarketDataInterface
type GlobalStatus struct {
	MarketSync *entity.SyncState `json:"market_sync"`          // Overall sync state of the market data system
	RateLimit  *entity.RateLimit `json:"rate_limit,omitempty"` // Global provider rate limit status
}

// RefreshPriceRequest is the payload for refreshing a security's price data.
// Used in: MarketDataHandler, MarketDataService, MarketDataInterface
type RefreshPriceRequest struct {
	SecurityID uuid.UUID `json:"security_id"`      // ID of the security to refresh
	Force      *string   `json:"force,omitempty"`   // Whether to ignore existing data and force a fresh fetch
	Full       *string   `json:"full,omitempty"`    // Whether to perform a full historical backfill
	From       *string   `json:"from,omitempty"`    // Start date for the refresh (YYYY-MM-DD)
	To         *string   `json:"to,omitempty"`      // End date for the refresh (YYYY-MM-DD)
}

// RefreshEventRequest is the payload for refreshing a security's corporate event data.
// Used in: MarketDataHandler, MarketDataService, MarketDataInterface
type RefreshEventRequest struct {
	SecurityID uuid.UUID `json:"security_id"`    // ID of the security to refresh
	Force      *string   `json:"force,omitempty"` // Whether to force a fresh fetch
}

// RefreshSymbolsRequest is the payload for refreshing binary security data for multiple symbols.
// Used in: MarketDataHandler, MarketDataService, MarketDataInterface
type RefreshSymbolsRequest struct {
	Symbols       []string `json:"symbols"`        // List of ticker symbols to refresh
	IncludePrices bool     `json:"include_prices"` // Whether to include price data in the refresh
	IncludeEvents bool     `json:"include_events"` // Whether to include event data in the refresh
	Force         *string  `json:"force,omitempty"` // Whether to force a fresh fetch
}

// MarketSyncRequest is the payload for synchronizing all market data.
// Used in: MarketDataHandler, MarketDataService, MarketDataInterface
type MarketSyncRequest struct {
	IncludePrices bool    `json:"include_prices"` // Whether to sync price data for all securities
	IncludeEvents bool    `json:"include_events"` // Whether to sync event data for all securities
	Force         *string `json:"force,omitempty"` // Whether to force a fresh fetch
	Full          bool    `json:"full"`           // Whether to perform a full historical sync
}
