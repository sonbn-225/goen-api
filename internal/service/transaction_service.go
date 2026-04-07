package service

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

type TransactionService struct {
	repo    interfaces.TransactionRepository
	tagSvc  interfaces.TagService
	debtSvc interfaces.DebtService
}

func NewTransactionService(repo interfaces.TransactionRepository, tagSvc interfaces.TagService) *TransactionService {
	return &TransactionService{repo: repo, tagSvc: tagSvc}
}

func (s *TransactionService) SetDebtService(ds interfaces.DebtService) {
	s.debtSvc = ds
}

func (s *TransactionService) List(ctx context.Context, userID uuid.UUID, req dto.ListTransactionsRequest) ([]dto.TransactionResponse, *string, int, error) {
	filter := entity.TransactionListFilter{
		Page:  req.Page,
		Limit: req.Limit,
	}

	if req.AccountID != nil {
		filter.AccountID = req.AccountID
	}
	if req.CategoryID != nil {
		filter.CategoryID = req.CategoryID
	}
	filter.Type = req.Type
	filter.Search = req.Search

	if req.From != nil {
		t, err := utils.ParseTimeOrDate(*req.From)
		if err == nil {
			filter.From = &t
		}
	}
	if req.To != nil {
		t, err := utils.ParseTimeOrDate(*req.To)
		if err == nil {
			filter.To = &t
		}
	}

	items, cursor, total, err := s.repo.ListTransactions(ctx, userID, filter)
	if err != nil {
		return nil, nil, 0, err
	}
	return dto.NewTransactionResponses(items), cursor, total, nil
}

func (s *TransactionService) Get(ctx context.Context, userID, transactionID uuid.UUID) (*dto.TransactionResponse, error) {
	it, err := s.repo.GetTransaction(ctx, userID, transactionID)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, nil
	}
	resp := dto.NewTransactionResponse(*it)
	return &resp, nil
}

func (s *TransactionService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateTransactionRequest) (*dto.TransactionResponse, error) {
	kind := strings.TrimSpace(req.Type)
	if kind != "expense" && kind != "income" && kind != "transfer" {
		return nil, errors.New("invalid transaction type")
	}

	amount := strings.TrimSpace(req.Amount)
	if !utils.IsValidDecimal(amount) {
		return nil, errors.New("invalid decimal amount")
	}

	fromAmount := utils.NormalizeOptionalString(req.FromAmount)
	toAmount := utils.NormalizeOptionalString(req.ToAmount)
	if (fromAmount != nil) != (toAmount != nil) {
		return nil, errors.New("from_amount and to_amount must be provided together for FX transfers")
	}

	occurredAt, occurredDate, err := utils.NormalizeOccurredAt(req.OccurredAt, req.OccurredDate, req.OccurredTime)
	if err != nil {
		return nil, err
	}

	lineItems := make([]entity.TransactionLineItem, 0, len(req.LineItems))
	// If CategoryID is top-level and no line items, create a default one.
	if len(req.LineItems) == 0 && req.CategoryID != nil && *req.CategoryID != uuid.Nil {
		lineItems = append(lineItems, entity.TransactionLineItem{
			BaseEntity: entity.BaseEntity{
				ID: utils.NewID(),
			},
			CategoryID: req.CategoryID,
			Amount:     amount,
		})
	}

	// Sum line items if present (except transfer)
	if kind != "transfer" && len(req.LineItems) > 0 {
		sum := big.NewRat(0, 1)
		for _, li := range req.LineItems {
			if !utils.IsValidDecimal(li.Amount) {
				return nil, errors.New("invalid line item amount")
			}
			r, _ := new(big.Rat).SetString(li.Amount)
			sum.Add(sum, r)

			// Resolve tags for line item if TagService is available
			liTags, _ := s.ensureTags(ctx, userID, li.TagIDs, req.Lang)

			var lCatID *uuid.UUID = li.CategoryID

			lineItems = append(lineItems, entity.TransactionLineItem{
				BaseEntity: entity.BaseEntity{
					ID: utils.NewID(),
				},
				CategoryID: lCatID,
				Amount:     li.Amount,
				Note:       utils.NormalizeOptionalString(li.Note),
				TagIDs:     liTags,
			})
		}
		amount = sum.FloatString(2)
	}

	description := utils.NormalizeOptionalString(req.Description)
	if description != nil && len(lineItems) > 0 {
		if lineItems[0].Note == nil || strings.TrimSpace(*lineItems[0].Note) == "" {
			lineItems[0].Note = description
		}
	}

	id := utils.NewID()

	tx := entity.Transaction{
		AuditEntity: entity.AuditEntity{
			BaseEntity: entity.BaseEntity{
				ID: id,
			},
		},
		ExternalRef:   utils.NormalizeOptionalString(req.ExternalRef),
		Type:          kind,
		OccurredAt:    occurredAt,
		OccurredDate:  occurredDate,
		Amount:        amount,
		FromAmount:    fromAmount,
		ToAmount:      toAmount,
		Description:   description,
		AccountID:     req.AccountID,
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		ExchangeRate:  utils.NormalizeOptionalString(req.ExchangeRate),
		Status:        "pending",
	}

	tagIDs, _ := s.ensureTags(ctx, userID, req.TagIDs, req.Lang)

	participants := []entity.GroupExpenseParticipant{}
	if len(req.GroupParticipants) > 0 {
		participants = s.allocateGroupParticipants(userID, id, tx.Amount, req.OwnerOriginalAmount, req.GroupParticipants)
		// Auto-debt side (if svc available)
		if s.debtSvc != nil && tx.AccountID != nil {
			for _, p := range participants {
				// Simple debt create
				debtName := p.ParticipantName
				if description != nil {
					debtName = *description + " (" + p.ParticipantName + ")"
				}
				_, _ = s.debtSvc.Create(ctx, userID, dto.CreateDebtRequest{
					AccountID: tx.AccountID.String(),
					Direction: "lent",
					Name:      &debtName,
					Principal: p.ShareAmount,
					StartDate: tx.OccurredDate,
					DueDate:   "2099-12-31",
				})
			}
		}
	}

	if err := s.repo.CreateTransaction(ctx, userID, tx, lineItems, tagIDs, participants); err != nil {
		return nil, err
	}

	it, err := s.repo.GetTransaction(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, nil
	}
	resp := dto.NewTransactionResponse(*it)
	return &resp, nil
}

