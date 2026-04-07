package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type SavingsRepository interface {
	CreateSavings(ctx context.Context, userID uuid.UUID, s entity.Savings) error
	GetSavings(ctx context.Context, userID uuid.UUID, savingsID uuid.UUID) (*entity.Savings, error)
	ListSavings(ctx context.Context, userID uuid.UUID) ([]entity.Savings, error)
	UpdateSavings(ctx context.Context, userID uuid.UUID, s entity.Savings) error
	DeleteSavings(ctx context.Context, userID uuid.UUID, savingsID uuid.UUID) error
}

type SavingsService interface {
	CreateSavings(ctx context.Context, userID uuid.UUID, req dto.CreateSavingsRequest) (*dto.SavingsResponse, error)
	GetSavings(ctx context.Context, userID, savingsID uuid.UUID) (*dto.SavingsResponse, error)
	ListSavings(ctx context.Context, userID uuid.UUID) ([]dto.SavingsResponse, error)
	PatchSavings(ctx context.Context, userID, savingsID uuid.UUID, req dto.PatchSavingsRequest) (*dto.SavingsResponse, error)
	DeleteSavings(ctx context.Context, userID, savingsID uuid.UUID) error
}

