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

// ListRotatingSavingsGroups godoc
// @Summary List rotating savings groups
// @Description List rotating savings groups owned by current user.
// @Tags rotating_savings
// @Produce json
// @Success 200 {array} domain.RotatingSavingsGroup
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /rotating-savings/groups [get]
func ListRotatingSavingsGroups(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		items, err := d.RotatingSavingsService.ListGroups(r.Context(), uid)
		if err != nil {
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// CreateRotatingSavingsGroup godoc
// @Summary Create rotating savings group
// @Description Create a new rotating savings group (hụi/họ) owned by current user.
// @Tags rotating_savings
// @Accept json
// @Produce json
// @Param body body services.CreateRotatingSavingsGroupRequest true "Create group request"
// @Success 200 {object} domain.RotatingSavingsGroup
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /rotating-savings/groups [post]
func CreateRotatingSavingsGroup(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		var req services.CreateRotatingSavingsGroupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
			return
		}

		item, err := d.RotatingSavingsService.CreateGroup(r.Context(), uid, req)
		if err != nil {
			apierror.Write(w, http.StatusBadRequest, "validation_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, item)
	}
}

// GetRotatingSavingsGroup godoc
// @Summary Get rotating savings group
// @Description Get a rotating savings group owned by current user.
// @Tags rotating_savings
// @Produce json
// @Param groupId path string true "Group ID"
// @Success 200 {object} domain.RotatingSavingsGroup
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /rotating-savings/groups/{groupId} [get]
func GetRotatingSavingsGroup(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		id := chi.URLParam(r, "groupId")
		if id == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "groupId is required", map[string]any{"field": "groupId"})
			return
		}

		item, err := d.RotatingSavingsService.GetGroup(r.Context(), uid, id)
		if err != nil {
			if errors.Is(err, domain.ErrRotatingSavingsGroupNotFound) {
				apierror.Write(w, http.StatusNotFound, "not_found", "group not found", nil)
				return
			}
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, item)
	}
}

// ListRotatingSavingsContributions godoc
// @Summary List rotating savings contributions
// @Description List contributions/payouts for a rotating savings group owned by current user.
// @Tags rotating_savings
// @Produce json
// @Param groupId path string true "Group ID"
// @Success 200 {array} domain.RotatingSavingsContribution
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /rotating-savings/groups/{groupId}/contributions [get]
func ListRotatingSavingsContributions(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		groupID := chi.URLParam(r, "groupId")
		if groupID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "groupId is required", map[string]any{"field": "groupId"})
			return
		}

		items, err := d.RotatingSavingsService.ListContributions(r.Context(), uid, groupID)
		if err != nil {
			if errors.Is(err, domain.ErrRotatingSavingsGroupNotFound) {
				apierror.Write(w, http.StatusNotFound, "not_found", "group not found", nil)
				return
			}
			apierror.Write(w, http.StatusInternalServerError, "internal_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// CreateRotatingSavingsContribution godoc
// @Summary Create rotating savings contribution
// @Description Create a contribution/payout. Server will create the underlying Transaction and link 1-1.
// @Tags rotating_savings
// @Accept json
// @Produce json
// @Param groupId path string true "Group ID"
// @Param body body services.CreateRotatingSavingsContributionRequest true "Create contribution request"
// @Success 200 {object} domain.RotatingSavingsContribution
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /rotating-savings/groups/{groupId}/contributions [post]
func CreateRotatingSavingsContribution(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			apierror.Write(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
			return
		}

		groupID := chi.URLParam(r, "groupId")
		if groupID == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "groupId is required", map[string]any{"field": "groupId"})
			return
		}

		var req services.CreateRotatingSavingsContributionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
			return
		}

		item, err := d.RotatingSavingsService.CreateContribution(r.Context(), uid, groupID, req)
		if err != nil {
			if errors.Is(err, domain.ErrRotatingSavingsGroupNotFound) {
				apierror.Write(w, http.StatusNotFound, "not_found", "group not found", nil)
				return
			}
			apierror.Write(w, http.StatusBadRequest, "validation_error", err.Error(), nil)
			return
		}

		writeJSON(w, http.StatusOK, item)
	}
}
