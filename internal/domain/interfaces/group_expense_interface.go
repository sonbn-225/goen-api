package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type GroupExpenseRepository interface {
	CreateGroupExpense(ctx context.Context, userID uuid.UUID, tx entity.Transaction, lineItems []entity.TransactionLineItem, tagIDs []uuid.UUID, participants []entity.GroupExpenseParticipant) error
	ListParticipantsByTransaction(ctx context.Context, userID, transactionID uuid.UUID) ([]entity.GroupExpenseParticipant, error)
	SettleParticipant(ctx context.Context, userID, participantID uuid.UUID, settlementTx entity.Transaction, settlementLineItems []entity.TransactionLineItem, settlementTagIDs []uuid.UUID) (settlementTransactionID uuid.UUID, err error)
	ListUniqueParticipantNames(ctx context.Context, userID uuid.UUID, limit int) ([]string, error)
	ListUnsettledParticipantsByName(ctx context.Context, userID uuid.UUID, name string) ([]entity.GroupExpenseParticipant, error)
	ListPublicParticipants(ctx context.Context, userID uuid.UUID) ([]string, error)
	ListPublicDebtsByParticipant(ctx context.Context, userID uuid.UUID, name string) ([]entity.PublicDebt, error)
}

type GroupExpenseService interface {
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateGroupExpenseRequest) (*dto.CreateGroupExpenseResponse, error)
	ListByTransaction(ctx context.Context, userID, transactionID uuid.UUID) ([]dto.GroupExpenseParticipantResponse, error)
	Settle(ctx context.Context, userID, participantID uuid.UUID, req dto.GroupExpenseSettleRequest) (*dto.TransactionResponse, error)
	ListUniqueParticipantNames(ctx context.Context, userID uuid.UUID, limit int) ([]string, error)
}

