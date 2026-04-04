package media

import "github.com/go-chi/chi/v5"

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/media/{bucket}/*", h.getMedia)
}
