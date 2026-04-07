package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type BudgetRepository interface {
	CreateBudget(ctx context.Context, userID uuid.UUID, b entity.Budget) error
	GetBudget(ctx context.Context, userID uuid.UUID, budgetID uuid.UUID) (*entity.Budget, error)
	ListBudgets(ctx context.Context, userID uuid.UUID) ([]entity.Budget, error)
	UpdateBudget(ctx context.Context, userID uuid.UUID, b entity.Budget) error
	DeleteBudget(ctx context.Context, userID uuid.UUID, budgetID uuid.UUID) error
	ComputeSpent(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID, startDate string, endDate string) (string, error)
}

type BudgetService interface {
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateBudgetRequest) (*dto.BudgetWithStatsResponse, error)
	Get(ctx context.Context, userID uuid.UUID, budgetID uuid.UUID) (*dto.BudgetWithStatsResponse, error)
	List(ctx context.Context, userID uuid.UUID) ([]dto.BudgetWithStatsResponse, error)
	Update(ctx context.Context, userID uuid.UUID, budgetID uuid.UUID, req dto.UpdateBudgetRequest) (*dto.BudgetWithStatsResponse, error)
	Delete(ctx context.Context, userID uuid.UUID, budgetID uuid.UUID) error
}

