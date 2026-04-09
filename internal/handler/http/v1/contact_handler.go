package v1
 
import (
	"encoding/json"
	"net/http"
 
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
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
 
		r.Route("/contacts", func(r chi.Router) {
			r.Get("/", h.List)
			r.Post("/", h.Create)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.Get)
				r.Patch("/", h.Update)
				r.Delete("/", h.Delete)
			})
		})
	})
}
 
// List godoc
// @Summary List Contacts
// @Description Retrieve a list of contacts for the current user
// @Tags Contacts
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.ContactResponse}
// @Failure 401 {object} response.ErrorEnvelope
// @Router /contacts [get]
func (h *ContactHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}
	items, err := h.svc.List(r.Context(), userID)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusOK, items)
}
 
// Create godoc
// @Summary Create Contact
// @Description Create a new contact, optionally linked to another Goen user
// @Tags Contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateContactRequest true "Contact Creation Payload"
// @Success 201 {object} response.SuccessEnvelope{data=dto.ContactResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /contacts [post]
func (h *ContactHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}
	var req dto.CreateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}
	res, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	response.WriteSuccess(w, http.StatusCreated, res)
}
 
// Get godoc
// @Summary Get Contact
// @Description Retrieve a specific contact by its ID
// @Tags Contacts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Contact ID"
// @Success 200 {object} response.SuccessEnvelope{data=dto.ContactResponse}
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /contacts/{id} [get]
func (h *ContactHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid contact id format", nil)
		return
	}
	res, err := h.svc.Get(r.Context(), userID, id)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	if res == nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "contact not found", nil)
		return
	}
	response.WriteSuccess(w, http.StatusOK, res)
}
 
// Update godoc
// @Summary Update Contact
// @Description Partially update specific contact properties
// @Tags Contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Contact ID"
// @Param request body dto.UpdateContactRequest true "Contact Update Payload"
// @Success 200 {object} response.SuccessEnvelope{data=dto.ContactResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /contacts/{id} [patch]
func (h *ContactHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid contact id format", nil)
		return
	}
	var req dto.UpdateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}
	res, err := h.svc.Update(r.Context(), userID, id, req)
	if err != nil {
		response.HandleError(w, err)
		return
	}
	if res == nil {
		response.WriteError(w, http.StatusNotFound, "not_found", "contact not found", nil)
		return
	}
	response.WriteSuccess(w, http.StatusOK, res)
}
 
// Delete godoc
// @Summary Delete Contact
// @Description Standardized soft-delete or link removal for a contact
// @Tags Contacts
// @Security BearerAuth
// @Param id path string true "Contact ID"
// @Success 204 "No Content"
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Router /contacts/{id} [delete]
func (h *ContactHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid contact id format", nil)
		return
	}
	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		response.HandleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
