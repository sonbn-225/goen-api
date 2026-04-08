package dto

import (
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

func NewAccountSettingsResponse(it entity.AccountSettings) AccountSettingsResponse {
	return AccountSettingsResponse{
		Color:      it.Color,
		Investment: mapInvestmentSettingsResponse(it.Investment),
		Savings:    mapSavingsSettingsResponse(it.Savings),
	}
}

func mapInvestmentSettingsResponse(it *entity.InvestmentSettings) *InvestmentSettingsResponse {
	if it == nil {
		return nil
	}
	return &InvestmentSettingsResponse{
		FeeSettings: it.FeeSettings,
		TaxSettings: it.TaxSettings,
	}
}

func mapSavingsSettingsResponse(it *entity.SavingsSettings) *SavingsSettingsResponse {
	if it == nil {
		return nil
	}
	return &SavingsSettingsResponse{
		TargetAmount: it.TargetAmount,
		TargetDate:   it.TargetDate,
	}
}
