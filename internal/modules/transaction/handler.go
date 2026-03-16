package transaction

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/platform/httpx"
	"github.com/sonbn-225/goen-api/internal/response"
)

// Handler handles HTTP requests for transactions.
type Handler struct {
	svc *Service
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
	r.With(authMiddleware).Get("/transactions/{transactionId}", h.Get)
	r.With(authMiddleware).Patch("/transactions/{transactionId}", h.Patch)
	r.With(authMiddleware).Delete("/transactions/{transactionId}", h.Delete)
}

// ListResponse is the response for listing transactions.
type ListResponse struct {
	Data       []domain.Transaction `json:"data"`
	NextCursor *string              `json:"next_cursor"`
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
	TagIDs        []string                `json:"tag_ids,omitempty"`
	LineItems     []CreateLineItemRequest `json:"line_items,omitempty"`
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
	limitStr := q.Get("limit")
	limit := 0
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil {
			limit = v
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

	items, next, err := h.svc.List(r.Context(), userID, ListRequest{
		AccountID:         accountPtr,
		CategoryID:        catPtr,
		Type:              typePtr,
		Search:            searchPtr,
		ExternalRefFamily: externalRefFamilyPtr,
		From:              fromPtr,
		To:                toPtr,
		Cursor:            cursorPtr,
		Limit:             limit,
	})
	if err != nil {
		httpx.WriteServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, ListResponse{Data: items, NextCursor: next})
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
		TagIDs:        body.TagIDs,
		LineItems:     body.LineItems,
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

