package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, userID uuid.UUID, tx entity.Transaction, lineItems []entity.TransactionLineItem, tagIDs []uuid.UUID, participants []entity.GroupExpenseParticipant) error
	GetTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) (*entity.Transaction, error)
	ListTransactions(ctx context.Context, userID uuid.UUID, filter entity.TransactionListFilter) ([]entity.Transaction, *string, int, error)
	PatchTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID, patch entity.TransactionPatch) (*entity.Transaction, error)
	BatchPatchTransactions(ctx context.Context, userID uuid.UUID, transactionIDs []uuid.UUID, patches map[uuid.UUID]entity.TransactionPatch, mode string) ([]uuid.UUID, []uuid.UUID, error)
	DeleteTransaction(ctx context.Context, userID uuid.UUID, transactionID uuid.UUID) error

	// Imported Transactions
	CreateImportedTransactions(ctx context.Context, userID uuid.UUID, items []entity.ImportedTransactionCreate) ([]entity.ImportedTransaction, error)
	ListImportedTransactions(ctx context.Context, userID uuid.UUID) ([]entity.ImportedTransaction, error)
	PatchImportedTransaction(ctx context.Context, userID uuid.UUID, importID uuid.UUID, patch entity.ImportedTransactionPatch) (*entity.ImportedTransaction, error)
	DeleteImportedTransaction(ctx context.Context, userID uuid.UUID, importID uuid.UUID) error
	DeleteAllImportedTransactions(ctx context.Context, userID uuid.UUID) (int64, error)

	// Import Mapping Rules
	UpsertImportMappingRules(ctx context.Context, userID uuid.UUID, rules []entity.ImportMappingRuleUpsert) ([]entity.ImportMappingRule, error)
	ListImportMappingRules(ctx context.Context, userID uuid.UUID) ([]entity.ImportMappingRule, error)
	DeleteImportMappingRule(ctx context.Context, userID uuid.UUID, ruleID uuid.UUID) error
	GetImportedTransaction(ctx context.Context, userID uuid.UUID, importID uuid.UUID) (*entity.ImportedTransaction, error)
}

type TransactionService interface {
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateTransactionRequest) (*dto.TransactionResponse, error)
	Get(ctx context.Context, userID, transactionID uuid.UUID) (*dto.TransactionResponse, error)
	List(ctx context.Context, userID uuid.UUID, req dto.ListTransactionsRequest) ([]dto.TransactionResponse, *string, int, error)
	Patch(ctx context.Context, userID, transactionID uuid.UUID, req dto.TransactionPatchRequest) (*dto.TransactionResponse, error)
	BatchPatch(ctx context.Context, userID uuid.UUID, req dto.BatchPatchRequest) (*dto.BatchPatchResult, error)
	Delete(ctx context.Context, userID, transactionID uuid.UUID) error

	// Imports
	StageImport(ctx context.Context, userID uuid.UUID, items []dto.StageImportedItem) (int, int, []string, error)
	ListImported(ctx context.Context, userID uuid.UUID) ([]dto.ImportedTransactionResponse, error)
	PatchImported(ctx context.Context, userID, importID uuid.UUID, patch entity.ImportedTransactionPatch) (*dto.ImportedTransactionResponse, error)
	DeleteImported(ctx context.Context, userID, importID uuid.UUID) error
	ClearImported(ctx context.Context, userID uuid.UUID) error

	// Rules
	UpsertMappingRules(ctx context.Context, userID uuid.UUID, inputs []dto.MappingRuleInput) ([]dto.ImportMappingRuleResponse, error)
	ListMappingRules(ctx context.Context, userID uuid.UUID) ([]dto.ImportMappingRuleResponse, error)
	DeleteMappingRule(ctx context.Context, userID, ruleID uuid.UUID) error

	// Create from Imports
	CreateFromImported(ctx context.Context, userID, importID uuid.UUID) (*dto.TransactionResponse, error)
	CreateManyFromImported(ctx context.Context, userID uuid.UUID, importIDs []uuid.UUID) (*dto.BatchImportResult, error)
	ApplyRulesAndCreate(ctx context.Context, userID uuid.UUID) (*dto.BatchImportResult, error)
}

