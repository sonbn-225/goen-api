package contact

import "github.com/go-chi/chi/v5"

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/contacts", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{contactId}", h.get)
		r.Patch("/{contactId}", h.patch)
		r.Delete("/{contactId}", h.delete)
	})
}
