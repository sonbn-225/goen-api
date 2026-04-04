package interfaces

import (
	"context"
	"time"

	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type DebtRepository interface {
	CreateDebt(ctx context.Context, debt entity.Debt) error
	GetDebt(ctx context.Context, userID string, debtID string) (*entity.Debt, error)
	ListDebts(ctx context.Context, userID string) ([]entity.Debt, error)
	UpdateDebt(ctx context.Context, userID string, debt entity.Debt) error
	DeleteDebt(ctx context.Context, userID string, debtID string) error

	CreatePaymentLink(ctx context.Context, userID string, link entity.DebtPaymentLink, newPrincipal string, newOutstandingPrincipal string, newAccruedInterest string, newStatus string, closedAt *time.Time) error
	ListPaymentLinks(ctx context.Context, userID string, debtID string) ([]entity.DebtPaymentLink, error)
	ListPaymentLinksByTransaction(ctx context.Context, userID string, transactionID string) ([]entity.DebtPaymentLink, error)

	CreateInstallment(ctx context.Context, userID string, inst entity.DebtInstallment) error
	ListInstallments(ctx context.Context, userID string, debtID string) ([]entity.DebtInstallment, error)
}

type DebtService interface {
	Create(ctx context.Context, userID string, req dto.CreateDebtRequest) (*dto.DebtResponse, error)
	Get(ctx context.Context, userID string, debtID string) (*dto.DebtResponse, error)
	List(ctx context.Context, userID string) ([]dto.DebtResponse, error)
	Update(ctx context.Context, userID string, debtID string, req dto.UpdateDebtRequest) (*dto.DebtResponse, error)
	Delete(ctx context.Context, userID string, debtID string) error

	AddPayment(ctx context.Context, userID string, debtID string, req dto.DebtPaymentRequest) (*dto.DebtResponse, error)
	ListPayments(ctx context.Context, userID string, debtID string) ([]dto.DebtPaymentLinkResponse, error)
}
