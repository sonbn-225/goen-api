package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type AccountService interface {
	CreateAccount(ctx context.Context, userID string, req CreateAccountRequest) (*domain.Account, error)
	ListAccounts(ctx context.Context, userID string) ([]domain.Account, error)
	GetAccount(ctx context.Context, userID string, accountID string) (*domain.Account, error)
	PatchAccount(ctx context.Context, userID string, accountID string, patch domain.AccountPatch) (*domain.Account, error)
	DeleteAccount(ctx context.Context, userID string, accountID string) error
	ListAccountBalances(ctx context.Context, userID string) ([]domain.AccountBalance, error)

	// UC-007 Shared Account
	ListAccountShares(ctx context.Context, userID string, accountID string) ([]domain.AccountShare, error)
	UpsertAccountShare(ctx context.Context, userID string, accountID string, login string, permission string) (*domain.AccountShare, error)
	RevokeAccountShare(ctx context.Context, userID string, accountID string, targetUserID string) error
}

type CreateAccountRequest struct {
	Name            string  `json:"name"`
	AccountNumber   *string `json:"account_number,omitempty"`
	Color           *string `json:"color,omitempty"`
	AccountType     string  `json:"account_type"`
	Currency        string  `json:"currency"`
	ParentAccountID *string `json:"parent_account_id,omitempty"`
}

func normalizeAccountColor(in *string) (*string, error) {
	v := strings.ToLower(strings.TrimSpace(toString(in)))
	if v == "" {
		return nil, nil
	}
	// Keep server-side allowlist in sync with the web predefined list.
	allowed := map[string]struct{}{
		"gray":   {},
		"red":    {},
		"pink":   {},
		"grape":  {},
		"violet": {},
		"indigo": {},
		"blue":   {},
		"cyan":   {},
		"teal":   {},
		"green":  {},
		"lime":   {},
		"yellow": {},
		"orange": {},
	}
	if _, ok := allowed[v]; !ok {
		return nil, ValidationError("color is invalid", map[string]any{"field": "color"})
	}
	return &v, nil
}

func toString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

type accountService struct {
	repo     domain.AccountRepository
	userRepo domain.UserRepository
}

func NewAccountService(repo domain.AccountRepository, userRepo domain.UserRepository) AccountService {
	return &accountService{repo: repo, userRepo: userRepo}
}

func (s *accountService) ListAccounts(ctx context.Context, userID string) ([]domain.Account, error) {
	return s.repo.ListAccountsForUser(ctx, userID)
}

func (s *accountService) GetAccount(ctx context.Context, userID string, accountID string) (*domain.Account, error) {
	return s.repo.GetAccountForUser(ctx, userID, accountID)
}

func (s *accountService) CreateAccount(ctx context.Context, userID string, req CreateAccountRequest) (*domain.Account, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ValidationError("name is required", map[string]any{"field": "name"})
	}

	accountType := strings.TrimSpace(req.AccountType)
	if !isValidAccountType(accountType) {
		return nil, ValidationError("account_type is invalid", map[string]any{"field": "account_type"})
	}

	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = s.defaultCurrencyForUser(ctx, userID)
	}
	if len(currency) != 3 {
		return nil, ValidationError("currency must be ISO4217", map[string]any{"field": "currency"})
	}

	accountNumber := normalizeOptionalString(req.AccountNumber)

	color, err := normalizeAccountColor(req.Color)
	if err != nil {
		return nil, err
	}

	parentID := normalizeOptionalString(req.ParentAccountID)
	if (accountType == "card" || accountType == "savings") && parentID == nil {
		return nil, ValidationError("parent_account_id is required", map[string]any{"field": "parent_account_id"})
	}
	if (accountType == "bank" || accountType == "wallet" || accountType == "cash" || accountType == "broker") && parentID != nil {
		return nil, ValidationError("parent_account_id must be empty", map[string]any{"field": "parent_account_id"})
	}

	// Note: business rule validation about parent type (card->bank, savings->bank|wallet)
	// requires reading parent account. MVP: enforce parent exists and is accessible, and enforce type.
	if parentID != nil {
		parent, err := s.repo.GetAccountForUser(ctx, userID, *parentID)
		if err != nil {
			if errors.Is(err, domain.ErrAccountNotFound) {
				return nil, NotFoundErrorWithCause("account not found", nil, err)
			}
			if errors.Is(err, domain.ErrAccountForbidden) {
				return nil, ForbiddenErrorWithCause("forbidden", nil, err)
			}
			return nil, err
		}
		switch accountType {
		case "card":
			if parent.AccountType != "bank" {
				return nil, ValidationError("parent account must be bank", map[string]any{"field": "parent_account_id"})
			}
		case "savings":
			if parent.AccountType != "bank" && parent.AccountType != "wallet" {
				return nil, ValidationError("parent account must be bank or wallet", map[string]any{"field": "parent_account_id"})
			}
		}
	}

	now := time.Now().UTC()
	id := uuid.NewString()

	account := domain.Account{
		ID:              id,
		Name:            name,
		AccountNumber:   accountNumber,
		Color:           color,
		AccountType:     accountType,
		Currency:        currency,
		ParentAccountID: parentID,
		Status:          "active",
		CreatedAt:       now,
		UpdatedAt:       now,
		CreatedBy:       &userID,
		UpdatedBy:       &userID,
	}

	if err := s.repo.CreateAccountWithOwner(ctx, account, userID); err != nil {
		return nil, err
	}

	return &account, nil
}

