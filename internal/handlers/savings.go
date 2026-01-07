package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/auth"
	"github.com/sonbn-225/goen-api/internal/domain"
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
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		items, err := d.SavingsService.ListInstruments(r.Context(), uid)
		if err != nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
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
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		var req services.CreateSavingsInstrumentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
			return
		}

		item, err := d.SavingsService.CreateInstrument(r.Context(), uid, req)
		if err != nil {
			apierror.Write(w, http.StatusBadRequest, "validation_error", err.Error(), nil)
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
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		id := chi.URLParam(r, "instrumentId")
		if id == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "instrumentId is required", map[string]any{"field": "instrumentId"})
			return
		}

		item, err := d.SavingsService.GetInstrument(r.Context(), uid, id)
		if err != nil {
			if errors.Is(err, domain.ErrSavingsInstrumentNotFound) {
				apierror.Write(w, http.StatusNotFound, "not_found", "savings instrument not found", nil)
				return
			}
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, item)
	}
}
