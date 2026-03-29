package transaction

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/platform/httpx"
	"github.com/sonbn-225/goen-api/internal/response"
)

// Handler handles HTTP requests for transactions.
type Handler struct {
	svc *Service
}

type ImportGoenV1Item struct {
	TransactionDate string  `json:"transaction_date"`
	Amount          string  `json:"amount"`
	CategoryID      *string `json:"category_id,omitempty"`
	Category        *string `json:"category,omitempty"`
	Description     *string `json:"description,omitempty"`
}

type ImportGoenV1Body struct {
	Items []ImportGoenV1Item `json:"items"`
}

type ImportGoenV1Response struct {
	Imported int      `json:"imported"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors,omitempty"`
}

type StageImportedGoenV1Body struct {
	Items []ImportedGoenV1StageItem `json:"items"`
}

type MapImportedGoenV1Body struct {
	MappedAccountID  *string `json:"mapped_account_id,omitempty"`
	MappedCategoryID *string `json:"mapped_category_id,omitempty"`
}

type CreateManyImportedGoenV1Body struct {
	ImportIDs []string `json:"import_ids"`
}

// Generic import/export request/response types (source-agnostic)
type StageImportedBody struct {
	Items []StageImportedItem `json:"items"` // Just items, source auto-detected or defaults to 'generic'
}

type MapImportedBody struct {
	MappedAccountID  *string `json:"mapped_account_id,omitempty"`
	MappedCategoryID *string `json:"mapped_category_id,omitempty"`
}

type CreateManyImportedBody struct {
	ImportIDs []string `json:"import_ids"`
}

type UpsertImportRulesBody struct {
	Rules []MappingRuleInput `json:"rules"`
}

type ImportRulesResponse struct {
	Data []domain.ImportMappingRule `json:"data"`
}

type ExportTransactionsQuery struct {
	AccountID *string
	From      *string
	To        *string
}

type ImportedTransactionResponse struct {
	Data []domain.ImportedTransaction `json:"data"`
}

type StagedImportResponse struct {
	Created int      `json:"created"`
	Skipped int      `json:"skipped"`
	Errors  []string `json:"errors,omitempty"`
}

type DeleteAllImportedResponse struct {
	Deleted int64 `json:"deleted"`
}

type ExportTransactionsResponse struct {
	Data []domain.ExportTransactionRow `json:"data"`
}

// NewHandler creates a new transaction handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers all transaction routes.
func (h *Handler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.With(authMiddleware).Get("/transactions", h.List)
	r.With(authMiddleware).Get("/transactions/", h.List)
	r.With(authMiddleware).Post("/transactions", h.Create)
	r.With(authMiddleware).Post("/transactions/", h.Create)
	r.With(authMiddleware).Patch("/transactions/batch", h.BatchPatch)
	r.With(authMiddleware).Get("/transactions/{transactionId}", h.Get)
	r.With(authMiddleware).Patch("/transactions/{transactionId}", h.Patch)
	r.With(authMiddleware).Delete("/transactions/{transactionId}", h.Delete)

	// Export endpoint (v2 format compatible with import)
	r.With(authMiddleware).Get("/transactions/export", h.ExportTransactions)

	// Generic import/export (source-agnostic)
	r.With(authMiddleware).Post("/transactions/import/stage", h.StageImported)
	r.With(authMiddleware).Get("/transactions/import/staged", h.ListImported)
	r.With(authMiddleware).Patch("/transactions/import/staged/{importId}", h.MapImported)
	r.With(authMiddleware).Post("/transactions/import/staged/{importId}/create", h.CreateFromImported)
	r.With(authMiddleware).Post("/transactions/import/staged/create-many", h.CreateManyFromImported)
	r.With(authMiddleware).Get("/transactions/import/rules", h.ListImportRules)
	r.With(authMiddleware).Post("/transactions/import/rules", h.UpsertImportRules)
	r.With(authMiddleware).Delete("/transactions/import/rules/{ruleId}", h.DeleteImportRule)
	r.With(authMiddleware).Post("/transactions/import/apply-rules-create", h.ApplyImportRulesAndCreate)
	r.With(authMiddleware).Delete("/transactions/import/staged/{importId}", h.DeleteImported)
	r.With(authMiddleware).Delete("/transactions/import/staged", h.DeleteAllImported)

	// Legacy goen-v1 import (keep for backward compatibility)
	r.With(authMiddleware).Post("/accounts/{accountId}/import/goen-v1", h.ImportGoenV1)
	r.With(authMiddleware).Post("/transactions/imported/goen-v1", h.StageImportedGoenV1)
	r.With(authMiddleware).Get("/transactions/imported/goen-v1", h.ListImportedGoenV1)
	r.With(authMiddleware).Patch("/transactions/imported/goen-v1/{importId}", h.MapImportedGoenV1)
	r.With(authMiddleware).Post("/transactions/imported/goen-v1/{importId}/create", h.CreateFromImportedGoenV1)
	r.With(authMiddleware).Post("/transactions/imported/goen-v1/create-many", h.CreateManyFromImportedGoenV1)
	r.With(authMiddleware).Delete("/transactions/imported/goen-v1/{importId}", h.DeleteImportedGoenV1)
}

