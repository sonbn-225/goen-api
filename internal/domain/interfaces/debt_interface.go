package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type DebtRepository interface {
	CreateDebt(ctx context.Context, debt entity.Debt) error
	GetDebt(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) (*entity.Debt, error)
	ListDebts(ctx context.Context, userID uuid.UUID) ([]entity.Debt, error)
	UpdateDebt(ctx context.Context, userID uuid.UUID, debt entity.Debt) error
	DeleteDebt(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) error

	CreatePaymentLink(ctx context.Context, userID uuid.UUID, link entity.DebtPaymentLink, newPrincipal string, newOutstandingPrincipal string, newAccruedInterest string, newStatus string, closedAt *time.Time) error
	ListPaymentLinks(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) ([]entity.DebtPaymentLink, error)
	ListPaymentLinksByTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) ([]entity.DebtPaymentLink, error)

	CreateInstallment(ctx context.Context, userID uuid.UUID, inst entity.DebtInstallment) error
	ListInstallments(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) ([]entity.DebtInstallment, error)
}

type DebtService interface {
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateDebtRequest) (*dto.DebtResponse, error)
	Get(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) (*dto.DebtResponse, error)
	List(ctx context.Context, userID uuid.UUID) ([]dto.DebtResponse, error)
	Update(ctx context.Context, userID uuid.UUID, debtID uuid.UUID, req dto.UpdateDebtRequest) (*dto.DebtResponse, error)
	Delete(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) error

	AddPayment(ctx context.Context, userID uuid.UUID, debtID uuid.UUID, req dto.DebtPaymentRequest) (*dto.DebtResponse, error)
	ListPayments(ctx context.Context, userID uuid.UUID, debtID uuid.UUID) ([]dto.DebtPaymentLinkResponse, error)
}

