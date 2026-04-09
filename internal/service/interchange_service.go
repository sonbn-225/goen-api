package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

type InterchangeService struct {
	repo  interfaces.InterchangeRepository
	txSvc interfaces.TransactionService
	db    *database.Postgres
}

func NewInterchangeService(
	repo interfaces.InterchangeRepository,
	txSvc interfaces.TransactionService,
	db *database.Postgres,
) *InterchangeService {
	return &InterchangeService{
		repo:  repo,
		txSvc: txSvc,
		db:    db,
	}
}

// StageImport implementation (generic)
func (s *InterchangeService) StageImport(ctx context.Context, userID uuid.UUID, resourceType string, source string, items []map[string]any) (int, int, []string, error) {
	if len(items) == 0 {
		return 0, 0, nil, nil
	}

	creates := make([]entity.StagedImportCreate, 0, len(items))
	for _, item := range items {
		extID := ""
		if val, ok := item["external_id"].(string); ok {
			extID = val
		}

		creates = append(creates, entity.StagedImportCreate{
			ResourceType: resourceType,
			Source:       source,
			ExternalID:   &extID,
			Data:         item,
			Metadata:     make(map[string]any),
		})
	}

	var staged []entity.StagedImport
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		var err error
		staged, err = s.repo.UpsertStagedImportsTx(ctx, tx, userID, creates)
		return err
	})
	if err != nil {
		return 0, 0, nil, err
	}

	return len(staged), 0, nil, nil
}

func (s *InterchangeService) ListStaged(ctx context.Context, userID uuid.UUID, resourceType string) ([]dto.StagedImportResponse, error) {
	items, err := s.repo.ListStagedImportsTx(ctx, nil, userID, resourceType)
	if err != nil {
		return nil, err
	}

	res := make([]dto.StagedImportResponse, len(items))
	for i, item := range items {
		res[i] = dto.StagedImportResponse{
			ID:           item.ID,
			ResourceType: item.ResourceType,
			Source:       item.Source,
			ExternalID:   item.ExternalID,
			Data:         item.Data,
			Metadata:     item.Metadata,
			Status:       item.Status,
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
		}
	}
	return res, nil
}

func (s *InterchangeService) PatchStaged(ctx context.Context, userID, id uuid.UUID, req dto.PatchStagedImportRequest) (*dto.StagedImportResponse, error) {
	patch := entity.StagedImportPatch{
		Metadata: req.Metadata,
		Status:   req.Status,
	}
	item, err := s.repo.PatchStagedImportTx(ctx, nil, userID, id, patch)
	if err != nil {
		return nil, err
	}
	return &dto.StagedImportResponse{
		ID:           item.ID,
		ResourceType: item.ResourceType,
		Source:       item.Source,
		ExternalID:   item.ExternalID,
		Data:         item.Data,
		Metadata:     item.Metadata,
		Status:       item.Status,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}, nil
}

func (s *InterchangeService) DeleteStaged(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.DeleteStagedImportTx(ctx, nil, userID, id)
}

func (s *InterchangeService) ClearStaged(ctx context.Context, userID uuid.UUID, resourceType string) error {
	_, err := s.repo.DeleteAllStagedImportsTx(ctx, nil, userID, resourceType)
	return err
}

func (s *InterchangeService) UpsertRules(ctx context.Context, userID uuid.UUID, resourceType string, rules []dto.MappingRuleInput) ([]dto.ImportMappingRuleResponse, error) {
	upserts := make([]entity.StagedImportRuleUpsert, len(rules))
	for i, r := range rules {
		upserts[i] = entity.StagedImportRuleUpsert{
			ResourceType: resourceType,
			RuleType:     r.RuleType,
			MatchKey:     r.MatchKey,
			MatchValue:   r.MatchValue,
			MappedID:     r.MappedID,
		}
	}
	var resItems []entity.StagedImportRule
	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		var err error
		resItems, err = s.repo.UpsertImportRulesTx(ctx, tx, userID, upserts)
		return err
	})
	if err != nil {
		return nil, err
	}

	res := make([]dto.ImportMappingRuleResponse, len(resItems))
	for i, item := range resItems {
		res[i] = dto.ImportMappingRuleResponse{
			ID:           item.ID,
			ResourceType: item.ResourceType,
			RuleType:     item.RuleType,
			MatchKey:     item.MatchKey,
			MatchValue:   item.MatchValue,
			MappedID:     item.MappedID,
		}
	}
	return res, nil
}

func (s *InterchangeService) ListRules(ctx context.Context, userID uuid.UUID, resourceType string) ([]dto.ImportMappingRuleResponse, error) {
	items, err := s.repo.ListImportRulesTx(ctx, nil, userID, resourceType)
	if err != nil {
		return nil, err
	}
	res := make([]dto.ImportMappingRuleResponse, len(items))
	for i, item := range items {
		res[i] = dto.ImportMappingRuleResponse{
			ID:           item.ID,
			ResourceType: item.ResourceType,
			RuleType:     item.RuleType,
			MatchKey:     item.MatchKey,
			MatchValue:   item.MatchValue,
			MappedID:     item.MappedID,
		}
	}
	return res, nil
}

func (s *InterchangeService) DeleteRule(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.DeleteImportRuleTx(ctx, nil, userID, id)
}

