package savings

import "github.com/go-chi/chi/v5"

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/savings/instruments", func(r chi.Router) {
		r.Get("/", h.list)
		r.Post("/", h.create)
		r.Get("/{instrumentId}", h.get)
		r.Patch("/{instrumentId}", h.patch)
		r.Delete("/{instrumentId}", h.delete)
	})
}
