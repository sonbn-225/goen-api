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
}

type AccountService interface {
	List(ctx context.Context, userID string) ([]entity.Account, error)
	Get(ctx context.Context, userID, accountID string) (*entity.Account, error)
	Create(ctx context.Context, userID string, req dto.CreateAccountRequest) (*entity.Account, error)
	Patch(ctx context.Context, userID, accountID string, patch entity.AccountPatch) (*entity.Account, error)
	Delete(ctx context.Context, userID, accountID string) error
	ListBalances(ctx context.Context, userID string) ([]entity.AccountBalance, error)
	ListShares(ctx context.Context, userID, accountID string) ([]entity.AccountShare, error)
	UpsertShare(ctx context.Context, userID, accountID, login, permission string) (*entity.AccountShare, error)
	RevokeShare(ctx context.Context, userID, accountID, targetUserID string) error
}
