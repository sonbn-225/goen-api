package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
)

type TransactionHandler struct {
	svc interfaces.TransactionService
}

func NewTransactionHandler(svc interfaces.TransactionService) *TransactionHandler {
	return &TransactionHandler{svc: svc}
}

func (h *TransactionHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))
		r.Get("/transactions", h.List)
		r.Post("/transactions", h.Create)
		r.Post("/transactions/batch-patch", h.BatchPatch)
		r.Get("/transactions/{transactionId}", h.Get)
		r.Patch("/transactions/{transactionId}", h.Patch)
		r.Delete("/transactions/{transactionId}", h.Delete)
	})
}

// List godoc
// @Summary List Transactions
// @Description Retrieve a paginated list of transactions for the authenticated user
// @Tags Transactions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object "items, next_cursor, total"
// @Failure 401 {object} response.ErrorEnvelope
// @Router /transactions [get]
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	// Filter from query params (Simplified mapping)
	req := dto.CreateTransactionRequest{
		// To be filled from query params if needed by service.
		// For now service LIST is stubbed anyway.
	}
	
	items, cursor, total, err := h.svc.List(r.Context(), userID, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"next_cursor": cursor,
		"total": total,
	})
}

// Create godoc
// @Summary Create Transaction
// @Description Create a new financial transaction with optional line items and tags
// @Tags Transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateTransactionRequest true "Transaction Creation Payload"
// @Success 201 {object} entity.Transaction
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /transactions [post]
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	var req dto.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
		return
	}

	tx, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusCreated, tx)
}

// Get godoc
// @Summary Get Transaction
// @Description Retrieve details of a specific transaction by its ID
// @Tags Transactions
// @Produce json
// @Security BearerAuth
// @Param transactionId path string true "Transaction ID"
// @Success 200 {object} entity.Transaction
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /transactions/{transactionId} [get]
func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	id := chi.URLParam(r, "transactionId")
	tx, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, tx)
}

// Patch godoc
// @Summary Update Transaction
// @Description Partially update properties of a transaction
// @Tags Transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param transactionId path string true "Transaction ID"
// @Param request body dto.TransactionPatchRequest true "Transaction Patch Payload"
// @Success 200 {object} entity.Transaction
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /transactions/{transactionId} [patch]
func (h *TransactionHandler) Patch(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	id := chi.URLParam(r, "transactionId")
	var req dto.TransactionPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
		return
	}

	tx, err := h.svc.Patch(r.Context(), userID, id, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, tx)
}

// BatchPatch godoc
// @Summary Batch Patch Transactions
// @Description Update multiple transactions at once (e.g. recategorize, retag)
// @Tags Transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.BatchPatchRequest true "Batch Patch Payload"
// @Success 200 {object} object
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /transactions/batch-patch [post]
func (h *TransactionHandler) BatchPatch(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	var req dto.BatchPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
		return
	}

	res, err := h.svc.BatchPatch(r.Context(), userID, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, res)
}

// Delete godoc
// @Summary Delete Transaction
// @Description Delete a specific transaction by its ID
// @Tags Transactions
// @Security BearerAuth
// @Param transactionId path string true "Transaction ID"
// @Success 204 "No Content"
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /transactions/{transactionId} [delete]
func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	id := chi.URLParam(r, "transactionId")
	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		response.WriteInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
