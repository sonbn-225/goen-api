package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// InterchangeRepository defines the persistence layer for staged imports, mapping rules, and future export tasks.
type InterchangeRepository interface {
	// Imports
	UpsertStagedImports(ctx context.Context, userID uuid.UUID, items []entity.StagedImportCreate) ([]entity.StagedImport, error)
	ListStagedImports(ctx context.Context, userID uuid.UUID, resourceType string) ([]entity.StagedImport, error)
	GetStagedImport(ctx context.Context, userID, id uuid.UUID) (*entity.StagedImport, error)
	PatchStagedImport(ctx context.Context, userID, id uuid.UUID, patch entity.StagedImportPatch) (*entity.StagedImport, error)
	DeleteStagedImport(ctx context.Context, userID, id uuid.UUID) error
	DeleteAllStagedImports(ctx context.Context, userID uuid.UUID, resourceType string) (int64, error)

	UpsertImportRules(ctx context.Context, userID uuid.UUID, rules []entity.StagedImportRuleUpsert) ([]entity.StagedImportRule, error)
	ListImportRules(ctx context.Context, userID uuid.UUID, resourceType string) ([]entity.StagedImportRule, error)
	DeleteImportRule(ctx context.Context, userID, id uuid.UUID) error
}

// InterchangeService defines the business logic for generic data interchange (Import & Export).
type InterchangeService interface {
	// Import Logic
	StageImport(ctx context.Context, userID uuid.UUID, resourceType string, source string, items []map[string]any) (int, int, []string, error)
	ListStaged(ctx context.Context, userID uuid.UUID, resourceType string) ([]dto.StagedImportResponse, error)
	PatchStaged(ctx context.Context, userID, id uuid.UUID, req dto.PatchStagedImportRequest) (*dto.StagedImportResponse, error)
	DeleteStaged(ctx context.Context, userID, id uuid.UUID) error
	ClearStaged(ctx context.Context, userID uuid.UUID, resourceType string) error

	UpsertRules(ctx context.Context, userID uuid.UUID, resourceType string, rules []dto.MappingRuleInput) ([]dto.ImportMappingRuleResponse, error)
	ListRules(ctx context.Context, userID uuid.UUID, resourceType string) ([]dto.ImportMappingRuleResponse, error)
	DeleteRule(ctx context.Context, userID, id uuid.UUID) error

	ApplyRulesAndCreate(ctx context.Context, userID uuid.UUID, resourceType string) (*dto.BatchImportResult, error)
	CreateManyFromStaged(ctx context.Context, userID uuid.UUID, resourceType string, ids []uuid.UUID) (*dto.BatchImportResult, error)

	// Export Logic
	ExportToCSV(ctx context.Context, userID uuid.UUID, resourceType string, filter any) ([]byte, string, error)
}