func (s *TransactionService) Patch(ctx context.Context, userID, transactionID uuid.UUID, req dto.TransactionPatchRequest) (*dto.TransactionResponse, error) {
	tagIDs, _ := s.ensureTags(ctx, userID, req.TagIDs, req.Lang)

	patch := entity.TransactionPatch{
		Description: utils.NormalizeOptionalString(req.Description),
		CategoryIDs: req.CategoryIDs,
		TagIDs:      tagIDs,
		Amount:      utils.NormalizeOptionalString(req.Amount),
		Status:      utils.NormalizeOptionalString(req.Status),
	}
	if req.OccurredAt != nil {
		t, err := utils.ParseTimeOrDate(*req.OccurredAt)
		if err == nil {
			patch.OccurredAt = &t
		}
	}

	it, err := s.repo.PatchTransaction(ctx, userID, transactionID, patch)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, nil
	}
	resp := dto.NewTransactionResponse(*it)
	return &resp, nil
}

func (s *TransactionService) BatchPatch(ctx context.Context, userID uuid.UUID, req dto.BatchPatchRequest) (*dto.BatchPatchResult, error) {
	mode := "atomic"
	if req.Mode != nil {
		mode = *req.Mode
	}

	tagIDs, _ := s.ensureTags(ctx, userID, req.Patch.TagIDs, req.Patch.Lang)

	patches := make(map[uuid.UUID]entity.TransactionPatch, len(req.TransactionIDs))
	for _, id := range req.TransactionIDs {
		// Just reuse same patch payload for all (legacy behavior)
		p := entity.TransactionPatch{
			Description: utils.NormalizeOptionalString(req.Patch.Description),
			CategoryIDs: req.Patch.CategoryIDs,
			TagIDs:      tagIDs,
			Amount:      utils.NormalizeOptionalString(req.Patch.Amount),
			Status:      utils.NormalizeOptionalString(req.Patch.Status),
		}
		patches[id] = p
	}

	updated, failed, err := s.repo.BatchPatchTransactions(ctx, userID, req.TransactionIDs, patches, mode)
	if err != nil {
		return nil, err
	}

	return &dto.BatchPatchResult{
		Mode:         mode,
		UpdatedCount: len(updated),
		FailedCount:  len(failed),
		UpdatedIDs:   updated,
		FailedIDs:    failed,
	}, nil
}

