package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type SecurityRepository interface {
	GetSecurity(ctx context.Context, securityID uuid.UUID) (*entity.Security, error)
	ListSecurities(ctx context.Context) ([]entity.Security, error)

	ListSecurityPrices(ctx context.Context, securityID uuid.UUID, from *string, to *string) ([]entity.SecurityPriceDaily, error)
	ListSecurityEvents(ctx context.Context, securityID uuid.UUID, from *string, to *string) ([]entity.SecurityEvent, error)
	GetSecurityEvent(ctx context.Context, securityEventID uuid.UUID) (*entity.SecurityEvent, error)
}

type SecurityService interface {
	GetSecurity(ctx context.Context, securityID uuid.UUID) (*dto.SecurityResponse, error)
	ListSecurities(ctx context.Context) ([]dto.SecurityResponse, error)

	ListSecurityPrices(ctx context.Context, securityID uuid.UUID, from, to *string) ([]dto.SecurityPriceDailyResponse, error)
	ListSecurityEvents(ctx context.Context, securityID uuid.UUID, from, to *string) ([]dto.SecurityEventResponse, error)
}
