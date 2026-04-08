package dto

import (
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

func NewTradeResponse(t entity.Trade) TradeResponse {
	return TradeResponse{
		ID:               t.ID,
		AccountID:        t.AccountID,
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

func NewHoldingResponse(h entity.Holding) HoldingResponse {
	return HoldingResponse{
		ID:              h.ID,
		AccountID:       h.AccountID,
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

