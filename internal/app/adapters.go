package app

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/modules/account"
	rotatingsavings "github.com/sonbn-225/goen-api/internal/modules/rotating_savings"
	"github.com/sonbn-225/goen-api/internal/modules/transaction"
)

type auditServiceAdapter struct {
	repo domain.AuditRepository
}

func (a *auditServiceAdapter) ListAuditEvents(ctx context.Context, userID, accountID string, limit int) ([]domain.AuditEvent, error) {
	return a.repo.ListAuditEventsForAccount(ctx, userID, accountID, limit)
}

type rotTxServiceAdapter struct {
	svc *transaction.Service
}

func (t *rotTxServiceAdapter) Create(ctx context.Context, userID string, req rotatingsavings.TxCreateRequest) (*domain.Transaction, error) {
	mapped := transaction.CreateRequest{
		Type:         req.Type,
		OccurredDate: req.OccurredDate,
		OccurredTime: req.OccurredTime,
		Amount:       req.Amount,
		Description:  req.Description,
		AccountID:    req.AccountID,
		Notes:        req.Notes,
	}
	return t.svc.Create(ctx, userID, mapped)
}

type investmentAccountServiceAdapter struct {
	svc *account.Service
}

func (a *investmentAccountServiceAdapter) GetAccountByID(ctx context.Context, userID, accountID string) (*domain.Account, error) {
	return a.svc.Get(ctx, userID, accountID)
}
