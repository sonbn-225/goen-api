package auth

import "github.com/go-chi/chi/v5"

func (h *Handler) RegisterPublicRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/signup", h.signup)
		r.Post("/signin", h.signin)
		r.Post("/register", h.register)
		r.Post("/login", h.login)
	})
}

func (h *Handler) RegisterProtectedRoutes(r chi.Router) {
	r.Post("/auth/refresh", h.refresh)
}
