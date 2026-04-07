package service

import (
	"context"
	"errors"
	"strings"

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

func (s *AccountService) List(ctx context.Context, userID uuid.UUID) ([]dto.AccountResponse, error) {
	items, err := s.repo.ListAccountsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dto.NewAccountResponses(items), nil
}

func (s *AccountService) Get(ctx context.Context, userID, accountID uuid.UUID) (*dto.AccountResponse, error) {
	it, err := s.repo.GetAccountForUser(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, nil
	}
	resp := dto.NewAccountResponse(*it)
	return &resp, nil
}

func (s *AccountService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateAccountRequest) (*dto.AccountResponse, error) {
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
	var parentID *uuid.UUID
	if req.ParentAccountID != nil && *req.ParentAccountID != "" {
		parsed, err := uuid.Parse(*req.ParentAccountID)
		if err != nil {
			return nil, errors.New("invalid parent account ID")
		}
		parentID = &parsed
	}

	// Basic validation for sub-accounts
	if (accountType == "card" || accountType == "savings") && parentID == nil {
		return nil, errors.New("parent account ID is required for card or savings accounts")
	}

	id := utils.NewID()

	account := entity.Account{
		AuditEntity: entity.AuditEntity{
			BaseEntity: entity.BaseEntity{
				ID: id,
			},
		},
		Name:            name,
		AccountNumber:   utils.NormalizeOptionalString(req.AccountNumber),
		Color:           color,
		AccountType:     accountType,
		Currency:        currency,
		ParentAccountID: parentID,
		Status:          "active",
	}

	if err := s.repo.CreateAccountWithOwner(ctx, account, userID); err != nil {
		return nil, err
	}

	resp := dto.NewAccountResponse(account)
	return &resp, nil
}

func (s *AccountService) Patch(ctx context.Context, userID, accountID uuid.UUID, req dto.PatchAccountRequest) (*dto.AccountResponse, error) {
	patch := entity.AccountPatch{
		Name:   req.Name,
		Color:  req.Color,
		Status: req.Status,
	}

	it, err := s.repo.PatchAccount(ctx, userID, accountID, patch)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, nil
	}
	resp := dto.NewAccountResponse(*it)
	return &resp, nil
}

func (s *AccountService) Delete(ctx context.Context, userID, accountID uuid.UUID) error {
	acc, err := s.repo.GetAccountForUser(ctx, userID, accountID)
	if err != nil {
		return err
	}
	if acc == nil {
		return errors.New("account not found")
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

func (s *AccountService) ListShares(ctx context.Context, userID, accountID uuid.UUID) ([]dto.AccountShareResponse, error) {
	if _, err := s.repo.GetAccountForUser(ctx, userID, accountID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListAccountShares(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}
	return dto.NewAccountShareResponses(items), nil
}

func (s *AccountService) UpsertShare(ctx context.Context, userID, accountID uuid.UUID, login, permission string) (*dto.AccountShareResponse, error) {
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

	it, err := s.repo.UpsertAccountShare(ctx, userID, accountID, target.ID, permission)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, nil
	}
	resp := dto.NewAccountShareResponse(*it)
	return &resp, nil
}

func (s *AccountService) RevokeShare(ctx context.Context, userID, accountID, targetUserID uuid.UUID) error {
	if _, err := s.repo.GetAccountForUser(ctx, userID, accountID); err != nil {
		return err
	}
	return s.repo.RevokeAccountShare(ctx, userID, accountID, targetUserID)
}

func (s *AccountService) ListAuditEvents(ctx context.Context, userID, accountID uuid.UUID, limit int) ([]dto.AccountAuditEventResponse, error) {
	if _, err := s.repo.GetAccountForUser(ctx, userID, accountID); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	items, err := s.repo.ListAccountAuditEvents(ctx, userID, accountID, limit)
	if err != nil {
		return nil, err
	}
	return dto.NewAccountAuditEventResponses(items), nil
}

func (s *AccountService) defaultCurrencyForUser(ctx context.Context, userID uuid.UUID) string {
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

