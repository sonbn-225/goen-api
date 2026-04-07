package dto

import (
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type RefreshOneResponse struct {
	Stream    string `json:"stream"`
	MessageID string `json:"message_id"`
}

type RefreshManyResponse struct {
	Stream     string   `json:"stream"`
	Enqueued   int      `json:"enqueued"`
	MessageIDs []string `json:"message_ids"`
	NotFound   []string `json:"not_found_symbols,omitempty"`
}

type SecurityStatus struct {
	SecurityID uuid.UUID         `json:"security_id"`
	Prices     *entity.SyncState `json:"prices_daily"`
	Events     *entity.SyncState `json:"security_events"`
	RateLimit  *entity.RateLimit `json:"rate_limit,omitempty"`
}

type GlobalStatus struct {
	MarketSync *entity.SyncState `json:"market_sync"`
	RateLimit  *entity.RateLimit `json:"rate_limit,omitempty"`
}

type RefreshPriceRequest struct {
	SecurityID uuid.UUID `json:"security_id"`
	Force      *string   `json:"force,omitempty"`
	Full       *string   `json:"full,omitempty"`
	From       *string   `json:"from,omitempty"`
	To         *string   `json:"to,omitempty"`
}

type RefreshEventRequest struct {
	SecurityID uuid.UUID `json:"security_id"`
	Force      *string   `json:"force,omitempty"`
}

type RefreshSymbolsRequest struct {
	Symbols       []string `json:"symbols"`
	IncludePrices bool     `json:"include_prices"`
	IncludeEvents bool     `json:"include_events"`
	Force         *string  `json:"force,omitempty"`
}

type MarketSyncRequest struct {
	IncludePrices bool    `json:"include_prices"`
	IncludeEvents bool    `json:"include_events"`
	Force         *string `json:"force,omitempty"`
	Full          bool    `json:"full"`
}
