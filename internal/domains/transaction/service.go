package transaction

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
	"github.com/sonbn-225/goen-api-v2/internal/core/money"
)

type service struct {
	repo Repository
}

var _ Service = (*service)(nil)

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, userID string, input CreateInput) (*Transaction, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "transaction", "operation", "create")
	logger.Info("transaction_create_started", "user_id", userID)

	if strings.TrimSpace(userID) == "" {
		logger.Warn("transaction_create_failed", "reason", "missing user context")
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}

	kind := strings.ToLower(strings.TrimSpace(input.Type))
	if kind == "" {
		return nil, apperrors.New(apperrors.KindValidation, "transaction type is required")
	}
	if kind != "expense" && kind != "income" && kind != "transfer" {
		return nil, apperrors.New(apperrors.KindValidation, "unsupported transaction type")
	}
	if !input.Amount.GreaterThan(money.Zero().Decimal) {
		return nil, apperrors.New(apperrors.KindValidation, "amount must be greater than zero")
	}

	lineItems, effectiveAmount, err := normalizeCreateLineItems(kind, input.LineItems, input.Note, input.Amount)
	if err != nil {
		return nil, err
	}

	ownerOriginalAmount := money.Zero()
	if input.OwnerOriginalAmount != nil {
		if input.OwnerOriginalAmount.LessThan(money.Zero().Decimal) {
			return nil, apperrors.New(apperrors.KindValidation, "owner_original_amount must be greater than or equal to zero")
		}
		ownerOriginalAmount = *input.OwnerOriginalAmount
	}

	accountID := normalizeOptionalString(input.AccountID)
	fromAccountID := normalizeOptionalString(input.FromAccountID)
	toAccountID := normalizeOptionalString(input.ToAccountID)

	if kind == "transfer" {
		if accountID != nil {
			return nil, apperrors.New(apperrors.KindValidation, "account_id must be empty for transfer")
		}
		if fromAccountID == nil {
			return nil, apperrors.New(apperrors.KindValidation, "from_account_id is required")
		}
		if toAccountID == nil {
			return nil, apperrors.New(apperrors.KindValidation, "to_account_id is required")
		}
		if *fromAccountID == *toAccountID {
			return nil, apperrors.New(apperrors.KindValidation, "from_account_id and to_account_id must be different")
		}
	} else {
		if accountID == nil {
			return nil, apperrors.New(apperrors.KindValidation, "account_id is required")
		}
		if fromAccountID != nil || toAccountID != nil {
			return nil, apperrors.New(apperrors.KindValidation, "from_account_id/to_account_id must be empty")
		}
	}

	if kind != "expense" {
		if len(input.GroupParticipants) > 0 {
			return nil, apperrors.New(apperrors.KindValidation, "group_participants are only supported for expense transactions")
		}
		if ownerOriginalAmount.GreaterThan(money.Zero().Decimal) {
			return nil, apperrors.New(apperrors.KindValidation, "owner_original_amount is only supported for expense transactions")
		}
	}

	createOpts, err := normalizeCreateOptions(input.GroupParticipants, ownerOriginalAmount, effectiveAmount)
	if err != nil {
		return nil, err
	}
	createOpts.LineItems = lineItems

	tx := &Transaction{
		ID:            uuid.NewString(),
		UserID:        userID,
		AccountID:     accountID,
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Type:          kind,
		Status:        "pending",
		Amount:        effectiveAmount,
		Note:          strings.TrimSpace(input.Note),
		OccurredAt:    time.Now().UTC(),
		CreatedAt:     time.Now().UTC(),
	}
	if len(lineItems) > 0 {
		tx.LineItems = toTransactionLineItems(lineItems)
	}

	if err := s.repo.Create(ctx, tx, createOpts); err != nil {
		logger.Error("transaction_create_failed", "error", err)
		return nil, passThroughOrWrapInternal("failed to create transaction", err)
	}
	logger.Info("transaction_create_succeeded", "transaction_id", tx.ID, "type", tx.Type)
	return tx, nil
}

