package v1
 
import (
	"encoding/json"
	"net/http"
 
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
)
 
type InvestmentHandler struct {
	svc interfaces.InvestmentService
}
 
func NewInvestmentHandler(svc interfaces.InvestmentService) *InvestmentHandler {
	return &InvestmentHandler{svc: svc}
}
 
func (h *InvestmentHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))
 
		r.Route("/investments/accounts", func(r chi.Router) {
			r.Get("/", h.ListInvestmentAccounts)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetInvestmentAccount)
				r.Patch("/", h.UpdateInvestmentAccountSettings)
 
				r.Route("/trades", func(r chi.Router) {
					r.Get("/", h.ListTrades)
					r.Post("/", h.CreateTrade)
					r.Patch("/{tradeId}", h.UpdateTrade)
					r.Delete("/{tradeId}", h.DeleteTrade)
				})
 
				r.Get("/holdings", h.ListHoldings)
				r.Get("/reports/realized-pnl", h.GetRealizedPNLReport)
				r.Get("/eligible-actions", h.ListEligibleActions)
				r.Post("/actions/{eventId}/claim", h.ClaimAction)
				r.Post("/backfill-cash", h.BackfillCash)
			})
		})
 
		r.Route("/investments/securities", func(r chi.Router) {
			r.Get("/", h.ListSecurities)
			r.Route("/{securityId}", func(r chi.Router) {
				r.Get("/", h.GetSecurity)
				r.Get("/prices-daily", h.ListSecurityPrices)
				r.Get("/events", h.ListSecurityEvents)
			})
		})
	})
}
 
// ListInvestmentAccounts godoc
// @Summary List Investment Accounts
// @Description Retrieve a list of investment accounts for the current user
// @Tags Investments
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.InvestmentAccountResponse}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/accounts [get]
func (h *InvestmentHandler) ListInvestmentAccounts(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
	accounts, err := h.svc.ListInvestmentAccounts(r.Context(), userID)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, accounts)
}
 
// GetInvestmentAccount godoc
// @Summary Get Investment Account
// @Description Retrieve a specific investment account by ID
// @Tags Investments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Success 200 {object} response.SuccessEnvelope{data=dto.InvestmentAccountResponse}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/accounts/{id} [get]
func (h *InvestmentHandler) GetInvestmentAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
	account, err := h.svc.GetInvestmentAccount(r.Context(), userID, id)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	if account == nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "investment account not found", nil)
		return
	}
	response.WriteSuccess(w, http.StatusOK, account)
}
 
// UpdateInvestmentAccountSettings godoc
// @Summary Update Investment Account Settings
// @Description Partially update settings like automated tracking for an investment account
// @Tags Investments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param request body dto.PatchInvestmentAccountRequest true "Update Payload"
// @Success 200 {object} response.SuccessEnvelope{data=dto.InvestmentAccountResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/accounts/{id} [patch]
func (h *InvestmentHandler) UpdateInvestmentAccountSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
	var req dto.PatchInvestmentAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	account, err := h.svc.UpdateInvestmentAccountSettings(r.Context(), userID, id, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	if account == nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "investment account not found", nil)
		return
	}
	response.WriteSuccess(w, http.StatusOK, account)
}
 
// ListTrades godoc
// @Summary List Trades
// @Description Retrieve trades associated with an investment account
// @Tags Investments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.TradeResponse}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/accounts/{id}/trades [get]
func (h *InvestmentHandler) ListTrades(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
	trades, err := h.svc.ListTrades(r.Context(), userID, id)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, trades)
}
 
// CreateTrade godoc
// @Summary Create Trade
// @Description Create a new trade for an investment account
// @Tags Investments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param request body dto.CreateTradeRequest true "Trade Payload"
// @Success 201 {object} response.SuccessEnvelope{data=dto.TradeResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/accounts/{id}/trades [post]
func (h *InvestmentHandler) CreateTrade(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
	var req dto.CreateTradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	trade, err := h.svc.CreateTrade(r.Context(), userID, id, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusCreated, trade)
}
 
// UpdateTrade godoc
// @Summary Update Trade
// @Description Update an existing trade record
// @Tags Investments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param tradeId path string true "Trade ID"
// @Param request body dto.CreateTradeRequest true "Trade Payload"
// @Success 200 {object} response.SuccessEnvelope{data=dto.TradeResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/accounts/{id}/trades/{tradeId} [patch]
func (h *InvestmentHandler) UpdateTrade(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
	tradeID, err := uuid.Parse(chi.URLParam(r, "tradeId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid trade id format", nil)
		return
	}
	var req dto.CreateTradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
	trade, err := h.svc.UpdateTrade(r.Context(), userID, id, tradeID, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	if trade == nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "trade not found", nil)
		return
	}
	response.WriteSuccess(w, http.StatusOK, trade)
}
 
