package transaction

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/modules/debt"
)

// CreateLineItemRequest contains line item create parameters.
type CreateLineItemRequest struct {
	CategoryID *string  `json:"category_id,omitempty"`
	TagIDs     []string `json:"tag_ids,omitempty"`
	Amount     string   `json:"amount"`
	Note       *string  `json:"note,omitempty"`
}

// CreateRequest contains transaction create parameters.
type CreateRequest struct {
	ClientID      *string                 `json:"client_id,omitempty"`
	ExternalRef   *string                 `json:"external_ref,omitempty"`
	Type          string                  `json:"type"`
	OccurredAt    *string                 `json:"occurred_at,omitempty"`
	OccurredDate  *string                 `json:"occurred_date,omitempty"`
	OccurredTime  *string                 `json:"occurred_time,omitempty"`
	Amount        string                  `json:"amount"`
	FromAmount    *string                 `json:"from_amount,omitempty"`
	ToAmount      *string                 `json:"to_amount,omitempty"`
	Description   *string                 `json:"description,omitempty"`
	AccountID     *string                 `json:"account_id,omitempty"`
	FromAccountID *string                 `json:"from_account_id,omitempty"`
	ToAccountID   *string                 `json:"to_account_id,omitempty"`
	ExchangeRate  *string                 `json:"exchange_rate,omitempty"`
	CategoryID    *string                 `json:"category_id,omitempty"`
	TagIDs              []string                `json:"tag_ids,omitempty"`
	LineItems           []CreateLineItemRequest `json:"line_items,omitempty"`
	GroupParticipants   []GroupParticipantInput `json:"group_participants,omitempty"`
	OwnerOriginalAmount *string                 `json:"owner_original_amount,omitempty"`
	Lang                string                  `json:"lang,omitempty"`
}

// ListRequest contains transaction list filters.
type ListRequest struct {
	AccountID         *string
	CategoryID        *string
	Type              *string
	Search            *string
	ExternalRefFamily *string
	From              *string
	To                *string
	Cursor            *string
	Page              int
	Limit             int
}

// PatchRequest contains transaction patch parameters.
type PatchRequest struct {
	Description       *string                  `json:"description,omitempty"`
	CategoryIDs       []string                 `json:"category_ids,omitempty"`
	TagIDs            []string                 `json:"tag_ids,omitempty"`
	Amount            *string                  `json:"amount,omitempty"`
	Status            *string                  `json:"status,omitempty"`
	OccurredAt        *string                  `json:"occurred_at,omitempty"`
	LineItems         *[]LineItemInput         `json:"line_items,omitempty"`
	GroupParticipants *[]GroupParticipantInput `json:"group_participants,omitempty"`
	Lang              string                   `json:"lang,omitempty"`
}

type BatchPatchRequest struct {
	TransactionIDs []string     `json:"transaction_ids"`
	Patch          PatchRequest `json:"patch"`
	Mode           *string      `json:"mode,omitempty"`
}

type BatchPatchResult struct {
	Mode         string   `json:"mode"`
	UpdatedCount int      `json:"updated_count"`
	FailedCount  int      `json:"failed_count"`
	UpdatedIDs   []string `json:"updated_ids,omitempty"`
	FailedIDs    []string `json:"failed_ids,omitempty"`
}

// LineItemInput is the line item payload for patch.
type LineItemInput struct {
	CategoryID *string  `json:"category_id,omitempty"`
	TagIDs     []string `json:"tag_ids,omitempty"`
	Amount     string   `json:"amount"`
	Note       *string  `json:"note,omitempty"`
}

// GroupParticipantInput is the group participant payload for patch.
type GroupParticipantInput struct {
	ParticipantName string `json:"participant_name"`
	OriginalAmount  string `json:"original_amount"`
	ShareAmount     string `json:"share_amount"`
}

type Service struct {
	repo         domain.TransactionRepository
	categoryRepo domain.CategoryRepository
	accountRepo  domain.AccountRepository
	tagService   TagService
	debtService  DebtService
}

func (s *Service) SetDebtService(ds DebtService) {
	s.debtService = ds
}

type DebtService interface {
	Create(ctx context.Context, userID string, req debt.CreateRequest) (*domain.Debt, error)
}

type TagService interface {
	GetOrCreateByName(ctx context.Context, userID, name, langHint string) (string, error)
}

// NewService creates a new transaction service.
func NewService(repo domain.TransactionRepository, categoryRepo domain.CategoryRepository, accountRepo domain.AccountRepository, tagService TagService) *Service {
	return &Service{repo: repo, categoryRepo: categoryRepo, accountRepo: accountRepo, tagService: tagService}
}

type ImportGoenV1Result struct {
	Imported int
	Skipped  int
	Errors   []string
}

type ImportedGoenV1StageItem struct {
	TransactionDate string         `json:"transaction_date"`
	Amount          string         `json:"amount"`
	Description     *string        `json:"description,omitempty"`
	TransactionType *string        `json:"transaction_type,omitempty"`
	AccountName     *string        `json:"account_name,omitempty"`
	CategoryName    *string        `json:"category,omitempty"`
	Raw             map[string]any `json:"raw,omitempty"`
}

type CreateImportedGoenV1BatchResult struct {
	Created int      `json:"created"`
	Skipped int      `json:"skipped"`
	Errors  []string `json:"errors,omitempty"`
}

// Generic import types (source-agnostic)
type StageImportedItem struct {
	TransactionDate string         `json:"transaction_date"`
	Amount          string         `json:"amount"`
	Description     *string        `json:"description,omitempty"`
	TransactionType *string        `json:"transaction_type,omitempty"`
	AccountName     *string        `json:"account_name,omitempty"`
	CategoryName    *string        `json:"category_name,omitempty"`
	Category        *string        `json:"category,omitempty"`
	Raw             map[string]any `json:"raw,omitempty"`
}

type StagedImportResult struct {
	Created int      `json:"created"`
	Skipped int      `json:"skipped"`
	Errors  []string `json:"errors,omitempty"`
}

type ExportTransactionsFilter struct {
	AccountID *string
	From      *time.Time
	To        *time.Time
}

type MappingRuleInput struct {
	Kind       string `json:"kind"`
	SourceName string `json:"source_name"`
	MappedID   string `json:"mapped_id"`
}

type ApplyImportRulesResult struct {
	UpdatedMappings int      `json:"updated_mappings"`
	Created         int      `json:"created"`
	Remaining       int      `json:"remaining"`
	Errors          []string `json:"errors,omitempty"`
}

