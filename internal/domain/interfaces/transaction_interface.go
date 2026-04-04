package interfaces

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, userID string, tx entity.Transaction, lineItems []entity.TransactionLineItem, tagIDs []string, participants []entity.GroupExpenseParticipant) error
	GetTransaction(ctx context.Context, userID string, transactionID string) (*entity.Transaction, error)
	ListTransactions(ctx context.Context, userID string, filter entity.TransactionListFilter) ([]entity.Transaction, *string, int, error)
	PatchTransaction(ctx context.Context, userID string, transactionID string, patch entity.TransactionPatch) (*entity.Transaction, error)
	BatchPatchTransactions(ctx context.Context, userID string, transactionIDs []string, patches map[string]entity.TransactionPatch, mode string) ([]string, []string, error)
	DeleteTransaction(ctx context.Context, userID string, transactionID string) error

	// Imported Transactions
	CreateImportedTransactions(ctx context.Context, userID string, items []entity.ImportedTransactionCreate) ([]entity.ImportedTransaction, error)
	ListImportedTransactions(ctx context.Context, userID string) ([]entity.ImportedTransaction, error)
	PatchImportedTransaction(ctx context.Context, userID string, importID string, patch entity.ImportedTransactionPatch) (*entity.ImportedTransaction, error)
	DeleteImportedTransaction(ctx context.Context, userID string, importID string) error
	DeleteAllImportedTransactions(ctx context.Context, userID string) (int64, error)

	// Import Mapping Rules
	UpsertImportMappingRules(ctx context.Context, userID string, rules []entity.ImportMappingRuleUpsert) ([]entity.ImportMappingRule, error)
	ListImportMappingRules(ctx context.Context, userID string) ([]entity.ImportMappingRule, error)
	DeleteImportMappingRule(ctx context.Context, userID string, ruleID string) error
}

type TransactionService interface {
	Create(ctx context.Context, userID string, req dto.CreateTransactionRequest) (*entity.Transaction, error)
	Get(ctx context.Context, userID, transactionID string) (*entity.Transaction, error)
	List(ctx context.Context, userID string, req dto.CreateTransactionRequest) ([]entity.Transaction, *string, int, error) // Note: Filter is DTO-ed
	Patch(ctx context.Context, userID, transactionID string, req dto.TransactionPatchRequest) (*entity.Transaction, error)
	BatchPatch(ctx context.Context, userID string, req dto.BatchPatchRequest) (*dto.BatchPatchResult, error)
	Delete(ctx context.Context, userID, transactionID string) error

	// Imports
	StageImport(ctx context.Context, userID string, items []dto.StageImportedItem) (int, int, []string, error)
	ListImported(ctx context.Context, userID string) ([]entity.ImportedTransaction, error)
	PatchImported(ctx context.Context, userID, importID string, patch entity.ImportedTransactionPatch) (*entity.ImportedTransaction, error)
	DeleteImported(ctx context.Context, userID, importID string) error
	ClearImported(ctx context.Context, userID string) error

	// Rules
	UpsertMappingRules(ctx context.Context, userID string, inputs []dto.MappingRuleInput) ([]entity.ImportMappingRule, error)
	ListMappingRules(ctx context.Context, userID string) ([]entity.ImportMappingRule, error)
	DeleteMappingRule(ctx context.Context, userID, ruleID string) error
}