func (s *service) Update(ctx context.Context, userID, transactionID string, input UpdateInput) (*Transaction, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "transaction", "operation", "update")
	logger.Info("transaction_update_started", "user_id", userID, "transaction_id", transactionID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	transactionID = strings.TrimSpace(transactionID)
	if transactionID == "" {
		return nil, apperrors.New(apperrors.KindValidation, "transaction id is required")
	}

	current, err := s.Get(ctx, userID, transactionID)
	if err != nil {
		return nil, err
	}

	normalized := UpdateInput{}
	if input.Note != nil {
		note := strings.TrimSpace(*input.Note)
		normalized.Note = &note
	}

	if input.LineItems != nil {
		items, err := normalizeUpdateLineItems(current.Type, *input.LineItems, normalized.Note)
		if err != nil {
			return nil, err
		}
		normalized.LineItems = &items
	}

	if input.GroupParticipants != nil {
		participants, err := normalizeUpdateGroupParticipants(current.Type, *input.GroupParticipants)
		if err != nil {
			return nil, err
		}
		normalized.GroupParticipants = &participants
	}

	if normalized.Note == nil && normalized.LineItems == nil && normalized.GroupParticipants == nil {
		return nil, apperrors.New(apperrors.KindValidation, "no fields to update")
	}

	updated, err := s.repo.Update(ctx, userID, transactionID, normalized)
	if err != nil {
		logger.Error("transaction_update_failed", "error", err)
		return nil, passThroughOrWrapInternal("failed to update transaction", err)
	}
	if updated == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "transaction not found")
	}

	logger.Info("transaction_update_succeeded", "transaction_id", updated.ID)
	return updated, nil
}

func normalizeCreateLineItems(kind string, lineItems []CreateTransactionLineItemInput, topLevelNote string, baseAmount money.Amount) ([]CreateTransactionLineItem, money.Amount, error) {
	if kind == "transfer" {
		if len(lineItems) > 0 {
			return nil, money.Amount{}, apperrors.New(apperrors.KindValidation, "line_items must be empty for transfer")
		}
		return nil, baseAmount, nil
	}

	if len(lineItems) == 0 {
		return nil, money.Amount{}, apperrors.New(apperrors.KindValidation, "line_items is required and must include at least one category")
	}

	sum := money.Zero().Decimal
	cleaned := make([]CreateTransactionLineItem, 0, len(lineItems))
	for _, lineItem := range lineItems {
		categoryID := normalizeOptionalString(lineItem.CategoryID)
		if categoryID == nil {
			return nil, money.Amount{}, apperrors.New(apperrors.KindValidation, "line_items.category_id is required")
		}
		if !lineItem.Amount.GreaterThan(money.Zero().Decimal) {
			return nil, money.Amount{}, apperrors.New(apperrors.KindValidation, "line_items.amount must be greater than zero")
		}

		cleaned = append(cleaned, CreateTransactionLineItem{
			CategoryID: categoryID,
			TagIDs:     normalizeTagIDs(lineItem.TagIDs),
			Amount:     lineItem.Amount,
			Note:       normalizeOptionalString(lineItem.Note),
		})
		sum = sum.Add(lineItem.Amount.Decimal)
	}

	note := strings.TrimSpace(topLevelNote)
	if note != "" && cleaned[0].Note == nil {
		cleaned[0].Note = &note
	}

	return cleaned, money.Amount{Decimal: sum}, nil
}

func normalizeCreateOptions(participants []CreateGroupExpenseParticipantInput, ownerOriginalAmount, totalAmount money.Amount) (CreateOptions, error) {
	if len(participants) == 0 {
		return CreateOptions{}, nil
	}

	cleaned := make([]CreateGroupExpenseParticipant, 0, len(participants))
	requiresShareCalculation := false
	totalOriginal := ownerOriginalAmount.Decimal

	for _, participant := range participants {
		name := strings.TrimSpace(participant.ParticipantName)
		if name == "" {
			return CreateOptions{}, apperrors.New(apperrors.KindValidation, "participant_name is required")
		}
		if !participant.OriginalAmount.GreaterThan(money.Zero().Decimal) {
			return CreateOptions{}, apperrors.New(apperrors.KindValidation, "participant original_amount must be greater than zero")
		}

		cleanedItem := CreateGroupExpenseParticipant{
			ParticipantName: name,
			OriginalAmount:  participant.OriginalAmount,
		}

		if participant.ShareAmount == nil || !participant.ShareAmount.GreaterThan(money.Zero().Decimal) {
			requiresShareCalculation = true
		} else {
			cleanedItem.ShareAmount = *participant.ShareAmount
		}

		totalOriginal = totalOriginal.Add(participant.OriginalAmount.Decimal)
		cleaned = append(cleaned, cleanedItem)
	}

	if requiresShareCalculation {
		if !totalOriginal.GreaterThan(money.Zero().Decimal) {
			return CreateOptions{}, apperrors.New(apperrors.KindValidation, "sum of owner_original_amount and participants original_amount must be greater than zero")
		}

		for i := range cleaned {
			calculated := totalAmount.Decimal.Mul(cleaned[i].OriginalAmount.Decimal).Div(totalOriginal).Round(2)
			if !calculated.GreaterThan(money.Zero().Decimal) {
				return CreateOptions{}, apperrors.New(apperrors.KindValidation, "calculated participant share_amount must be greater than zero")
			}
			cleaned[i].ShareAmount = money.Amount{Decimal: calculated}
		}
	}

	for _, item := range cleaned {
		if !item.ShareAmount.GreaterThan(money.Zero().Decimal) {
			return CreateOptions{}, apperrors.New(apperrors.KindValidation, "participant share_amount must be greater than zero")
		}
	}

	return CreateOptions{GroupParticipants: cleaned}, nil
}

