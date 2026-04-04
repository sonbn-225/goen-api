package transaction

import "github.com/go-chi/chi/v5"

func NewModule(deps ModuleDeps) *Module {
	svc := deps.Service
	if svc == nil {
		svc = NewService(deps.Repo)
	}
	h := NewHandler(svc)
	return &Module{Service: svc, Handler: h}
}

func (m *Module) RegisterRoutes(r chi.Router) {
	m.Handler.RegisterRoutes(r)
}
