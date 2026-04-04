package rotatingsavings

import "github.com/go-chi/chi/v5"

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/rotating-savings/groups", func(r chi.Router) {
		r.Get("/", h.listGroups)
		r.Post("/", h.createGroup)
		r.Get("/{groupId}", h.getGroup)
		r.Patch("/{groupId}", h.updateGroup)
		r.Delete("/{groupId}", h.deleteGroup)
		r.Get("/{groupId}/contributions", h.listContributions)
		r.Post("/{groupId}/contributions", h.createContribution)
		r.Delete("/{groupId}/contributions/{contributionId}", h.deleteContribution)
	})
}
