package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
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

		r.Route("/transactions", func(r chi.Router) {
			r.Get("/", h.List)
			r.Post("/", h.Create)
			r.Post("/batch-patch", h.BatchPatch)
			r.Route("/{transactionId}", func(r chi.Router) {
				r.Get("/", h.Get)
				r.Patch("/", h.Patch)
				r.Delete("/", h.Delete)
			})
		})
	})
}

// List godoc
// @Summary List Transactions
// @Description Retrieve a paginated list of transactions for the authenticated user
// @Tags Transactions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.TransactionResponse,meta=object}
// @Failure 401 {object} response.ErrorEnvelope
// @Router /transactions [get]
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	normalize := func(s string) *string {
		return utils.NormalizeOptionalString(&s)
	}

	var accID *uuid.UUID
	if val := q.Get("account_id"); val != "" {
		if id, err := uuid.Parse(val); err == nil {
			accID = &id
		}
	}
	var catID *uuid.UUID
	if val := q.Get("category_id"); val != "" {
		if id, err := uuid.Parse(val); err == nil {
			catID = &id
		}
	}

	req := dto.ListTransactionsRequest{
		AccountID:  accID,
		CategoryID: catID,
		Type:       (*entity.TransactionType)(normalize(q.Get("type"))),
		Search:     normalize(q.Get("search")),
		From:       normalize(q.Get("from")),
		To:         normalize(q.Get("to")),
		Page:       page,
		Limit:      limit,
	}

	items, cursor, total, err := h.svc.List(r.Context(), userID, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	totalPages := 0
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}

	response.WriteSuccess(w, http.StatusOK, dto.ListTransactionsResponse{
		Data:       items,
		TotalCount: total,
		TotalPages: totalPages,
		NextCursor: cursor,
		Page:       page,
		Limit:      limit,
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
// @Success 201 {object} response.SuccessEnvelope{data=dto.TransactionResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /transactions [post]
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}

	var req dto.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.HandleError(w, apperr.BadRequest("validation_error", "invalid json"))
		return
	}

	tx, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusCreated, tx)
}

// Get godoc
// @Summary Get Transaction
// @Description Retrieve details of a specific transaction by its ID
// @Tags Transactions
// @Produce json
// @Security BearerAuth
// @Param transactionId path string true "Transaction ID"
// @Success 200 {object} response.SuccessEnvelope{data=dto.TransactionResponse}
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /transactions/{transactionId} [get]
func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}

	transactionID, err := uuid.Parse(chi.URLParam(r, "transactionId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid transaction id format", nil)
		return
	}

	tx, err := h.svc.Get(r.Context(), userID, transactionID)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	if tx == nil {
		response.HandleError(w, apperr.ErrNotFound)
		return
	}

	response.WriteSuccess(w, http.StatusOK, tx)
}

// Patch godoc
// @Summary Update Transaction
// @Description Update fields of an existing transaction
// @Tags Transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param transactionId path string true "Transaction ID"
// @Param request body dto.TransactionPatchRequest true "Transaction Patch Payload"
// @Success 200 {object} response.SuccessEnvelope{data=dto.TransactionResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /transactions/{transactionId} [patch]
func (h *TransactionHandler) Patch(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}

	transactionID, err := uuid.Parse(chi.URLParam(r, "transactionId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid transaction id format", nil)
		return
	}

	var req dto.TransactionPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.HandleError(w, apperr.BadRequest("validation_error", "invalid json"))
		return
	}

	tx, err := h.svc.Patch(r.Context(), userID, transactionID, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	if tx == nil {
		response.HandleError(w, apperr.ErrNotFound)
		return
	}

	response.WriteSuccess(w, http.StatusOK, tx)
}

// Delete godoc
// @Summary Delete Transaction
// @Description Remove a transaction record permanently
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
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}

	transactionID, err := uuid.Parse(chi.URLParam(r, "transactionId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid transaction id format", nil)
		return
	}

	if err := h.svc.Delete(r.Context(), userID, transactionID); err != nil {
		response.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// BatchPatch godoc
// @Summary Batch Update Transactions
// @Description Update multiple transactions at once (e.g. category assignment)
// @Tags Transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.BatchPatchRequest true "Batch Patch Payload"
// @Success 200 {object} response.SuccessEnvelope{data=dto.BatchPatchResult}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /transactions/batch-patch [post]
func (h *TransactionHandler) BatchPatch(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}

	var req dto.BatchPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
		return
	}

	res, err := h.svc.BatchPatch(r.Context(), userID, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, res)
}



