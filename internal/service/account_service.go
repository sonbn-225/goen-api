package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

type AccountService struct {
	repo            interfaces.AccountRepository
	userRepo        interfaces.UserRepository
	transactionRepo interfaces.TransactionRepository
	db              *database.Postgres
}

func NewAccountService(
	repo interfaces.AccountRepository,
	userRepo interfaces.UserRepository,
	transactionRepo interfaces.TransactionRepository,
	db *database.Postgres,
) *AccountService {
	return &AccountService{
		repo:            repo,
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		db:              db,
	}
}

func (s *AccountService) List(ctx context.Context, userID uuid.UUID) ([]dto.AccountResponse, error) {
	items, err := s.repo.ListAccountsForUserTx(ctx, nil, userID)
	if err != nil {
		return nil, err
	}
	return dto.NewAccountResponses(items), nil
}

func (s *AccountService) Get(ctx context.Context, userID, accountID uuid.UUID) (*dto.AccountResponse, error) {
	it, err := s.repo.GetAccountForUserTx(ctx, nil, userID, accountID)
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
		AccountType:     entity.AccountType(accountType),
		Currency:        currency,
		ParentAccountID: parentID,
		Status:          entity.AccountStatusActive,
		Settings:        mapAccountSettingsEntity(req.Settings),
	}

	if err := s.repo.CreateAccountWithOwnerTx(ctx, nil, account, userID); err != nil {
		return nil, err
	}

	resp := dto.NewAccountResponse(account)
	return &resp, nil
}

func (s *AccountService) Patch(ctx context.Context, userID, accountID uuid.UUID, req dto.PatchAccountRequest) (*dto.AccountResponse, error) {
	var status *entity.AccountStatus
	if req.Status != nil {
		s := entity.AccountStatus(*req.Status)
		status = &s
	}

	var it *entity.Account
	var err error

	// We use db.WithTx here because PatchAccountTx requires a transaction
	err = s.db.WithTx(ctx, func(tx pgx.Tx) error {
		cur, err := s.repo.GetAccountForUserTx(ctx, tx, userID, accountID)
		if err != nil {
			return err
		}
		if cur == nil {
			return errors.New("account not found")
		}

		patch := entity.AccountPatch{
			Name:     req.Name,
			Status:   status,
			Settings: utils.MapPtr(mapAccountSettingsEntity(req.Settings)),
		}
		
		if req.Settings == nil {
			patch.Settings = nil
		}

		it, err = s.repo.PatchAccountTx(ctx, tx, userID, accountID, patch)
		return err
	})

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
	acc, err := s.repo.GetAccountForUserTx(ctx, nil, userID, accountID)
	if err != nil {
		return err
	}
	if acc == nil {
		return errors.New("account not found")
	}

	if acc.AccountType == entity.AccountTypeCash {
		return errors.New("cash account cannot be deleted; should be closed instead")
	}

	// Khi xóa hẳn tài khoản, chúng ta cũng thực hiện xóa mềm toàn bộ các giao dịch liên quan
	// để tránh tình trạng "thất thoát" số dư (số dư biến mất khỏi tài khoản nguồn nhưng không có đích).
	return s.db.WithTx(ctx, func(tx pgx.Tx) error {
		// 1. Xóa tất cả giao dịch liên quan
		if err := s.transactionRepo.DeleteTransactionsByAccountTx(ctx, tx, userID, accountID); err != nil {
			return err
		}

		// 2. Xóa tài khoản (cập nhật status thành 'deleted' và set deleted_at)
		return s.repo.DeleteAccountTx(ctx, tx, userID, accountID)
	})
}

func (s *AccountService) ListShares(ctx context.Context, userID, accountID uuid.UUID) ([]dto.AccountShareResponse, error) {
	if _, err := s.repo.GetAccountForUserTx(ctx, nil, userID, accountID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListAccountSharesTx(ctx, nil, userID, accountID)
	if err != nil {
		return nil, err
	}
	return dto.NewAccountShareResponses(items), nil
}

func (s *AccountService) UpsertShare(ctx context.Context, userID, accountID uuid.UUID, login, permission string) (*dto.AccountShareResponse, error) {
	var it *entity.AccountShare
	var err error

	err = s.db.WithTx(ctx, func(tx pgx.Tx) error {
		if _, err := s.repo.GetAccountForUserTx(ctx, tx, userID, accountID); err != nil {
			return err
		}

		var target *entity.UserWithPassword
		if strings.Contains(login, "@") {
			target, err = s.userRepo.FindUserByEmail(ctx, strings.ToLower(login))
		} else if len(login) > 0 && login[0] != '+' && !strings.ContainsAny(login, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
			target, err = s.userRepo.FindUserByPhone(ctx, login)
		} else {
			target, err = s.userRepo.FindUserByUsername(ctx, login)
		}

		if err != nil {
			return err
		}
		if target.ID == userID {
			return errors.New("cannot share with yourself")
		}

		it, err = s.repo.UpsertAccountShareTx(ctx, tx, userID, accountID, target.ID, permission)
		return err
	})

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
	return s.db.WithTx(ctx, func(tx pgx.Tx) error {
		if _, err := s.repo.GetAccountForUserTx(ctx, tx, userID, accountID); err != nil {
			return err
		}
		return s.repo.RevokeAccountShareTx(ctx, tx, userID, accountID, targetUserID)
	})
}

func (s *AccountService) ListAuditEvents(ctx context.Context, userID, accountID uuid.UUID, limit int) ([]dto.AccountAuditEventResponse, error) {
	if _, err := s.repo.GetAccountForUserTx(ctx, nil, userID, accountID); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	items, err := s.repo.ListAccountAuditEventsTx(ctx, nil, userID, accountID, limit)
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
	switch entity.AccountType(t) {
	case entity.AccountTypeBank, entity.AccountTypeWallet, entity.AccountTypeCash,
		entity.AccountTypeBroker, entity.AccountTypeCard, entity.AccountTypeSavings:
		return true
	default:
		return false
	}
}

func mapAccountSettingsEntity(it *dto.AccountSettingsRequest) entity.AccountSettings {
	if it == nil {
		return entity.AccountSettings{}
	}
	return entity.AccountSettings{
		Color:      it.Color,
		Investment: mapInvestmentSettingsEntity(it.Investment),
		Savings:    mapSavingsSettingsEntity(it.Savings),
	}
}

func mapInvestmentSettingsEntity(it *dto.InvestmentSettingsRequest) *entity.InvestmentSettings {
	if it == nil {
		return nil
	}
	return &entity.InvestmentSettings{
		FeeSettings: it.FeeSettings,
		TaxSettings: it.TaxSettings,
	}
}

func mapSavingsSettingsEntity(it *dto.SavingsSettingsRequest) *entity.SavingsSettings {
	if it == nil {
		return nil
	}
	autoRenew := false
	if it.AutoRenew != nil {
		autoRenew = *it.AutoRenew
	}
	return &entity.SavingsSettings{
		Principal:    it.Principal,
		InterestRate: it.InterestRate,
		TermMonths:   it.TermMonths,
		StartDate:    it.StartDate,
		MaturityDate: it.MaturityDate,
		AutoRenew:    autoRenew,
	}
}
