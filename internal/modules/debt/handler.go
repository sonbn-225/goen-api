package debt

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/httpapi"
	"github.com/sonbn-225/goen-api/internal/response"
)

// Handler handles HTTP requests for debts.
type Handler struct {
	svc *Service
}

// NewHandler creates a new debt handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers all debt routes.
func (h *Handler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.With(authMiddleware).Get("/debts", h.List)
	r.With(authMiddleware).Get("/debts/", h.List)
	r.With(authMiddleware).Post("/debts", h.Create)
	r.With(authMiddleware).Post("/debts/", h.Create)
	r.With(authMiddleware).Get("/debts/{debtId}", h.Get)
	r.With(authMiddleware).Get("/debts/{debtId}/payments", h.ListPayments)
	r.With(authMiddleware).Post("/debts/{debtId}/payments", h.CreatePayment)
	r.With(authMiddleware).Get("/debts/{debtId}/installments", h.ListInstallments)
	r.With(authMiddleware).Post("/debts/{debtId}/installments", h.CreateInstallment)
	r.With(authMiddleware).Get("/transactions/{transactionId}/debt-links", h.ListDebtLinksForTransaction)
}

// List handles GET /debts
// @Summary List debts
// @Description List debts owned by current user.
// @Tags debts
// @Produce json
// @Success 200 {array} domain.Debt
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /debts [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, items)
}

// Create handles POST /debts
// @Summary Create debt
// @Description Create a new debt.
// @Tags debts
// @Accept json
// @Produce json
// @Param body body CreateRequest true "Create debt request"
// @Success 200 {object} domain.Debt
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /debts [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	item, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, item)
}

// Get handles GET /debts/{debtId}
// @Summary Get debt
// @Description Get a single debt.
// @Tags debts
// @Produce json
// @Param debtId path string true "Debt ID"
// @Success 200 {object} domain.Debt
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /debts/{debtId} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	id := chi.URLParam(r, "debtId")
	if id == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "debtId is required", map[string]any{"field": "debtId"})
		return
	}

	item, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, item)
}

// ListPayments handles GET /debts/{debtId}/payments
// @Summary List debt payments
// @Description List payment links for a debt.
// @Tags debts
// @Produce json
// @Param debtId path string true "Debt ID"
// @Success 200 {array} domain.DebtPaymentLink
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /debts/{debtId}/payments [get]
func (h *Handler) ListPayments(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	debtID := chi.URLParam(r, "debtId")
	if debtID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "debtId is required", map[string]any{"field": "debtId"})
		return
	}

	items, err := h.svc.ListPayments(r.Context(), userID, debtID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, items)
}

// CreatePayment handles POST /debts/{debtId}/payments
// @Summary Create debt payment link
// @Description Link a transaction as a payment/collection for a debt.
// @Tags debts
// @Accept json
// @Produce json
// @Param debtId path string true "Debt ID"
// @Param body body CreatePaymentRequest true "Create debt payment request"
// @Success 200 {object} domain.DebtPaymentLink
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /debts/{debtId}/payments [post]
func (h *Handler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	debtID := chi.URLParam(r, "debtId")
	if debtID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "debtId is required", map[string]any{"field": "debtId"})
		return
	}

	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	item, err := h.svc.CreatePayment(r.Context(), userID, debtID, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, item)
}

// ListInstallments handles GET /debts/{debtId}/installments
// @Summary List debt installments
// @Description List installment schedule for a debt.
// @Tags debts
// @Produce json
// @Param debtId path string true "Debt ID"
// @Success 200 {array} domain.DebtInstallment
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /debts/{debtId}/installments [get]
func (h *Handler) ListInstallments(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	debtID := chi.URLParam(r, "debtId")
	if debtID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "debtId is required", map[string]any{"field": "debtId"})
		return
	}

	items, err := h.svc.ListInstallments(r.Context(), userID, debtID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, items)
}

// CreateInstallment handles POST /debts/{debtId}/installments
// @Summary Create debt installment
// @Description Create a single installment schedule row for a debt.
// @Tags debts
// @Accept json
// @Produce json
// @Param debtId path string true "Debt ID"
// @Param body body CreateInstallmentRequest true "Create debt installment request"
// @Success 200 {object} domain.DebtInstallment
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /debts/{debtId}/installments [post]
func (h *Handler) CreateInstallment(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	debtID := chi.URLParam(r, "debtId")
	if debtID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "debtId is required", map[string]any{"field": "debtId"})
		return
	}

	var req CreateInstallmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	item, err := h.svc.CreateInstallment(r.Context(), userID, debtID, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, item)
}

// ListDebtLinksForTransaction handles GET /transactions/{transactionId}/debt-links
// @Summary List debt links for transaction
// @Description List debt payment links for a transaction.
// @Tags debts
// @Produce json
// @Param transactionId path string true "Transaction ID"
// @Success 200 {array} domain.DebtPaymentLink
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /transactions/{transactionId}/debt-links [get]
func (h *Handler) ListDebtLinksForTransaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	txID := chi.URLParam(r, "transactionId")
	if txID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "transactionId is required", map[string]any{"field": "transactionId"})
		return
	}

	items, err := h.svc.ListPaymentsByTransaction(r.Context(), userID, txID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	var se *apperrors.Error
	if errors.As(err, &se) {
		response.WriteError(w, se.HTTPStatus(), string(se.Kind), se.Message, se.Details)
		return
	}
	response.WriteInternalError(w, err)
}
