package account

import (
	"context"
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

func (s *service) Create(ctx context.Context, userID string, input CreateInput) (*Account, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "account", "operation", "create")
	logger.Info("account_create_started", "user_id", userID)

	if strings.TrimSpace(userID) == "" {
		logger.Warn("account_create_failed", "reason", "missing user context")
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, apperrors.New(apperrors.KindValidation, "account name is required")
	}

	accountType := strings.ToLower(strings.TrimSpace(input.Type))
	if !isValidAccountType(accountType) {
		return nil, apperrors.New(apperrors.KindValidation, "account_type is invalid")
	}

	currency := strings.ToUpper(strings.TrimSpace(input.Currency))
	if currency == "" {
		currency = s.defaultCurrencyForUser(ctx, userID)
	}
	if len(currency) != 3 {
		return nil, apperrors.New(apperrors.KindValidation, "currency must be ISO4217")
	}

	accountNumber := normalizeOptionalString(input.AccountNumber)
	color, err := normalizeAccountColor(input.Color)
	if err != nil {
		return nil, err
	}
	parentID := normalizeOptionalString(input.ParentAccountID)

	if (accountType == "card" || accountType == "savings") && parentID == nil {
		return nil, apperrors.New(apperrors.KindValidation, "parent_account_id is required")
	}
	if (accountType == "bank" || accountType == "wallet" || accountType == "cash" || accountType == "broker") && parentID != nil {
		return nil, apperrors.New(apperrors.KindValidation, "parent_account_id must be empty")
	}

	if parentID != nil {
		parent, err := s.repo.GetByID(ctx, userID, *parentID)
		if err != nil {
			logger.Error("account_create_failed", "error", err)
			return nil, apperrors.Wrap(apperrors.KindInternal, "failed to validate parent account", err)
		}
		if parent == nil {
			return nil, apperrors.New(apperrors.KindNotFound, "account not found")
		}
		switch accountType {
		case "card":
			if parent.Type != "bank" {
				return nil, apperrors.New(apperrors.KindValidation, "parent account must be bank")
			}
		case "savings":
			if parent.Type != "bank" && parent.Type != "wallet" {
				return nil, apperrors.New(apperrors.KindValidation, "parent account must be bank or wallet")
			}
		}
	}

	now := time.Now().UTC()
	account := &Account{
		ID:              uuid.NewString(),
		UserID:          userID,
		Name:            name,
		Type:            accountType,
		Currency:        currency,
		ParentAccountID: parentID,
		AccountNumber:   accountNumber,
		Color:           color,
		Status:          "active",
		Balance:         money.Zero(),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repo.Create(ctx, account); err != nil {
		logger.Error("account_create_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to create account", err)
	}
	logger.Info("account_create_succeeded", "account_id", account.ID)
	return account, nil
}

func (s *service) List(ctx context.Context, userID string) ([]Account, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "account", "operation", "list")
	logger.Info("account_list_started", "user_id", userID)

	if strings.TrimSpace(userID) == "" {
		logger.Warn("account_list_failed", "reason", "missing user context")
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	items, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		logger.Error("account_list_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to list accounts", err)
	}
	logger.Info("account_list_succeeded", "count", len(items))
	return items, nil
}

func (s *service) Get(ctx context.Context, userID, accountID string) (*Account, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "account", "operation", "get")
	logger.Info("account_get_started", "user_id", userID, "account_id", accountID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	if strings.TrimSpace(accountID) == "" {
		return nil, apperrors.New(apperrors.KindValidation, "account id is required")
	}

	acc, err := s.repo.GetByID(ctx, userID, accountID)
	if err != nil {
		logger.Error("account_get_failed", "error", err)
		return nil, apperrors.Wrap(apperrors.KindInternal, "failed to get account", err)
	}
	if acc == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "account not found")
	}

	logger.Info("account_get_succeeded", "account_id", acc.ID)
	return acc, nil
}

func (s *service) Delete(ctx context.Context, userID, accountID string) error {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "account", "operation", "delete")
	logger.Info("account_delete_started", "user_id", userID, "account_id", accountID)

	if strings.TrimSpace(userID) == "" {
		return apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	if strings.TrimSpace(accountID) == "" {
		return apperrors.New(apperrors.KindValidation, "accountId is required")
	}

	acc, err := s.repo.GetByID(ctx, userID, accountID)
	if err != nil {
		logger.Error("account_delete_failed", "error", err)
		return apperrors.Wrap(apperrors.KindInternal, "failed to get account", err)
	}
	if acc == nil {
		return apperrors.New(apperrors.KindNotFound, "account not found")
	}

	isOwner, err := s.repo.IsOwner(ctx, userID, accountID)
	if err != nil {
		logger.Error("account_delete_failed", "error", err)
		return apperrors.Wrap(apperrors.KindInternal, "failed to check account permission", err)
	}
	if !isOwner {
		return apperrors.New(apperrors.KindForbidden, "forbidden")
	}

	if acc.Type == "cash" {
		return apperrors.New(apperrors.KindValidation, "cash account cannot be deleted; close it to archive")
	}

	hasTransfers, err := s.repo.HasRelatedTransferTransactionsForAccount(ctx, accountID)
	if err != nil {
		logger.Error("account_delete_failed", "error", err)
		return apperrors.Wrap(apperrors.KindInternal, "failed to validate account transactions", err)
	}
	if hasTransfers {
		return apperrors.New(apperrors.KindValidation, "account has related transfer transactions and cannot be deleted; close it to archive")
	}

	deleted, err := s.repo.Delete(ctx, userID, accountID)
	if err != nil {
		logger.Error("account_delete_failed", "error", err)
		return apperrors.Wrap(apperrors.KindInternal, "failed to delete account", err)
	}
	if !deleted {
		return apperrors.New(apperrors.KindNotFound, "account not found")
	}

	logger.Info("account_delete_succeeded", "account_id", accountID)
	return nil
}

func (s *service) defaultCurrencyForUser(ctx context.Context, userID string) string {
	const fallback = "VND"
	currency, err := s.repo.GetDefaultCurrency(ctx, userID)
	if err != nil {
		return fallback
	}
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if len(currency) != 3 {
		return fallback
	}
	return currency
}

func isValidAccountType(t string) bool {
	switch t {
	case "bank", "wallet", "cash", "broker", "card", "savings":
		return true
	default:
		return false
	}
}

func normalizeOptionalString(s string) *string {
	v := strings.TrimSpace(s)
	if v == "" {
		return nil
	}
	return &v
}

func normalizeAccountColor(in string) (*string, error) {
	v := strings.TrimSpace(in)
	if v == "" {
		return nil, nil
	}
	if !isHexColor(v) {
		return nil, apperrors.New(apperrors.KindValidation, "color is invalid")
	}
	lower := strings.ToLower(v)
	return &lower, nil
}

func isHexColor(s string) bool {
	if len(s) == 0 || s[0] != '#' {
		return false
	}
	rest := s[1:]
	if len(rest) != 3 && len(rest) != 6 {
		return false
	}
	for _, c := range rest {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
