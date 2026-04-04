package interfaces

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type GroupExpenseRepository interface {
	CreateGroupExpense(ctx context.Context, userID string, tx entity.Transaction, lineItems []entity.TransactionLineItem, tagIDs []string, participants []entity.GroupExpenseParticipant) error
	ListParticipantsByTransaction(ctx context.Context, userID, transactionID string) ([]entity.GroupExpenseParticipant, error)
	SettleParticipant(ctx context.Context, userID, participantID string, settlementTx entity.Transaction, settlementLineItems []entity.TransactionLineItem, settlementTagIDs []string) (settlementTransactionID string, err error)
	ListUniqueParticipantNames(ctx context.Context, userID string, limit int) ([]string, error)
	ListUnsettledParticipantsByName(ctx context.Context, userID string, name string) ([]entity.GroupExpenseParticipant, error)
}

type GroupExpenseService interface {
	Create(ctx context.Context, userID string, req dto.CreateGroupExpenseRequest) (*dto.CreateGroupExpenseResponse, error)
	ListByTransaction(ctx context.Context, userID, transactionID string) ([]entity.GroupExpenseParticipant, error)
	Settle(ctx context.Context, userID, participantID string, req dto.GroupExpenseSettleRequest) (*entity.Transaction, error)
	ListUniqueParticipantNames(ctx context.Context, userID string, limit int) ([]string, error)
}
