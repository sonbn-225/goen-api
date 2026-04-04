package investment

import "github.com/go-chi/chi/v5"

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/investment-accounts", func(r chi.Router) {
		r.Get("/", h.listInvestmentAccounts)
		r.Get("/{investmentAccountId}", h.getInvestmentAccount)
		r.Patch("/{investmentAccountId}", h.patchInvestmentAccount)
		r.Post("/{investmentAccountId}/trades", h.createTrade)
		r.Get("/{investmentAccountId}/trades", h.listTrades)
		r.Get("/{investmentAccountId}/holdings", h.listHoldings)
	})

	r.Route("/securities", func(r chi.Router) {
		r.Get("/", h.listSecurities)
		r.Get("/{securityId}", h.getSecurity)
		r.Get("/{securityId}/prices-daily", h.listSecurityPrices)
		r.Get("/{securityId}/events", h.listSecurityEvents)
	})
}
