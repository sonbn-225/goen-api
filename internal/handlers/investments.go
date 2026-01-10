package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/services"
)

// ListInvestmentAccounts godoc
// @Summary List investment accounts
// @Description List investment accounts accessible by current user.
// @Tags investments
// @Produce json
// @Success 200 {array} domain.InvestmentAccount
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /investment-accounts [get]
func ListInvestmentAccounts(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		items, err := d.InvestmentService.ListInvestmentAccounts(r.Context(), uid)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// CreateInvestmentAccount godoc
// @Summary Create investment account
// @Description Create a 1-1 investment extension for a broker account.
// @Tags investments
// @Accept json
// @Produce json
// @Param body body services.CreateInvestmentAccountRequest true "Create investment account request"
// @Success 200 {object} domain.InvestmentAccount
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 403 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /investment-accounts [post]
func CreateInvestmentAccount(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		var req services.CreateInvestmentAccountRequest
		if ok := decodeJSON(w, r, &req); !ok {
			return
		}

		item, err := d.InvestmentService.CreateInvestmentAccount(r.Context(), uid, req)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, item)
	}
}

// GetInvestmentAccount godoc
// @Summary Get investment account
// @Description Get a single investment account.
// @Tags investments
// @Produce json
// @Param investmentAccountId path string true "Investment account ID"
// @Success 200 {object} domain.InvestmentAccount
// @Failure 401 {object} apierror.Envelope
// @Failure 403 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /investment-accounts/{investmentAccountId} [get]
func GetInvestmentAccount(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		id := chi.URLParam(r, "investmentAccountId")
		if id == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "investmentAccountId is required", map[string]any{"field": "investmentAccountId"})
			return
		}

		item, err := d.InvestmentService.GetInvestmentAccount(r.Context(), uid, id)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, item)
	}
}

// ListSecurities godoc
// @Summary List securities
// @Description List securities from global catalog.
// @Tags investments
// @Produce json
// @Success 200 {array} domain.Security
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /securities [get]
func ListSecurities(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		items, err := d.InvestmentService.ListSecurities(r.Context(), uid)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// GetSecurity godoc
// @Summary Get security
// @Description Get a single security.
// @Tags investments
// @Produce json
// @Param securityId path string true "Security ID"
// @Success 200 {object} domain.Security
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /securities/{securityId} [get]
func GetSecurity(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		id := chi.URLParam(r, "securityId")
		if id == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "securityId is required", map[string]any{"field": "securityId"})
			return
		}

		item, err := d.InvestmentService.GetSecurity(r.Context(), uid, id)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, item)
	}
}

// ListSecurityPricesDaily godoc
// @Summary List daily prices for security
// @Description Read-only daily prices (populated by market data service).
// @Tags investments
// @Produce json
// @Param securityId path string true "Security ID"
// @Param from query string false "From date (YYYY-MM-DD)"
// @Param to query string false "To date (YYYY-MM-DD)"
// @Success 200 {array} domain.SecurityPriceDaily
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /securities/{securityId}/prices-daily [get]
func ListSecurityPricesDaily(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		securityID := chi.URLParam(r, "securityId")
		if securityID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "securityId is required", map[string]any{"field": "securityId"})
			return
		}

		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")

		var fromPtr *string
		if from != "" {
			fromPtr = &from
		}
		var toPtr *string
		if to != "" {
			toPtr = &to
		}

		items, err := d.InvestmentService.ListSecurityPrices(r.Context(), uid, securityID, fromPtr, toPtr)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// ListSecurityEvents godoc
// @Summary List security events
// @Description Read-only security events (populated by market data service).
// @Tags investments
// @Produce json
// @Param securityId path string true "Security ID"
// @Param from query string false "From date (YYYY-MM-DD)"
// @Param to query string false "To date (YYYY-MM-DD)"
// @Success 200 {array} domain.SecurityEvent
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /securities/{securityId}/events [get]
func ListSecurityEvents(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		securityID := chi.URLParam(r, "securityId")
		if securityID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "securityId is required", map[string]any{"field": "securityId"})
			return
		}

		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")

		var fromPtr *string
		if from != "" {
			fromPtr = &from
		}
		var toPtr *string
		if to != "" {
			toPtr = &to
		}

		items, err := d.InvestmentService.ListSecurityEvents(r.Context(), uid, securityID, fromPtr, toPtr)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// ListTrades godoc
