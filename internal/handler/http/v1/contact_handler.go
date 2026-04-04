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

type ContactHandler struct {
	svc interfaces.ContactService
}

func NewContactHandler(svc interfaces.ContactService) *ContactHandler {
	return &ContactHandler{svc: svc}
}

func (h *ContactHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))
		r.Get("/contacts", h.List)
		r.Post("/contacts", h.Create)
		r.Get("/contacts/{id}", h.Get)
		r.Patch("/contacts/{id}", h.Update)
		r.Delete("/contacts/{id}", h.Delete)
	})
}

func (h *ContactHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, items)
}

func (h *ContactHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	var req dto.CreateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}
	res, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusCreated, res)
}

func (h *ContactHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	res, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, res)
}

func (h *ContactHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	var req dto.UpdateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}
	res, err := h.svc.Update(r.Context(), userID, id, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, res)
}

func (h *ContactHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		response.WriteInternalError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