// Create creates a new transaction.
func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*domain.Transaction, error) {
	kind := strings.TrimSpace(req.Type)
	if kind != "expense" && kind != "income" && kind != "transfer" {
		return nil, apperrors.Validation("type is invalid", nil)
	}

	amount := strings.TrimSpace(req.Amount)
	if amount == "" {
		return nil, apperrors.Validation("amount is required", nil)
	}
	if !isValidDecimal(amount) {
		return nil, apperrors.Validation("amount must be a decimal string", nil)
	}

	fromAmount := normalizeOptionalString(req.FromAmount)
	toAmount := normalizeOptionalString(req.ToAmount)
	if fromAmount != nil {
		v := strings.TrimSpace(*fromAmount)
		if v == "" {
			fromAmount = nil
		} else {
			if !isValidDecimal(v) {
				return nil, apperrors.Validation("from_amount must be a decimal string", nil)
			}
			fromAmount = &v
		}
	}
	if toAmount != nil {
		v := strings.TrimSpace(*toAmount)
		if v == "" {
			toAmount = nil
		} else {
			if !isValidDecimal(v) {
				return nil, apperrors.Validation("to_amount must be a decimal string", nil)
			}
			toAmount = &v
		}
	}
	if (fromAmount != nil) != (toAmount != nil) {
		return nil, apperrors.Validation("from_amount and to_amount must be provided together", nil)
	}

	occurredAt, occurredDate, err := normalizeOccurredAt(req.OccurredAt, req.OccurredDate, req.OccurredTime)
	if err != nil {
		return nil, err
	}

	lineItems := make([]domain.TransactionLineItem, 0, len(req.LineItems))

	if kind == "transfer" {
		if len(req.LineItems) > 0 {
			return nil, apperrors.Validation("line_items must be empty for transfer", nil)
		}
		if req.CategoryID != nil && strings.TrimSpace(*req.CategoryID) != "" {
			return nil, apperrors.Validation("category_id must be empty for transfer", nil)
		}
	}

	if kind != "transfer" && len(req.LineItems) == 0 && (req.CategoryID == nil || strings.TrimSpace(*req.CategoryID) == "") {
		return nil, apperrors.Validation("line_items is required and must include at least one category", nil)
	}

	// If CategoryID is provided at top-level and no lineItems, create a default lineItem.
	if len(req.LineItems) == 0 && req.CategoryID != nil && strings.TrimSpace(*req.CategoryID) != "" {
		catID := strings.TrimSpace(*req.CategoryID)
		lineItems = append(lineItems, domain.TransactionLineItem{
			ID:         uuid.NewString(),
			CategoryID: &catID,
			Amount:     amount,
		})
	}

	if len(req.LineItems) > 0 {
		sum := big.NewRat(0, 1)
		for _, li := range req.LineItems {
			if kind != "transfer" {
				if li.CategoryID == nil || strings.TrimSpace(*li.CategoryID) == "" {
					return nil, apperrors.Validation("line_items.category_id is required", nil)
				}
			}

			liAmt := strings.TrimSpace(li.Amount)
			if liAmt == "" {
				return nil, apperrors.Validation("line_items.amount is required", nil)
			}
			if !isValidDecimal(liAmt) {
				return nil, apperrors.Validation("line_items.amount must be a decimal string", nil)
			}
			r, ok := new(big.Rat).SetString(liAmt)
			if !ok {
				return nil, apperrors.Validation("line_items.amount must be a decimal string", nil)
			}
			sum.Add(sum, r)

			lineItems = append(lineItems, domain.TransactionLineItem{
				ID:         uuid.NewString(),
				CategoryID: normalizeOptionalString(li.CategoryID),
				Amount:     liAmt,
				Note:       normalizeOptionalString(li.Note),
			})
		}

		if kind != "transfer" {
			amount = sum.FloatString(2)
		}
	}

	if kind != "transfer" && len(lineItems) == 0 {
		return nil, apperrors.Validation("line_items is required and must include at least one category", nil)
	}

	description := normalizeOptionalString(req.Description)
	if description != nil {
		if len(lineItems) > 0 {
			if lineItems[0].Note == nil || strings.TrimSpace(*lineItems[0].Note) == "" {
				lineItems[0].Note = description
			}
		} else if kind == "transfer" {
			lineItems = append(lineItems, domain.TransactionLineItem{
				ID:         uuid.NewString(),
				CategoryID: nil,
				Amount:     amount,
				Note:       description,
			})
		}
	}

	now := time.Now().UTC()
	id := uuid.NewString()

	tx := domain.Transaction{
		ID:            id,
		ClientID:      normalizeOptionalString(req.ClientID),
		ExternalRef:   normalizeOptionalString(req.ExternalRef),
		Type:          kind,
		OccurredAt:    occurredAt,
		OccurredDate:  occurredDate,
		Amount:        amount,
		FromAmount:    fromAmount,
		ToAmount:      toAmount,
		Description:   nil,
		AccountID:     normalizeOptionalString(req.AccountID),
		FromAccountID: normalizeOptionalString(req.FromAccountID),
		ToAccountID:   normalizeOptionalString(req.ToAccountID),
		ExchangeRate:  normalizeOptionalString(req.ExchangeRate),
		Status:        "pending",
		CreatedAt:     now,
		UpdatedAt:     now,
		CreatedBy:     &userID,
		UpdatedBy:     &userID,
	}

	if err := validateTransactionLinkage(tx); err != nil {
		return nil, err
	}

	tagIDs, err := s.ensureTags(ctx, userID, req.TagIDs, req.Lang)
	if err != nil {
		return nil, err
	}

	for i := range lineItems {
		var reqLi *CreateLineItemRequest
		if i < len(req.LineItems) {
			reqLi = &req.LineItems[i]
		} else if i == 0 && req.CategoryID != nil {
			// case where CategoryID was top-level
		}

		if reqLi != nil {
			liTags, err := s.ensureTags(ctx, userID, reqLi.TagIDs, req.Lang)
			if err != nil {
				return nil, err
			}
			lineItems[i].TagIDs = liTags
		}
	}

	participants := []domain.GroupExpenseParticipant{}
	if len(req.GroupParticipants) > 0 {
		totalPaid, ok := new(big.Rat).SetString(tx.Amount)
		if ok {
			type person struct {
				name        string
				original    *big.Rat
				originalStr string
			}
			involved := []person{}
			if req.OwnerOriginalAmount != nil && *req.OwnerOriginalAmount != "" {
				r, ok := new(big.Rat).SetString(*req.OwnerOriginalAmount)
				if ok && r.Sign() > 0 {
					involved = append(involved, person{name: "owner", original: r, originalStr: *req.OwnerOriginalAmount})
				}
			}
			for _, p := range req.GroupParticipants {
				r, ok := new(big.Rat).SetString(p.OriginalAmount)
				if ok && r.Sign() > 0 {
					involved = append(involved, person{name: p.ParticipantName, original: r, originalStr: p.OriginalAmount})
				}
			}

			if len(involved) > 0 {
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
						rounded := roundRat(raw, 2)
						shares = append(shares, rounded)
						allocated.Add(allocated, rounded)
					} else {
						last := new(big.Rat).Sub(totalPaid, allocated)
						shares = append(shares, roundRat(last, 2))
					}
				}
				now := time.Now().UTC()
				for i, p := range involved {
					if p.name == "owner" {
						continue
					}
					share := shares[i]
					part := domain.GroupExpenseParticipant{
						ID:              uuid.NewString(),
						UserID:          userID,
						TransactionID:   id,
						ParticipantName: p.name,
						OriginalAmount:  p.originalStr,
						ShareAmount:     formatRatDecimalScale(share, 2),
						IsSettled:       false,
						CreatedAt:       now,
						UpdatedAt:       now,
					}
					participants = append(participants, part)

					// Create debt
					if s.debtService != nil && tx.AccountID != nil {
						debtName := p.name
						if tx.Description != nil && *tx.Description != "" {
							debtName = *tx.Description + " (" + p.name + ")"
						}
						_, _ = s.debtService.Create(ctx, userID, debt.CreateRequest{
							AccountID:    *tx.AccountID,
							Direction:    "lent",
							Name:         &debtName,
							Principal:    part.ShareAmount,
							StartDate:    tx.OccurredDate,
							DueDate:      "2099-12-31",
							Status:       pointer("active"),
							InterestRate: pointer("0"),
						})
					}
				}
			}
		}
	}

	if err := s.repo.CreateTransaction(ctx, userID, tx, lineItems, tagIDs, participants); err != nil {
		if errors.Is(err, apperrors.ErrTransactionForbidden) {
			return nil, apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		return nil, err
	}

	created, err := s.repo.GetTransaction(ctx, userID, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrTransactionNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "transaction not found", err)
		}
		if errors.Is(err, apperrors.ErrTransactionForbidden) {
			return nil, apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		return nil, err
	}
	return created, nil
}

