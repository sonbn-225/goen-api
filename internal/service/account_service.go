package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

type AccountService struct {
	repo     interfaces.AccountRepository
	userRepo interfaces.UserRepository
}

func NewAccountService(repo interfaces.AccountRepository, userRepo interfaces.UserRepository) *AccountService {
	return &AccountService{repo: repo, userRepo: userRepo}
}

func (s *AccountService) List(ctx context.Context, userID string) ([]entity.Account, error) {
	return s.repo.ListAccountsForUser(ctx, userID)
}

func (s *AccountService) Get(ctx context.Context, userID, accountID string) (*entity.Account, error) {
	return s.repo.GetAccountForUser(ctx, userID, accountID)
}

func (s *AccountService) Create(ctx context.Context, userID string, req dto.CreateAccountRequest) (*entity.Account, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("account name is required")
	}

	accountType := strings.TrimSpace(req.AccountType)
	if !isValidAccountType(accountType) {
		return nil, errors.New("invalid account type")
	}

	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = s.defaultCurrencyForUser(ctx, userID)
	}

	color := utils.NormalizeOptionalString(req.Color)
	parentID := utils.NormalizeOptionalString(req.ParentAccountID)

	// Basic validation for sub-accounts
	if (accountType == "card" || accountType == "savings") && parentID == nil {
		return nil, errors.New("parent account ID is required for card or savings accounts")
	}

	now := time.Now().UTC()
	id := uuid.NewString()

	account := entity.Account{
		ID:              id,
		Name:            name,
		AccountNumber:   utils.NormalizeOptionalString(req.AccountNumber),
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

func (s *AccountService) Patch(ctx context.Context, userID, accountID string, patch entity.AccountPatch) (*entity.Account, error) {
	return s.repo.PatchAccount(ctx, userID, accountID, patch)
}

func (s *AccountService) Delete(ctx context.Context, userID, accountID string) error {
	acc, err := s.repo.GetAccountForUser(ctx, userID, accountID)
	if err != nil {
		return err
	}

	if acc.AccountType == "cash" {
		return errors.New("cash account cannot be deleted; should be closed instead")
	}

	hasTransfers, err := s.repo.HasRelatedTransferTransactionsForAccount(ctx, accountID)
	if err != nil {
		return err
	}
	if hasTransfers {
		return errors.New("account has related transfer transactions and cannot be deleted")
	}

	return s.repo.DeleteAccount(ctx, userID, accountID)
}

func (s *AccountService) ListBalances(ctx context.Context, userID string) ([]entity.AccountBalance, error) {
	return s.repo.ListAccountBalancesForUser(ctx, userID)
}

func (s *AccountService) ListShares(ctx context.Context, userID, accountID string) ([]entity.AccountShare, error) {
	if _, err := s.repo.GetAccountForUser(ctx, userID, accountID); err != nil {
		return nil, err
	}
	return s.repo.ListAccountShares(ctx, userID, accountID)
}

func (s *AccountService) UpsertShare(ctx context.Context, userID, accountID, login, permission string) (*entity.AccountShare, error) {
	if _, err := s.repo.GetAccountForUser(ctx, userID, accountID); err != nil {
		return nil, err
	}

	var target *entity.UserWithPassword
	var err error
	if strings.Contains(login, "@") {
		target, err = s.userRepo.FindUserByEmail(ctx, strings.ToLower(login))
	} else if len(login) > 0 && login[0] != '+' && !strings.ContainsAny(login, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		target, err = s.userRepo.FindUserByPhone(ctx, login)
	} else {
		target, err = s.userRepo.FindUserByUsername(ctx, login)
	}

	if err != nil {
		return nil, err
	}
	if target.ID == userID {
		return nil, errors.New("cannot share with yourself")
	}

	return s.repo.UpsertAccountShare(ctx, userID, accountID, target.ID, permission)
}

func (s *AccountService) RevokeShare(ctx context.Context, userID, accountID, targetUserID string) error {
	if _, err := s.repo.GetAccountForUser(ctx, userID, accountID); err != nil {
		return err
	}
	return s.repo.RevokeAccountShare(ctx, userID, accountID, targetUserID)
}

func (s *AccountService) defaultCurrencyForUser(ctx context.Context, userID string) string {
	u, err := s.userRepo.FindUserByID(ctx, userID)
	if err != nil || u == nil || u.Settings == nil {
		return "VND"
	}
	if settings, ok := u.Settings.(map[string]any); ok {
		if cur, ok := settings["default_currency"].(string); ok {
			return strings.ToUpper(cur)
		}
	}
	return "VND"
}

func isValidAccountType(t string) bool {
	switch t {
	case "bank", "wallet", "cash", "broker", "card", "savings":
		return true
	default:
		return false
	}
}
