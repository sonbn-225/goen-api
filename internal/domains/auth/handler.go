package auth

import (
	"net/http"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/httpx"
	"github.com/sonbn-225/goen-api-v2/internal/core/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// signup godoc
// @Summary Signup
// @Description Create a new user and return access token.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SignupRequest true "Signup request"
// @Success 200 {object} response.Envelope{data=AuthResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 409 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /auth/signup [post]
func (h *Handler) signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	result, err := h.service.Signup(r.Context(), req)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.WriteData(w, http.StatusOK, result)
}

// signin godoc
// @Summary Signin
// @Description Authenticate by email/phone/username and return access token.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SigninRequest true "Signin request"
// @Success 200 {object} response.Envelope{data=AuthResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /auth/signin [post]
func (h *Handler) signin(w http.ResponseWriter, r *http.Request) {
	var req SigninRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	result, err := h.service.Signin(r.Context(), req)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.WriteData(w, http.StatusOK, result)
}

// register godoc
// @Summary Register (legacy)
// @Description Legacy register endpoint for backward compatibility.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register request"
// @Success 201 {object} response.Envelope{data=AuthResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 409 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /auth/register [post]
func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	result, err := h.service.Signup(r.Context(), SignupRequest{Email: req.Email, Password: req.Password})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.WriteData(w, http.StatusCreated, result)
}

// login godoc
// @Summary Login (legacy)
// @Description Legacy login endpoint for backward compatibility.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} response.Envelope{data=AuthResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /auth/login [post]
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	result, err := h.service.Signin(r.Context(), SigninRequest{Login: req.Email, Password: req.Password})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.WriteData(w, http.StatusOK, result)
}

// refresh godoc
// @Summary Refresh Access Token
// @Description Refresh access token for current authenticated user.
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Envelope{data=AuthResponse}
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /auth/refresh [post]
func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	result, err := h.service.Refresh(r.Context(), userID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.WriteData(w, http.StatusOK, result)
}