// Get retrieves a transaction by ID.
func (s *Service) Get(ctx context.Context, userID, transactionID string) (*domain.Transaction, error) {
	tx, err := s.repo.GetTransaction(ctx, userID, transactionID)
	if err != nil {
		if errors.Is(err, apperrors.ErrTransactionNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "transaction not found", err)
		}
		if errors.Is(err, apperrors.ErrTransactionForbidden) {
			return nil, apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		return nil, err
	}
	return tx, nil
}

// List returns transactions matching the filter.
func (s *Service) List(ctx context.Context, userID string, req ListRequest) ([]domain.Transaction, *string, int, error) {
	filter := domain.TransactionListFilter{
		AccountID:         normalizeOptionalString(req.AccountID),
		CategoryID:        normalizeOptionalString(req.CategoryID),
		Type:              normalizeOptionalString(req.Type),
		Search:            normalizeOptionalString(req.Search),
		ExternalRefFamily: normalizeOptionalString(req.ExternalRefFamily),
		Cursor:            normalizeOptionalString(req.Cursor),
		Page:              req.Page,
		Limit:             req.Limit,
	}

	if req.From != nil {
		v := strings.TrimSpace(*req.From)
		if v != "" {
			t, err := parseTimeOrDate(v)
			if err != nil {
				return nil, nil, 0, apperrors.Validation("from is invalid", nil)
			}
			filter.From = &t
		}
	}
	if req.To != nil {
		v := strings.TrimSpace(*req.To)
		if v != "" {
			t, err := parseTimeOrDate(v)
			if err != nil {
				return nil, nil, 0, apperrors.Validation("to is invalid", nil)
			}
			filter.To = &t
		}
	}

	return s.repo.ListTransactions(ctx, userID, filter)
}

// Patch updates transaction fields.
func (s *Service) Patch(ctx context.Context, userID, transactionID string, req PatchRequest) (*domain.Transaction, error) {
	patch, err := s.buildTransactionPatch(ctx, userID, transactionID, req)
	if err != nil {
		return nil, err
	}

	tx, err := s.repo.PatchTransaction(ctx, userID, transactionID, patch)
	if err != nil {
		if errors.Is(err, apperrors.ErrTransactionNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "transaction not found", err)
		}
		if errors.Is(err, apperrors.ErrTransactionForbidden) {
			return nil, apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		return nil, err
	}
	return tx, nil
}

