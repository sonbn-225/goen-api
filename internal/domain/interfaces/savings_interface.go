package interfaces

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type SavingsRepository interface {
	CreateSavings(ctx context.Context, userID string, s entity.Savings) error
	GetSavings(ctx context.Context, userID string, savingsID string) (*entity.Savings, error)
	ListSavings(ctx context.Context, userID string) ([]entity.Savings, error)
	UpdateSavings(ctx context.Context, userID string, s entity.Savings) error
	DeleteSavings(ctx context.Context, userID string, savingsID string) error
}

type SavingsService interface {
	CreateSavings(ctx context.Context, userID string, req dto.CreateSavingsRequest) (*dto.SavingsResponse, error)
	GetSavings(ctx context.Context, userID, savingsID string) (*dto.SavingsResponse, error)
	ListSavings(ctx context.Context, userID string) ([]dto.SavingsResponse, error)
	PatchSavings(ctx context.Context, userID, savingsID string, req dto.PatchSavingsRequest) (*dto.SavingsResponse, error)
	DeleteSavings(ctx context.Context, userID, savingsID string) error
}
