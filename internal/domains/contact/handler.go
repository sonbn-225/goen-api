package contact

import (
	"net/http"

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

// list godoc
// @Summary List Contacts
// @Description List contacts for current authenticated user.
// @Tags contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Envelope{data=[]Contact,meta=response.Meta}
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /contacts [get]
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

// create godoc
// @Summary Create Contact
// @Description Create a new contact for current authenticated user.
// @Tags contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateContactRequest true "Create contact request"
// @Success 201 {object} response.Envelope{data=Contact}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /contacts [post]
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	var req CreateContactRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	created, err := h.service.Create(r.Context(), userID, CreateInput(req))
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusCreated, created)
}

// get godoc
// @Summary Get Contact
// @Description Get contact details by contact id for current authenticated user.
// @Tags contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contactId path string true "Contact ID"
// @Success 200 {object} response.Envelope{data=Contact}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /contacts/{contactId} [get]
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	contactID := chi.URLParam(r, "contactId")
	if contactID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "contactId is required"))
		return
	}

	item, err := h.service.Get(r.Context(), userID, contactID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, item)
}

// patch godoc
// @Summary Update Contact
// @Description Partially update contact details.
// @Tags contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contactId path string true "Contact ID"
// @Param request body PatchContactRequest true "Patch contact request"
// @Success 200 {object} response.Envelope{data=Contact}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /contacts/{contactId} [patch]
func (h *Handler) patch(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	contactID := chi.URLParam(r, "contactId")
	if contactID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "contactId is required"))
		return
	}

	var req PatchContactRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	updated, err := h.service.Update(r.Context(), userID, contactID, UpdateInput(req))
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, updated)
}

// delete godoc
// @Summary Delete Contact
// @Description Delete a contact by contact id.
// @Tags contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contactId path string true "Contact ID"
// @Success 200 {object} response.Envelope{data=map[string]any}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /contacts/{contactId} [delete]
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	contactID := chi.URLParam(r, "contactId")
	if contactID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "contactId is required"))
		return
	}

	if err := h.service.Delete(r.Context(), userID, contactID); err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, map[string]any{"deleted": true})
}