func (s *Service) buildTransactionPatch(ctx context.Context, userID, transactionID string, req PatchRequest) (domain.TransactionPatch, error) {
	cur, err := s.repo.GetTransaction(ctx, userID, transactionID)
	if err != nil {
		if errors.Is(err, apperrors.ErrTransactionNotFound) {
			return domain.TransactionPatch{}, apperrors.Wrap(apperrors.KindNotFound, "transaction not found", err)
		}
		if errors.Is(err, apperrors.ErrTransactionForbidden) {
			return domain.TransactionPatch{}, apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		return domain.TransactionPatch{}, err
	}

	patch := domain.TransactionPatch{
		Description: normalizeOptionalString(req.Description),
		CategoryIDs: req.CategoryIDs,
		TagIDs:      req.TagIDs,
		Amount:      normalizeOptionalString(req.Amount),
	}

	if req.Status != nil {
		nextStatus := strings.TrimSpace(*req.Status)
		switch nextStatus {
		case "pending", "posted", "cancelled":
			patch.Status = &nextStatus
		default:
			return domain.TransactionPatch{}, apperrors.Validation("status must be one of: pending, posted, cancelled", map[string]any{"field": "status"})
		}
	}

	if req.Amount != nil {
		a := strings.TrimSpace(*req.Amount)
		if a == "" {
			return domain.TransactionPatch{}, apperrors.Validation("amount is required", nil)
		}
		if !isValidDecimal(a) {
			return domain.TransactionPatch{}, apperrors.Validation("amount must be a decimal string", nil)
		}
	}

	if cur.Type == "transfer" {
		if req.LineItems != nil {
			return domain.TransactionPatch{}, apperrors.Validation("line_items must be empty for transfer", nil)
		}
		if len(req.CategoryIDs) > 0 {
			return domain.TransactionPatch{}, apperrors.Validation("category_ids must be empty for transfer", nil)
		}
	}

	// Convert LineItemInput → domain.TransactionLineItem
	if req.LineItems != nil {
		if cur.Type != "transfer" && len(*req.LineItems) == 0 {
			return domain.TransactionPatch{}, apperrors.Validation("line_items is required and must include at least one category", nil)
		}

		sum := big.NewRat(0, 1)
		items := make([]domain.TransactionLineItem, len(*req.LineItems))
		for i, li := range *req.LineItems {
			if cur.Type != "transfer" {
				if li.CategoryID == nil || strings.TrimSpace(*li.CategoryID) == "" {
					return domain.TransactionPatch{}, apperrors.Validation("line_items.category_id is required", nil)
				}
			}

			liAmt := strings.TrimSpace(li.Amount)
			if liAmt == "" {
				return domain.TransactionPatch{}, apperrors.Validation("line_items.amount is required", nil)
			}
			if !isValidDecimal(liAmt) {
				return domain.TransactionPatch{}, apperrors.Validation("line_items.amount must be a decimal string", nil)
			}
			r, ok := new(big.Rat).SetString(liAmt)
			if !ok {
				return domain.TransactionPatch{}, apperrors.Validation("line_items.amount must be a decimal string", nil)
			}
			sum.Add(sum, r)

			liID := uuid.NewString()
			liTags, err := s.ensureTags(ctx, userID, li.TagIDs, req.Lang)
			if err != nil {
				return domain.TransactionPatch{}, err
			}
			items[i] = domain.TransactionLineItem{
				ID:         liID,
				CategoryID: li.CategoryID,
				TagIDs:     liTags,
				Amount:     liAmt,
				Note:       li.Note,
			}
		}

		if cur.Type != "transfer" {
			summed := sum.FloatString(2)
			patch.Amount = &summed
		}
		patch.LineItems = &items
	} else if req.Amount != nil && cur.Type != "transfer" {
		return domain.TransactionPatch{}, apperrors.Validation("line_items is required when updating amount", nil)
	}

	// Convert GroupParticipantInput → domain.GroupExpenseParticipant
	if req.GroupParticipants != nil {
		now := time.Now().UTC()
		parts := make([]domain.GroupExpenseParticipant, len(*req.GroupParticipants))
		for i, g := range *req.GroupParticipants {
			parts[i] = domain.GroupExpenseParticipant{
				ID:              uuid.NewString(),
				UserID:          userID,
				TransactionID:   transactionID,
				ParticipantName: g.ParticipantName,
				OriginalAmount:  g.OriginalAmount,
				ShareAmount:     g.ShareAmount,
				IsSettled:       false,
				CreatedAt:       now,
				UpdatedAt:       now,
			}
		}
		patch.GroupParticipants = &parts
	}

	if req.OccurredAt != nil {
		v := strings.TrimSpace(*req.OccurredAt)
		if v != "" {
			t, err := parseTimeOrDate(v)
			if err != nil {
				return domain.TransactionPatch{}, apperrors.Validation("occurred_at is invalid", nil)
			}
			patch.OccurredAt = &t
		}
	}

	if req.TagIDs != nil {
		resolved, err := s.ensureTags(ctx, userID, req.TagIDs, req.Lang)
		if err != nil {
			return domain.TransactionPatch{}, err
		}
		patch.TagIDs = resolved
	}

	return patch, nil
}

func (s *Service) ensureTags(ctx context.Context, userID string, inputs []string, lang string) ([]string, error) {
	if len(inputs) == 0 {
		return []string{}, nil
	}

	if lang == "" {
		lang = "en"
	}

	out := make([]string, 0, len(inputs))
	for _, input := range inputs {
		trimmed := strings.TrimSpace(input)
		if trimmed == "" {
			continue
		}

		// If it's a UUID, assume it's an existing ID.
		if _, err := uuid.Parse(trimmed); err == nil {
			out = append(out, trimmed)
			continue
		}

		// Otherwise, it's a target label name to get-or-create.
		id, err := s.tagService.GetOrCreateByName(ctx, userID, trimmed, lang)
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}

func (s *Service) BatchPatch(ctx context.Context, userID string, req BatchPatchRequest) (*BatchPatchResult, error) {
	if len(req.TransactionIDs) == 0 {
		return nil, apperrors.Validation("transaction_ids is required", map[string]any{"field": "transaction_ids"})
	}

	hasPatch := req.Patch.Description != nil ||
		req.Patch.Amount != nil ||
		req.Patch.Status != nil ||
		req.Patch.OccurredAt != nil ||
		len(req.Patch.CategoryIDs) > 0 ||
		len(req.Patch.TagIDs) > 0 ||
		req.Patch.LineItems != nil ||
		req.Patch.GroupParticipants != nil
	if !hasPatch {
		return nil, apperrors.Validation("patch payload is required", map[string]any{"field": "patch"})
	}

	mode := "atomic"
	if req.Mode != nil {
		requestedMode := strings.ToLower(strings.TrimSpace(*req.Mode))
		if requestedMode != "" {
			if requestedMode != "atomic" && requestedMode != "partial" {
				return nil, apperrors.Validation("mode must be one of: atomic, partial", map[string]any{"field": "mode"})
			}
			mode = requestedMode
		}
	}

	result := &BatchPatchResult{Mode: mode}

	patchesByID := make(map[string]domain.TransactionPatch, len(req.TransactionIDs))
	preparedIDs := make([]string, 0, len(req.TransactionIDs))
	failedValidationIDs := make([]string, 0)

	seen := make(map[string]struct{}, len(req.TransactionIDs))
	for _, id := range req.TransactionIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}

		patch, err := s.buildTransactionPatch(ctx, userID, trimmed, req.Patch)
		if err != nil {
			if mode == "atomic" {
				return nil, apperrors.Validation("atomic batch patch validation failed", map[string]any{"field": "transaction_ids", "transaction_id": trimmed})
			}
			failedValidationIDs = append(failedValidationIDs, trimmed)
			continue
		}

		patchesByID[trimmed] = patch
		preparedIDs = append(preparedIDs, trimmed)
	}

	if len(preparedIDs) == 0 {
		result.UpdatedIDs = []string{}
		result.FailedIDs = failedValidationIDs
		result.UpdatedCount = 0
		result.FailedCount = len(failedValidationIDs)
		return result, nil
	}

	updatedIDs, failedIDs, err := s.repo.BatchPatchTransactions(ctx, userID, preparedIDs, patchesByID, mode)
	if err != nil {
		if errors.Is(err, apperrors.ErrTransactionNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "transaction not found", err)
		}
		if errors.Is(err, apperrors.ErrTransactionForbidden) {
			return nil, apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		return nil, err
	}

	result.UpdatedIDs = updatedIDs
	result.FailedIDs = append(failedValidationIDs, failedIDs...)
	result.UpdatedCount = len(updatedIDs)
	result.FailedCount = len(result.FailedIDs)
	return result, nil
}

