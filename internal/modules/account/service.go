package account

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/domain"
)

// CreateAccountRequest contains create account parameters.
type CreateAccountRequest struct {
	Name            string  `json:"name"`
	AccountNumber   *string `json:"account_number,omitempty"`
	Color           *string `json:"color,omitempty"`
	AccountType     string  `json:"account_type"`
	Currency        string  `json:"currency"`
	ParentAccountID *string `json:"parent_account_id,omitempty"`
}

// Service handles account business logic.
type Service struct {
	repo     domain.AccountRepository
	userRepo domain.UserRepository
}

// NewService creates a new account service.
func NewService(repo domain.AccountRepository, userRepo domain.UserRepository) *Service {
	return &Service{repo: repo, userRepo: userRepo}
}

// List returns accounts accessible by the user.
func (s *Service) List(ctx context.Context, userID string) ([]domain.Account, error) {
	return s.repo.ListAccountsForUser(ctx, userID)
}

// Get retrieves a single account.
func (s *Service) Get(ctx context.Context, userID, accountID string) (*domain.Account, error) {
	acc, err := s.repo.GetAccountForUser(ctx, userID, accountID)
	if err != nil {
		if errors.Is(err, apperrors.ErrAccountNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "account not found", err)
		}
		if errors.Is(err, apperrors.ErrAccountForbidden) {
			return nil, apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		return nil, err
	}
	return acc, nil
}

