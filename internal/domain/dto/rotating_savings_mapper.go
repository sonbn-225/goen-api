package dto

import (
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

func NewRotatingSavingsGroupResponse(g entity.RotatingSavingsGroup) RotatingSavingsGroupResponse {
	return RotatingSavingsGroupResponse{
		ID:                  g.ID,
		UserID:              g.UserID,
		AccountID:           g.AccountID,
		Name:                g.Name,
		Currency:            g.Currency,
		MemberCount:         g.MemberCount,
		UserSlots:           g.UserSlots,
		ContributionAmount:  g.ContributionAmount,
		PayoutCycleNo:       g.PayoutCycleNo,
		FixedInterestAmount: g.FixedInterestAmount,
		CycleFrequency:      g.CycleFrequency,
		StartDate:           g.StartDate,
		Status:              g.Status,
	}
}

func NewRotatingSavingsGroupResponses(items []entity.RotatingSavingsGroup) []RotatingSavingsGroupResponse {
	out := make([]RotatingSavingsGroupResponse, len(items))
	for i, it := range items {
		out[i] = NewRotatingSavingsGroupResponse(it)
	}
	return out
}

func NewRotatingSavingsContributionResponse(c entity.RotatingSavingsContribution) RotatingSavingsContributionResponse {
	return RotatingSavingsContributionResponse{
		ID:                  c.ID,
		GroupID:             c.GroupID,
		TransactionID:       c.TransactionID,
		Kind:                c.Kind,
		CycleNo:             c.CycleNo,
		DueDate:             c.DueDate,
		Amount:              c.Amount,
		SlotsTaken:          c.SlotsTaken,
		CollectedFeePerSlot: c.CollectedFeePerSlot,
		Note:                c.Note,
	}
}

func NewRotatingSavingsContributionResponses(items []entity.RotatingSavingsContribution) []RotatingSavingsContributionResponse {
	out := make([]RotatingSavingsContributionResponse, len(items))
	for i, it := range items {
		out[i] = NewRotatingSavingsContributionResponse(it)
	}
	return out
}

func NewRotatingSavingsAuditLogResponse(l entity.RotatingSavingsAuditLog) RotatingSavingsAuditLogResponse {
	return RotatingSavingsAuditLogResponse{
		ID:        l.ID,
		UserID:    l.UserID,
		GroupID:   l.GroupID,
		Action:    l.Action,
		Details:   l.Details,
		CreatedAt: l.CreatedAt,
	}
}

func NewRotatingSavingsAuditLogResponses(items []entity.RotatingSavingsAuditLog) []RotatingSavingsAuditLogResponse {
	out := make([]RotatingSavingsAuditLogResponse, len(items))
	for i, it := range items {
		out[i] = NewRotatingSavingsAuditLogResponse(it)
	}
	return out
}

func NewRotatingSavingsAuditLogResponsesFromUnified(items []entity.AuditLog) []RotatingSavingsAuditLogResponse {
	out := make([]RotatingSavingsAuditLogResponse, len(items))
	for i, l := range items {
		resourceID := l.ResourceID
		out[i] = RotatingSavingsAuditLogResponse{
			ID:        l.ID,
			UserID:    l.ActorUserID,
			GroupID:   &resourceID,
			Action:    entity.RotatingSavingsAuditAction(l.Action),
			Details:   l.Metadata,
			CreatedAt: l.OccurredAt,
		}
	}
	return out
}