// ListResponse is the response for listing transactions.
type ListResponse struct {
	Data       []domain.Transaction `json:"data"`
	NextCursor *string              `json:"next_cursor"`
	Total      int                  `json:"total,omitempty"`
	TotalPages int                  `json:"total_pages,omitempty"`
	Page       int                  `json:"page,omitempty"`
	Limit      int                  `json:"limit,omitempty"`
}

// CreateBody is the request body for creating a transaction.
type CreateBody struct {
	ClientID      *string                 `json:"client_id,omitempty"`
	ExternalRef   *string                 `json:"external_ref,omitempty"`
	Type          string                  `json:"type"`
	OccurredAt    *string                 `json:"occurred_at,omitempty"`
	OccurredDate  *string                 `json:"occurred_date,omitempty"`
	OccurredTime  *string                 `json:"occurred_time,omitempty"`
	Amount        string                  `json:"amount"`
	FromAmount    *string                 `json:"from_amount,omitempty"`
	ToAmount      *string                 `json:"to_amount,omitempty"`
	Description   *string                 `json:"description,omitempty"`
	AccountID     *string                 `json:"account_id,omitempty"`
	FromAccountID *string                 `json:"from_account_id,omitempty"`
	ToAccountID   *string                 `json:"to_account_id,omitempty"`
	ExchangeRate  *string                 `json:"exchange_rate,omitempty"`
	CategoryID    *string                 `json:"category_id,omitempty"`
	TagIDs                []string                `json:"tag_ids,omitempty"`
	LineItems             []CreateLineItemRequest `json:"line_items,omitempty"`
	GroupParticipants     []GroupParticipantInput `json:"group_participants,omitempty"`
	OwnerOriginalAmount   *string                 `json:"owner_original_amount,omitempty"`
	Lang                  string                  `json:"lang,omitempty"`
}

type BatchPatchBody struct {
	TransactionIDs []string     `json:"transaction_ids"`
	Patch          PatchRequest `json:"patch"`
	Mode           *string      `json:"mode,omitempty"`
}