func normalizeUpdateLineItems(kind string, lineItems []UpdateTransactionLineItemInput, topLevelNote *string) ([]UpdateTransactionLineItemInput, error) {
	if kind == "transfer" {
		if len(lineItems) > 0 {
			return nil, apperrors.New(apperrors.KindValidation, "line_items must be empty for transfer")
		}
		return []UpdateTransactionLineItemInput{}, nil
	}

	if len(lineItems) == 0 {
		return nil, apperrors.New(apperrors.KindValidation, "line_items is required and must include at least one category")
	}

	cleaned := make([]UpdateTransactionLineItemInput, 0, len(lineItems))
	for _, lineItem := range lineItems {
		categoryID := normalizeOptionalString(lineItem.CategoryID)
		if categoryID == nil {
			return nil, apperrors.New(apperrors.KindValidation, "line_items.category_id is required")
		}
		if !lineItem.Amount.GreaterThan(money.Zero().Decimal) {
			return nil, apperrors.New(apperrors.KindValidation, "line_items.amount must be greater than zero")
		}
		cleaned = append(cleaned, UpdateTransactionLineItemInput{
			CategoryID: categoryID,
			TagIDs:     normalizeTagIDs(lineItem.TagIDs),
			Amount:     lineItem.Amount,
			Note:       normalizeOptionalString(lineItem.Note),
		})
	}

	if topLevelNote != nil {
		note := strings.TrimSpace(*topLevelNote)
		if note != "" && cleaned[0].Note == nil {
			cleaned[0].Note = &note
		}
	}

	return cleaned, nil
}

func normalizeUpdateGroupParticipants(kind string, participants []UpdateGroupExpenseParticipantInput) ([]UpdateGroupExpenseParticipantInput, error) {
	if kind != "expense" {
		return nil, apperrors.New(apperrors.KindValidation, "group_participants are only supported for expense transactions")
	}

	cleaned := make([]UpdateGroupExpenseParticipantInput, 0, len(participants))
	for _, participant := range participants {
		name := strings.TrimSpace(participant.ParticipantName)
		if name == "" {
			return nil, apperrors.New(apperrors.KindValidation, "participant_name is required")
		}
		if !participant.OriginalAmount.GreaterThan(money.Zero().Decimal) {
			return nil, apperrors.New(apperrors.KindValidation, "participant original_amount must be greater than zero")
		}
		if !participant.ShareAmount.GreaterThan(money.Zero().Decimal) {
			return nil, apperrors.New(apperrors.KindValidation, "participant share_amount must be greater than zero")
		}

		cleaned = append(cleaned, UpdateGroupExpenseParticipantInput{
			ParticipantName: name,
			OriginalAmount:  participant.OriginalAmount,
			ShareAmount:     participant.ShareAmount,
		})
	}

	return cleaned, nil
}

func toTransactionLineItems(items []CreateTransactionLineItem) []TransactionLineItem {
	out := make([]TransactionLineItem, 0, len(items))
	for _, item := range items {
		out = append(out, TransactionLineItem{
			CategoryID: item.CategoryID,
			TagIDs:     item.TagIDs,
			Amount:     item.Amount.String(),
			Note:       item.Note,
		})
	}
	return out
}

func (s *service) List(ctx context.Context, userID string, filter ListFilter) ([]Transaction, int, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "transaction", "operation", "list")
	logger.Info("transaction_list_started", "user_id", userID)

	if strings.TrimSpace(userID) == "" {
		logger.Warn("transaction_list_failed", "reason", "missing user context")
		return nil, 0, apperrors.New(apperrors.KindUnauth, "missing user context")
	}

	normalizedFilter := normalizeListFilter(filter)
	if err := validateListFilter(normalizedFilter); err != nil {
		return nil, 0, err
	}
	items, totalCount, err := s.repo.ListByUser(ctx, userID, normalizedFilter)
	if err != nil {
		logger.Error("transaction_list_failed", "error", err)
		return nil, 0, apperrors.Wrap(apperrors.KindInternal, "failed to list transactions", err)
	}
	logger.Info("transaction_list_succeeded", "count", len(items))
	return items, totalCount, nil
}

