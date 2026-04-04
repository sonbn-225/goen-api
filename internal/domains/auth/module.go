package auth

import "github.com/go-chi/chi/v5"

func NewModule(deps ModuleDeps) *Module {
	svc := deps.Service
	if svc == nil {
		svc = NewService(deps.UserRepo, deps.Hasher, deps.Issuer, deps.AccessTTLMinutes)
	}
	h := NewHandler(svc)
	return &Module{Service: svc, Handler: h}
}

func (m *Module) RegisterPublicRoutes(r chi.Router) {
	m.Handler.RegisterPublicRoutes(r)
}

func (m *Module) RegisterProtectedRoutes(r chi.Router) {
	m.Handler.RegisterProtectedRoutes(r)
}
