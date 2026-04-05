package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
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

			r.Route("/transactions", func(r chi.Router) {
			r.Get("/", h.List)
			r.Post("/", h.Create)
			r.Post("/batch-patch", h.BatchPatch)
			r.Route("/import", func(r chi.Router) {
				r.Post("/stage", h.StageImport)
				r.Get("/staged", h.ListImported)
				r.Delete("/staged", h.ClearImported)
				r.Post("/apply-rules-create", h.ApplyRulesAndCreate)
				r.Route("/staged/{importId}", func(r chi.Router) {
					r.Patch("/", h.PatchImported)
					r.Delete("/", h.DeleteImported)
					r.Post("/create", h.CreateFromImported)
				})
				r.Post("/staged/create-many", h.CreateManyFromImported)
				r.Route("/rules", func(r chi.Router) {
					r.Get("/", h.ListMappingRules)
					r.Post("/", h.UpsertMappingRules)
					r.Delete("/{ruleId}", h.DeleteMappingRule)
				})
			})
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

	response.WriteSuccessWithMeta(w, http.StatusOK, items, map[string]any{
		"next_cursor": cursor,
		"total":       total,
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
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	transactionID := chi.URLParam(r, "transactionId")
	tx, err := h.svc.Get(r.Context(), userID, transactionID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	if tx == nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "transaction not found", nil)
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
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	transactionID := chi.URLParam(r, "transactionId")
	var req dto.TransactionPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
		return
	}

	tx, err := h.svc.Patch(r.Context(), userID, transactionID, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	if tx == nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "transaction not found", nil)
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
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	transactionID := chi.URLParam(r, "transactionId")
	if err := h.svc.Delete(r.Context(), userID, transactionID); err != nil {
		response.WriteInternalError(w, err)
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

	response.WriteSuccess(w, http.StatusOK, res)
}

// StageImport godoc
// @Summary Stage Imported Transactions
// @Description Stage generic transaction data for manual mapping and review
// @Tags Transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.StageImportRequest true "Import Payload"
// @Success 200 {object} response.SuccessEnvelope{data=object}
// @Router /transactions/import/stage [post]
func (h *TransactionHandler) StageImport(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	var req dto.StageImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
		return
	}

	staged, skipped, errors, err := h.svc.StageImport(r.Context(), userID, req.Items)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, map[string]any{
		"staged_count":  staged,
		"skipped_count": skipped,
		"errors":        errors,
	})
}

// ListImported godoc
// @Summary List Staged Imports
// @Description Retrieve all transactions currently in the staging area
// @Tags Transactions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.ImportedTransactionResponse}
// @Router /transactions/import/staged [get]
func (h *TransactionHandler) ListImported(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	items, err := h.svc.ListImported(r.Context(), userID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, items)
}

// PatchImported godoc
// @Summary Update Staged Import Mapping
// @Description Manually map a staged import to an account and category
// @Tags Transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param importId path string true "Import ID"
// @Param request body dto.PatchImportedRequest true "Mapping Payload"
// @Success 200 {object} response.SuccessEnvelope{data=dto.ImportedTransactionResponse}
// @Router /transactions/import/staged/{importId} [patch]
func (h *TransactionHandler) PatchImported(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	importID := chi.URLParam(r, "importId")
	var req dto.PatchImportedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
		return
	}

	patch := entity.ImportedTransactionPatch{
		MappedAccountID:  req.MappedAccountID,
		MappedCategoryID: req.MappedCategoryID,
	}

	res, err := h.svc.PatchImported(r.Context(), userID, importID, patch)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, res)
}

// DeleteImported godoc
// @Summary Delete Staged Import
// @Description Remove an item from the staging area without creating a transaction
// @Tags Transactions
// @Security BearerAuth
// @Param importId path string true "Import ID"
// @Success 204 "No Content"
// @Router /transactions/import/staged/{importId} [delete]
func (h *TransactionHandler) DeleteImported(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	importID := chi.URLParam(r, "importId")
	if err := h.svc.DeleteImported(r.Context(), userID, importID); err != nil {
		response.WriteInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ClearImported godoc
// @Summary Clear All Staged Imports
// @Description Remove all items from the staging area
// @Tags Transactions
// @Security BearerAuth
// @Success 204 "No Content"
// @Router /transactions/import/staged [delete]
func (h *TransactionHandler) ClearImported(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	if err := h.svc.ClearImported(r.Context(), userID); err != nil {
		response.WriteInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateFromImported godoc
// @Summary Create Transaction from Staged Import
// @Description Finalize a staged import by creating a real transaction
// @Tags Transactions
// @Produce json
// @Security BearerAuth
// @Param importId path string true "Import ID"
// @Success 201 {object} response.SuccessEnvelope{data=dto.TransactionResponse}
// @Router /transactions/import/staged/{importId}/create [post]
func (h *TransactionHandler) CreateFromImported(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	importID := chi.URLParam(r, "importId")
	tx, err := h.svc.CreateFromImported(r.Context(), userID, importID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusCreated, tx)
}

// CreateManyFromImported godoc
// @Summary Create Multiple Transactions from Staged Imports
// @Description Finalize multiple staged imports at once
// @Tags Transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateManyImportedRequest true "Import IDs Payload"
// @Success 200 {object} response.SuccessEnvelope{data=dto.BatchImportResult}
// @Router /transactions/import/staged/create-many [post]
func (h *TransactionHandler) CreateManyFromImported(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	var req dto.CreateManyImportedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
		return
	}

	res, err := h.svc.CreateManyFromImported(r.Context(), userID, req.ImportIDs)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, res)
}

// ListMappingRules godoc
// @Summary List Import Mapping Rules
// @Description Retrieve all auto-mapping rules for the authenticated user
// @Tags Transactions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.ImportMappingRuleResponse}
// @Router /transactions/import/rules [get]
func (h *TransactionHandler) ListMappingRules(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	rules, err := h.svc.ListMappingRules(r.Context(), userID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, rules)
}

// UpsertMappingRules godoc
// @Summary Upsert Import Mapping Rules
// @Description Create or update multiple auto-mapping rules
// @Tags Transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpsertMappingRulesRequest true "Rules Payload"
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.ImportMappingRuleResponse}
// @Router /transactions/import/rules [post]
func (h *TransactionHandler) UpsertMappingRules(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	var req dto.UpsertMappingRulesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
		return
	}

	res, err := h.svc.UpsertMappingRules(r.Context(), userID, req.Rules)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, res)
}

// DeleteMappingRule godoc
// @Summary Delete Import Mapping Rule
// @Description Remove an auto-mapping rule
// @Tags Transactions
// @Security BearerAuth
// @Param ruleId path string true "Rule ID"
// @Success 204 "No Content"
// @Router /transactions/import/rules/{ruleId} [delete]
func (h *TransactionHandler) DeleteMappingRule(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	ruleID := chi.URLParam(r, "ruleId")
	if err := h.svc.DeleteMappingRule(r.Context(), userID, ruleID); err != nil {
		response.WriteInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ApplyRulesAndCreate godoc
// @Summary Apply Rules and Create Transactions
// @Description Automatically map all staged imports using rules and create transactions for fully mapped items
// @Tags Transactions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessEnvelope{data=dto.BatchImportResult}
// @Router /transactions/import/apply-rules-create [post]
func (h *TransactionHandler) ApplyRulesAndCreate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	res, err := h.svc.ApplyRulesAndCreate(r.Context(), userID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, res)
}
