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

	// UC-007 Shared Account
	ListAccountShares(ctx context.Context, userID string, accountID string) ([]domain.AccountShare, error)
	UpsertAccountShare(ctx context.Context, userID string, accountID string, login string, permission string) (*domain.AccountShare, error)
	RevokeAccountShare(ctx context.Context, userID string, accountID string, targetUserID string) error
}

type CreateAccountRequest struct {
	Name            string  `json:"name"`
	AccountType     string  `json:"account_type"`
	Currency        string  `json:"currency"`
	ParentAccountID *string `json:"parent_account_id,omitempty"`
}

type accountService struct {
	repo domain.AccountRepository
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
		return nil, errors.New("name is required")
	}

	accountType := strings.TrimSpace(req.AccountType)
	if !isValidAccountType(accountType) {
		return nil, errors.New("account_type is invalid")
	}

	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if len(currency) != 3 {
		return nil, errors.New("currency must be ISO4217")
	}

	parentID := normalizeOptionalString(req.ParentAccountID)
	if (accountType == "card" || accountType == "savings") && parentID == nil {
		return nil, errors.New("parent_account_id is required")
	}
	if (accountType == "bank" || accountType == "wallet" || accountType == "cash" || accountType == "broker") && parentID != nil {
		return nil, errors.New("parent_account_id must be empty")
	}

	// Note: business rule validation about parent type (card->bank, savings->bank|wallet)
	// requires reading parent account. MVP: enforce parent exists and is accessible, and enforce type.
	if parentID != nil {
		parent, err := s.repo.GetAccountForUser(ctx, userID, *parentID)
		if err != nil {
			return nil, err
		}
		switch accountType {
		case "card":
			if parent.AccountType != "bank" {
				return nil, errors.New("parent account must be bank")
			}
		case "savings":
			if parent.AccountType != "bank" && parent.AccountType != "wallet" {
				return nil, errors.New("parent account must be bank or wallet")
			}
		}
	}

	now := time.Now().UTC()
	id := uuid.NewString()

	account := domain.Account{
		ID:              id,
		Name:            name,
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

func (s *accountService) ListAccountShares(ctx context.Context, userID string, accountID string) ([]domain.AccountShare, error) {
	// Ensure account exists & user can access it at all (avoid leaking)
	if _, err := s.repo.GetAccountForUser(ctx, userID, accountID); err != nil {
		return nil, err
	}
	return s.repo.ListAccountShares(ctx, userID, accountID)
}

func (s *accountService) UpsertAccountShare(ctx context.Context, userID string, accountID string, login string, permission string) (*domain.AccountShare, error) {
	login = strings.TrimSpace(login)
	permission = strings.TrimSpace(permission)
	if login == "" {
		return nil, domain.ErrAccountShareInvalidInput
	}
	if permission != "viewer" && permission != "editor" {
		return nil, domain.ErrAccountShareInvalidInput
	}

	// Ensure account exists & user can access it at all (avoid leaking)
	if _, err := s.repo.GetAccountForUser(ctx, userID, accountID); err != nil {
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
		return nil, err
	}
	if target == nil {
		return nil, domain.ErrUserNotFound
	}
	if target.ID == userID {
		return nil, domain.ErrAccountShareInvalidInput
	}

	return s.repo.UpsertAccountShare(ctx, userID, accountID, target.ID, permission)
}

func (s *accountService) RevokeAccountShare(ctx context.Context, userID string, accountID string, targetUserID string) error {
	if strings.TrimSpace(targetUserID) == "" {
		return domain.ErrAccountShareInvalidInput
	}

	// Ensure account exists & user can access it at all (avoid leaking)
	if _, err := s.repo.GetAccountForUser(ctx, userID, accountID); err != nil {
		return err
	}

	return s.repo.RevokeAccountShare(ctx, userID, accountID, targetUserID)
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
