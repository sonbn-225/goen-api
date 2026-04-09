package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
)

type SecurityService struct {
	repo interfaces.SecurityRepository
}

func NewSecurityService(repo interfaces.SecurityRepository) *SecurityService {
	return &SecurityService{repo: repo}
}

func (s *SecurityService) GetSecurity(ctx context.Context, securityID uuid.UUID) (*dto.SecurityResponse, error) {
	it, err := s.repo.GetSecurityTx(ctx, nil, securityID)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, nil
	}
	resp := dto.NewSecurityResponse(*it)
	return &resp, nil
}

func (s *SecurityService) ListSecurities(ctx context.Context) ([]dto.SecurityResponse, error) {
	items, err := s.repo.ListSecuritiesTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return dto.NewSecurityResponses(items), nil
}

func (s *SecurityService) ListSecurityPrices(ctx context.Context, securityID uuid.UUID, from, to *string) ([]dto.SecurityPriceDailyResponse, error) {
	items, err := s.repo.ListSecurityPricesTx(ctx, nil, securityID, from, to)
	if err != nil {
		return nil, err
	}
	return dto.NewSecurityPriceDailyResponses(items), nil
}

func (s *SecurityService) ListSecurityEvents(ctx context.Context, securityID uuid.UUID, from, to *string) ([]dto.SecurityEventResponse, error) {
	items, err := s.repo.ListSecurityEventsTx(ctx, nil, securityID, from, to)
	if err != nil {
		return nil, err
	}
	return dto.NewSecurityEventResponses(items), nil
}