func (s *Service) ImportGoenV1(ctx context.Context, userID, accountID string, items []ImportGoenV1Item) (*ImportGoenV1Result, error) {
	if strings.TrimSpace(accountID) == "" {
		return nil, apperrors.Validation("accountId is required", map[string]any{"field": "accountId"})
	}
	if len(items) == 0 {
		return nil, apperrors.Validation("items is required", map[string]any{"field": "items"})
	}

	result := &ImportGoenV1Result{Errors: make([]string, 0)}

	for i, item := range items {
		linePrefix := fmt.Sprintf("item[%d]", i)

		dateStr := strings.TrimSpace(item.TransactionDate)
		if dateStr == "" {
			result.Skipped++
			result.Errors = append(result.Errors, linePrefix+": transaction_date is required")
			continue
		}
		if _, err := time.Parse("2006-01-02", dateStr); err != nil {
			result.Skipped++
			result.Errors = append(result.Errors, linePrefix+": transaction_date must be YYYY-MM-DD")
			continue
		}

		amountRaw := strings.TrimSpace(item.Amount)
		if !isValidDecimal(amountRaw) {
			result.Skipped++
			result.Errors = append(result.Errors, linePrefix+": amount must be decimal string")
			continue
		}

		r, ok := new(big.Rat).SetString(amountRaw)
		if !ok {
			result.Skipped++
			result.Errors = append(result.Errors, linePrefix+": amount must be decimal string")
			continue
		}

		var catID string
		if item.CategoryID != nil && strings.TrimSpace(*item.CategoryID) != "" {
			catID = strings.TrimSpace(*item.CategoryID)
		} else if item.Category != nil && strings.TrimSpace(*item.Category) != "" {
			// NOTE: Category names are no longer supported for lookups since the name column was removed.
			// CSV imports must use category_id instead. Use the API to list available category IDs.
			result.Skipped++
			result.Errors = append(result.Errors, linePrefix+": category_id is required (category name lookups no longer supported)")
			continue
		} else {
			result.Skipped++
			result.Errors = append(result.Errors, linePrefix+": category mapping is required (category_id or category)")
			continue
		}

		kind := "income"
		if r.Sign() < 0 {
			kind = "expense"
			r.Neg(r)
		}

		amountAbs := r.FloatString(2)
		desc := normalizeOptionalString(item.Description)
		_, createErr := s.Create(ctx, userID, CreateRequest{
			Type:         kind,
			OccurredDate: &dateStr,
			Amount:       amountAbs,
			Description:  desc,
			AccountID:    &accountID,
			LineItems: []CreateLineItemRequest{
				{
					CategoryID: &catID,
					Amount:     amountAbs,
				},
			},
		})
		if createErr != nil {
			result.Skipped++
			result.Errors = append(result.Errors, linePrefix+": "+createErr.Error())
			continue
		}

		result.Imported++
	}

	return result, nil
}

func (s *Service) StageImported(ctx context.Context, userID string, source string, items []StageImportedItem) ([]domain.ImportedTransaction, error) {
	if len(items) == 0 {
		return []domain.ImportedTransaction{}, nil
	}

	creates := make([]domain.ImportedTransactionCreate, 0, len(items))
	for i, item := range items {
		linePrefix := fmt.Sprintf("item[%d]", i)
		dateStr := strings.TrimSpace(item.TransactionDate)
		if dateStr == "" {
			return nil, apperrors.Validation(linePrefix+": transaction_date is required", map[string]any{"field": "transaction_date"})
		}
		normalizedDate, err := normalizeImportDate(dateStr)
		if err != nil {
			return nil, apperrors.Validation(linePrefix+": transaction_date must be YYYY-MM-DD", map[string]any{"field": "transaction_date"})
		}

		amountRaw := strings.TrimSpace(item.Amount)
		if !isValidDecimal(amountRaw) {
			return nil, apperrors.Validation(linePrefix+": amount must be decimal string", map[string]any{"field": "amount"})
		}
		r, ok := new(big.Rat).SetString(amountRaw)
		if !ok {
			return nil, apperrors.Validation(linePrefix+": amount must be decimal string", map[string]any{"field": "amount"})
		}

		var txType *string
		if item.TransactionType != nil {
			t := strings.ToLower(strings.TrimSpace(*item.TransactionType))
			if t != "" {
				txType = &t
			}
		}
		if txType == nil {
			if r.Sign() < 0 {
				v := "expense"
				txType = &v
			} else {
				v := "income"
				txType = &v
			}
		}

		categoryName := normalizeOptionalString(item.CategoryName)
		if categoryName == nil {
			categoryName = normalizeOptionalString(item.Category)
		}

		creates = append(creates, domain.ImportedTransactionCreate{
			Source:               source,
			TransactionDate:      normalizedDate,
			Amount:               amountRaw,
			Description:          normalizeOptionalString(item.Description),
			TransactionType:      txType,
			ImportedAccountName:  normalizeOptionalString(item.AccountName),
			ImportedCategoryName: categoryName,
			RawPayload:           item.Raw,
		})
	}
	return s.repo.CreateImportedTransactions(ctx, userID, creates)
}