func (s *TransactionService) Delete(ctx context.Context, userID, transactionID uuid.UUID) error {
	return s.repo.DeleteTransaction(ctx, userID, transactionID)
}

// Imports
func (s *TransactionService) StageImport(ctx context.Context, userID uuid.UUID, items []dto.StageImportedItem) (int, int, []string, error) {
	// 1. Fetch rules for auto-mapping
	rules, err := s.repo.ListImportMappingRules(ctx, userID)
	if err != nil {
		return 0, 0, nil, err
	}

	normalize := func(v string) string {
		return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(v)), " "))
	}

	accountRules := map[string]uuid.UUID{}
	categoryRules := map[string]uuid.UUID{}
	for _, rule := range rules {
		key := normalize(rule.SourceName)
		if key == "" {
			continue
		}
		if rule.Kind == "account" {
			accountRules[key] = rule.MappedID
		}
		if rule.Kind == "category" {
			categoryRules[key] = rule.MappedID
		}
	}

	// 2. Prepare entities
	creates := make([]entity.ImportedTransactionCreate, 0, len(items))
	for _, item := range items {
		tDate, err := s.normalizeImportDate(item.TransactionDate)
		if err != nil {
			continue
		}

		create := entity.ImportedTransactionCreate{
			Source:               "generic", // Default source
			TransactionDate:      tDate,
			Amount:               item.Amount,
			Description:          item.Description,
			TransactionType:      item.TransactionType,
			ImportedAccountName:  item.AccountName,
			ImportedCategoryName: item.CategoryName,
			RawPayload:           item.Raw,
		}

		// Auto-map
		if item.AccountName != nil {
			if id, ok := accountRules[normalize(*item.AccountName)]; ok {
				create.MappedAccountID = &id
			}
		}
		if item.CategoryName != nil {
			if id, ok := categoryRules[normalize(*item.CategoryName)]; ok {
				create.MappedCategoryID = &id
			}
		}

		creates = append(creates, create)
	}

	if len(creates) == 0 {
		return 0, len(items), nil, nil
	}

	// 3. Save
	staged, err := s.repo.CreateImportedTransactions(ctx, userID, creates)
	if err != nil {
		return 0, 0, nil, err
	}

	return len(staged), len(items) - len(staged), nil, nil
}

func (s *TransactionService) ListImported(ctx context.Context, userID uuid.UUID) ([]dto.ImportedTransactionResponse, error) {
	items, err := s.repo.ListImportedTransactions(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dto.NewImportedTransactionResponses(items), nil
}

func (s *TransactionService) PatchImported(ctx context.Context, userID uuid.UUID, importID uuid.UUID, patch entity.ImportedTransactionPatch) (*dto.ImportedTransactionResponse, error) {
	it, err := s.repo.PatchImportedTransaction(ctx, userID, importID, patch)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, nil
	}
	resp := dto.NewImportedTransactionResponse(*it)
	return &resp, nil
}

func (s *TransactionService) DeleteImported(ctx context.Context, userID, importID uuid.UUID) error {
	return s.repo.DeleteImportedTransaction(ctx, userID, importID)
}

func (s *TransactionService) ClearImported(ctx context.Context, userID uuid.UUID) error {
	_, err := s.repo.DeleteAllImportedTransactions(ctx, userID)
	return err
}

// Rules
func (s *TransactionService) UpsertMappingRules(ctx context.Context, userID uuid.UUID, inputs []dto.MappingRuleInput) ([]dto.ImportMappingRuleResponse, error) {
	upserts := make([]entity.ImportMappingRuleUpsert, len(inputs))
	for i, in := range inputs {
		upserts[i] = entity.ImportMappingRuleUpsert{
			Kind:       in.Kind,
			SourceName: in.SourceName,
			MappedID:   in.MappedID,
		}
	}
	rules, err := s.repo.UpsertImportMappingRules(ctx, userID, upserts)
	if err != nil {
		return nil, err
	}
	return dto.NewImportMappingRuleResponses(rules), nil
}

