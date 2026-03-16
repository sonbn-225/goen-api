package report

import (
	"github.com/sonbn-225/goen-api/internal/domain"
)

type Module struct {
	Service *Service
	Handler *Handler
}

type ModuleDeps struct {
	ReportRepo  domain.ReportRepository
	AccountRepo domain.AccountRepository
}

func NewModule(deps ModuleDeps) *Module {
	svc := NewService(deps.ReportRepo, deps.AccountRepo)
	h := NewHandler(svc)

	return &Module{
		Service: svc,
		Handler: h,
	}
}