// Create creates a new account.
func (s *Service) Create(ctx context.Context, userID string, req CreateAccountRequest) (*domain.Account, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperrors.Validation("name is required", map[string]any{"field": "name"})
	}

	accountType := strings.TrimSpace(req.AccountType)
	if !isValidAccountType(accountType) {
		return nil, apperrors.Validation("account_type is invalid", map[string]any{"field": "account_type"})
	}

	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = s.defaultCurrencyForUser(ctx, userID)
	}
	if len(currency) != 3 {
		return nil, apperrors.Validation("currency must be ISO4217", map[string]any{"field": "currency"})
	}

	accountNumber := normalizeOptionalString(req.AccountNumber)

	color, err := normalizeAccountColor(req.Color)
	if err != nil {
		return nil, err
	}

	parentID := normalizeOptionalString(req.ParentAccountID)
	if (accountType == "card" || accountType == "savings") && parentID == nil {
		return nil, apperrors.Validation("parent_account_id is required", map[string]any{"field": "parent_account_id"})
	}
	if (accountType == "bank" || accountType == "wallet" || accountType == "cash" || accountType == "broker") && parentID != nil {
		return nil, apperrors.Validation("parent_account_id must be empty", map[string]any{"field": "parent_account_id"})
	}

	if parentID != nil {
		parent, err := s.repo.GetAccountForUser(ctx, userID, *parentID)
		if err != nil {
			if errors.Is(err, apperrors.ErrAccountNotFound) {
				return nil, apperrors.Wrap(apperrors.KindNotFound, "account not found", err)
			}
			if errors.Is(err, apperrors.ErrAccountForbidden) {
				return nil, apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
			}
			return nil, err
		}
		switch accountType {
		case "card":
			if parent.AccountType != "bank" {
				return nil, apperrors.Validation("parent account must be bank", map[string]any{"field": "parent_account_id"})
			}
		case "savings":
			if parent.AccountType != "bank" && parent.AccountType != "wallet" {
				return nil, apperrors.Validation("parent account must be bank or wallet", map[string]any{"field": "parent_account_id"})
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

// Patch updates an account.
func (s *Service) Patch(ctx context.Context, userID, accountID string, patch domain.AccountPatch) (*domain.Account, error) {
	if strings.TrimSpace(accountID) == "" {
		return nil, apperrors.Validation("accountId is required", map[string]any{"field": "accountId"})
	}
	if patch.Name == nil && patch.Status == nil && patch.Color == nil {
		return nil, apperrors.Validation("at least one field is required", nil)
	}
	if patch.Color != nil {
		if _, err := normalizeAccountColor(patch.Color); err != nil {
			return nil, err
		}
	}

	acc, err := s.repo.PatchAccount(ctx, userID, accountID, patch)
	if err != nil {
		if errors.Is(err, apperrors.ErrAccountForbidden) {
			return nil, apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		if errors.Is(err, apperrors.ErrAccountNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "account not found", err)
		}
		return nil, err
	}
	return acc, nil
}

// Delete soft-deletes an account.
func (s *Service) Delete(ctx context.Context, userID, accountID string) error {
	if strings.TrimSpace(accountID) == "" {
		return apperrors.Validation("accountId is required", map[string]any{"field": "accountId"})
	}

	err := s.repo.DeleteAccount(ctx, userID, accountID)
	if err != nil {
		if errors.Is(err, apperrors.ErrAccountForbidden) {
			return apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		if errors.Is(err, apperrors.ErrAccountNotFound) {
			return apperrors.Wrap(apperrors.KindNotFound, "account not found", err)
		}
		return err
	}
	return nil
}

// ListBalances returns account balances.
func (s *Service) ListBalances(ctx context.Context, userID string) ([]domain.AccountBalance, error) {
	return s.repo.ListAccountBalancesForUser(ctx, userID)
}

// ListShares returns shares for an account.
func (s *Service) ListShares(ctx context.Context, userID, accountID string) ([]domain.AccountShare, error) {
	if _, err := s.repo.GetAccountForUser(ctx, userID, accountID); err != nil {
		if errors.Is(err, apperrors.ErrAccountForbidden) {
			return nil, apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		if errors.Is(err, apperrors.ErrAccountNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "account not found", err)
		}
		return nil, err
	}
	return s.repo.ListAccountShares(ctx, userID, accountID)
}

// UpsertShare creates or updates a share.
func (s *Service) UpsertShare(ctx context.Context, userID, accountID, login, permission string) (*domain.AccountShare, error) {
	login = strings.TrimSpace(login)
	permission = strings.TrimSpace(permission)
	if login == "" {
		return nil, apperrors.Validation("login is required", map[string]any{"field": "login"})
	}
	if permission != "viewer" && permission != "editor" {
		return nil, apperrors.Validation("permission must be viewer or editor", map[string]any{"field": "permission"})
	}

	if _, err := s.repo.GetAccountForUser(ctx, userID, accountID); err != nil {
		if errors.Is(err, apperrors.ErrAccountForbidden) {
			return nil, apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		if errors.Is(err, apperrors.ErrAccountNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "account not found", err)
		}
		return nil, err
	}

	var target *domain.UserWithPassword
	var err error
	if strings.Contains(login, "@") {
		target, err = s.userRepo.FindUserByEmail(ctx, strings.ToLower(login))
	} else {
		target, err = s.userRepo.FindUserByPhone(ctx, login)
	}
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return nil, apperrors.Wrap(apperrors.KindValidation, "user not found", err)
		}
		return nil, err
	}
	if target == nil {
		return nil, apperrors.Validation("user not found", nil)
	}
	if target.ID == userID {
		return nil, apperrors.Validation("cannot share with yourself", nil)
	}

	item, err := s.repo.UpsertAccountShare(ctx, userID, accountID, target.ID, permission)
	if err != nil {
		if errors.Is(err, apperrors.ErrAccountForbidden) || errors.Is(err, apperrors.ErrAccountShareForbidden) {
			return nil, apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		if errors.Is(err, apperrors.ErrAccountNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "account not found", err)
		}
		return nil, err
	}
	return item, nil
}

// RevokeShare removes a share.
func (s *Service) RevokeShare(ctx context.Context, userID, accountID, targetUserID string) error {
	if strings.TrimSpace(targetUserID) == "" {
		return apperrors.Validation("userId is required", map[string]any{"field": "userId"})
	}

	if _, err := s.repo.GetAccountForUser(ctx, userID, accountID); err != nil {
		if errors.Is(err, apperrors.ErrAccountForbidden) {
			return apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		if errors.Is(err, apperrors.ErrAccountNotFound) {
			return apperrors.Wrap(apperrors.KindNotFound, "account not found", err)
		}
		return err
	}

	err := s.repo.RevokeAccountShare(ctx, userID, accountID, targetUserID)
	if err != nil {
		if errors.Is(err, apperrors.ErrAccountForbidden) || errors.Is(err, apperrors.ErrAccountShareForbidden) {
			return apperrors.Wrap(apperrors.KindForbidden, "forbidden", err)
		}
		if errors.Is(err, apperrors.ErrAccountNotFound) {
			return apperrors.Wrap(apperrors.KindNotFound, "account not found", err)
		}
		return err
	}
	return nil
}

func (s *Service) defaultCurrencyForUser(ctx context.Context, userID string) string {
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

func toString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func normalizeAccountColor(in *string) (*string, error) {
	v := strings.ToLower(strings.TrimSpace(toString(in)))
	if v == "" {
		return nil, nil
	}
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
		return nil, apperrors.Validation("color is invalid", map[string]any{"field": "color"})
	}
	return &v, nil
}

