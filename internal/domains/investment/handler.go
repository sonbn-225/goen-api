package investment

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/httpx"
	"github.com/sonbn-225/goen-api-v2/internal/core/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// listInvestmentAccounts godoc
// @Summary List Investment Accounts
// @Description List investment accounts for current authenticated user.
// @Tags investments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Envelope{data=[]InvestmentAccount,meta=response.Meta}
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investment-accounts [get]
func (h *Handler) listInvestmentAccounts(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	items, err := h.service.ListInvestmentAccounts(r.Context(), userID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteList(w, http.StatusOK, items, response.Meta{Total: len(items)})
}

// getInvestmentAccount godoc
// @Summary Get Investment Account
// @Description Get investment account detail for current authenticated user.
// @Tags investments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param investmentAccountId path string true "Investment Account ID"
// @Success 200 {object} response.Envelope{data=InvestmentAccount}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investment-accounts/{investmentAccountId} [get]
func (h *Handler) getInvestmentAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	investmentAccountID := chi.URLParam(r, "investmentAccountId")
	if investmentAccountID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "investmentAccountId is required"))
		return
	}

	item, err := h.service.GetInvestmentAccount(r.Context(), userID, investmentAccountID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, item)
}

// patchInvestmentAccount godoc
// @Summary Patch Investment Account Settings
// @Description Patch fee/tax settings for an investment account.
// @Tags investments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param investmentAccountId path string true "Investment Account ID"
// @Param request body PatchInvestmentAccountRequest true "Patch request"
// @Success 200 {object} response.Envelope{data=InvestmentAccount}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investment-accounts/{investmentAccountId} [patch]
func (h *Handler) patchInvestmentAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	investmentAccountID := chi.URLParam(r, "investmentAccountId")
	if investmentAccountID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "investmentAccountId is required"))
		return
	}

	var req PatchInvestmentAccountRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	item, err := h.service.UpdateInvestmentAccountSettings(r.Context(), userID, investmentAccountID, UpdateInvestmentAccountSettingsInput{
		FeeSettings: req.FeeSettings,
		TaxSettings: req.TaxSettings,
	})
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, item)
}

// createTrade godoc
// @Summary Create Trade
// @Description Create a trade in an investment account and update holding snapshot.
// @Tags investments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param investmentAccountId path string true "Investment Account ID"
// @Param request body CreateTradeRequest true "Create trade request"
// @Success 201 {object} response.Envelope{data=Trade}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investment-accounts/{investmentAccountId}/trades [post]
func (h *Handler) createTrade(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	investmentAccountID := chi.URLParam(r, "investmentAccountId")
	if investmentAccountID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "investmentAccountId is required"))
		return
	}

	var req CreateTradeRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	created, err := h.service.CreateTrade(r.Context(), userID, investmentAccountID, CreateTradeInput(req))
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusCreated, created)
}

// listTrades godoc
// @Summary List Trades
// @Description List trades for an investment account.
// @Tags investments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param investmentAccountId path string true "Investment Account ID"
// @Success 200 {object} response.Envelope{data=[]Trade,meta=response.Meta}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investment-accounts/{investmentAccountId}/trades [get]
func (h *Handler) listTrades(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	investmentAccountID := chi.URLParam(r, "investmentAccountId")
	if investmentAccountID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "investmentAccountId is required"))
		return
	}

	items, err := h.service.ListTrades(r.Context(), userID, investmentAccountID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteList(w, http.StatusOK, items, response.Meta{Total: len(items)})
}

// listHoldings godoc
// @Summary List Holdings
// @Description List holdings for an investment account.
// @Tags investments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param investmentAccountId path string true "Investment Account ID"
// @Success 200 {object} response.Envelope{data=[]Holding,meta=response.Meta}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investment-accounts/{investmentAccountId}/holdings [get]
func (h *Handler) listHoldings(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	investmentAccountID := chi.URLParam(r, "investmentAccountId")
	if investmentAccountID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "investmentAccountId is required"))
		return
	}

	items, err := h.service.ListHoldings(r.Context(), userID, investmentAccountID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteList(w, http.StatusOK, items, response.Meta{Total: len(items)})
}

// listSecurities godoc
// @Summary List Securities
// @Description List all securities.
// @Tags investments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Envelope{data=[]Security,meta=response.Meta}
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /securities [get]
func (h *Handler) listSecurities(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	items, err := h.service.ListSecurities(r.Context(), userID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteList(w, http.StatusOK, items, response.Meta{Total: len(items)})
}

// getSecurity godoc
// @Summary Get Security
// @Description Get security detail.
// @Tags investments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param securityId path string true "Security ID"
// @Success 200 {object} response.Envelope{data=Security}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /securities/{securityId} [get]
func (h *Handler) getSecurity(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	securityID := chi.URLParam(r, "securityId")
	if securityID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "securityId is required"))
		return
	}

	item, err := h.service.GetSecurity(r.Context(), userID, securityID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, item)
}

// listSecurityPrices godoc
// @Summary List Security Prices Daily
// @Description List daily prices for a security.
// @Tags investments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param securityId path string true "Security ID"
// @Param from query string false "From date (YYYY-MM-DD)"
// @Param to query string false "To date (YYYY-MM-DD)"
// @Success 200 {object} response.Envelope{data=[]SecurityPriceDaily,meta=response.Meta}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /securities/{securityId}/prices-daily [get]
func (h *Handler) listSecurityPrices(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	securityID := chi.URLParam(r, "securityId")
	if securityID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "securityId is required"))
		return
	}

	from, err := optionalDateQuery(r.URL.Query().Get("from"), "from")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	to, err := optionalDateQuery(r.URL.Query().Get("to"), "to")
	if err != nil {
		response.WriteError(w, err)
		return
	}

	items, err := h.service.ListSecurityPrices(r.Context(), userID, securityID, from, to)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteList(w, http.StatusOK, items, response.Meta{Total: len(items)})
}

// listSecurityEvents godoc
// @Summary List Security Events
// @Description List security events.
// @Tags investments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param securityId path string true "Security ID"
// @Param from query string false "From date (YYYY-MM-DD)"
// @Param to query string false "To date (YYYY-MM-DD)"
// @Success 200 {object} response.Envelope{data=[]SecurityEvent,meta=response.Meta}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /securities/{securityId}/events [get]
func (h *Handler) listSecurityEvents(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	securityID := chi.URLParam(r, "securityId")
	if securityID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "securityId is required"))
		return
	}

	from, err := optionalDateQuery(r.URL.Query().Get("from"), "from")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	to, err := optionalDateQuery(r.URL.Query().Get("to"), "to")
	if err != nil {
		response.WriteError(w, err)
		return
	}

	items, err := h.service.ListSecurityEvents(r.Context(), userID, securityID, from, to)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteList(w, http.StatusOK, items, response.Meta{Total: len(items)})
}

func optionalDateQuery(raw, field string) (*string, error) {
	if raw == "" {
		return nil, nil
	}
	if _, err := time.Parse("2006-01-02", raw); err != nil {
		return nil, apperrors.New(apperrors.KindValidation, field+" must be YYYY-MM-DD")
	}
	v := raw
	return &v, nil
}
