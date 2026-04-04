package profile

import "github.com/go-chi/chi/v5"

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/profile/me", h.me)
	r.Patch("/profile/me", h.patchProfile)
	r.Post("/profile/me/avatar", h.uploadAvatar)
	r.Post("/profile/me/change-password", h.changePassword)
}
