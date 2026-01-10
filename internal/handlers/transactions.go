package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/services"
)

type CreateTransactionBody struct {
	ClientID      *string                                     `json:"client_id,omitempty"`
	ExternalRef   *string                                     `json:"external_ref,omitempty"`
	Type          string                                      `json:"type"`
	OccurredAt    *string                                     `json:"occurred_at,omitempty"`
	OccurredDate  *string                                     `json:"occurred_date,omitempty"`
	OccurredTime  *string                                     `json:"occurred_time,omitempty"`
	Amount        string                                      `json:"amount"`
	Currency      *string                                     `json:"currency,omitempty"`
	FromAmount    *string                                     `json:"from_amount,omitempty"`
	ToAmount      *string                                     `json:"to_amount,omitempty"`
	Description   *string                                     `json:"description,omitempty"`
	AccountID     *string                                     `json:"account_id,omitempty"`
	FromAccountID *string                                     `json:"from_account_id,omitempty"`
	ToAccountID   *string                                     `json:"to_account_id,omitempty"`
	ExchangeRate  *string                                     `json:"exchange_rate,omitempty"`
	Counterparty  *string                                     `json:"counterparty,omitempty"`
	Notes         *string                                     `json:"notes,omitempty"`
	TagIDs        []string                                    `json:"tag_ids,omitempty"`
	LineItems     []services.CreateTransactionLineItemRequest `json:"line_items,omitempty"`
}

type TransactionListResponse struct {
	Data       []domain.Transaction `json:"data"`
	NextCursor *string              `json:"next_cursor"`
}

// ListTransactions godoc
// @Summary List transactions
// @Description List transactions visible to current user; supports filtering by account and time range.
// @Tags transactions
// @Produce json
// @Param account_id query string false "Filter by account id"
// @Param from query string false "From (YYYY-MM-DD or RFC3339)"
// @Param to query string false "To (YYYY-MM-DD or RFC3339)"
// @Param cursor query string false "Cursor"
// @Param limit query int false "Limit (max 200)"
// @Success 200 {object} handlers.TransactionListResponse
// @Failure 401 {object} apierror.Envelope
// @Failure 400 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /transactions [get]
func ListTransactions(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		q := r.URL.Query()
		accountID := q.Get("account_id")
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

		items, next, err := d.TransactionService.List(r.Context(), uid, services.ListTransactionsRequest{
			AccountID: accountPtr,
			From:      fromPtr,
			To:        toPtr,
			Cursor:    cursorPtr,
			Limit:     limit,
		})
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, TransactionListResponse{Data: items, NextCursor: next})
	}
}

// GetTransaction godoc
// @Summary Get transaction
// @Description Get a single transaction if visible to current user.
// @Tags transactions
// @Produce json
// @Param transactionId path string true "Transaction ID"
// @Success 200 {object} domain.Transaction
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /transactions/{transactionId} [get]
func GetTransaction(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		id := chi.URLParam(r, "transactionId")
		if id == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "transactionId is required", map[string]any{"field": "transactionId"})
			return
		}

		tx, err := d.TransactionService.Get(r.Context(), uid, id)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, tx)
	}
}

// CreateTransaction godoc
// @Summary Create transaction
// @Description Create a new transaction (expense/income/transfer).
// @Tags transactions
// @Accept json
// @Produce json
// @Param X-Client-Id header string false "Client instance ID (recommended)"
// @Param body body handlers.CreateTransactionBody true "Create transaction request"
// @Success 200 {object} domain.Transaction
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 403 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /transactions [post]
func CreateTransaction(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		var body CreateTransactionBody
		if ok := decodeJSON(w, r, &body); !ok {
			return
		}
		if body.Currency != nil && strings.TrimSpace(*body.Currency) != "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "currency is not supported for transactions (omit currency)", map[string]any{"field": "currency"})
			return
		}

		req := services.CreateTransactionRequest{
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
			Counterparty:  body.Counterparty,
			Notes:         body.Notes,
			TagIDs:        body.TagIDs,
			LineItems:     body.LineItems,
		}

		tx, err := d.TransactionService.Create(r.Context(), uid, req)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, tx)
	}
}

// PatchTransaction godoc
// @Summary Patch transaction
// @Description Merge-patch transaction fields (MVP: description/notes/counterparty).
// @Tags transactions
// @Accept json
// @Produce json
// @Param transactionId path string true "Transaction ID"
// @Param body body services.PatchTransactionRequest true "Patch transaction request"
// @Success 200 {object} domain.Transaction
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 403 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /transactions/{transactionId} [patch]
func PatchTransaction(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		id := chi.URLParam(r, "transactionId")
		if id == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "transactionId is required", map[string]any{"field": "transactionId"})
			return
		}

		var req services.PatchTransactionRequest
		if ok := decodeJSON(w, r, &req); !ok {
			return
		}

		tx, err := d.TransactionService.Patch(r.Context(), uid, id, req)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, tx)
	}
}

// DeleteTransaction godoc
// @Summary Delete transaction
// @Description Soft delete a transaction.
// @Tags transactions
// @Produce json
// @Param transactionId path string true "Transaction ID"
// @Success 204
// @Failure 401 {object} apierror.Envelope
// @Failure 403 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /transactions/{transactionId} [delete]
func DeleteTransaction(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		id := chi.URLParam(r, "transactionId")
		if id == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "transactionId is required", map[string]any{"field": "transactionId"})
			return
		}

		err := d.TransactionService.Delete(r.Context(), uid, id)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
