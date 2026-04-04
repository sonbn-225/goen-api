package interfaces

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type AccountRepository interface {
	CreateAccountWithOwner(ctx context.Context, account entity.Account, ownerUserID string) error
	ListAccountsForUser(ctx context.Context, userID string) ([]entity.Account, error)
	GetAccountForUser(ctx context.Context, userID string, accountID string) (*entity.Account, error)
	PatchAccount(ctx context.Context, actorUserID string, accountID string, patch entity.AccountPatch) (*entity.Account, error)
	DeleteAccount(ctx context.Context, actorUserID string, accountID string) error
	HasRelatedTransferTransactionsForAccount(ctx context.Context, accountID string) (bool, error)
	ListAccountBalancesForUser(ctx context.Context, userID string) ([]entity.AccountBalance, error)
	ListAccountShares(ctx context.Context, actorUserID string, accountID string) ([]entity.AccountShare, error)
	UpsertAccountShare(ctx context.Context, actorUserID string, accountID string, targetUserID string, permission string) (*entity.AccountShare, error)
	RevokeAccountShare(ctx context.Context, actorUserID string, accountID string, targetUserID string) error
	ListAccountAuditEvents(ctx context.Context, actorUserID string, accountID string, limit int) ([]entity.AccountAuditEvent, error)
}

type AccountService interface {
	List(ctx context.Context, userID string) ([]dto.AccountResponse, error)
	Get(ctx context.Context, userID, accountID string) (*dto.AccountResponse, error)
	Create(ctx context.Context, userID string, req dto.CreateAccountRequest) (*dto.AccountResponse, error)
	Patch(ctx context.Context, userID, accountID string, req dto.PatchAccountRequest) (*dto.AccountResponse, error)
	Delete(ctx context.Context, userID, accountID string) error
	ListShares(ctx context.Context, userID, accountID string) ([]dto.AccountShareResponse, error)
	UpsertShare(ctx context.Context, userID, accountID, login, permission string) (*dto.AccountShareResponse, error)
	RevokeShare(ctx context.Context, userID, accountID, targetUserID string) error
	ListAuditEvents(ctx context.Context, userID, accountID string, limit int) ([]dto.AccountAuditEventResponse, error)
}
