package account

import (
	"net/http"
	"strings"

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

// create godoc
// @Summary Create Account
// @Description Create a new account for the current authenticated user.
// @Tags accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateAccountRequest true "Create account request"
// @Success 201 {object} response.Envelope{data=Account}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /accounts [post]
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	var req CreateAccountRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	in := CreateInput{
		Name:            req.Name,
		Type:            req.Type,
		Currency:        req.Currency,
		ParentAccountID: req.ParentAccount,
		AccountNumber:   req.AccountNumber,
		Color:           req.Color,
	}
	if strings.TrimSpace(in.Type) == "" {
		in.Type = req.AccountType
	}

	account, err := h.service.Create(r.Context(), userID, in)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusCreated, account)
}

// list godoc
// @Summary List Accounts
// @Description List accounts for current authenticated user.
// @Tags accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Envelope{data=[]Account,meta=response.Meta}
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /accounts [get]
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

// get godoc
// @Summary Get Account
// @Description Get account details by account id for current authenticated user.
// @Tags accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Success 200 {object} response.Envelope{data=Account}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /accounts/{accountId} [get]
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	accountID := chi.URLParam(r, "accountId")
	if accountID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "accountId is required"))
		return
	}

	acc, err := h.service.Get(r.Context(), userID, accountID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, acc)
}

// delete godoc
// @Summary Delete Account
// @Description Soft-delete an account (owner-only).
// @Tags accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 403 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /accounts/{accountId} [delete]
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	accountID := chi.URLParam(r, "accountId")
	if accountID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "accountId is required"))
		return
	}

	if err := h.service.Delete(r.Context(), userID, accountID); err != nil {
		response.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
