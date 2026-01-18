// Package group_expense provides group spending functionality (splits + settlements).
package group_expense

import (
	"context"

	"github.com/sonbn-225/goen-api/internal/domain"
)

type Module struct {
	Service *Service
	Handler *Handler
}

type TransactionServiceInterface interface {
	Get(ctx context.Context, userID, transactionID string) (*domain.Transaction, error)
}

type ModuleDeps struct {
	Repo  domain.GroupExpenseRepository
	TxSvc TransactionServiceInterface
}

func NewModule(deps ModuleDeps) *Module {
	svc := NewService(deps.TxSvc, deps.Repo)
	h := NewHandler(svc)
	return &Module{Service: svc, Handler: h}
}
