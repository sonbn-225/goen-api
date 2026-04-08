package entity

import (
	"github.com/google/uuid"
)

// StagedImport represents a generic record from an external source awaiting processing.
type StagedImport struct {
	AuditEntity
	UserID       uuid.UUID      `json:"-"`                     // Owner of the import
	ResourceType string         `json:"resource_type"`         // e.g., "transaction"
	Source       string         `json:"source"`                // e.g., "manual_csv", "bank_api"
	ExternalID   *string        `json:"external_id,omitempty"` // ID from the original source
	Data         map[string]any `json:"data"`                  // Raw data fields (JSONB)
	Metadata     map[string]any `json:"metadata"`              // Enrichment/Mapping data (JSONB)
	Status       string         `json:"status"`                // e.g., "pending", "processed", "error"
}

// StagedImportRule defines automatic mapping logic for staged imports.
type StagedImportRule struct {
	AuditEntity
	UserID       uuid.UUID             `json:"-"`             // Owner of the rule
	ResourceType string                `json:"resource_type"` // e.g., "transaction"
	RuleType     ImportMappingRuleKind `json:"rule_type"`     // e.g., "account", "category"
	MatchKey     string                `json:"match_key"`     // The field in Data to match (default "source_name")
	MatchValue   string                `json:"match_value"`   // The value to look for
	MappedID     uuid.UUID             `json:"mapped_id"`     // The system ID to assign
}

// StagedImportCreate defines parameters for creating a new StagedImport.
type StagedImportCreate struct {
	ResourceType string
	Source       string
	ExternalID   *string
	Data         map[string]any
	Metadata     map[string]any
}

// StagedImportRuleUpsert defines parameters for rule creation/update.
type StagedImportRuleUpsert struct {
	ResourceType string
	RuleType     ImportMappingRuleKind
	MatchKey     string
	MatchValue   string
	MappedID     uuid.UUID
}

// StagedImportPatch defines the fields that can be updated for a StagedImport.
type StagedImportPatch struct {
	Metadata map[string]any // Merges with existing metadata
	Status   *string
}

// ExportTransactionRow represents a single transaction record formatted for CSV/Excel export.
type ExportTransactionRow struct {
	ID           uuid.UUID `json:"id"`
	Description  *string   `json:"description"`
	Amount       string    `json:"amount"`
	Type         string    `json:"type"`
	OccurredDate string    `json:"occurred_date"`
	AccountName  *string   `json:"account_name"`
	CategoryName *string   `json:"category_name"`
	TagName      *string   `json:"tag_name"`
	ExternalRef  *string   `json:"external_ref"`
}