func (s *service) Get(ctx context.Context, userID, transactionID string) (*Transaction, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "transaction", "operation", "get")
	logger.Info("transaction_get_started", "user_id", userID, "transaction_id", transactionID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	if strings.TrimSpace(transactionID) == "" {
		return nil, apperrors.New(apperrors.KindValidation, "transaction id is required")
	}

	item, err := s.repo.GetByID(ctx, userID, transactionID)
	if err != nil {
		logger.Error("transaction_get_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to get transaction", err)
	}
	if item == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "transaction not found")
	}

	logger.Info("transaction_get_succeeded", "transaction_id", item.ID)
	return item, nil
}

func (s *service) BatchPatchStatus(ctx context.Context, userID string, req BatchPatchRequest) (*BatchPatchResult, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "transaction", "operation", "batch_patch_status")
	logger.Info("transaction_batch_patch_started", "user_id", userID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	if len(req.TransactionIDs) == 0 {
		return nil, apperrors.New(apperrors.KindValidation, "transaction_ids is required")
	}

	status := strings.ToLower(strings.TrimSpace(req.Patch.Status))
	if status != "pending" && status != "posted" && status != "cancelled" {
		return nil, apperrors.New(apperrors.KindValidation, "unsupported status")
	}

	seen := make(map[string]struct{}, len(req.TransactionIDs))
	cleanIDs := make([]string, 0, len(req.TransactionIDs))
	for _, id := range req.TransactionIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		cleanIDs = append(cleanIDs, trimmed)
	}
	if len(cleanIDs) == 0 {
		return nil, apperrors.New(apperrors.KindValidation, "transaction_ids is required")
	}

	updatedIDs, err := s.repo.BatchPatchStatus(ctx, userID, cleanIDs, status)
	if err != nil {
		logger.Error("transaction_batch_patch_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to patch transaction status", err)
	}

	updatedMap := make(map[string]struct{}, len(updatedIDs))
	for _, id := range updatedIDs {
		updatedMap[id] = struct{}{}
	}
	failedIDs := make([]string, 0)
	for _, id := range cleanIDs {
		if _, ok := updatedMap[id]; !ok {
			failedIDs = append(failedIDs, id)
		}
	}

	result := &BatchPatchResult{
		UpdatedCount: len(updatedIDs),
		FailedCount:  len(failedIDs),
		UpdatedIDs:   updatedIDs,
		FailedIDs:    failedIDs,
	}
	logger.Info("transaction_batch_patch_succeeded", "updated_count", result.UpdatedCount, "failed_count", result.FailedCount)

	return result, nil
}

func (s *service) ListGroupParticipantsByTransaction(ctx context.Context, userID, transactionID string) ([]GroupExpenseParticipant, error) {
	if _, err := s.Get(ctx, userID, transactionID); err != nil {
		return nil, err
	}

	items, err := s.repo.ListGroupParticipantsByTransaction(ctx, userID, transactionID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to list group participants", err)
	}

	return items, nil
}

func normalizeOptionalString(v *string) *string {
	if v == nil {
		return nil
	}
	s := strings.TrimSpace(*v)
	if s == "" {
		return nil
	}
	return &s
}

func normalizeTagIDs(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	items := make([]string, 0, len(values))
	for _, v := range values {
		id := strings.TrimSpace(v)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		items = append(items, id)
	}
	if len(items) == 0 {
		return nil
	}
	return items
}

func passThroughOrWrapInternal(message string, err error) error {
	if err == nil {
		return nil
	}
	var appErr *apperrors.Error
	if errors.As(err, &appErr) {
		return err
	}
	return apperrors.Wrap(apperrors.KindInternal, message, err)
}

func normalizeListFilter(filter ListFilter) ListFilter {
	filter.AccountID = normalizeOptionalString(filter.AccountID)
	filter.Status = normalizeOptionalString(filter.Status)
	filter.Search = normalizeOptionalString(filter.Search)
	filter.Type = normalizeOptionalString(filter.Type)
	filter.ExternalRefFamily = normalizeOptionalString(filter.ExternalRefFamily)

	if filter.Status != nil {
		s := strings.ToLower(*filter.Status)
		filter.Status = &s
	}
	if filter.Type != nil {
		s := strings.ToLower(*filter.Type)
		filter.Type = &s
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 200 {
		filter.Limit = 200
	}

	return filter
}

func validateListFilter(filter ListFilter) error {
	if filter.Status != nil {
		s := *filter.Status
		if s != "pending" && s != "posted" && s != "cancelled" {
			return apperrors.New(apperrors.KindValidation, "status must be one of: pending, posted, cancelled")
		}
	}
	if filter.Type != nil {
		t := *filter.Type
		if t != "expense" && t != "income" && t != "transfer" {
			return apperrors.New(apperrors.KindValidation, "type must be one of: expense, income, transfer")
		}
	}
	if filter.From != nil && filter.To != nil && filter.From.After(*filter.To) {
		return apperrors.New(apperrors.KindValidation, "from must be less than or equal to to")
	}
	return nil
}
