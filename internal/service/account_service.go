package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
	"github.com/sonbn-225/goen-api/internal/pkg/validation"
	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
)

type AccountService struct {
	repo            interfaces.AccountRepository
	userRepo        interfaces.UserRepository
	transactionRepo interfaces.TransactionRepository
	auditSvc        interfaces.AuditService
	db              *database.Postgres
}

func NewAccountService(
	repo interfaces.AccountRepository,
	userRepo interfaces.UserRepository,
	transactionRepo interfaces.TransactionRepository,
	auditSvc interfaces.AuditService,
	db *database.Postgres,
) *AccountService {
	return &AccountService{
		repo:            repo,
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		auditSvc:        auditSvc,
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
		return nil, apperr.BadRequest("missing_name", "account name is required").WithDetail("field", "name")
	}

	accountType := strings.TrimSpace(req.AccountType)
	if !validation.IsValidAccountType(accountType) {
		return nil, apperr.BadRequest("invalid_type", "invalid account type").
			WithDetail("field", "account_type").
			WithDetail("value", accountType)
	}

	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = s.defaultCurrencyForUser(ctx, userID)
	}

	var parentID *uuid.UUID
	if req.ParentAccountID != nil && *req.ParentAccountID != "" {
		parsed, err := uuid.Parse(*req.ParentAccountID)
		if err != nil {
			return nil, apperr.BadRequest("invalid_parent_id", "invalid parent account ID").
				WithDetail("field", "parent_account_id").
				WithDetail("value", *req.ParentAccountID)
		}
		parentID = &parsed
	}

	// Basic validation for sub-accounts
	if (accountType == "card" || accountType == "savings") && parentID == nil {
		return nil, apperr.BadRequest("missing_parent", "parent account ID is required for card or savings accounts").
			WithDetail("field", "parent_account_id").
			WithDetail("account_type", accountType)
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

	_ = s.auditSvc.Record(ctx, nil, userID, &account.ID, entity.ResourceAccount, entity.ActionCreated, account.ID, nil, account)

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
			return apperr.NotFound("account not found").WithDetail("account_id", accountID)
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
		if err != nil {
			return err
		}

		_ = s.auditSvc.Record(ctx, tx, userID, &accountID, entity.ResourceAccount, entity.ActionUpdated, accountID, cur, it)
		return nil
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
		return apperr.NotFound("account not found")
	}

	if acc.AccountType == entity.AccountTypeCash {
		return apperr.BadRequest("cannot_delete_cash", "cash account cannot be deleted; should be closed instead")
	}

	// Khi xóa hẳn tài khoản, chúng ta cũng thực hiện xóa mềm toàn bộ các giao dịch liên quan
	// để tránh tình trạng "thất thoát" số dư (số dư biến mất khỏi tài khoản nguồn nhưng không có đích).
	return s.db.WithTx(ctx, func(tx pgx.Tx) error {
		// 1. Xóa tất cả giao dịch liên quan
		if err := s.transactionRepo.DeleteTransactionsByAccountTx(ctx, tx, userID, accountID); err != nil {
			return err
		}

		// 2. Xóa tài khoản (cập nhật status thành 'deleted' và set deleted_at)
		if err := s.repo.DeleteAccountTx(ctx, tx, userID, accountID); err != nil {
			return err
		}

		return s.auditSvc.Record(ctx, tx, userID, &accountID, entity.ResourceAccount, entity.ActionDeleted, accountID, acc, nil)
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

		target, err := ResolveUserByLoginTx(ctx, tx, s.userRepo, login)
		if err != nil {
			return err
		}
		if target == nil {
			return apperr.NotFound("user not found").WithDetail("login", login)
		}
		if target.ID == userID {
			return apperr.BadRequest("cannot_share_with_self", "cannot share with yourself").WithDetail("user_id", userID)
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


func (s *AccountService) defaultCurrencyForUser(ctx context.Context, userID uuid.UUID) string {
	u, err := s.userRepo.FindUserByIDTx(ctx, nil, userID)
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