// DeleteTrade godoc
// @Summary Delete Trade
// @Description Delete a trade from an investment account
// @Tags Investments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param tradeId path string true "Trade ID"
// @Success 200 {object} response.SuccessEnvelope{data=map[string]string}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/accounts/{id}/trades/{tradeId} [delete]
func (h *InvestmentHandler) DeleteTrade(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
	tradeID, err := uuid.Parse(chi.URLParam(r, "tradeId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid trade id format", nil)
		return
	}
	err = h.svc.DeleteTrade(r.Context(), userID, id, tradeID)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "Trade deleted"})
}
 
// ListHoldings godoc
// @Summary List Holdings
// @Description Retrieve current holdings for an investment account
// @Tags Investments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.HoldingResponse}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/accounts/{id}/holdings [get]
func (h *InvestmentHandler) ListHoldings(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
	holdings, err := h.svc.ListHoldings(r.Context(), userID, id)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, holdings)
}
 
// GetRealizedPNLReport godoc
// @Summary Get Realized P&L Report
// @Description Realized profit/loss report for an investment account
// @Tags Investments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Success 200 {object} response.SuccessEnvelope{data=dto.RealizedPNLReport}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/accounts/{id}/reports/realized-pnl [get]
func (h *InvestmentHandler) GetRealizedPNLReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
	report, err := h.svc.GetRealizedPNLReport(r.Context(), userID, id)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, report)
}
 
func (h *InvestmentHandler) ListEligibleActions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
	items, err := h.svc.ListEligibleCorporateActions(r.Context(), userID, id)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, items)
}
 
func (h *InvestmentHandler) ClaimAction(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
	eventID, err := uuid.Parse(chi.URLParam(r, "eventId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid event id format", nil)
		return
	}
 
	var req dto.ClaimCorporateActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", nil)
		return
	}
 
	trade, err := h.svc.ClaimCorporateAction(r.Context(), userID, id, eventID, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, trade)
}
 
func (h *InvestmentHandler) BackfillCash(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid account id format", nil)
		return
	}
	result, err := h.svc.BackfillTradePrincipalTransactions(r.Context(), userID, id)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, result)
}
 
// ListSecurities godoc
// @Summary List Securities
// @Description Retrieve list of all available securities in the system
// @Tags Securities
// @Produce json
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.SecurityResponse}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/securities [get]
func (h *InvestmentHandler) ListSecurities(w http.ResponseWriter, r *http.Request) {
	securities, err := h.svc.ListSecurities(r.Context())
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, securities)
}
 
// GetSecurity godoc
// @Summary Get Security
// @Description Retrieve details of a specific security by ID
// @Tags Securities
// @Produce json
// @Param securityId path string true "Security ID"
// @Success 200 {object} response.SuccessEnvelope{data=dto.SecurityResponse}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/securities/{securityId} [get]
func (h *InvestmentHandler) GetSecurity(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "securityId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid security id format", nil)
		return
	}
	security, err := h.svc.GetSecurity(r.Context(), id)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	if security == nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "security not found", nil)
		return
	}
	response.WriteSuccess(w, http.StatusOK, security)
}
 
// ListSecurityPrices godoc
// @Summary List Security Prices
// @Description Retrieve historical/daily prices for a security between dates
// @Tags Securities
// @Produce json
// @Param securityId path string true "Security ID"
// @Param from query string false "From Date (YYYY-MM-DD)"
// @Param to query string false "To Date (YYYY-MM-DD)"
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.SecurityPriceDailyResponse}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/securities/{securityId}/prices-daily [get]
func (h *InvestmentHandler) ListSecurityPrices(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "securityId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid security id format", nil)
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
 
	prices, err := h.svc.ListSecurityPrices(r.Context(), id, fromPtr, toPtr)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, prices)
}
 
// ListSecurityEvents godoc
// @Summary List Security Events
// @Description Retrieve events (dividends, splits, etc) mapping to a security
// @Tags Securities
// @Produce json
// @Param securityId path string true "Security ID"
// @Param from query string false "From Date (YYYY-MM-DD)"
// @Param to query string false "To Date (YYYY-MM-DD)"
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.SecurityEventResponse}
// @Failure 500 {object} response.ErrorEnvelope
// @Router /investments/securities/{securityId}/events [get]
func (h *InvestmentHandler) ListSecurityEvents(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "securityId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid security id format", nil)
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
 
	events, err := h.svc.ListSecurityEvents(r.Context(), id, fromPtr, toPtr)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, events)
}
