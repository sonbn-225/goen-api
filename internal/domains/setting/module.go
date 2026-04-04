package setting

import "github.com/go-chi/chi/v5"

func NewModule(deps ModuleDeps) *Module {
	h := NewHandler(deps.Service)
	return &Module{Handler: h}
}

func (m *Module) RegisterRoutes(r chi.Router) {
	m.Handler.RegisterRoutes(r)
}
