package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

// StagedImportResponse represents a single item in the staging area.
type StagedImportResponse struct {
	ID           uuid.UUID      `json:"id"`
	ResourceType string         `json:"resource_type"`
	Source       string         `json:"source"`
	ExternalID   *string        `json:"external_id,omitempty"`
	Data         map[string]any `json:"data"`
	Metadata     map[string]any `json:"metadata"`
	Status       string         `json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// StageImportRequest is the payload for staging multiple items.
type StageImportRequest struct {
	Source string           `json:"source"`
	Items  []map[string]any `json:"items"`
}

// PatchStagedImportRequest is the payload for updating a staged item.
type PatchStagedImportRequest struct {
	Metadata map[string]any `json:"metadata,omitempty"`
	Status   *string        `json:"status,omitempty"`
}

// MappingRuleInput is the payload for creating or updating an import mapping rule.
type MappingRuleInput struct {
	RuleType   entity.ImportMappingRuleKind `json:"rule_type"`   // e.g., "account", "category"
	MatchKey   string                       `json:"match_key"`   // Field to match in source data
	MatchValue string                       `json:"match_value"` // Value to match
	MappedID   uuid.UUID                    `json:"mapped_id"`   // System ID to map to
}

// ImportMappingRuleResponse represents an existing mapping rule.
type ImportMappingRuleResponse struct {
	ID           uuid.UUID                    `json:"id"`
	ResourceType string                       `json:"resource_type"`
	RuleType     entity.ImportMappingRuleKind `json:"rule_type"`
	MatchKey     string                       `json:"match_key"`
	MatchValue   string                       `json:"match_value"`
	MappedID     uuid.UUID                    `json:"mapped_id"`
}

// BatchImportResult represents the outcome of a batch promotion operation.
type BatchImportResult struct {
	Created int      `json:"created"`
	Skipped int      `json:"skipped"`
	Errors  []string `json:"errors,omitempty"`
}

// CreateManyImportedRequest is the payload for final processing of multiple staged items.
type CreateManyImportedRequest struct {
	IDs []uuid.UUID `json:"ids"`
}

// UpsertMappingRulesRequest is the payload for bulk creating or updating mapping rules.
type UpsertMappingRulesRequest struct {
	Rules []MappingRuleInput `json:"rules"`
}
