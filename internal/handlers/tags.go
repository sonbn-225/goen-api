package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/apierror"
	"github.com/sonbn-225/goen-api/internal/services"
)

// ListTags godoc
// @Summary List tags
// @Description List tags owned by current user.
// @Tags tags
// @Produce json
// @Success 200 {array} domain.Tag
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /tags [get]
func ListTags(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		items, err := d.TagService.List(r.Context(), uid)
		if err != nil {
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// CreateTag godoc
// @Summary Create tag
// @Description Create a new tag owned by current user.
// @Tags tags
// @Accept json
// @Produce json
// @Param body body services.CreateTagRequest true "Create tag request"
// @Success 200 {object} domain.Tag
// @Failure 400 {object} apierror.Envelope
// @Failure 401 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /tags [post]
func CreateTag(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		var req services.CreateTagRequest
		if ok := decodeJSON(w, r, &req); !ok {
			return
		}

		t, err := d.TagService.Create(r.Context(), uid, req)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, t)
	}
}

// GetTag godoc
// @Summary Get tag
// @Description Get a single tag owned by current user.
// @Tags tags
// @Produce json
// @Param tagId path string true "Tag ID"
// @Success 200 {object} domain.Tag
// @Failure 401 {object} apierror.Envelope
// @Failure 404 {object} apierror.Envelope
// @Failure 500 {object} apierror.Envelope
// @Router /tags/{tagId} [get]
func GetTag(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := requireUserID(w, r)
		if !ok {
			return
		}

		id := chi.URLParam(r, "tagId")
		if id == "" {
			apierror.Write(w, http.StatusBadRequest, "validation_error", "tagId is required", map[string]any{"field": "tagId"})
			return
		}

		t, err := d.TagService.Get(r.Context(), uid, id)
		if err != nil {
			if writeServiceError(w, err) {
				return
			}
			writeInternalError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, t)
	}
}
