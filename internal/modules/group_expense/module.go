// Package group_expense provides group spending functionality (splits + settlements).
package group_expense

import (
	"github.com/sonbn-225/goen-api/internal/domain"
)

type Module struct {
	Service *Service
	Handler *Handler
}

type ModuleDeps struct {
	Repo    domain.GroupExpenseRepository
	TxSvc   TransactionServiceInterface
	DebtSvc DebtServiceInterface
}

func NewModule(deps ModuleDeps) *Module {
	svc := NewService(deps.TxSvc, deps.DebtSvc, deps.Repo)
	h := NewHandler(svc)
	return &Module{Service: svc, Handler: h}
}

