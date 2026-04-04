package debt

import "github.com/go-chi/chi/v5"

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/debts", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{debtId}", h.get)
		r.Get("/{debtId}/payments", h.listPayments)
		r.Post("/{debtId}/payments", h.createPayment)
		r.Get("/{debtId}/installments", h.listInstallments)
		r.Post("/{debtId}/installments", h.createInstallment)
	})
	r.Get("/transactions/{transactionId}/debt-links", h.listDebtLinksForTransaction)
}