// List handles GET /transactions
// @Summary List transactions
// @Description List transactions visible to current user.
// @Tags transactions
// @Produce json
// @Param account_id query string false "Filter by account id"
// @Param from query string false "From (YYYY-MM-DD or RFC3339)"
// @Param to query string false "To (YYYY-MM-DD or RFC3339)"
// @Param cursor query string false "Cursor"
// @Param limit query int false "Limit (max 200)"
// @Success 200 {object} ListResponse
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /transactions [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	q := r.URL.Query()
	accountID := q.Get("account_id")
	categoryID := q.Get("category_id")
	txType := q.Get("type")
	search := q.Get("search")
	externalRefFamily := q.Get("external_ref_family")
	from := q.Get("from")
	to := q.Get("to")
	cursor := q.Get("cursor")
	pageStr := q.Get("page")
	limitStr := q.Get("limit")
	limit := 0
	page := 0
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil {
			limit = v
		}
	}
	if pageStr != "" {
		if v, err := strconv.Atoi(pageStr); err == nil {
			page = v
		}
	}

	var accountPtr *string
	if accountID != "" {
		accountPtr = &accountID
	}
	var catPtr *string
	if categoryID != "" {
		catPtr = &categoryID
	}
	var typePtr *string
	if txType != "" {
		typePtr = &txType
	}
	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}
	var externalRefFamilyPtr *string
	if externalRefFamily != "" {
		externalRefFamilyPtr = &externalRefFamily
	}
	var fromPtr *string
	if from != "" {
		fromPtr = &from
	}
	var toPtr *string
	if to != "" {
		toPtr = &to
	}
	var cursorPtr *string
	if cursor != "" {
		cursorPtr = &cursor
	}

	items, next, total, err := h.svc.List(r.Context(), userID, ListRequest{
		AccountID:         accountPtr,
		CategoryID:        catPtr,
		Type:              typePtr,
		Search:            searchPtr,
		ExternalRefFamily: externalRefFamilyPtr,
		From:              fromPtr,
		To:                toPtr,
		Cursor:            cursorPtr,
		Page:              page,
		Limit:             limit,
	})
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	effectiveLimit := limit
	if effectiveLimit <= 0 {
		effectiveLimit = 50
	}
	if effectiveLimit > 200 {
		effectiveLimit = 200
	}
	out := ListResponse{Data: items, NextCursor: next}
	if page > 0 {
		totalPages := 1
		if limit > 0 {
			totalPages = 0
			if total > 0 {
				totalPages = (total + effectiveLimit - 1) / effectiveLimit
			}
			if totalPages <= 0 {
				totalPages = 1
			}
		}
		out.Total = total
		out.TotalPages = totalPages
		out.Page = page
		out.Limit = limit
	}

	response.WriteJSON(w, http.StatusOK, out)
}

// Create handles POST /transactions
// @Summary Create transaction
// @Description Create a new transaction (expense/income/transfer).
// @Tags transactions
// @Accept json
// @Produce json
// @Param body body CreateBody true "Create transaction request"
// @Success 200 {object} domain.Transaction
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /transactions [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid request body", nil)
		return
	}

	// Currency is derived from account(s) and must not be provided by clients.
	// We intentionally do NOT include this field in CreateBody (Swagger schema),
	// but we still validate and reject if a client sends it.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(bodyBytes, &raw); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}
	if v, ok := raw["currency"]; ok && len(v) > 0 && string(v) != "null" {
		var cur string
		if err := json.Unmarshal(v, &cur); err != nil {
			response.WriteError(w, http.StatusBadRequest, "validation_error", "currency must be a string", map[string]any{"field": "currency"})
			return
		}
		if strings.TrimSpace(cur) != "" {
			response.WriteError(w, http.StatusBadRequest, "validation_error", "currency is not supported for transactions (omit currency)", map[string]any{"field": "currency"})
			return
		}
	}

	var body CreateBody
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	req := CreateRequest{
		ClientID:      body.ClientID,
		ExternalRef:   body.ExternalRef,
		Type:          body.Type,
		OccurredAt:    body.OccurredAt,
		OccurredDate:  body.OccurredDate,
		OccurredTime:  body.OccurredTime,
		Amount:        body.Amount,
		FromAmount:    body.FromAmount,
		ToAmount:      body.ToAmount,
		Description:   body.Description,
		AccountID:     body.AccountID,
		FromAccountID: body.FromAccountID,
		ToAccountID:   body.ToAccountID,
		ExchangeRate:  body.ExchangeRate,
		CategoryID:    body.CategoryID,
		TagIDs:              body.TagIDs,
		LineItems:           body.LineItems,
		GroupParticipants:   body.GroupParticipants,
		OwnerOriginalAmount: body.OwnerOriginalAmount,
		Lang:                body.Lang,
	}

	tx, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, tx)
}

// Get handles GET /transactions/{transactionId}
// @Summary Get transaction
// @Description Get a single transaction.
// @Tags transactions
// @Produce json
// @Param transactionId path string true "Transaction ID"
// @Success 200 {object} domain.Transaction
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /transactions/{transactionId} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	id := chi.URLParam(r, "transactionId")
	if id == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "transactionId is required", map[string]any{"field": "transactionId"})
		return
	}

	tx, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, tx)
}

