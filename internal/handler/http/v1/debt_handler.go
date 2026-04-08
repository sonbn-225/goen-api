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
 
type DebtHandler struct {
	svc interfaces.DebtService
}
 
func NewDebtHandler(svc interfaces.DebtService) *DebtHandler {
	return &DebtHandler{svc: svc}
}
 
func (h *DebtHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))
 
		r.Route("/debts", func(r chi.Router) {
			r.Get("/", h.List)
			r.Post("/", h.Create)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.Get)
				r.Patch("/", h.Update)
				r.Delete("/", h.Delete)
				r.Post("/payments", h.AddPayment)
				r.Get("/payments", h.ListPayments)
				r.Post("/repay", h.Repay)
			})
		})
	})
}
 
// List godoc
// @Summary List Debts
// @Description Retrieve a list of debts for the current user
// @Tags Debts
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.DebtResponse}
// @Failure 401 {object} response.ErrorEnvelope
// @Router /debts [get]
func (h *DebtHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, items)
}
 
// Create godoc
// @Summary Create Debt
// @Description Create a new debt record linked to a contact
// @Tags Debts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateDebtRequest true "Debt Creation Payload"
// @Success 201 {object} response.SuccessEnvelope{data=dto.DebtResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /debts [post]
func (h *DebtHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
	var req dto.CreateDebtRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}
	res, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusCreated, res)
}
 
// Get godoc
// @Summary Get Debt
// @Description Retrieve a specific debt by its ID
// @Tags Debts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Debt ID"
// @Success 200 {object} response.SuccessEnvelope{data=dto.DebtResponse}
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /debts/{id} [get]
func (h *DebtHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid debt id format", nil)
		return
	}
	res, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	if res == nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "debt not found", nil)
		return
	}
	response.WriteSuccess(w, http.StatusOK, res)
}
 
// Update godoc
// @Summary Update Debt
// @Description Partially update specific debt properties
// @Tags Debts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Debt ID"
// @Param request body dto.UpdateDebtRequest true "Debt Update Payload"
// @Success 200 {object} response.SuccessEnvelope{data=dto.DebtResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /debts/{id} [patch]
func (h *DebtHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid debt id format", nil)
		return
	}
	var req dto.UpdateDebtRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}
	res, err := h.svc.Update(r.Context(), userID, id, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	if res == nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "debt not found", nil)
		return
	}
	response.WriteSuccess(w, http.StatusOK, res)
}
 
// Delete godoc
// @Summary Delete Debt
// @Description Delete a debt record by ID
// @Tags Debts
// @Security BearerAuth
// @Param id path string true "Debt ID"
// @Success 204 "No Content"
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /debts/{id} [delete]
func (h *DebtHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid debt id format", nil)
		return
	}
	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		response.HandleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
 
// AddPayment godoc
// @Summary Add Debt Payment
// @Description Add a new transaction representing a payment toward a specific debt
// @Tags Debts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Debt ID"
// @Param request body dto.DebtPaymentRequest true "Debt Payment Payload"
// @Success 200 {object} response.SuccessEnvelope{data=dto.DebtResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /debts/{id}/payments [post]
func (h *DebtHandler) AddPayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid debt id format", nil)
		return
	}
	var req dto.DebtPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}
	res, err := h.svc.AddPayment(r.Context(), userID, id, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	if res == nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "debt not found", nil)
		return
	}
	response.WriteSuccess(w, http.StatusOK, res)
}
 
// ListPayments godoc
// @Summary List Debt Payments
// @Description Retrieve all payments associated with a specific debt
// @Tags Debts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Debt ID"
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.DebtPaymentLinkResponse}
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /debts/{id}/payments [get]
func (h *DebtHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid debt id format", nil)
		return
	}
	items, err := h.svc.ListPayments(r.Context(), userID, id)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, items)
}

// Repay godoc
// @Summary Repay Debt (1-step)
// @Description Automatically create a payment transaction and link it to the debt
// @Tags Debts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Debt ID"
// @Param request body dto.DebtRepayRequest true "Repayment Payload"
// @Success 200 {object} response.SuccessEnvelope{data=dto.DebtResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /debts/{id}/repay [post]
func (h *DebtHandler) Repay(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid debt id format", nil)
		return
	}
	var req dto.DebtRepayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}
	res, err := h.svc.Repay(r.Context(), userID, id, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	if res == nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "debt not found", nil)
		return
	}
	response.WriteSuccess(w, http.StatusOK, res)
}
