package investment

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/httpapi"
	"github.com/sonbn-225/goen-api/internal/response"
)

// Handler handles HTTP requests for investment operations.
type Handler struct {
	svc *Service
}

// NewHandler creates a new investment handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers all investment routes on the given router.
// This allows the module to self-register its routes.
func (h *Handler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.With(authMiddleware).Get("/investment-accounts", h.ListInvestmentAccounts)
	r.With(authMiddleware).Get("/investment-accounts/{investmentAccountId}", h.GetInvestmentAccount)
	r.With(authMiddleware).Get("/investment-accounts/{investmentAccountId}/trades", h.ListTrades)
	r.With(authMiddleware).Get("/investment-accounts/{investmentAccountId}/holdings", h.ListHoldings)
	r.With(authMiddleware).Get("/securities", h.ListSecurities)
	r.With(authMiddleware).Get("/securities/{securityId}", h.GetSecurity)
	r.With(authMiddleware).Get("/securities/{securityId}/prices-daily", h.ListSecurityPrices)
	r.With(authMiddleware).Get("/securities/{securityId}/events", h.ListSecurityEvents)
}

// ListInvestmentAccounts handles GET /investment-accounts
func (h *Handler) ListInvestmentAccounts(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	accounts, err := h.svc.ListInvestmentAccounts(r.Context(), userID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, accounts)
}

// GetInvestmentAccount handles GET /investment-accounts/{investmentAccountId}
func (h *Handler) GetInvestmentAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	id := chi.URLParam(r, "investmentAccountId")
	if id == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "investmentAccountId is required", nil)
		return
	}

	account, err := h.svc.GetInvestmentAccount(r.Context(), userID, id)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	if account == nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "investment account not found", nil)
		return
	}

	response.WriteJSON(w, http.StatusOK, account)
}

// ListSecurities handles GET /securities
func (h *Handler) ListSecurities(w http.ResponseWriter, r *http.Request) {
	_, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	securities, err := h.svc.ListSecurities(r.Context())
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, securities)
}

// GetSecurity handles GET /securities/{securityId}
func (h *Handler) GetSecurity(w http.ResponseWriter, r *http.Request) {
	_, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	id := chi.URLParam(r, "securityId")
	if id == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "securityId is required", nil)
		return
	}

	security, err := h.svc.GetSecurity(r.Context(), id)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	if security == nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "security not found", nil)
		return
	}

	response.WriteJSON(w, http.StatusOK, security)
}

// ListTrades handles GET /investment-accounts/{investmentAccountId}/trades
func (h *Handler) ListTrades(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	brokerAccountID := chi.URLParam(r, "investmentAccountId")
	if brokerAccountID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "investmentAccountId is required", nil)
		return
	}

	trades, err := h.svc.ListTrades(r.Context(), userID, brokerAccountID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, trades)
}

// ListHoldings handles GET /investment-accounts/{investmentAccountId}/holdings
func (h *Handler) ListHoldings(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	brokerAccountID := chi.URLParam(r, "investmentAccountId")
	if brokerAccountID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "investmentAccountId is required", nil)
		return
	}

	holdings, err := h.svc.ListHoldings(r.Context(), userID, brokerAccountID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, holdings)
}

// ListSecurityPrices handles GET /securities/{securityId}/prices-daily
func (h *Handler) ListSecurityPrices(w http.ResponseWriter, r *http.Request) {
	_, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	securityID := chi.URLParam(r, "securityId")
	if securityID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "securityId is required", nil)
		return
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	var fromPtr, toPtr *string
	if from != "" {
		fromPtr = &from
	}
	if to != "" {
		toPtr = &to
	}

	prices, err := h.svc.ListSecurityPrices(r.Context(), securityID, fromPtr, toPtr)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, prices)
}

// ListSecurityEvents handles GET /securities/{securityId}/events
func (h *Handler) ListSecurityEvents(w http.ResponseWriter, r *http.Request) {
	_, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	securityID := chi.URLParam(r, "securityId")
	if securityID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "securityId is required", nil)
		return
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	var fromPtr, toPtr *string
	if from != "" {
		fromPtr = &from
	}
	if to != "" {
		toPtr = &to
	}

	events, err := h.svc.ListSecurityEvents(r.Context(), securityID, fromPtr, toPtr)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, events)
}