// Patch handles PATCH /transactions/{transactionId}
// @Summary Patch transaction
// @Description Merge-patch transaction fields.
// @Tags transactions
// @Accept json
// @Produce json
// @Param transactionId path string true "Transaction ID"
// @Param body body PatchRequest true "Patch transaction request"
// @Success 200 {object} domain.Transaction
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /transactions/{transactionId} [patch]
func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	id := chi.URLParam(r, "transactionId")
	if id == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "transactionId is required", map[string]any{"field": "transactionId"})
		return
	}

	var req PatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	tx, err := h.svc.Patch(r.Context(), userID, id, req)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, tx)
}

// BatchPatch handles PATCH /transactions/batch
// @Summary Batch patch transactions
// @Description Apply patch payload to multiple transactions in one API call.
// @Tags transactions
// @Accept json
// @Produce json
// @Param body body BatchPatchBody true "Batch patch request"
// @Success 200 {object} BatchPatchResult
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /transactions/batch [patch]
func (h *Handler) BatchPatch(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	var body BatchPatchBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	result, err := h.svc.BatchPatch(r.Context(), userID, BatchPatchRequest{
		TransactionIDs: body.TransactionIDs,
		Patch:          body.Patch,
		Mode:           body.Mode,
	})
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, result)
}

// ImportGoenV1 handles POST /accounts/{accountId}/import/goen-v1
func (h *Handler) ImportGoenV1(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	accountID := strings.TrimSpace(chi.URLParam(r, "accountId"))
	if accountID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "accountId is required", map[string]any{"field": "accountId"})
		return
	}

	var body ImportGoenV1Body
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	result, err := h.svc.ImportGoenV1(r.Context(), userID, accountID, body.Items)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, ImportGoenV1Response{
		Imported: result.Imported,
		Skipped:  result.Skipped,
		Errors:   result.Errors,
	})
}

func (h *Handler) StageImportedGoenV1(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	var body StageImportedGoenV1Body
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	items, err := h.svc.StageImportedGoenV1(r.Context(), userID, body.Items)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) ListImportedGoenV1(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	items, err := h.svc.ListImportedGoenV1(r.Context(), userID)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *Handler) MapImportedGoenV1(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	importID := strings.TrimSpace(chi.URLParam(r, "importId"))
	if importID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "importId is required", map[string]any{"field": "importId"})
		return
	}

	var body MapImportedGoenV1Body
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	item, err := h.svc.MapImportedGoenV1(r.Context(), userID, importID, body.MappedAccountID, body.MappedCategoryID)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) CreateFromImportedGoenV1(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	importID := strings.TrimSpace(chi.URLParam(r, "importId"))
	if importID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "importId is required", map[string]any{"field": "importId"})
		return
	}

	tx, err := h.svc.CreateFromImportedGoenV1(r.Context(), userID, importID)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, tx)
}

func (h *Handler) CreateManyFromImportedGoenV1(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	var body CreateManyImportedGoenV1Body
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	result, err := h.svc.CreateManyFromImportedGoenV1(r.Context(), userID, body.ImportIDs)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) DeleteImportedGoenV1(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	importID := strings.TrimSpace(chi.URLParam(r, "importId"))
	if importID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "importId is required", map[string]any{"field": "importId"})
		return
	}

	if err := h.svc.DeleteImportedGoenV1(r.Context(), userID, importID); err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// Generic Import/Export Handlers (source-agnostic)
// ============================================================================

// ExportTransactions handles GET /transactions/export
func (h *Handler) ExportTransactions(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	q := r.URL.Query()
	accountID := q.Get("account_id")
	from := q.Get("from")
	to := q.Get("to")

	var accountPtr *string
	if accountID != "" {
		accountPtr = &accountID
	}
	var fromPtr, toPtr *time.Time
	if from != "" {
		t, err := parseTimeOrDateHandler(from)
		if err != nil {
			response.WriteError(w, http.StatusBadRequest, "validation_error", "from is invalid", map[string]any{"field": "from"})
			return
		}
		fromPtr = &t
	}
	if to != "" {
		t, err := parseTimeOrDateHandler(to)
		if err != nil {
			response.WriteError(w, http.StatusBadRequest, "validation_error", "to is invalid", map[string]any{"field": "to"})
			return
		}
		toPtr = &t
	}

	rows, err := h.svc.ExportTransactions(r.Context(), userID, ExportTransactionsFilter{
		AccountID: accountPtr,
		From:      fromPtr,
		To:        toPtr,
	})
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, ExportTransactionsResponse{Data: rows})
}

