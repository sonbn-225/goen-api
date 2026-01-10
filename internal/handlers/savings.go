package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/services"
)

// ListSavingsInstruments godoc
// @Summary List savings instruments
// @Description List savings instruments accessible to current user.
// @Tags savings
// @Produce json
// @Success 200 {array} domain.SavingsInstrument
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /savings/instruments [get]
func ListSavingsInstruments(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		items, err := d.SavingsService.ListInstruments(r.Context(), uid)
		if err != nil {
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// CreateSavingsInstrument godoc
// @Summary Create savings instrument
// @Description Create a new SavingsInstrument (1-1 extension of a savings Account).
// @Tags savings
// @Accept json
// @Produce json
// @Param body body services.CreateSavingsInstrumentRequest true "Create savings instrument request"
// @Success 200 {object} domain.SavingsInstrument
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /savings/instruments [post]
func CreateSavingsInstrument(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		var req services.CreateSavingsInstrumentRequest
		if ok := decodeJSON(w, r, &req); !ok {
			return
		}

		item, err := d.SavingsService.CreateInstrument(r.Context(), uid, req)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	}
}

// GetSavingsInstrument godoc
// @Summary Get savings instrument
// @Description Get a single SavingsInstrument accessible to current user.
// @Tags savings
// @Produce json
// @Param instrumentId path string true "Savings instrument ID"
// @Success 200 {object} domain.SavingsInstrument
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /savings/instruments/{instrumentId} [get]
func GetSavingsInstrument(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		id := chi.URLParam(r, "instrumentId")
		if id == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "instrumentId is required", map[string]any{"field": "instrumentId"})
			return
		}

		item, err := d.SavingsService.GetInstrument(r.Context(), uid, id)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, item)
	}
}

// PatchSavingsInstrument godoc
// @Summary Patch savings instrument
// @Description Patch fields of an existing SavingsInstrument accessible to current user.
// @Tags savings
// @Accept json
// @Produce json
// @Param instrumentId path string true "Savings instrument ID"
// @Param body body services.PatchSavingsInstrumentRequest true "Patch savings instrument request"
// @Success 200 {object} domain.SavingsInstrument
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /savings/instruments/{instrumentId} [patch]
func PatchSavingsInstrument(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		id := chi.URLParam(r, "instrumentId")
		if id == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "instrumentId is required", map[string]any{"field": "instrumentId"})
			return
		}

		var req services.PatchSavingsInstrumentRequest
		if ok := decodeJSON(w, r, &req); !ok {
			return
		}

		item, err := d.SavingsService.PatchInstrument(r.Context(), uid, id, req)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	}
}

// DeleteSavingsInstrument godoc
// @Summary Delete savings instrument
// @Description Delete an existing SavingsInstrument accessible to current user.
// @Tags savings
// @Produce json
// @Param instrumentId path string true "Savings instrument ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /savings/instruments/{instrumentId} [delete]
func DeleteSavingsInstrument(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		id := chi.URLParam(r, "instrumentId")
		if id == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "instrumentId is required", map[string]any{"field": "instrumentId"})
			return
		}

		if err := d.SavingsService.DeleteInstrument(r.Context(), uid, id); err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"deleted": id})
	}
}
