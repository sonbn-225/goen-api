package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
)

type InvestmentHandler struct {
	svc interfaces.InvestmentService
}

func NewInvestmentHandler(svc interfaces.InvestmentService) *InvestmentHandler {
	return &InvestmentHandler{svc: svc}
}

func (h *InvestmentHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Get("/investment-accounts", h.ListInvestmentAccounts)
	r.Get("/investment-accounts/{id}", h.GetInvestmentAccount)
	r.Patch("/investment-accounts/{id}", h.UpdateInvestmentAccountSettings)

	r.Get("/investment-accounts/{id}/trades", h.ListTrades)
	r.Post("/investment-accounts/{id}/trades", h.CreateTrade)
	r.Put("/investment-accounts/{id}/trades/{tradeId}", h.UpdateTrade)
	r.Delete("/investment-accounts/{id}/trades/{tradeId}", h.DeleteTrade)

	r.Get("/investment-accounts/{id}/holdings", h.ListHoldings)
	r.Get("/investment-accounts/{id}/reports/realized-pnl", h.GetRealizedPNLReport)

	r.Get("/securities", h.ListSecurities)
	r.Get("/securities/{securityId}", h.GetSecurity)
	r.Get("/securities/{securityId}/prices-daily", h.ListSecurityPrices)
	r.Get("/securities/{securityId}/events", h.ListSecurityEvents)
}

func (h *InvestmentHandler) ListInvestmentAccounts(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	accounts, err := h.svc.ListInvestmentAccounts(r.Context(), userID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, accounts)
}

func (h *InvestmentHandler) GetInvestmentAccount(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	account, err := h.svc.GetInvestmentAccount(r.Context(), userID, id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, account)
}

func (h *InvestmentHandler) UpdateInvestmentAccountSettings(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	var req dto.PatchInvestmentAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	account, err := h.svc.UpdateInvestmentAccountSettings(r.Context(), userID, id, req)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, account)
}

func (h *InvestmentHandler) ListTrades(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	trades, err := h.svc.ListTrades(r.Context(), userID, id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, trades)
}

func (h *InvestmentHandler) CreateTrade(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	var req dto.CreateTradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	trade, err := h.svc.CreateTrade(r.Context(), userID, id, req)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusCreated, trade)
}

func (h *InvestmentHandler) UpdateTrade(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	tradeID := chi.URLParam(r, "tradeId")
	var req dto.CreateTradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	trade, err := h.svc.UpdateTrade(r.Context(), userID, id, tradeID, req)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, trade)
}

func (h *InvestmentHandler) DeleteTrade(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	tradeID := chi.URLParam(r, "tradeId")
	err := h.svc.DeleteTrade(r.Context(), userID, id, tradeID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]string{"message": "Trade deleted"})
}

func (h *InvestmentHandler) ListHoldings(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	holdings, err := h.svc.ListHoldings(r.Context(), userID, id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, holdings)
}

func (h *InvestmentHandler) GetRealizedPNLReport(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	report, err := h.svc.GetRealizedPNLReport(r.Context(), userID, id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, report)
}

func (h *InvestmentHandler) ListSecurities(w http.ResponseWriter, r *http.Request) {
	securities, err := h.svc.ListSecurities(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, securities)
}

func (h *InvestmentHandler) GetSecurity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "securityId")
	security, err := h.svc.GetSecurity(r.Context(), id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, security)
}

func (h *InvestmentHandler) ListSecurityPrices(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "securityId")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	var fromPtr, toPtr *string
	if from != "" { fromPtr = &from }
	if to != "" { toPtr = &to }
	
	prices, err := h.svc.ListSecurityPrices(r.Context(), id, fromPtr, toPtr)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, prices)
}

func (h *InvestmentHandler) ListSecurityEvents(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "securityId")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	var fromPtr, toPtr *string
	if from != "" { fromPtr = &from }
	if to != "" { toPtr = &to }

	events, err := h.svc.ListSecurityEvents(r.Context(), id, fromPtr, toPtr)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
		return
	}
	response.WriteJSON(w, http.StatusOK, events)
}
