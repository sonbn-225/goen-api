package v1

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
)

type AuthHandler struct {
	svc interfaces.AuthService
}

func NewAuthHandler(svc interfaces.AuthService) *AuthHandler {
	return &AuthHandler{
		svc: svc,
	}
}

func (h *AuthHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/signup", h.Signup)
		r.Post("/signin", h.Signin)
		r.Post("/refresh", h.Refresh)
		r.Post("/logout", h.Logout)
	})
}

// Signup godoc
// @Summary User Signup
// @Description Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.SignupRequest true "Signup request"
// @Success 201 {object} response.SuccessEnvelope{data=dto.AuthResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /auth/signup [post]
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req dto.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "failed to decode request", nil)
		return
	}

	res, err := h.svc.Signup(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusCreated, res)
}

// Signin godoc
// @Summary User Signin
// @Description Authenticate a user and return tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.SigninRequest true "Signin request"
// @Success 200 {object} response.SuccessEnvelope{data=dto.AuthResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /auth/signin [post]
func (h *AuthHandler) Signin(w http.ResponseWriter, r *http.Request) {
	var req dto.SigninRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "failed to decode request", nil)
		return
	}

	res, err := h.svc.Signin(r.Context(), req)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, res)
}

// Refresh godoc
// @Summary Refresh Token
// @Description Get a new access token using an existing session
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshRequest true "Refresh request"
// @Success 200 {object} response.SuccessEnvelope{data=dto.AuthResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "failed to decode request", nil)
		return
	}

	res, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, res)
}

// Logout godoc
// @Summary User Logout
// @Description Logout current user session on the server side
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshRequest true "Logout request"
// @Success 200 {object} response.SuccessEnvelope{data=map[string]string}
// @Failure 400 {object} response.ErrorEnvelope
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "failed to decode request", nil)
		return
	}

	if err := h.svc.Logout(r.Context(), req.RefreshToken); err != nil {
		response.HandleError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "logout successful"})
}