func (s *accountService) defaultCurrencyForUser(ctx context.Context, userID string) string {
	// Fallbacks are intentionally conservative: always return a 3-letter code.
	const fallback = "VND"
	if s.userRepo == nil {
		return fallback
	}
	u, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil || u == nil || u.Settings == nil {
		return fallback
	}
	settings, ok := u.Settings.(map[string]any)
	if !ok {
		return fallback
	}
	v, ok := settings["default_currency"]
	if !ok {
		return fallback
	}
	cur, ok := v.(string)
	if !ok {
		return fallback
	}
	cur = strings.ToUpper(strings.TrimSpace(cur))
	if len(cur) != 3 {
		return fallback
	}
	return cur
}

func (s *accountService) PatchAccount(ctx context.Context, userID string, accountID string, patch domain.AccountPatch) (*domain.Account, error) {
	if strings.TrimSpace(accountID) == "" {
		return nil, ValidationErrorWithCause("invalid account input", nil, domain.ErrAccountInvalidInput)
	}
	if patch.Name == nil && patch.Status == nil && patch.Color == nil {
		return nil, ValidationErrorWithCause("invalid account input", nil, domain.ErrAccountInvalidInput)
	}
	if patch.Color != nil {
		if _, err := normalizeAccountColor(patch.Color); err != nil {
			return nil, ValidationErrorWithCause("invalid account input", nil, domain.ErrAccountInvalidInput)
		}
	}
	acc, err := s.repo.PatchAccount(ctx, userID, accountID, patch)
	if err != nil {
		if errors.Is(err, domain.ErrAccountForbidden) {
			return nil, ForbiddenErrorWithCause("forbidden", nil, err)
		}
		if errors.Is(err, domain.ErrAccountNotFound) {
			return nil, NotFoundErrorWithCause("account not found", nil, err)
		}
		return nil, err
	}
	return acc, nil
}

func (s *accountService) DeleteAccount(ctx context.Context, userID string, accountID string) error {
	if strings.TrimSpace(accountID) == "" {
		return ValidationErrorWithCause("invalid account input", nil, domain.ErrAccountInvalidInput)
	}
	err := s.repo.DeleteAccount(ctx, userID, accountID)
	if err != nil {
		if errors.Is(err, domain.ErrAccountForbidden) {
			return ForbiddenErrorWithCause("forbidden", nil, err)
		}
		if errors.Is(err, domain.ErrAccountNotFound) {
			return NotFoundErrorWithCause("account not found", nil, err)
		}
		return err
	}
	return nil
}

func (s *accountService) ListAccountBalances(ctx context.Context, userID string) ([]domain.AccountBalance, error) {
	return s.repo.ListAccountBalancesForUser(ctx, userID)
}

