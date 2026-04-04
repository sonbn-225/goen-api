package debt

import (
	"net/http"

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

// list godoc
// @Summary List Debts
// @Description List debts for current authenticated user.
// @Tags debts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Envelope{data=[]Debt,meta=response.Meta}
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /debts [get]
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	items, err := h.service.List(r.Context(), userID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteList(w, http.StatusOK, items, response.Meta{Total: len(items)})
}

// create godoc
// @Summary Create Debt
// @Description Create a new debt for current authenticated user.
// @Tags debts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateDebtRequest true "Create debt request"
// @Success 201 {object} response.Envelope{data=Debt}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /debts [post]
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	var req CreateDebtRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	created, err := h.service.Create(r.Context(), userID, CreateInput(req))
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusCreated, created)
}

// get godoc
// @Summary Get Debt
// @Description Get debt details by debt id for current authenticated user.
// @Tags debts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param debtId path string true "Debt ID"
// @Success 200 {object} response.Envelope{data=Debt}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /debts/{debtId} [get]
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	debtID := chi.URLParam(r, "debtId")
	if debtID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "debtId is required"))
		return
	}

	item, err := h.service.Get(r.Context(), userID, debtID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, item)
}

// listPayments godoc
// @Summary List Debt Payments
// @Description List payment links for a debt.
// @Tags debts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param debtId path string true "Debt ID"
// @Success 200 {object} response.Envelope{data=[]DebtPaymentLink,meta=response.Meta}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /debts/{debtId}/payments [get]
func (h *Handler) listPayments(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	debtID := chi.URLParam(r, "debtId")
	if debtID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "debtId is required"))
		return
	}

	items, err := h.service.ListPayments(r.Context(), userID, debtID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteList(w, http.StatusOK, items, response.Meta{Total: len(items)})
}

// createPayment godoc
// @Summary Create Debt Payment Link
// @Description Link a transaction as payment or top-up for a debt.
// @Tags debts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param debtId path string true "Debt ID"
// @Param request body CreateDebtPaymentRequest true "Create debt payment request"
// @Success 201 {object} response.Envelope{data=DebtPaymentLink}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /debts/{debtId}/payments [post]
func (h *Handler) createPayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	debtID := chi.URLParam(r, "debtId")
	if debtID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "debtId is required"))
		return
	}

	var req CreateDebtPaymentRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	created, err := h.service.CreatePayment(r.Context(), userID, debtID, CreatePaymentInput(req))
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusCreated, created)
}

// listInstallments godoc
// @Summary List Debt Installments
// @Description List installment schedule for a debt.
// @Tags debts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param debtId path string true "Debt ID"
// @Success 200 {object} response.Envelope{data=[]DebtInstallment,meta=response.Meta}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /debts/{debtId}/installments [get]
func (h *Handler) listInstallments(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	debtID := chi.URLParam(r, "debtId")
	if debtID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "debtId is required"))
		return
	}

	items, err := h.service.ListInstallments(r.Context(), userID, debtID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteList(w, http.StatusOK, items, response.Meta{Total: len(items)})
}

// createInstallment godoc
// @Summary Create Debt Installment
// @Description Create an installment row for a debt.
// @Tags debts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param debtId path string true "Debt ID"
// @Param request body CreateDebtInstallmentRequest true "Create debt installment request"
// @Success 201 {object} response.Envelope{data=DebtInstallment}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /debts/{debtId}/installments [post]
func (h *Handler) createInstallment(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	debtID := chi.URLParam(r, "debtId")
	if debtID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "debtId is required"))
		return
	}

	var req CreateDebtInstallmentRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	created, err := h.service.CreateInstallment(r.Context(), userID, debtID, CreateInstallmentInput(req))
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusCreated, created)
}

// listDebtLinksForTransaction godoc
// @Summary List Debt Links For Transaction
// @Description List debt payment links associated with a transaction.
// @Tags debts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param transactionId path string true "Transaction ID"
// @Success 200 {object} response.Envelope{data=[]DebtPaymentLink,meta=response.Meta}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /transactions/{transactionId}/debt-links [get]
func (h *Handler) listDebtLinksForTransaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	transactionID := chi.URLParam(r, "transactionId")
	if transactionID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "transactionId is required"))
		return
	}

	items, err := h.service.ListPaymentsByTransaction(r.Context(), userID, transactionID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteList(w, http.StatusOK, items, response.Meta{Total: len(items)})
}
