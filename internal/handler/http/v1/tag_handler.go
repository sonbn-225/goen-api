package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
)

type TagHandler struct {
	svc interfaces.TagService
}

func NewTagHandler(svc interfaces.TagService) *TagHandler {
	return &TagHandler{svc: svc}
}

func (h *TagHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))
		r.Get("/tags", h.List)
		r.Post("/tags", h.Create)
		r.Get("/tags/{tagId}", h.Get)
	})
}

// List godoc
// @Summary List Tags
// @Description Retrieve all tags owned by the current user
// @Tags Tags
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessEnvelope{data=[]entity.Tag}
// @Failure 401 {object} response.ErrorEnvelope
// @Router /tags [get]
func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, items)
}

// Create godoc
// @Summary Create Tag
// @Description Create a new tag for labeling transactions
// @Tags Tags
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateTagRequest true "Tag Creation Payload"
// @Success 201 {object} response.SuccessEnvelope{data=entity.Tag}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /tags [post]
func (h *TagHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	var req dto.CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json body", nil)
		return
	}

	t, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusCreated, t)
}

// Get godoc
// @Summary Get Tag
// @Description Retrieve a specific tag by its ID
// @Tags Tags
// @Produce json
// @Security BearerAuth
// @Param tagId path string true "Tag ID"
// @Success 200 {object} response.SuccessEnvelope{data=entity.Tag}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /tags/{tagId} [get]
func (h *TagHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}

	id := chi.URLParam(r, "tagId")
	if id == "" {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "tagId is required", nil)
		return
	}

	t, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, t)
}
