package tag

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/httpapi"
	"github.com/sonbn-225/goen-api/internal/response"
	"github.com/sonbn-225/goen-api/internal/apperrors"
)

// Handler handles HTTP requests for tags.
type Handler struct {
	svc *Service
}

// NewHandler creates a new tag handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers all tag routes.
func (h *Handler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.With(authMiddleware).Get("/tags", h.List)
	r.With(authMiddleware).Get("/tags/", h.List)
	r.With(authMiddleware).Post("/tags", h.Create)
	r.With(authMiddleware).Post("/tags/", h.Create)
	r.With(authMiddleware).Get("/tags/{tagId}", h.Get)
}

// List handles GET /tags
// @Summary List tags
// @Description List tags owned by current user.
// @Tags tags
// @Produce json
// @Success 200 {array} domain.Tag
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /tags [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, items)
}

// Create handles POST /tags
// @Summary Create tag
// @Description Create a new tag.
// @Tags tags
// @Accept json
// @Produce json
// @Param body body CreateTagRequest true "Create tag request"
// @Success 200 {object} domain.Tag
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /tags [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	var req CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	t, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, t)
}

// Get handles GET /tags/{tagId}
// @Summary Get tag
// @Description Get a single tag.
// @Tags tags
// @Produce json
// @Param tagId path string true "Tag ID"
// @Success 200 {object} domain.Tag
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /tags/{tagId} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized", nil)
		return
	}

	id := chi.URLParam(r, "tagId")
	if id == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "tagId is required", map[string]any{"field": "tagId"})
		return
	}

	t, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, t)
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	var se *apperrors.Error
	if errors.As(err, &se) {
		response.WriteError(w, se.HTTPStatus(), string(se.Kind), se.Message, se.Details)
		return
	}
	response.WriteInternalError(w, err)
}
