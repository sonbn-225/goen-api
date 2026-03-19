package app

import (
	"context"
	"strings"

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
	var mergedDescription *string
	if req.Description != nil && strings.TrimSpace(*req.Description) != "" {
		v := strings.TrimSpace(*req.Description)
		mergedDescription = &v
	} else if req.Notes != nil && strings.TrimSpace(*req.Notes) != "" {
		v := strings.TrimSpace(*req.Notes)
		mergedDescription = &v
	}

	mapped := transaction.CreateRequest{
		Type:         req.Type,
		OccurredDate: req.OccurredDate,
		OccurredTime: req.OccurredTime,
		Amount:       req.Amount,
		CategoryID:   req.CategoryID,
		Description:  mergedDescription,
		AccountID:    req.AccountID,
	}
	return t.svc.Create(ctx, userID, mapped)
}

func (t *rotTxServiceAdapter) Delete(ctx context.Context, userID, transactionID string) error {
	return t.svc.Delete(ctx, userID, transactionID)
}

type investmentAccountServiceAdapter struct {
	svc *account.Service
}

func (a *investmentAccountServiceAdapter) GetAccountByID(ctx context.Context, userID, accountID string) (*domain.Account, error) {
	return a.svc.Get(ctx, userID, accountID)
}