func (s *InterchangeService) CreateManyFromStaged(ctx context.Context, userID uuid.UUID, resourceType string, ids []uuid.UUID) (*dto.BatchImportResult, error) {
	if resourceType != "transaction" {
		return nil, fmt.Errorf("unsupported resource type for promotion: %s", resourceType)
	}

	result := &dto.BatchImportResult{Errors: []string{}}
	for _, id := range ids {
		staged, err := s.repo.GetStagedImportTx(ctx, nil, userID, id)
		if err != nil || staged == nil {
			result.Errors = append(result.Errors, fmt.Sprintf("ID %s: not found", id))
			continue
		}

		req, err := s.mapStagedToTransactionRequest(staged)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("ID %s: %v", id, err))
			continue
		}

		_, err = s.txSvc.Create(ctx, userID, *req)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("ID %s: %v", id, err))
			continue
		}

		// Mark as processed
		status := "processed"
		_, _ = s.repo.PatchStagedImportTx(ctx, nil, userID, id, entity.StagedImportPatch{Status: &status})
		result.Created++
	}

	return result, nil
}

func (s *InterchangeService) ApplyRulesAndCreate(ctx context.Context, userID uuid.UUID, resourceType string) (*dto.BatchImportResult, error) {
	if resourceType != "transaction" {
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	stagedItems, err := s.repo.ListStagedImportsTx(ctx, nil, userID, resourceType)
	if err != nil {
		return nil, err
	}
	rules, err := s.repo.ListImportRulesTx(ctx, nil, userID, resourceType)
	if err != nil {
		return nil, err
	}

	promotableIDs := []uuid.UUID{}
	for _, item := range stagedItems {
		if item.Status != "pending" {
			continue
		}

		// Apply rules to metadata
		updated := false
		for _, rule := range rules {
			matchVal, ok := item.Data[rule.MatchKey].(string)
			if !ok && rule.MatchKey == "" {
				// Default match key is "source_name"
				matchVal, ok = item.Data["source_name"].(string)
			}

			if ok && matchVal == rule.MatchValue {
				if item.Metadata == nil {
					item.Metadata = make(map[string]any)
				}
				key := fmt.Sprintf("mapped_%s_id", rule.RuleType)
				item.Metadata[key] = rule.MappedID.String()
				updated = true
			}
		}

		if updated {
			_, _ = s.repo.PatchStagedImportTx(ctx, nil, userID, item.ID, entity.StagedImportPatch{Metadata: item.Metadata})
		}

		// Check if fully mapped
		if s.canCreateFromMetadata(item.Metadata) {
			promotableIDs = append(promotableIDs, item.ID)
		}
	}

	return s.CreateManyFromStaged(ctx, userID, resourceType, promotableIDs)
}

// Export Logic
func (s *InterchangeService) ExportToCSV(ctx context.Context, userID uuid.UUID, resourceType string, filter any) ([]byte, string, error) {
	var records [][]string
	var filename string
	timestamp := utils.Now().Format("20060102_150405")

	switch resourceType {
	case "transaction":
		f, ok := filter.(entity.TransactionListFilter)
		if !ok {
			f = entity.TransactionListFilter{}
		}
		rows, err := s.txSvc.ListForExport(ctx, userID, f)
		if err != nil {
			return nil, "", err
		}

		filename = fmt.Sprintf("goen_transactions_%s.csv", timestamp)
		records = append(records, []string{"ID", "Date", "Description", "Amount", "Type", "Account", "Category", "Tags", "External Ref"})
		for _, r := range rows {
			desc := ""
			if r.Description != nil {
				desc = *r.Description
			}
			acc := ""
			if r.AccountName != nil {
				acc = *r.AccountName
			}
			cat := ""
			if r.CategoryName != nil {
				cat = *r.CategoryName
			}
			tag := ""
			if r.TagName != nil {
				tag = *r.TagName
			}
			ext := ""
			if r.ExternalRef != nil {
				ext = *r.ExternalRef
			}

			records = append(records, []string{
				r.ID.String(),
				r.OccurredDate,
				desc,
				r.Amount,
				r.Type,
				acc,
				cat,
				tag,
				ext,
			})
		}

	default:
		return nil, "", fmt.Errorf("unsupported resource type for export: %s", resourceType)
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.WriteAll(records); err != nil {
		return nil, "", err
	}

	return buf.Bytes(), filename, nil
}

// Internal mappers
func (s *InterchangeService) mapStagedToTransactionRequest(staged *entity.StagedImport) (*dto.CreateTransactionRequest, error) {
	amount, _ := staged.Data["amount"].(string)
	occurredDate, _ := staged.Data["transaction_date"].(string)
	description, _ := staged.Data["description"].(string)

	// Default to expense if not specified
	txType := entity.TransactionTypeExpense
	if t, ok := staged.Data["type"].(string); ok {
		txType = entity.TransactionType(t)
	}

	req := &dto.CreateTransactionRequest{
		ExternalRef:  staged.ExternalID,
		Type:         txType,
		Amount:       amount,
		OccurredDate: &occurredDate,
		Description:  &description,
	}

	// Map specific fields from Metadata (populated by rules or manual patch)
	if accIDStr, ok := staged.Metadata["mapped_account_id"].(string); ok {
		id, err := uuid.Parse(accIDStr)
		if err == nil {
			req.AccountID = &id
		}
	}
	if catIDStr, ok := staged.Metadata["mapped_category_id"].(string); ok {
		id, err := uuid.Parse(catIDStr)
		if err == nil {
			req.CategoryID = &id
		}
	}

	return req, nil
}

func (s *InterchangeService) canCreateFromMetadata(metadata map[string]any) bool {
	_, hasAcc := metadata["mapped_account_id"]
	return hasAcc
}
