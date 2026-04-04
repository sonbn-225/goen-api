package report

import "github.com/go-chi/chi/v5"

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/reports/dashboard", h.getDashboardReport)
}