func (s *Service) StageImportedGoenV1(ctx context.Context, userID string, items []ImportedGoenV1StageItem) ([]domain.ImportedTransaction, error) {
	if len(items) == 0 {
		return nil, apperrors.Validation("items is required", map[string]any{"field": "items"})
	}

	genericItems := make([]StageImportedItem, len(items))
	for i, item := range items {
		genericItems[i] = StageImportedItem{
			TransactionDate: item.TransactionDate,
			Amount:          item.Amount,
			Description:     item.Description,
			TransactionType: item.TransactionType,
			AccountName:     item.AccountName,
			CategoryName:    item.CategoryName,
			Raw:             item.Raw,
		}
	}

	return s.StageImported(ctx, userID, "goen_v1", genericItems)
}

func (s *Service) ListImportedGoenV1(ctx context.Context, userID string) ([]domain.ImportedTransaction, error) {
	items, err := s.repo.ListImportedTransactions(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ImportedTransaction, 0)
	for _, it := range items {
		if it.Source == "goen_v1" {
			out = append(out, it)
		}
	}
	return out, nil
}

func (s *Service) MapImportedGoenV1(ctx context.Context, userID, importID string, accountID, categoryID *string) (*domain.ImportedTransaction, error) {
	return s.repo.PatchImportedTransaction(ctx, userID, importID, domain.ImportedTransactionPatch{
		MappedAccountID:  accountID,
		MappedCategoryID: categoryID,
	})
}

func (s *Service) CreateFromImportedGoenV1(ctx context.Context, userID, importID string) (*domain.Transaction, error) {
	it, err := s.repo.GetImportedTransaction(ctx, userID, importID)
	if err != nil {
		return nil, err
	}

	req, err := s.buildCreateRequestFromImported(ctx, userID, *it)
	if err != nil {
		return nil, err
	}

	tx, err := s.Create(ctx, userID, *req)
	if err != nil {
		return nil, err
	}

	if err := s.repo.DeleteImportedTransaction(ctx, userID, importID); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *Service) CreateManyFromImportedGoenV1(ctx context.Context, userID string, importIDs []string) (*CreateImportedGoenV1BatchResult, error) {
	result := &CreateImportedGoenV1BatchResult{Errors: []string{}}
	for _, id := range importIDs {
		tx, err := s.CreateFromImportedGoenV1(ctx, userID, id)
		if err != nil {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("Import %s: %v", id, err))
			continue
		}
		if tx != nil {
			result.Created++
		}
	}
	return result, nil
}

func (s *Service) DeleteImportedGoenV1(ctx context.Context, userID, importID string) error {
	return s.repo.DeleteImportedTransaction(ctx, userID, importID)
}

func (s *Service) buildCreateRequestFromImported(ctx context.Context, userID string, it domain.ImportedTransaction) (*CreateRequest, error) {
	if it.MappedAccountID == nil || strings.TrimSpace(*it.MappedAccountID) == "" {
		return nil, apperrors.Validation("mapped_account_id is required", map[string]any{"field": "mapped_account_id"})
	}
	if it.MappedCategoryID == nil || strings.TrimSpace(*it.MappedCategoryID) == "" {
		return nil, apperrors.Validation("mapped_category_id is required", map[string]any{"field": "mapped_category_id"})
	}

	r, ok := new(big.Rat).SetString(strings.TrimSpace(it.Amount))
	if !ok {
		return nil, apperrors.Validation("amount is invalid", map[string]any{"field": "amount"})
	}

	if it.TransactionType != nil && strings.EqualFold(strings.TrimSpace(*it.TransactionType), "transfer") {
		toAccountName := rawPayloadString(it.RawPayload, "to_account_name", "toAccountName")
		if toAccountName == nil || strings.TrimSpace(*toAccountName) == "" {
			return nil, apperrors.Validation("to_account_name is required for transfer import", map[string]any{"field": "to_account_name"})
		}

		toAccountID, err := s.resolveUserAccountIDByName(ctx, userID, *toAccountName)
		if err != nil {
			return nil, err
		}
		if toAccountID == nil {
			return nil, apperrors.Validation("cannot map to_account_name to active account", map[string]any{"field": "to_account_name"})
		}
		if *toAccountID == strings.TrimSpace(*it.MappedAccountID) {
			return nil, apperrors.Validation("from_account_id and to_account_id must be different", map[string]any{"field": "to_account_name"})
		}

		amountAbsRat := new(big.Rat).Set(r)
		if amountAbsRat.Sign() < 0 {
			amountAbsRat.Neg(amountAbsRat)
		}
		amountAbs := amountAbsRat.FloatString(2)

		return &CreateRequest{
			Type:          "transfer",
			OccurredDate:  &it.TransactionDate,
			Amount:        amountAbs,
			Description:   it.Description,
			FromAccountID: it.MappedAccountID,
			ToAccountID:   toAccountID,
		}, nil
	}

	kind := "income"
	if it.TransactionType != nil {
		t := strings.ToLower(strings.TrimSpace(*it.TransactionType))
		if t == "expense" || t == "income" {
			kind = t
		}
	}
	if kind != "expense" && kind != "income" {
		if r.Sign() < 0 {
			kind = "expense"
		} else {
			kind = "income"
		}
	}
	if r.Sign() < 0 {
		r.Neg(r)
	}
	amountAbs := r.FloatString(2)

	return &CreateRequest{
		Type:         kind,
		OccurredDate: &it.TransactionDate,
		Amount:       amountAbs,
		Description:  it.Description,
		AccountID:    it.MappedAccountID,
		LineItems: []CreateLineItemRequest{{
			CategoryID: it.MappedCategoryID,
			Amount:     amountAbs,
		}},
	}, nil
}

// ============================================================================
// Generic Import/Export (source-agnostic) - supports v1, v2, or other sources
// ============================================================================

