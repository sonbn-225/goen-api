package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
)

type InterchangeHandler struct {
	svc interfaces.InterchangeService
}

func NewInterchangeHandler(svc interfaces.InterchangeService) *InterchangeHandler {
	return &InterchangeHandler{svc: svc}
}

func (h *InterchangeHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))

		r.Route("/imports/{resourceType}", func(r chi.Router) {
			r.Post("/stage", h.StageImport)
			r.Get("/staged", h.ListStaged)
			r.Delete("/staged", h.ClearStaged)
			r.Post("/apply-rules-create", h.ApplyRulesAndCreate)
			r.Route("/staged/{id}", func(r chi.Router) {
				r.Patch("/", h.PatchStaged)
				r.Delete("/", h.DeleteStaged)
			})
			r.Post("/staged/create-many", h.CreateManyFromStaged)
			r.Route("/rules", func(r chi.Router) {
				r.Get("/", h.ListRules)
				r.Post("/", h.UpsertRules)
				r.Delete("/{id}", h.DeleteRule)
			})
		})

		r.Route("/exports/{resourceType}", func(r chi.Router) {
			r.Get("/", h.Export)
		})
	})
}

// StageImport godoc
// @Summary Stage generic data for manual mapping
func (h *InterchangeHandler) StageImport(w http.ResponseWriter, r *http.Request) {
	resourceType := chi.URLParam(r, "resourceType")
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}

	var req dto.StageImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
		return
	}

	stagedCount, skipped, errors, err := h.svc.StageImport(r.Context(), userID, resourceType, req.Source, req.Items)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, map[string]any{
		"staged_count":  stagedCount,
		"skipped_count": skipped,
		"errors":        errors,
	})
}

// ListStaged godoc
func (h *InterchangeHandler) ListStaged(w http.ResponseWriter, r *http.Request) {
	resourceType := chi.URLParam(r, "resourceType")
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}

	data, err := h.svc.ListStaged(r.Context(), userID, resourceType)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, data)
}

// PatchStaged godoc
func (h *InterchangeHandler) PatchStaged(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid id format", nil)
		return
	}

	var req dto.PatchStagedImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
		return
	}

	res, err := h.svc.PatchStaged(r.Context(), userID, id, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, res)
}

// DeleteStaged godoc
func (h *InterchangeHandler) DeleteStaged(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid id format", nil)
		return
	}

	if err := h.svc.DeleteStaged(r.Context(), userID, id); err != nil {
		response.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ClearStaged godoc
func (h *InterchangeHandler) ClearStaged(w http.ResponseWriter, r *http.Request) {
	resourceType := chi.URLParam(r, "resourceType")
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}

	if err := h.svc.ClearStaged(r.Context(), userID, resourceType); err != nil {
		response.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListRules godoc
func (h *InterchangeHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	resourceType := chi.URLParam(r, "resourceType")
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}

	data, err := h.svc.ListRules(r.Context(), userID, resourceType)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, data)
}

// UpsertRules godoc
func (h *InterchangeHandler) UpsertRules(w http.ResponseWriter, r *http.Request) {
	resourceType := chi.URLParam(r, "resourceType")
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}

	var req dto.UpsertMappingRulesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
		return
	}

	res, err := h.svc.UpsertRules(r.Context(), userID, resourceType, req.Rules)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, res)
}

// DeleteRule godoc
func (h *InterchangeHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid id format", nil)
		return
	}

	if err := h.svc.DeleteRule(r.Context(), userID, id); err != nil {
		response.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ApplyRulesAndCreate godoc
func (h *InterchangeHandler) ApplyRulesAndCreate(w http.ResponseWriter, r *http.Request) {
	resourceType := chi.URLParam(r, "resourceType")
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}

	res, err := h.svc.ApplyRulesAndCreate(r.Context(), userID, resourceType)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, res)
}

// CreateManyFromStaged godoc
func (h *InterchangeHandler) CreateManyFromStaged(w http.ResponseWriter, r *http.Request) {
	resourceType := chi.URLParam(r, "resourceType")
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}

	var req dto.CreateManyImportedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "invalid json", nil)
		return
	}

	res, err := h.svc.CreateManyFromStaged(r.Context(), userID, resourceType, req.IDs)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, res)
}

// Export godoc
func (h *InterchangeHandler) Export(w http.ResponseWriter, r *http.Request) {
	resourceType := chi.URLParam(r, "resourceType")
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("user not found in context"))
		return
	}

	// Parse filters from query params if needed
	// For now, let's just pass empty filter
	data, filename, err := h.svc.ExportToCSV(r.Context(), userID, resourceType, nil)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Write(data)
}
