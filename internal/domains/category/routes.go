package category

import "github.com/go-chi/chi/v5"

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/categories", func(r chi.Router) {
		r.Get("/", h.list)
		r.Get("/{categoryId}", h.get)
	})
}
