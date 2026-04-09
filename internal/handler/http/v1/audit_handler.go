package v1

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/handler/middleware"
	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
	"github.com/sonbn-225/goen-api/internal/pkg/config"
	"github.com/sonbn-225/goen-api/internal/pkg/response"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

type AuditHandler struct {
	svc interfaces.AuditService
}

func NewAuditHandler(svc interfaces.AuditService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

func (h *AuditHandler) RegisterRoutes(r chi.Router, cfg *config.Config) {
	r.Route("/audit", func(r chi.Router) {
		r.Get("/", h.List)
	})
}

// List trả về danh sách nhật ký kiểm soát với các bộ lọc.
// @Summary List Audit Logs
// @Description Fetch unified audit logs for the user with optional filters
// @Tags Audit
// @Security BearerAuth
// @Param resource_type query string false "Resource type (account, transaction, etc.)"
// @Param resource_id query string false "Specific resource UUID"
// @Param action query string false "Action type (created, updated, deleted)"
// @Param limit query int false "Max results"
// @Success 200 {array} dto.AuditLogResponse
// @Router /audit [get]
func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.HandleError(w, apperr.Unauthorized("unauthorized", "user not found in context"))
		return
	}

	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(q.Get("offset"))

	req := dto.AuditLogFilterRequest{
		ResourceType: (*entity.AuditResourceType)(utils.NormalizeOptionalString(ptr(q.Get("resource_type")))),
		Action:       (*entity.AuditAction)(utils.NormalizeOptionalString(ptr(q.Get("action")))),
		Limit:        limit,
		Offset:       offset,
	}

	if ridStr := q.Get("resource_id"); ridStr != "" {
		if id, err := uuid.Parse(ridStr); err == nil {
			req.ResourceID = &id
		}
	}
	if aidStr := q.Get("account_id"); aidStr != "" {
		if id, err := uuid.Parse(aidStr); err == nil {
			req.AccountID = &id
		}
	}
	if uidStr := q.Get("actor_user_id"); uidStr != "" {
		if id, err := uuid.Parse(uidStr); err == nil {
			req.ActorUserID = &id
		}
	}

	logs, err := h.svc.List(r.Context(), userID, req.ToDomain())
	if err != nil {
		response.HandleError(w, err)
		return
	}

	response.WriteSuccess(w, http.StatusOK, dto.NewAuditLogResponses(logs))
}

func ptr[T any](v T) *T {
	return &v
}