func (s *accountService) ListAccountShares(ctx context.Context, userID string, accountID string) ([]domain.AccountShare, error) {
	// Ensure account exists & user can access it at all (avoid leaking)
	if _, err := s.repo.GetAccountForUser(ctx, userID, accountID); err != nil {
		if errors.Is(err, domain.ErrAccountForbidden) {
			return nil, ForbiddenErrorWithCause("forbidden", nil, err)
		}
		if errors.Is(err, domain.ErrAccountNotFound) {
			return nil, NotFoundErrorWithCause("account not found", nil, err)
		}
		return nil, err
	}
	return s.repo.ListAccountShares(ctx, userID, accountID)
}

func (s *accountService) UpsertAccountShare(ctx context.Context, userID string, accountID string, login string, permission string) (*domain.AccountShare, error) {
	login = strings.TrimSpace(login)
	permission = strings.TrimSpace(permission)
	if login == "" {
		return nil, InvalidRequestErrorWithCause("invalid request", nil, domain.ErrAccountShareInvalidInput)
	}
	if permission != "viewer" && permission != "editor" {
		return nil, InvalidRequestErrorWithCause("invalid request", nil, domain.ErrAccountShareInvalidInput)
	}

	// Ensure account exists & user can access it at all (avoid leaking)
	if _, err := s.repo.GetAccountForUser(ctx, userID, accountID); err != nil {
		if errors.Is(err, domain.ErrAccountForbidden) {
			return nil, ForbiddenErrorWithCause("forbidden", nil, err)
		}
		if errors.Is(err, domain.ErrAccountNotFound) {
			return nil, NotFoundErrorWithCause("account not found", nil, err)
		}
		return nil, err
	}

	// Lookup target user
	var target *domain.UserWithPassword
	var err error
	if strings.Contains(login, "@") {
		target, err = s.userRepo.FindUserByEmail(ctx, strings.ToLower(login))
	} else {
		target, err = s.userRepo.FindUserByPhone(ctx, login)
	}
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, InvalidRequestErrorWithCause("user not found", nil, err)
		}
		return nil, err
	}
	if target == nil {
		return nil, InvalidRequestErrorWithCause("user not found", nil, domain.ErrUserNotFound)
	}
	if target.ID == userID {
		return nil, InvalidRequestErrorWithCause("invalid request", nil, domain.ErrAccountShareInvalidInput)
	}

	item, err := s.repo.UpsertAccountShare(ctx, userID, accountID, target.ID, permission)
	if err != nil {
		if errors.Is(err, domain.ErrAccountForbidden) || errors.Is(err, domain.ErrAccountShareForbidden) {
			return nil, ForbiddenErrorWithCause("forbidden", nil, err)
		}
		if errors.Is(err, domain.ErrAccountNotFound) {
			return nil, NotFoundErrorWithCause("account not found", nil, err)
		}
		return nil, err
	}
	return item, nil
}

func (s *accountService) RevokeAccountShare(ctx context.Context, userID string, accountID string, targetUserID string) error {
	if strings.TrimSpace(targetUserID) == "" {
		return InvalidRequestErrorWithCause("invalid request", nil, domain.ErrAccountShareInvalidInput)
	}

	// Ensure account exists & user can access it at all (avoid leaking)
	if _, err := s.repo.GetAccountForUser(ctx, userID, accountID); err != nil {
		if errors.Is(err, domain.ErrAccountForbidden) {
			return ForbiddenErrorWithCause("forbidden", nil, err)
		}
		if errors.Is(err, domain.ErrAccountNotFound) {
			return NotFoundErrorWithCause("account not found", nil, err)
		}
		return err
	}

	err := s.repo.RevokeAccountShare(ctx, userID, accountID, targetUserID)
	if err != nil {
		if errors.Is(err, domain.ErrAccountForbidden) || errors.Is(err, domain.ErrAccountShareForbidden) {
			return ForbiddenErrorWithCause("forbidden", nil, err)
		}
		if errors.Is(err, domain.ErrAccountNotFound) {
			return NotFoundErrorWithCause("account not found", nil, err)
		}
		return err
	}
	return nil
}

func isValidAccountType(t string) bool {
	switch t {
	case "bank", "wallet", "cash", "broker", "card", "savings":
		return true
	default:
		return false
	}
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
