package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/auth"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/services"
)

// ListDebts godoc
// @Summary List debts
// @Description List debts owned by current user.
// @Tags debts
// @Produce json
// @Success 200 {array} domain.Debt
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /debts [get]
func ListDebts(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		items, err := d.DebtService.List(r.Context(), uid)
		if err != nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// CreateDebt godoc
// @Summary Create debt
// @Description Create a new debt owned by current user.
// @Tags debts
// @Accept json
// @Produce json
// @Param body body services.CreateDebtRequest true "Create debt request"
// @Success 200 {object} domain.Debt
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /debts [post]
func CreateDebt(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		var req services.CreateDebtRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
			return
		}

		item, err := d.DebtService.Create(r.Context(), uid, req)
		if err != nil {
			apierror.Write(w, http.StatusBadRequest, "validation_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, item)
	}
}

// GetDebt godoc
// @Summary Get debt
// @Description Get a single debt owned by current user.
// @Tags debts
// @Produce json
// @Param debtId path string true "Debt ID"
// @Success 200 {object} domain.Debt
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /debts/{debtId} [get]
func GetDebt(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		id := chi.URLParam(r, "debtId")
		if id == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "debtId is required", map[string]any{"field": "debtId"})
			return
		}

		item, err := d.DebtService.Get(r.Context(), uid, id)
		if err != nil {
			if errors.Is(err, domain.ErrDebtNotFound) {
				apierror.Write(w, http.StatusNotFound, "not_found", "debt not found", nil)
				return
			}
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, item)
	}
}

// ListDebtPayments godoc
// @Summary List debt payments
// @Description List payment links for a debt.
// @Tags debts
// @Produce json
// @Param debtId path string true "Debt ID"
// @Success 200 {array} domain.DebtPaymentLink
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /debts/{debtId}/payments [get]
func ListDebtPayments(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		debtID := chi.URLParam(r, "debtId")
		if debtID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "debtId is required", map[string]any{"field": "debtId"})
			return
		}

		items, err := d.DebtService.ListPayments(r.Context(), uid, debtID)
		if err != nil {
			if errors.Is(err, domain.ErrDebtNotFound) {
				apierror.Write(w, http.StatusNotFound, "not_found", "debt not found", nil)
				return
			}
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// CreateDebtPayment godoc
// @Summary Create debt payment link
// @Description Link a transaction as a payment/collection for a debt and update outstanding.
// @Tags debts
// @Accept json
// @Produce json
// @Param debtId path string true "Debt ID"
// @Param body body services.CreateDebtPaymentRequest true "Create debt payment request"
// @Success 200 {object} domain.DebtPaymentLink
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /debts/{debtId}/payments [post]
func CreateDebtPayment(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		debtID := chi.URLParam(r, "debtId")
		if debtID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "debtId is required", map[string]any{"field": "debtId"})
			return
		}

		var req services.CreateDebtPaymentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
			return
		}

		item, err := d.DebtService.CreatePayment(r.Context(), uid, debtID, req)
		if err != nil {
			if errors.Is(err, domain.ErrDebtNotFound) {
				apierror.Write(w, http.StatusNotFound, "not_found", "debt not found", nil)
				return
			}
			if errors.Is(err, domain.ErrTransactionNotFound) {
				apierror.Write(w, http.StatusNotFound, "not_found", "transaction not found", nil)
				return
			}
			apierror.Write(w, http.StatusBadRequest, "validation_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, item)
	}
}

// ListDebtInstallments godoc
// @Summary List debt installments
// @Description List installment schedule for a debt.
// @Tags debts
// @Produce json
// @Param debtId path string true "Debt ID"
// @Success 200 {array} domain.DebtInstallment
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /debts/{debtId}/installments [get]
func ListDebtInstallments(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		debtID := chi.URLParam(r, "debtId")
		if debtID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "debtId is required", map[string]any{"field": "debtId"})
			return
		}

		items, err := d.DebtService.ListInstallments(r.Context(), uid, debtID)
		if err != nil {
			if errors.Is(err, domain.ErrDebtNotFound) {
				apierror.Write(w, http.StatusNotFound, "not_found", "debt not found", nil)
				return
			}
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// CreateDebtInstallment godoc
// @Summary Create debt installment
// @Description Create a single installment schedule row for a debt.
// @Tags debts
// @Accept json
// @Produce json
// @Param debtId path string true "Debt ID"
// @Param body body services.CreateDebtInstallmentRequest true "Create debt installment request"
// @Success 200 {object} domain.DebtInstallment
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /debts/{debtId}/installments [post]
func CreateDebtInstallment(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		debtID := chi.URLParam(r, "debtId")
		if debtID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "debtId is required", map[string]any{"field": "debtId"})
			return
		}

		var req services.CreateDebtInstallmentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
			return
		}

		item, err := d.DebtService.CreateInstallment(r.Context(), uid, debtID, req)
		if err != nil {
			if errors.Is(err, domain.ErrDebtNotFound) {
				apierror.Write(w, http.StatusNotFound, "not_found", "debt not found", nil)
				return
			}
			apierror.Write(w, http.StatusBadRequest, "validation_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, item)
	}
}

// ListDebtLinksForTransaction godoc
// @Summary List debt links for transaction
// @Description List debt payment links for a transaction accessible by current user.
// @Tags debts
// @Produce json
// @Param transactionId path string true "Transaction ID"
// @Success 200 {array} domain.DebtPaymentLink
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /transactions/{transactionId}/debt-links [get]
func ListDebtLinksForTransaction(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		txID := chi.URLParam(r, "transactionId")
		if txID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "transactionId is required", map[string]any{"field": "transactionId"})
			return
		}

		items, err := d.DebtService.ListPaymentsByTransaction(r.Context(), uid, txID)
		if err != nil {
			if errors.Is(err, domain.ErrTransactionNotFound) {
				apierror.Write(w, http.StatusNotFound, "not_found", "transaction not found", nil)
				return
			}
			apierror.Write(w, http.StatusBadRequest, "validation_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}
