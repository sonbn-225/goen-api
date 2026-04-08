package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, userID uuid.UUID, tx entity.Transaction, lineItems []entity.TransactionLineItem, tagIDs []uuid.UUID) error
	CreateTransactionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, txEntity entity.Transaction, lineItems []entity.TransactionLineItem, tagIDs []uuid.UUID) error
	GetTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) (*entity.Transaction, error)
	ListTransactions(ctx context.Context, userID uuid.UUID, filter entity.TransactionListFilter) ([]entity.Transaction, *string, int, error)
	PatchTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID, patch entity.TransactionPatch) (*entity.Transaction, error)
	BatchPatchTransactions(ctx context.Context, userID uuid.UUID, transactionIDs []uuid.UUID, patches map[uuid.UUID]entity.TransactionPatch, mode string) ([]uuid.UUID, []uuid.UUID, error)
	DeleteTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) error
	DeleteTransactionTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, transactionID uuid.UUID) error
}

type TransactionService interface {
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateTransactionRequest) (*dto.TransactionResponse, error)
	Get(ctx context.Context, userID, transactionID uuid.UUID) (*dto.TransactionResponse, error)
	List(ctx context.Context, userID uuid.UUID, req dto.ListTransactionsRequest) ([]dto.TransactionResponse, *string, int, error)
	Patch(ctx context.Context, userID, transactionID uuid.UUID, req dto.TransactionPatchRequest) (*dto.TransactionResponse, error)
	BatchPatch(ctx context.Context, userID uuid.UUID, req dto.BatchPatchRequest) (*dto.BatchPatchResult, error)
	Delete(ctx context.Context, userID, transactionID uuid.UUID) error
	ListForExport(ctx context.Context, userID uuid.UUID, filter entity.TransactionListFilter) ([]entity.ExportTransactionRow, error)
}
