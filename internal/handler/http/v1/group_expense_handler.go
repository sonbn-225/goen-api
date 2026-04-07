package v1
 
import (
	"encoding/json"
	"net/http"
	"strconv"
 
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
)
 
type GroupExpenseHandler struct {
	svc interfaces.GroupExpenseService
}
 
func NewGroupExpenseHandler(svc interfaces.GroupExpenseService) *GroupExpenseHandler {
	return &GroupExpenseHandler{svc: svc}
}
 
func (h *GroupExpenseHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(cfg))
 
		r.Route("/group-expenses", func(r chi.Router) {
			r.Post("/", h.Create)
			r.Get("/participants/{transactionId}", h.ListByTransaction)
			r.Post("/settle/{participantId}", h.Settle)
			r.Get("/names", h.ListNames)
		})
	})
}
 
// Create godoc
// @Summary Create Group Expense
// @Description Create a shared transaction splitting cost among participants
// @Tags GroupExpenses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateGroupExpenseRequest true "Group Expense Creation Payload"
// @Success 201 {object} response.SuccessEnvelope{data=dto.CreateGroupExpenseResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /group-expenses [post]
func (h *GroupExpenseHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}
 
	var req dto.CreateGroupExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "failed to decode request", nil)
		return
	}
 
	res, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
 
	response.WriteSuccess(w, http.StatusCreated, res)
}
 
// ListByTransaction godoc
// @Summary List Group Expense Participants
// @Description Retrieve participant portions by an existing transaction ID
// @Tags GroupExpenses
// @Produce json
// @Security BearerAuth
// @Param transactionId path string true "Transaction ID"
// @Success 200 {object} response.SuccessEnvelope{data=[]dto.GroupExpenseParticipantResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /group-expenses/participants/{transactionId} [get]
func (h *GroupExpenseHandler) ListByTransaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}
 
	txID, err := uuid.Parse(chi.URLParam(r, "transactionId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid transaction id format", nil)
		return
	}
 
	items, err := h.svc.ListByTransaction(r.Context(), userID, txID)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
 
	response.WriteSuccess(w, http.StatusOK, items)
}
 
// Settle godoc
// @Summary Settle Group Expense
// @Description Complete settlement of a specific group expense participant by ID
// @Tags GroupExpenses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param participantId path string true "Participant ID"
// @Param request body dto.GroupExpenseSettleRequest true "Settlement details"
// @Success 200 {object} response.SuccessEnvelope{data=dto.TransactionResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Router /group-expenses/settle/{participantId} [post]
func (h *GroupExpenseHandler) Settle(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}
 
	pID, err := uuid.Parse(chi.URLParam(r, "participantId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_id", "invalid participant id format", nil)
		return
	}
 
	var req dto.GroupExpenseSettleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid_request", "failed to decode request", nil)
		return
	}
 
	tx, err := h.svc.Settle(r.Context(), userID, pID, req)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
 
	response.WriteSuccess(w, http.StatusOK, tx)
}
 
// ListNames godoc
// @Summary List Group Expense Participant Names
// @Description Retrieve a list of unique names previously used in group expenses
// @Tags GroupExpenses
// @Produce json
// @Security BearerAuth
// @Param limit query integer false "Limit"
// @Success 200 {object} response.SuccessEnvelope{data=[]string}
// @Failure 401 {object} response.ErrorEnvelope
// @Router /group-expenses/names [get]
func (h *GroupExpenseHandler) ListNames(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "user not found in context", nil)
		return
	}
 
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}
 
	names, err := h.svc.ListUniqueParticipantNames(r.Context(), userID, limit)
	if err != nil {
		response.WriteInternalError(w, err)
		return
	}
 
	response.WriteSuccess(w, http.StatusOK, names)
}
