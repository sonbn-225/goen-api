package setting

import "github.com/go-chi/chi/v5"

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Patch("/settings/me", h.patchSettings)
}
