package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type AccountRepository interface {
	CreateAccountWithOwner(ctx context.Context, account entity.Account, ownerUserID uuid.UUID) error
	ListAccountsForUser(ctx context.Context, userID uuid.UUID) ([]entity.Account, error)
	GetAccountForUser(ctx context.Context, userID uuid.UUID, accountID uuid.UUID) (*entity.Account, error)
	PatchAccount(ctx context.Context, actorUserID uuid.UUID, accountID uuid.UUID, patch entity.AccountPatch) (*entity.Account, error)
	DeleteAccount(ctx context.Context, actorUserID uuid.UUID, accountID uuid.UUID) error
	HasRelatedTransferTransactionsForAccount(ctx context.Context, accountID uuid.UUID) (bool, error)
	ListAccountBalancesForUser(ctx context.Context, userID uuid.UUID) ([]entity.AccountBalance, error)
	ListAccountShares(ctx context.Context, actorUserID uuid.UUID, accountID uuid.UUID) ([]entity.AccountShare, error)
	UpsertAccountShare(ctx context.Context, actorUserID uuid.UUID, accountID uuid.UUID, targetUserID uuid.UUID, permission string) (*entity.AccountShare, error)
	RevokeAccountShare(ctx context.Context, actorUserID uuid.UUID, accountID uuid.UUID, targetUserID uuid.UUID) error
	ListAccountAuditEvents(ctx context.Context, actorUserID uuid.UUID, accountID uuid.UUID, limit int) ([]entity.AccountAuditEvent, error)
	RecordAccountAuditEvent(ctx context.Context, event entity.AccountAuditEvent) error
}

type AccountService interface {
	List(ctx context.Context, userID uuid.UUID) ([]dto.AccountResponse, error)
	Get(ctx context.Context, userID, accountID uuid.UUID) (*dto.AccountResponse, error)
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateAccountRequest) (*dto.AccountResponse, error)
	Patch(ctx context.Context, userID, accountID uuid.UUID, req dto.PatchAccountRequest) (*dto.AccountResponse, error)
	Delete(ctx context.Context, userID, accountID uuid.UUID) error
	ListShares(ctx context.Context, userID, accountID uuid.UUID) ([]dto.AccountShareResponse, error)
	UpsertShare(ctx context.Context, userID, accountID uuid.UUID, login, permission string) (*dto.AccountShareResponse, error)
	RevokeShare(ctx context.Context, userID, accountID, targetUserID uuid.UUID) error
	ListAuditEvents(ctx context.Context, userID, accountID uuid.UUID, limit int) ([]dto.AccountAuditEventResponse, error)
}

