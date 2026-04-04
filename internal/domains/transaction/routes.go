package transaction

import "github.com/go-chi/chi/v5"

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/transactions", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Patch("/batch", h.batchPatchStatus)
		r.Get("/{transactionId}", h.get)
		r.Patch("/{transactionId}", h.update)
		r.Get("/{transactionId}/group-expense-participants", h.listGroupParticipants)
	})
}
