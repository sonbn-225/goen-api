package service

import (
	"context"
	"errors"
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
	repo        interfaces.TransactionRepository
	tagSvc      interfaces.TagService
	debtSvc     interfaces.DebtService
}

func NewTransactionService(repo interfaces.TransactionRepository, tagSvc interfaces.TagService) *TransactionService {
	return &TransactionService{repo: repo, tagSvc: tagSvc}
}

func (s *TransactionService) SetDebtService(ds interfaces.DebtService) {
	s.debtSvc = ds
}

func (s *TransactionService) List(ctx context.Context, userID string, req dto.CreateTransactionRequest) ([]dto.TransactionResponse, *string, int, error) {
	filter := entity.TransactionListFilter{
		// Map from req fields (to be refined)
	}
	items, cursor, total, err := s.repo.ListTransactions(ctx, userID, filter)
	if err != nil {
		return nil, nil, 0, err
	}
	return dto.NewTransactionResponses(items), cursor, total, nil
}

func (s *TransactionService) Get(ctx context.Context, userID, transactionID string) (*dto.TransactionResponse, error) {
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

func (s *TransactionService) Create(ctx context.Context, userID string, req dto.CreateTransactionRequest) (*dto.TransactionResponse, error) {
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
	if len(req.LineItems) == 0 && req.CategoryID != nil && strings.TrimSpace(*req.CategoryID) != "" {
		catID := strings.TrimSpace(*req.CategoryID)
		lineItems = append(lineItems, entity.TransactionLineItem{
			ID:         uuid.NewString(),
			CategoryID: &catID,
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

			lineItems = append(lineItems, entity.TransactionLineItem{
				ID:         uuid.NewString(),
				CategoryID: utils.NormalizeOptionalString(li.CategoryID),
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

	now := time.Now().UTC()
	id := uuid.NewString()

	tx := entity.Transaction{
		ID:            id,
		ClientID:      utils.NormalizeOptionalString(req.ClientID),
		ExternalRef:   utils.NormalizeOptionalString(req.ExternalRef),
		Type:          kind,
		OccurredAt:    occurredAt,
		OccurredDate:  occurredDate,
		Amount:        amount,
		FromAmount:    fromAmount,
		ToAmount:      toAmount,
		Description:   nil,
		AccountID:     utils.NormalizeOptionalString(req.AccountID),
		FromAccountID: utils.NormalizeOptionalString(req.FromAccountID),
		ToAccountID:   utils.NormalizeOptionalString(req.ToAccountID),
		ExchangeRate:  utils.NormalizeOptionalString(req.ExchangeRate),
		Status:        "pending",
		CreatedAt:     now,
		UpdatedAt:     now,
		CreatedBy:     &userID,
		UpdatedBy:     &userID,
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
					AccountID: *tx.AccountID,
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

func (s *TransactionService) Patch(ctx context.Context, userID, transactionID string, req dto.TransactionPatchRequest) (*dto.TransactionResponse, error) {
	// Simplified patch for now (similar to repo logic)
	patch := entity.TransactionPatch{
		Description: utils.NormalizeOptionalString(req.Description),
		CategoryIDs: req.CategoryIDs,
		TagIDs:      req.TagIDs,
		Amount:      utils.NormalizeOptionalString(req.Amount),
		Status:      utils.NormalizeOptionalString(req.Status),
	}
	if req.OccurredAt != nil {
		t, err := utils.ParseTimeOrDate(*req.OccurredAt)
		if err == nil {
			patch.OccurredAt = &t
		}
	}
	// Note: LineItems and Participants replace-all could be added easily.

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

func (s *TransactionService) BatchPatch(ctx context.Context, userID string, req dto.BatchPatchRequest) (*dto.BatchPatchResult, error) {
	mode := "atomic"
	if req.Mode != nil {
		mode = *req.Mode
	}

	patches := make(map[string]entity.TransactionPatch, len(req.TransactionIDs))
	for _, id := range req.TransactionIDs {
		// Just reuse same patch payload for all (legacy behavior)
		p := entity.TransactionPatch{
			Description: utils.NormalizeOptionalString(req.Patch.Description),
			CategoryIDs: req.Patch.CategoryIDs,
			TagIDs:      req.Patch.TagIDs,
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

func (s *TransactionService) Delete(ctx context.Context, userID, transactionID string) error {
	return s.repo.DeleteTransaction(ctx, userID, transactionID)
}

// Stubs for Imports
func (s *TransactionService) StageImport(ctx context.Context, userID string, items []dto.StageImportedItem) (int, int, []string, error) {
	return 0, 0, nil, nil
}
func (s *TransactionService) ListImported(ctx context.Context, userID string) ([]dto.ImportedTransactionResponse, error) {
	return nil, nil
}
func (s *TransactionService) PatchImported(ctx context.Context, userID, importID string, patch entity.ImportedTransactionPatch) (*dto.ImportedTransactionResponse, error) {
	return nil, nil
}
func (s *TransactionService) DeleteImported(ctx context.Context, userID, importID string) error {
	return nil
}
func (s *TransactionService) ClearImported(ctx context.Context, userID string) error {
	return nil
}
func (s *TransactionService) UpsertMappingRules(ctx context.Context, userID string, inputs []dto.MappingRuleInput) ([]dto.ImportMappingRuleResponse, error) {
	return nil, nil
}
func (s *TransactionService) ListMappingRules(ctx context.Context, userID string) ([]dto.ImportMappingRuleResponse, error) {
	return nil, nil
}
func (s *TransactionService) DeleteMappingRule(ctx context.Context, userID, ruleID string) error {
	return nil
}

// Helpers
func (s *TransactionService) ensureTags(ctx context.Context, userID string, inputs []string, lang string) ([]string, error) {
	if s.tagSvc == nil || len(inputs) == 0 { return inputs, nil }
	if lang == "" { lang = "en" }
	out := make([]string, 0, len(inputs))
	for _, input := range inputs {
		trimmed := strings.TrimSpace(input)
		if trimmed == "" { continue }
		if _, err := uuid.Parse(trimmed); err == nil {
			out = append(out, trimmed)
			continue
		}
		id, err := s.tagSvc.GetOrCreateByName(ctx, userID, trimmed, lang)
		if err == nil { out = append(out, id) } else { out = append(out, trimmed) }
	}
	return out, nil
}

func (s *TransactionService) allocateGroupParticipants(userID, txID, txAmt string, ownerAmt *string, inputs []dto.GroupParticipantInput) []entity.GroupExpenseParticipant {
	totalPaid, ok := new(big.Rat).SetString(txAmt)
	if !ok { return nil }

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

	if len(involved) == 0 { return nil }

	sumOriginal := new(big.Rat)
	for _, p := range involved { sumOriginal.Add(sumOriginal, p.original) }

	shares := make([]*big.Rat, 0, len(involved))
	allocated := new(big.Rat)
	for i, p := range involved {
		if i < len(involved)-1 {
			raw := new(big.Rat).Mul(totalPaid, p.original)
			raw.Quo(raw, sumOriginal)
			rounded := roundRat(raw, 2)
			shares = append(shares, rounded)
			allocated.Add(allocated, rounded)
		} else {
			last := new(big.Rat).Sub(totalPaid, allocated) // remainder
			shares = append(shares, roundRat(last, 2))
		}
	}

	now := time.Now().UTC()
	out := []entity.GroupExpenseParticipant{}
	for i, p := range involved {
		if p.name == "owner" { continue }
		out = append(out, entity.GroupExpenseParticipant{
			ID: uuid.NewString(), UserID: userID, TransactionID: txID,
			ParticipantName: p.name, OriginalAmount: p.origStr,
			ShareAmount: shares[i].FloatString(2),
			IsSettled: false, CreatedAt: now, UpdatedAt: now,
		})
	}
	return out
}

func roundRat(r *big.Rat, scale int) *big.Rat {
	// Simplified rounding: truncation of extra digits + maybe adding something?
	// Real financial rounding is more complex. Legacy used something similar.
	fStr := r.FloatString(scale)
	rounded, _ := new(big.Rat).SetString(fStr)
	return rounded
}