// StageImported handles POST /transactions/import/stage
func (h *Handler) StageImported(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	var body StageImportedBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	// Always use 'generic' source - flexible for any CSV format
	items, err := h.svc.StageImported(r.Context(), userID, "generic", body.Items)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, ImportedTransactionResponse{Data: items})
}

// ListImported handles GET /transactions/import/staged
func (h *Handler) ListImported(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	// List all staged imports regardless of source
	items, err := h.svc.ListImported(r.Context(), userID, nil)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, ImportedTransactionResponse{Data: items})
}

// MapImported handles PATCH /transactions/import/staged/{importId}
func (h *Handler) MapImported(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	importID := strings.TrimSpace(chi.URLParam(r, "importId"))
	if importID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "importId is required", map[string]any{"field": "importId"})
		return
	}

	var body MapImportedBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	item, err := h.svc.MapImported(r.Context(), userID, importID, body.MappedAccountID, body.MappedCategoryID)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, item)
}

// CreateFromImported handles POST /transactions/import/staged/{importId}/create
func (h *Handler) CreateFromImported(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	importID := strings.TrimSpace(chi.URLParam(r, "importId"))
	if importID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "importId is required", map[string]any{"field": "importId"})
		return
	}

	tx, err := h.svc.CreateFromImported(r.Context(), userID, importID)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, tx)
}

// CreateManyFromImported handles POST /transactions/import/staged/create-many
func (h *Handler) CreateManyFromImported(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	var body CreateManyImportedBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	result, err := h.svc.CreateManyFromImported(r.Context(), userID, body.ImportIDs)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, result)
}

// DeleteImported handles DELETE /transactions/import/staged/{importId}
func (h *Handler) DeleteImported(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	importID := strings.TrimSpace(chi.URLParam(r, "importId"))
	if importID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "importId is required", map[string]any{"field": "importId"})
		return
	}

	if err := h.svc.DeleteImported(r.Context(), userID, importID); err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteAllImported handles DELETE /transactions/import/staged
func (h *Handler) DeleteAllImported(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	deleted, err := h.svc.DeleteAllImported(r.Context(), userID)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, DeleteAllImportedResponse{Deleted: deleted})
}

// ListImportRules handles GET /transactions/import/rules
func (h *Handler) ListImportRules(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	items, err := h.svc.ListImportRules(r.Context(), userID)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, ImportRulesResponse{Data: items})
}

// UpsertImportRules handles POST /transactions/import/rules
func (h *Handler) UpsertImportRules(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	var body UpsertImportRulesBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	items, err := h.svc.UpsertImportRules(r.Context(), userID, body.Rules)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, ImportRulesResponse{Data: items})
}

// DeleteImportRule handles DELETE /transactions/import/rules/{ruleId}
func (h *Handler) DeleteImportRule(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	ruleID := strings.TrimSpace(chi.URLParam(r, "ruleId"))
	if ruleID == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "ruleId is required", map[string]any{"field": "ruleId"})
		return
	}

	if err := h.svc.DeleteImportRule(r.Context(), userID, ruleID); err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ApplyImportRulesAndCreate handles POST /transactions/import/apply-rules-create
func (h *Handler) ApplyImportRulesAndCreate(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	result, err := h.svc.ApplyImportRulesAndCreate(r.Context(), userID)
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, result)
}

// Delete handles DELETE /transactions/{transactionId}
// @Summary Delete transaction
// @Description Soft delete a transaction.
// @Tags transactions
// @Param transactionId path string true "Transaction ID"
// @Success 204
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /transactions/{transactionId} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteUnauthorized(w)
		return
	}

	id := chi.URLParam(r, "transactionId")
	if id == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "transactionId is required", map[string]any{"field": "transactionId"})
		return
	}

	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// parseTimeOrDateHandler parses a time string in RFC3339 or YYYY-MM-DD format
func parseTimeOrDateHandler(v string) (time.Time, error) {
	if strings.Contains(v, "T") {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return time.Time{}, err
		}
		return t.UTC(), nil
	}
	d, err := time.Parse("2006-01-02", v)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC), nil
}



