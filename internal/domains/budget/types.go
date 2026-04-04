package budget

import (
	"context"
	"time"

	"github.com/sonbn-225/goen-api-v2/internal/domains/category"
)

type Budget struct {
	ID                    string    `json:"id"`
	UserID                string    `json:"user_id"`
	Name                  *string   `json:"name,omitempty"`
	Period                string    `json:"period"`
	PeriodStart           *string   `json:"period_start,omitempty"`
	PeriodEnd             *string   `json:"period_end,omitempty"`
	Amount                string    `json:"amount"`
	AlertThresholdPercent *int      `json:"alert_threshold_percent,omitempty"`
	RolloverMode          *string   `json:"rollover_mode,omitempty"`
	CategoryID            *string   `json:"category_id,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type WithStats struct {
	Budget
	Spent       string `json:"spent"`
	Remaining   string `json:"remaining"`
	PercentUsed int    `json:"percent_used"`
}

type CreateInput struct {
	Name                  *string `json:"name,omitempty"`
	Period                string  `json:"period"`
	PeriodStart           *string `json:"period_start,omitempty"`
	PeriodEnd             *string `json:"period_end,omitempty"`
	Amount                string  `json:"amount"`
	AlertThresholdPercent *int    `json:"alert_threshold_percent,omitempty"`
	RolloverMode          *string `json:"rollover_mode,omitempty"`
	CategoryID            *string `json:"category_id,omitempty"`
}

type Repository interface {
	Create(ctx context.Context, userID string, input Budget) error
	GetByID(ctx context.Context, userID, budgetID string) (*Budget, error)
	ListByUser(ctx context.Context, userID string) ([]Budget, error)
	ComputeSpent(ctx context.Context, userID, categoryID, startDate, endDate string) (string, error)
}

type CategoryRepository interface {
	GetByID(ctx context.Context, userID, categoryID string) (*category.Category, error)
}

type Service interface {
	Create(ctx context.Context, userID string, input CreateInput) (*WithStats, error)
	Get(ctx context.Context, userID, budgetID string) (*WithStats, error)
	List(ctx context.Context, userID string) ([]WithStats, error)
}

type ModuleDeps struct {
	Repo         Repository
	CategoryRepo CategoryRepository
	Service      Service
}

type Module struct {
	Service Service
	Handler *Handler
}