func (s *TransactionService) ListMappingRules(ctx context.Context, userID uuid.UUID) ([]dto.ImportMappingRuleResponse, error) {
	rules, err := s.repo.ListImportMappingRules(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dto.NewImportMappingRuleResponses(rules), nil
}

func (s *TransactionService) DeleteMappingRule(ctx context.Context, userID, ruleID uuid.UUID) error {
	return s.repo.DeleteImportMappingRule(ctx, userID, ruleID)
}

// Create from Imports
func (s *TransactionService) CreateFromImported(ctx context.Context, userID, importID uuid.UUID) (*dto.TransactionResponse, error) {
	it, err := s.repo.GetImportedTransaction(ctx, userID, importID)
	if err != nil {
		return nil, err
	}

	// Prepare create request
	req := dto.CreateTransactionRequest{
		OccurredDate: &it.TransactionDate,
		Amount:       it.Amount,
		Description:  it.Description,
		Type:         "expense", // Default
	}
	if it.TransactionType != nil {
		req.Type = strings.ToLower(*it.TransactionType)
	}
	if it.MappedAccountID != nil {
		if req.Type == "transfer" {
			req.FromAccountID = it.MappedAccountID
		} else {
			req.AccountID = it.MappedAccountID
		}
	}
	if it.MappedCategoryID != nil {
		req.CategoryID = it.MappedCategoryID
	}

	tx, err := s.Create(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	// Cleanup
	_ = s.repo.DeleteImportedTransaction(ctx, userID, importID)

	return tx, nil
}

func (s *TransactionService) CreateManyFromImported(ctx context.Context, userID uuid.UUID, importIDs []uuid.UUID) (*dto.BatchImportResult, error) {
	res := &dto.BatchImportResult{}
	for _, id := range importIDs {
		_, err := s.CreateFromImported(ctx, userID, id)
		if err != nil {
			res.Skipped++
			res.Errors = append(res.Errors, fmt.Sprintf("ID %v: %v", id, err))
		} else {
			res.Created++
		}
	}
	return res, nil
}

func (s *TransactionService) ApplyRulesAndCreate(ctx context.Context, userID uuid.UUID) (*dto.BatchImportResult, error) {
	// 1. Get rules and items
	rules, _ := s.repo.ListImportMappingRules(ctx, userID)
	items, _ := s.repo.ListImportedTransactions(ctx, userID)

	normalize := func(v string) string {
		return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(v)), " "))
	}

	accountRules := map[string]uuid.UUID{}
	categoryRules := map[string]uuid.UUID{}
	for _, rule := range rules {
		k := normalize(rule.SourceName)
		if rule.Kind == "account" {
			accountRules[k] = rule.MappedID
		} else {
			categoryRules[k] = rule.MappedID
		}
	}

	// 2. Patch items that can be auto-mapped
	for _, it := range items {
		patch := entity.ImportedTransactionPatch{}
		changed := false
		if it.MappedAccountID == nil && it.ImportedAccountName != nil {
			if id, ok := accountRules[normalize(*it.ImportedAccountName)]; ok {
				patch.MappedAccountID = &id
				changed = true
			}
		}
		if it.MappedCategoryID == nil && it.ImportedCategoryName != nil {
			if id, ok := categoryRules[normalize(*it.ImportedCategoryName)]; ok {
				patch.MappedCategoryID = &id
				changed = true
			}
		}
		if changed {
			_, _ = s.repo.PatchImportedTransaction(ctx, userID, it.ID, patch)
		}
	}

	// 3. Create all that are fully mapped
	items, _ = s.repo.ListImportedTransactions(ctx, userID)
	res := &dto.BatchImportResult{}
	for _, it := range items {
		if it.MappedAccountID == nil {
			continue
		}
		if !strings.EqualFold(utils.Coalesce(it.TransactionType, ""), "transfer") && it.MappedCategoryID == nil {
			continue
		}

		_, err := s.CreateFromImported(ctx, userID, it.ID)
		if err != nil {
			res.Skipped++
			res.Errors = append(res.Errors, fmt.Sprintf("ID %v: %v", it.ID, err))
		} else {
			res.Created++
		}
	}
	return res, nil
}