// @Summary List trades
// @Description List trades for a broker investment account.
// @Tags investments
// @Produce json
// @Param investmentAccountId path string true "Investment account ID"
// @Success 200 {array} domain.Trade
// @Failure 401 {object} apierror.Envelope
// @Failure 403 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /investment-accounts/{investmentAccountId}/trades [get]
func ListTrades(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		investmentAccountID := chi.URLParam(r, "investmentAccountId")
		if investmentAccountID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "investmentAccountId is required", map[string]any{"field": "investmentAccountId"})
			return
		}

		items, err := d.InvestmentService.ListTrades(r.Context(), uid, investmentAccountID)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// CreateTrade godoc
// @Summary Create trade
// @Description Create a trade; optionally auto-creates fee/tax expense transactions.
// @Tags investments
// @Accept json
// @Produce json
// @Param investmentAccountId path string true "Investment account ID"
// @Param body body services.CreateTradeRequest true "Create trade request"
// @Success 200 {object} domain.Trade
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 403 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /investment-accounts/{investmentAccountId}/trades [post]
func CreateTrade(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		investmentAccountID := chi.URLParam(r, "investmentAccountId")
		if investmentAccountID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "investmentAccountId is required", map[string]any{"field": "investmentAccountId"})
			return
		}

		var req services.CreateTradeRequest
		if ok := decodeJSON(w, r, &req); !ok {
			return
		}

		item, err := d.InvestmentService.CreateTrade(r.Context(), uid, investmentAccountID, req)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, item)
	}
}

// ListHoldings godoc
// @Summary List holdings
// @Description List holdings for a broker investment account.
// @Tags investments
// @Produce json
// @Param investmentAccountId path string true "Investment account ID"
// @Success 200 {array} domain.Holding
// @Failure 401 {object} apierror.Envelope
// @Failure 403 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /investment-accounts/{investmentAccountId}/holdings [get]
func ListHoldings(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		investmentAccountID := chi.URLParam(r, "investmentAccountId")
		if investmentAccountID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "investmentAccountId is required", map[string]any{"field": "investmentAccountId"})
			return
		}

		items, err := d.InvestmentService.ListHoldings(r.Context(), uid, investmentAccountID)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// ListSecurityEventElections godoc
// @Summary List security event elections
// @Description List elections for a broker investment account.
// @Tags investments
// @Produce json
// @Param investmentAccountId path string true "Investment account ID"
// @Param status query string false "Filter by status: draft|confirmed|cancelled"
// @Success 200 {array} domain.SecurityEventElection
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 403 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /investment-accounts/{investmentAccountId}/security-event-elections [get]
func ListSecurityEventElections(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		investmentAccountID := chi.URLParam(r, "investmentAccountId")
		if investmentAccountID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "investmentAccountId is required", map[string]any{"field": "investmentAccountId"})
			return
		}

		status := r.URL.Query().Get("status")
		var statusPtr *string
		if status != "" {
			statusPtr = &status
		}

		items, err := d.InvestmentService.ListSecurityEventElections(r.Context(), uid, investmentAccountID, statusPtr)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// UpsertSecurityEventElection godoc
// @Summary Upsert security event election
// @Description Create or update an election for the given broker investment account.
// @Tags investments
// @Accept json
// @Produce json
// @Param investmentAccountId path string true "Investment account ID"
// @Param body body services.UpsertSecurityEventElectionRequest true "Upsert election request"
// @Success 200 {object} domain.SecurityEventElection
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 403 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /investment-accounts/{investmentAccountId}/security-event-elections [post]
func UpsertSecurityEventElection(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		investmentAccountID := chi.URLParam(r, "investmentAccountId")
		if investmentAccountID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "investmentAccountId is required", map[string]any{"field": "investmentAccountId"})
			return
		}

		var req services.UpsertSecurityEventElectionRequest
		if ok := decodeJSON(w, r, &req); !ok {
			return
		}

		item, err := d.InvestmentService.UpsertSecurityEventElection(r.Context(), uid, investmentAccountID, req)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, item)
	}
}