// ExportTransactions exports transactions in a portable CSV format (goen-v2 compatible)
func (s *Service) ExportTransactions(ctx context.Context, userID string, filter ExportTransactionsFilter) ([]domain.ExportTransactionRow, error) {
	transactions, _, _, err := s.repo.ListTransactions(ctx, userID, domain.TransactionListFilter{
		AccountID: filter.AccountID,
		From:      filter.From,
		To:        filter.To,
		Limit:     10000, // Export limit
	})
	if err != nil {
		return nil, err
	}

	accounts, err := s.accountRepo.ListAccountsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	accIDToName := make(map[string]string)
	for _, a := range accounts {
		accIDToName[a.ID] = a.Name
	}

	/*
	categories, err := s.categoryRepo.ListCategories(ctx, userID)
	if err != nil {
		return nil, err
	}
	*/
	catIDToName := make(map[string]string)
	// Categories don't have a simple name field anymore, but for export we'll leave it empty 
	// or you can implement label resolution here if needed.
	/*
	for _, c := range categories {
		catIDToName[c.ID] = c.Name
	}
	*/


	rows := make([]domain.ExportTransactionRow, 0, len(transactions))
	for _, tx := range transactions {
		row := domain.ExportTransactionRow{
			TransactionID:   tx.ID,
			TransactionDate: tx.OccurredDate,
			Amount:          tx.Amount,
			Description:     tx.Description,
			TransactionType: tx.Type,
		}

		if tx.Type == "expense" || tx.Type == "income" {
			if tx.AccountID != nil {
				row.AccountID = *tx.AccountID
				row.AccountName = accIDToName[*tx.AccountID]
			}
			if len(tx.LineItems) > 0 {
				if tx.LineItems[0].CategoryID != nil {
					catID := *tx.LineItems[0].CategoryID
					row.CategoryID = &catID
					if name, ok := catIDToName[catID]; ok {
						row.CategoryName = &name
					}
				}
			}
		} else if tx.Type == "transfer" {
			if tx.FromAccountID != nil {
				accID := *tx.FromAccountID
				row.FromAccountID = &accID
				name := accIDToName[accID]
				row.FromAccountName = &name
			}
			if tx.ToAccountID != nil {
				accID := *tx.ToAccountID
				row.ToAccountID = &accID
				name := accIDToName[accID]
				row.ToAccountName = &name
			}
		}
		rows = append(rows, row)
	}

	return rows, nil
}

// ListImported lists staged imported transactions (optionally filtered by source)
func (s *Service) ListImported(ctx context.Context, userID string, source *string) ([]domain.ImportedTransaction, error) {
	items, err := s.repo.ListImportedTransactions(ctx, userID)
	if err != nil {
		return nil, err
	}

	if source == nil {
		return items, nil
	}

	filtered := make([]domain.ImportedTransaction, 0)
	for _, it := range items {
		if it.Source == *source {
			filtered = append(filtered, it)
		}
	}
	return filtered, nil
}

// MapImported updates mapping for a staged imported transaction
func (s *Service) MapImported(ctx context.Context, userID, importID string, accountID, categoryID *string) (*domain.ImportedTransaction, error) {
	return s.repo.PatchImportedTransaction(ctx, userID, importID, domain.ImportedTransactionPatch{
		MappedAccountID:  accountID,
		MappedCategoryID: categoryID,
	})
}

// CreateFromImported creates a transaction from a single staged import and deletes the import
func (s *Service) CreateFromImported(ctx context.Context, userID, importID string) (*domain.Transaction, error) {
	it, err := s.repo.GetImportedTransaction(ctx, userID, importID)
	if err != nil {
		return nil, err
	}

	req, err := s.buildCreateRequestFromImported(ctx, userID, *it)
	if err != nil {
		return nil, err
	}

	tx, err := s.Create(ctx, userID, *req)
	if err != nil {
		return nil, err
	}

	if err := s.repo.DeleteImportedTransaction(ctx, userID, importID); err != nil {
		return nil, err
	}

	return tx, nil
}

// CreateManyFromImported creates transactions from multiple staged imports
func (s *Service) CreateManyFromImported(ctx context.Context, userID string, importIDs []string) (*StagedImportResult, error) {
	result := &StagedImportResult{Errors: []string{}}
	for _, id := range importIDs {
		tx, err := s.CreateFromImported(ctx, userID, id)
		if err != nil {
			result.Skipped++
			result.Errors = append(result.Errors, fmt.Sprintf("Import %s: %v", id, err))
			continue
		}
		if tx != nil {
			result.Created++
		}
	}
	return result, nil
}

// DeleteImported deletes a staged imported transaction
func (s *Service) DeleteImported(ctx context.Context, userID, importID string) error {
	return s.repo.DeleteImportedTransaction(ctx, userID, importID)
}

func (s *Service) DeleteAllImported(ctx context.Context, userID string) (int64, error) {
	return s.repo.DeleteAllImportedTransactions(ctx, userID)
}

func (s *Service) ListImportRules(ctx context.Context, userID string) ([]domain.ImportMappingRule, error) {
	return s.repo.ListImportMappingRules(ctx, userID)
}

func (s *Service) UpsertImportRules(ctx context.Context, userID string, inputs []MappingRuleInput) ([]domain.ImportMappingRule, error) {
	upserts := make([]domain.ImportMappingRuleUpsert, 0, len(inputs))
	for _, input := range inputs {
		upserts = append(upserts, domain.ImportMappingRuleUpsert{
			Kind:       input.Kind,
			SourceName: input.SourceName,
			MappedID:   input.MappedID,
		})
	}

	return s.repo.UpsertImportMappingRules(ctx, userID, upserts)
}

func (s *Service) DeleteImportRule(ctx context.Context, userID, ruleID string) error {
	return s.repo.DeleteImportMappingRule(ctx, userID, ruleID)
}

func (s *Service) ApplyImportRulesAndCreate(ctx context.Context, userID string) (*ApplyImportRulesResult, error) {
	rules, err := s.repo.ListImportMappingRules(ctx, userID)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.ListImportedTransactions(ctx, userID)
	if err != nil {
		return nil, err
	}

	normalize := func(v string) string {
		return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(v)), " "))
	}

	accountRules := map[string]string{}
	categoryRules := map[string]string{}
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

	result := &ApplyImportRulesResult{Errors: make([]string, 0)}

	for _, item := range items {
		patch := domain.ImportedTransactionPatch{}
		hasPatch := false

		if item.MappedAccountID == nil && item.ImportedAccountName != nil {
			if mappedID := accountRules[normalize(*item.ImportedAccountName)]; mappedID != "" {
				patch.MappedAccountID = &mappedID
				hasPatch = true
			}
		}
		if item.MappedCategoryID == nil && item.ImportedCategoryName != nil {
			if mappedID := categoryRules[normalize(*item.ImportedCategoryName)]; mappedID != "" {
				patch.MappedCategoryID = &mappedID
				hasPatch = true
			}
		}

		if !hasPatch {
			continue
		}
		if _, err := s.repo.PatchImportedTransaction(ctx, userID, item.ID, patch); err != nil {
			result.Errors = append(result.Errors, item.ID+": "+err.Error())
			continue
		}
		result.UpdatedMappings++
	}

	itemsAfterPatch, err := s.repo.ListImportedTransactions(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, item := range itemsAfterPatch {
		if item.MappedAccountID == nil {
			continue
		}
		isTransfer := item.TransactionType != nil && strings.EqualFold(strings.TrimSpace(*item.TransactionType), "transfer")
		if !isTransfer && item.MappedCategoryID == nil {
			continue
		}
		if _, err := s.CreateFromImported(ctx, userID, item.ID); err != nil {
			result.Errors = append(result.Errors, item.ID+": "+err.Error())
			continue
		}
		result.Created++
	}

	remaining, err := s.repo.ListImportedTransactions(ctx, userID)
	if err != nil {
		return nil, err
	}
	result.Remaining = len(remaining)

	return result, nil
}

