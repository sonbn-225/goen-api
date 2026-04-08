package dto

import (
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

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
