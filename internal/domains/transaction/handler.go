package transaction

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/httpx"
	"github.com/sonbn-225/goen-api-v2/internal/core/money"
	"github.com/sonbn-225/goen-api-v2/internal/core/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// create godoc
// @Summary Create Transaction
// @Description Create a new transaction for the current authenticated user.
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateTransactionRequest true "Create transaction request"
// @Success 201 {object} response.Envelope{data=Transaction}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /transactions [post]
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	var req CreateTransactionRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	in := CreateInput{
		AccountID:           req.AccountID,
		FromAccountID:       req.FromAccountID,
		ToAccountID:         req.ToAccountID,
		Type:                req.Type,
		Amount:              req.Amount,
		Note:                req.Note,
		OwnerOriginalAmount: cloneAmountPointer(req.OwnerOriginalAmount),
	}

	if len(req.LineItems) > 0 {
		in.LineItems = make([]CreateTransactionLineItemInput, 0, len(req.LineItems))
		for _, lineItem := range req.LineItems {
			in.LineItems = append(in.LineItems, CreateTransactionLineItemInput{
				CategoryID: lineItem.CategoryID,
				TagIDs:     lineItem.TagIDs,
				Amount:     lineItem.Amount,
				Note:       cloneStringPointer(lineItem.Note),
			})
		}
	}

	if len(req.GroupParticipants) > 0 {
		in.GroupParticipants = make([]CreateGroupExpenseParticipantInput, 0, len(req.GroupParticipants))
		for _, participant := range req.GroupParticipants {
			in.GroupParticipants = append(in.GroupParticipants, CreateGroupExpenseParticipantInput{
				ParticipantName: participant.ParticipantName,
				OriginalAmount:  participant.OriginalAmount,
				ShareAmount:     cloneAmountPointer(participant.ShareAmount),
			})
		}
	}

	tx, err := h.service.Create(r.Context(), userID, in)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusCreated, tx)
}

// update godoc
// @Summary Update Transaction
// @Description Update transaction fields for current authenticated user.
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param transactionId path string true "Transaction ID"
// @Param request body UpdateTransactionRequest true "Update transaction request"
// @Success 200 {object} response.Envelope{data=Transaction}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /transactions/{transactionId} [patch]
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateTransactionRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	in := UpdateInput{Note: cloneStringPointer(req.Note)}
	if req.LineItems != nil {
		items := make([]UpdateTransactionLineItemInput, 0, len(*req.LineItems))
		for _, lineItem := range *req.LineItems {
			items = append(items, UpdateTransactionLineItemInput{
				CategoryID: lineItem.CategoryID,
				TagIDs:     lineItem.TagIDs,
				Amount:     lineItem.Amount,
				Note:       cloneStringPointer(lineItem.Note),
			})
		}
		in.LineItems = &items
	}

	if req.GroupParticipants != nil {
		participants := make([]UpdateGroupExpenseParticipantInput, 0, len(*req.GroupParticipants))
		for _, participant := range *req.GroupParticipants {
			participants = append(participants, UpdateGroupExpenseParticipantInput{
				ParticipantName: participant.ParticipantName,
				OriginalAmount:  participant.OriginalAmount,
				ShareAmount:     participant.ShareAmount,
			})
		}
		in.GroupParticipants = &participants
	}

	updated, err := h.service.Update(r.Context(), userID, transactionID, in)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, updated)
}

// list godoc
// @Summary List Transactions
// @Description List transactions for current authenticated user.
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit (max 200)"
// @Param account_id query string false "Filter by account id"
// @Param status query string false "Filter by status"
// @Param search query string false "Search by external_ref"
// @Param from query string false "From date/time (YYYY-MM-DD or RFC3339)"
// @Param to query string false "To date/time (YYYY-MM-DD or RFC3339)"
// @Param type query string false "Filter by transaction type"
// @Param external_ref_family query string false "Filter by external ref family"
// @Success 200 {object} response.Envelope{data=[]Transaction,meta=response.Meta}
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /transactions [get]
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	q := r.URL.Query()
	limit, err := parseLimitQuery(q.Get("limit"))
	if err != nil {
		response.WriteError(w, err)
		return
	}

	from, err := parseDateTimeQuery(q.Get("from"), "from")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	to, err := parseDateTimeQuery(q.Get("to"), "to")
	if err != nil {
		response.WriteError(w, err)
		return
	}

	items, totalCount, err := h.service.List(r.Context(), userID, ListFilter{
		AccountID:         optionalQueryString(q.Get("account_id")),
		Status:            optionalQueryString(q.Get("status")),
		Search:            optionalQueryString(q.Get("search")),
		From:              from,
		To:                to,
		Type:              optionalQueryString(q.Get("type")),
		ExternalRefFamily: optionalQueryString(q.Get("external_ref_family")),
		Limit:             limit,
	})
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteList(w, http.StatusOK, items, response.Meta{Total: totalCount, Limit: limit})
}

// get godoc
// @Summary Get Transaction
// @Description Get transaction details by id for current authenticated user.
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param transactionId path string true "Transaction ID"
// @Success 200 {object} response.Envelope{data=Transaction}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /transactions/{transactionId} [get]
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
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

	item, err := h.service.Get(r.Context(), userID, transactionID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, item)
}

// batchPatchStatus godoc
// @Summary Batch Update Transaction Status
// @Description Update status for multiple transactions in one call.
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BatchPatchTransactionsRequest true "Batch status patch request"
// @Success 200 {object} response.Envelope{data=BatchPatchResult}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /transactions/batch [patch]
func (h *Handler) batchPatchStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	var req BatchPatchTransactionsRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	input := BatchPatchRequest{
		TransactionIDs: req.TransactionIDs,
		Patch: BatchPatchData{
			Status: req.Patch.Status,
		},
	}

	result, err := h.service.BatchPatchStatus(r.Context(), userID, input)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, result)
}

// listGroupParticipants godoc
// @Summary List Group Expense Participants For Transaction
// @Description List group expense participants associated with a transaction.
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param transactionId path string true "Transaction ID"
// @Success 200 {object} response.Envelope{data=[]GroupExpenseParticipant,meta=response.Meta}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /transactions/{transactionId}/group-expense-participants [get]
func (h *Handler) listGroupParticipants(w http.ResponseWriter, r *http.Request) {
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

	items, err := h.service.ListGroupParticipantsByTransaction(r.Context(), userID, transactionID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteList(w, http.StatusOK, items, response.Meta{Total: len(items)})
}

func optionalQueryString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func parseLimitQuery(raw string) (int, error) {
	if raw == "" {
		return 50, nil
	}
	limit, err := strconv.Atoi(raw)
	if err != nil {
		return 0, apperrors.New(apperrors.KindValidation, "limit must be an integer")
	}
	if limit <= 0 {
		return 0, apperrors.New(apperrors.KindValidation, "limit must be greater than zero")
	}
	if limit > 200 {
		limit = 200
	}
	return limit, nil
}

func parseDateTimeQuery(raw, field string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return &t, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, apperrors.New(apperrors.KindValidation, field+" must be RFC3339 or YYYY-MM-DD")
	}
	return &t, nil
}

func cloneAmountPointer(v *money.Amount) *money.Amount {
	if v == nil {
		return nil
	}
	cloned := *v
	return &cloned
}

func cloneStringPointer(v *string) *string {
	if v == nil {
		return nil
	}
	cloned := *v
	return &cloned
}