// Delete soft-deletes a transaction.
func (s *Service) Delete(ctx context.Context, userID, transactionID string) error {
	err := s.repo.DeleteTransaction(ctx, userID, transactionID)
	if err != nil {
		if errors.Is(err, apperrors.ErrTransactionNotFound) {
			return apperrors.Wrap(apperrors.KindNotFound, "transaction not found", err)
		}
		if errors.Is(err, apperrors.ErrTransactionForbidden) {
			return apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		return err
	}
	return nil
}

func validateTransactionLinkage(tx domain.Transaction) error {
	switch tx.Type {
	case "expense", "income":
		if tx.AccountID == nil {
			return apperrors.Validation("account_id is required", nil)
		}
		if tx.FromAccountID != nil || tx.ToAccountID != nil {
			return apperrors.Validation("from_account_id/to_account_id must be empty", nil)
		}
	case "transfer":
		if tx.FromAccountID == nil {
			return apperrors.Validation("from_account_id is required", nil)
		}
		if tx.ToAccountID == nil {
			return apperrors.Validation("to_account_id is required", nil)
		}
		if tx.AccountID != nil {
			return apperrors.Validation("account_id must be empty", nil)
		}
	default:
		return apperrors.Validation("type is invalid", nil)
	}
	return nil
}

func normalizeOccurredAt(occurredAt, occurredDate, occurredTime *string) (time.Time, string, error) {
	if occurredAt != nil {
		v := strings.TrimSpace(*occurredAt)
		if v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return time.Time{}, "", apperrors.Validation("occurred_at is invalid", nil)
			}
			return t.UTC(), t.UTC().Format("2006-01-02"), nil
		}
	}

	if occurredDate == nil || strings.TrimSpace(*occurredDate) == "" {
		return time.Time{}, "", apperrors.Validation("occurred_date is required", nil)
	}
	d, err := time.Parse("2006-01-02", strings.TrimSpace(*occurredDate))
	if err != nil {
		return time.Time{}, "", apperrors.Validation("occurred_date is invalid", nil)
	}

	h := 0
	m := 0
	if occurredTime != nil {
		v := strings.TrimSpace(*occurredTime)
		if v != "" {
			tm, err := time.Parse("15:04", v)
			if err != nil {
				return time.Time{}, "", apperrors.Validation("occurred_time is invalid", nil)
			}
			h = tm.Hour()
			m = tm.Minute()
		}
	}

	t := time.Date(d.Year(), d.Month(), d.Day(), h, m, 0, 0, time.UTC)
	return t, t.Format("2006-01-02"), nil
}

func parseTimeOrDate(v string) (time.Time, error) {
	if strings.Contains(v, "T") {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return time.Time{}, err
		}
		return t.UTC(), nil
	}
	d, err := time.Parse("2006-01-02", v)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC), nil
}

func normalizeImportDate(v string) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", errors.New("empty date")
	}

	layouts := []string{
		"2006-01-02",
		"2006/01/02",
		"02/01/2006",
		"2/1/2006",
		"02-01-2006",
		"2-1-2006",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		time.RFC3339,
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, v); err == nil {
			return t.UTC().Format("2006-01-02"), nil
		}
	}

	return "", errors.New("invalid date format")
}

func normalizeOptionalString(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}

func normalizeLookupName(v string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(v)), " "))
}

func rawPayloadString(raw map[string]any, keys ...string) *string {
	if len(raw) == 0 {
		return nil
	}
	for _, key := range keys {
		if val, ok := raw[key]; ok {
			switch typed := val.(type) {
			case string:
				trimmed := strings.TrimSpace(typed)
				if trimmed != "" {
					return &trimmed
				}
			}
		}
	}
	return nil
}

func (s *Service) resolveUserAccountIDByName(ctx context.Context, userID, accountName string) (*string, error) {
	if s.accountRepo == nil {
		return nil, nil
	}
	needle := normalizeLookupName(accountName)
	if needle == "" {
		return nil, nil
	}
	accounts, err := s.accountRepo.ListAccountsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	byID := map[string]domain.Account{}
	for _, account := range accounts {
		byID[account.ID] = account
	}

	for _, account := range accounts {
		if strings.TrimSpace(account.Status) != "active" {
			continue
		}
		if normalizeLookupName(account.Name) == needle {
			id := account.ID
			return &id, nil
		}
	}

	for _, account := range accounts {
		if strings.TrimSpace(account.Status) != "active" {
			continue
		}
		for _, candidate := range accountLookupCandidates(account, byID) {
			if candidate == needle {
				id := account.ID
				return &id, nil
			}
		}
	}
	return nil, nil
}

func accountLookupCandidates(account domain.Account, byID map[string]domain.Account) []string {
	name := strings.TrimSpace(account.Name)
	if name == "" {
		return nil
	}

	out := []string{normalizeLookupName(name)}
	if account.ParentAccountID != nil {
		if parent, ok := byID[*account.ParentAccountID]; ok {
			parentName := strings.TrimSpace(parent.Name)
			if parentName != "" {
				out = append(out,
					normalizeLookupName(parentName+" / "+name),
					normalizeLookupName(parentName+" - "+name),
					normalizeLookupName(parentName+" > "+name),
				)
			}
		}
	}

	return out
}

func normalizeTagIDs(ids []string) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		v := strings.TrimSpace(id)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func isValidDecimal(s string) bool {
	_, ok := new(big.Rat).SetString(s)
	return ok
}

func roundRat(r *big.Rat, scale int) *big.Rat {
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

func formatRatDecimalScale(r *big.Rat, scale int) string {
	rr := roundRat(r, scale)
	return rr.FloatString(scale)
}

func pointer[T any](v T) *T {
	return &v
}