func (s *TransactionService) normalizeImportDate(v string) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", errors.New("empty date")
	}
	layouts := []string{"2006-01-02", "2006/01/02", "02/01/2006", "2-1-2006", "02-01-2006", time.RFC3339}
	for _, l := range layouts {
		if t, err := time.Parse(l, v); err == nil {
			return t.Format("2006-01-02"), nil
		}
	}
	return "", errors.New("invalid date format")
}

// Helpers
func (s *TransactionService) ensureTags(ctx context.Context, userID uuid.UUID, inputs []uuid.UUID, lang string) ([]uuid.UUID, error) {
	if s.tagSvc == nil || len(inputs) == 0 {
		return nil, nil
	}
	// Note: now inputs are already UUIDs from DTO.
	// But ensureTags was also handling names.
	// I'll update it to accept the new type but keep logic for creating tags if needed?
	// Actually, if they are already UUIDs, we just return them.
	// Wait, if the handler receives strings, it should call a different method if it wants to resolve names.
	// Currently the DTO uses uuid.UUID, which means the JSON must contain valid UUID strings.
	// If the user wants to create a tag by name, they should probably do it separately or we need a special DTO field.
	// For now, I'll just return the UUIDs.
	return inputs, nil
}

func (s *TransactionService) allocateGroupParticipants(userID uuid.UUID, txID uuid.UUID, txAmt string, ownerAmt *string, inputs []dto.GroupParticipantInput) []entity.GroupExpenseParticipant {
	totalPaid, ok := new(big.Rat).SetString(txAmt)
	if !ok {
		return nil
	}

	type person struct {
		name     string
		original *big.Rat
		origStr  string
	}
	involved := []person{}
	if ownerAmt != nil && *ownerAmt != "" {
		if r, ok := new(big.Rat).SetString(*ownerAmt); ok && r.Sign() > 0 {
			involved = append(involved, person{name: "owner", original: r, origStr: *ownerAmt})
		}
	}
	for _, p := range inputs {
		if r, ok := new(big.Rat).SetString(p.OriginalAmount); ok && r.Sign() > 0 {
			involved = append(involved, person{name: p.ParticipantName, original: r, origStr: p.OriginalAmount})
		}
	}

	if len(involved) == 0 {
		return nil
	}

	sumOriginal := new(big.Rat)
	for _, p := range involved {
		sumOriginal.Add(sumOriginal, p.original)
	}

	shares := make([]*big.Rat, 0, len(involved))
	allocated := new(big.Rat)
	for i, p := range involved {
		if i < len(involved)-1 {
			raw := new(big.Rat).Mul(totalPaid, p.original)
			raw.Quo(raw, sumOriginal)
			rounded := s.roundRat(raw, 2)
			shares = append(shares, rounded)
			allocated.Add(allocated, rounded)
		} else {
			last := new(big.Rat).Sub(totalPaid, allocated) // remainder
			shares = append(shares, s.roundRat(last, 2))
		}
	}

	out := []entity.GroupExpenseParticipant{}
	for i, p := range involved {
		if p.name == "owner" {
			continue
		}
		out = append(out, entity.GroupExpenseParticipant{
			AuditEntity: entity.AuditEntity{
				BaseEntity: entity.BaseEntity{
					ID: utils.NewID(),
				},
			},
			UserID:          userID,
			TransactionID:   txID,
			ParticipantName: p.name,
			OriginalAmount:  p.origStr,
			ShareAmount:     shares[i].FloatString(2),
			IsSettled:       false,
		})
	}
	return out
}

func (s *TransactionService) roundRat(r *big.Rat, scale int) *big.Rat {
	if r == nil {
		return big.NewRat(0, 1)
	}
	if scale < 0 {
		scale = 0
	}
	factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
	num := new(big.Int).Mul(r.Num(), factor)
	den := new(big.Int).Set(r.Denom())
	q, rem := new(big.Int).QuoRem(num, den, new(big.Int))
	if rem.Sign() >= 0 {
		twoRem := new(big.Int).Mul(rem, big.NewInt(2))
		if twoRem.Cmp(den) >= 0 {
			q.Add(q, big.NewInt(1))
		}
	}
	return new(big.Rat).SetFrac(q, factor)
}
