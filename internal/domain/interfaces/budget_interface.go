package interfaces

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type BudgetRepository interface {
	CreateBudget(ctx context.Context, userID string, b entity.Budget) error
	GetBudget(ctx context.Context, userID string, budgetID string) (*entity.Budget, error)
	ListBudgets(ctx context.Context, userID string) ([]entity.Budget, error)
	UpdateBudget(ctx context.Context, userID string, b entity.Budget) error
	DeleteBudget(ctx context.Context, userID string, budgetID string) error
	ComputeSpent(ctx context.Context, userID string, categoryID string, startDate string, endDate string) (string, error)
}

type BudgetService interface {
	Create(ctx context.Context, userID string, req dto.CreateBudgetRequest) (*dto.BudgetWithStatsResponse, error)
	Get(ctx context.Context, userID string, budgetID string) (*dto.BudgetWithStatsResponse, error)
	List(ctx context.Context, userID string) ([]dto.BudgetWithStatsResponse, error)
	Update(ctx context.Context, userID string, budgetID string, req dto.UpdateBudgetRequest) (*dto.BudgetWithStatsResponse, error)
	Delete(ctx context.Context, userID string, budgetID string) error
}
