package rotatingsavings

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
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

// listGroups godoc
// @Summary List rotating savings groups
// @Description List rotating savings groups owned by current user.
// @Tags rotating_savings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Envelope{data=[]GroupSummary}
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/groups [get]
func (h *Handler) listGroups(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	items, err := h.service.ListGroups(r.Context(), userID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.WriteData(w, http.StatusOK, items)
}

// createGroup godoc
// @Summary Create rotating savings group
// @Description Create a new rotating savings group.
// @Tags rotating_savings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateGroupRequest true "Create rotating savings group request"
// @Success 200 {object} response.Envelope{data=RotatingSavingsGroup}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/groups [post]
func (h *Handler) createGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}

	var req CreateGroupRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	contributionAmount, err := parseFlexibleFloat(req.ContributionAmount)
	if err != nil {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, fmt.Sprintf("contribution_amount %s", err.Error())))
		return
	}

	var fixedInterest *float64
	if req.FixedInterestAmount != nil {
		v, err := parseFlexibleFloat(req.FixedInterestAmount)
		if err != nil {
			response.WriteError(w, apperrors.New(apperrors.KindValidation, fmt.Sprintf("fixed_interest_amount %s", err.Error())))
			return
		}
		fixedInterest = &v
	}

	created, err := h.service.CreateGroup(r.Context(), userID, CreateGroupInput{
		AccountID:           req.AccountID,
		Name:                req.Name,
		MemberCount:         req.MemberCount,
		UserSlots:           req.UserSlots,
		ContributionAmount:  contributionAmount,
		FixedInterestAmount: fixedInterest,
		CycleFrequency:      req.CycleFrequency,
		StartDate:           req.StartDate,
		Status:              req.Status,
	})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.WriteData(w, http.StatusOK, created)
}

// getGroup godoc
// @Summary Get rotating savings group detail
// @Description Get rotating savings group detail with schedule, contributions, and audit logs.
// @Tags rotating_savings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param groupId path string true "Group ID"
// @Success 200 {object} response.Envelope{data=GroupDetailResponse}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/groups/{groupId} [get]
func (h *Handler) getGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}
	groupID := strings.TrimSpace(chi.URLParam(r, "groupId"))
	if groupID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "groupId is required"))
		return
	}

	item, err := h.service.GetGroupDetail(r.Context(), userID, groupID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.WriteData(w, http.StatusOK, item)
}

// updateGroup godoc
// @Summary Update rotating savings group
// @Description Update mutable fields of rotating savings group.
// @Tags rotating_savings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param groupId path string true "Group ID"
// @Param body body UpdateGroupRequest true "Update rotating savings group request"
// @Success 200 {object} response.Envelope{data=RotatingSavingsGroup}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/groups/{groupId} [patch]
func (h *Handler) updateGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}
	groupID := strings.TrimSpace(chi.URLParam(r, "groupId"))
	if groupID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "groupId is required"))
		return
	}

	var req UpdateGroupRequest
	if err := httpx.DecodeJSONAllowEmpty(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	updated, err := h.service.UpdateGroup(r.Context(), userID, groupID, UpdateGroupInput(req))
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.WriteData(w, http.StatusOK, updated)
}

// listContributions godoc
// @Summary List rotating savings contributions
// @Description List contributions/payouts of a rotating savings group.
// @Tags rotating_savings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param groupId path string true "Group ID"
// @Success 200 {object} response.Envelope{data=[]RotatingSavingsContribution}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/groups/{groupId}/contributions [get]
func (h *Handler) listContributions(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}
	groupID := strings.TrimSpace(chi.URLParam(r, "groupId"))
	if groupID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "groupId is required"))
		return
	}

	items, err := h.service.ListContributions(r.Context(), userID, groupID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.WriteData(w, http.StatusOK, items)
}

// createContribution godoc
// @Summary Create rotating savings contribution
// @Description Create a contribution or payout for rotating savings group.
// @Tags rotating_savings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param groupId path string true "Group ID"
// @Param body body CreateContributionRequest true "Create rotating savings contribution request"
// @Success 200 {object} response.Envelope{data=RotatingSavingsContribution}
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/groups/{groupId}/contributions [post]
func (h *Handler) createContribution(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}
	groupID := strings.TrimSpace(chi.URLParam(r, "groupId"))
	if groupID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "groupId is required"))
		return
	}

	var req CreateContributionRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, apperrors.Wrap(apperrors.KindValidation, "invalid request body", err))
		return
	}

	amount, err := parseFlexibleFloat(req.Amount)
	if err != nil {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, fmt.Sprintf("amount %s", err.Error())))
		return
	}
	feePerSlot, err := parseFlexibleFloat(req.CollectedFeePerSlot)
	if err != nil {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, fmt.Sprintf("collected_fee_per_slot %s", err.Error())))
		return
	}

	created, err := h.service.CreateContribution(r.Context(), userID, groupID, CreateContributionInput{
		Kind:                req.Kind,
		AccountID:           req.AccountID,
		OccurredDate:        req.OccurredDate,
		OccurredTime:        req.OccurredTime,
		Amount:              amount,
		SlotsTaken:          req.SlotsTaken,
		CollectedFeePerSlot: feePerSlot,
		CycleNo:             req.CycleNo,
		DueDate:             req.DueDate,
		Note:                req.Note,
	})
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.WriteData(w, http.StatusOK, created)
}

// deleteContribution godoc
// @Summary Delete rotating savings contribution
// @Description Delete a contribution/payout from rotating savings group.
// @Tags rotating_savings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param groupId path string true "Group ID"
// @Param contributionId path string true "Contribution ID"
// @Success 204 {string} string ""
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/groups/{groupId}/contributions/{contributionId} [delete]
func (h *Handler) deleteContribution(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}
	groupID := strings.TrimSpace(chi.URLParam(r, "groupId"))
	contributionID := strings.TrimSpace(chi.URLParam(r, "contributionId"))
	if groupID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "groupId is required"))
		return
	}
	if contributionID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "contributionId is required"))
		return
	}

	if err := h.service.DeleteContribution(r.Context(), userID, groupID, contributionID); err != nil {
		response.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// deleteGroup godoc
// @Summary Delete rotating savings group
// @Description Delete rotating savings group and linked contributions.
// @Tags rotating_savings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param groupId path string true "Group ID"
// @Success 204 {string} string ""
// @Failure 400 {object} response.ErrorEnvelope
// @Failure 401 {object} response.ErrorEnvelope
// @Failure 404 {object} response.ErrorEnvelope
// @Failure 500 {object} response.ErrorEnvelope
// @Router /rotating-savings/groups/{groupId} [delete]
func (h *Handler) deleteGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpx.UserIDFromContext(r.Context())
	if !ok {
		response.WriteError(w, apperrors.New(apperrors.KindUnauth, "unauthorized"))
		return
	}
	groupID := strings.TrimSpace(chi.URLParam(r, "groupId"))
	if groupID == "" {
		response.WriteError(w, apperrors.New(apperrors.KindValidation, "groupId is required"))
		return
	}

	if err := h.service.DeleteGroup(r.Context(), userID, groupID); err != nil {
		response.WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseFlexibleFloat(raw any) (float64, error) {
	if raw == nil {
		return 0, fmt.Errorf("is required")
	}
	switch v := raw.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return 0, fmt.Errorf("is invalid")
		}
		return f, nil
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return 0, fmt.Errorf("is required")
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, fmt.Errorf("is invalid")
		}
		return f, nil
	default:
		return 0, fmt.Errorf("is invalid")
	}
}
