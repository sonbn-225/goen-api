package savings

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
// @Summary List Savings Instruments
// @Description List savings instruments accessible to current authenticated user.
// @Tags savings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Envelope{data=[]SavingsInstrument}
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /savings/instruments [get]
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

	response.WriteData(w, http.StatusOK, items)
}

// create godoc
// @Summary Create Savings Instrument
// @Description Create a new savings instrument.
// @Tags savings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateSavingsInstrumentRequest true "Create savings instrument request"
// @Success 201 {object} response.Envelope{data=SavingsInstrument}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /savings/instruments [post]
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	var req CreateSavingsInstrumentRequest
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
// @Summary Get Savings Instrument
// @Description Get a savings instrument by id.
// @Tags savings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param instrumentId path string true "Savings instrument ID"
// @Success 200 {object} response.Envelope{data=SavingsInstrument}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /savings/instruments/{instrumentId} [get]
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	instrumentID := chi.URLParam(r, "instrumentId")
	if instrumentID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "instrumentId is required"))
		return
	}

	item, err := h.service.Get(r.Context(), userID, instrumentID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, item)
}

// patch godoc
// @Summary Patch Savings Instrument
// @Description Patch mutable fields of savings instrument.
// @Tags savings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param instrumentId path string true "Savings instrument ID"
// @Param body body PatchSavingsInstrumentRequest true "Patch savings instrument request"
// @Success 200 {object} response.Envelope{data=SavingsInstrument}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /savings/instruments/{instrumentId} [patch]
func (h *Handler) patch(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	instrumentID := chi.URLParam(r, "instrumentId")
	if instrumentID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "instrumentId is required"))
		return
	}

	var req PatchSavingsInstrumentRequest
	if err := httpx.DecodeJSONAllowEmpty(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	updated, err := h.service.Patch(r.Context(), userID, instrumentID, PatchInput(req))
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, updated)
}

// delete godoc
// @Summary Delete Savings Instrument
// @Description Delete savings instrument by id.
// @Tags savings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param instrumentId path string true "Savings instrument ID"
// @Success 200 {object} response.Envelope{data=map[string]string}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /savings/instruments/{instrumentId} [delete]
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	instrumentID := chi.URLParam(r, "instrumentId")
	if instrumentID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "instrumentId is required"))
		return
	}

	if err := h.service.Delete(r.Context(), userID, instrumentID); err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteData(w, http.StatusOK, map[string]string{"deleted": instrumentID})
}
