package public

import (
	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type Module struct {
	handler *Handler
}

func New(userRepo domain.UserRepository, accountRepo domain.AccountRepository, groupExpenseRepo domain.GroupExpenseRepository) *Module {
	svc := NewService(userRepo, accountRepo, groupExpenseRepo)
	return &Module{
		handler: NewHandler(svc),
	}
}

func (m *Module) RegisterRoutes(r chi.Router) {
	r.Route("/public", func(r chi.Router) {
		r.Get("/u/{userId}/profile", m.handler.GetProfile)
		r.Get("/u/{userId}/payment-info", m.handler.GetPaymentInfo)
		r.Get("/u/{userId}/participants", m.handler.ListParticipants)
		r.Get("/u/{userId}/debts", m.handler.ListDebts)
	})
}
