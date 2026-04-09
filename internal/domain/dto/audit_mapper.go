package dto

import (
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

func NewAuditLogResponse(l entity.AuditLog) AuditLogResponse {
	return AuditLogResponse{
		ID:           l.ID,
		OccurredAt:   l.OccurredAt,
		ActorUserID:  l.ActorUserID,
		AccountID:    l.AccountID,
		ResourceType: l.ResourceType,
		ResourceID:   l.ResourceID,
		Action:       l.Action,
		Metadata:     l.Metadata,
	}
}

func NewAuditLogResponses(logs []entity.AuditLog) []AuditLogResponse {
	resps := make([]AuditLogResponse, len(logs))
	for i, l := range logs {
		resps[i] = NewAuditLogResponse(l)
	}
	return resps
}

func (r AuditLogFilterRequest) ToDomain() entity.AuditLogFilter {
	return entity.AuditLogFilter{
		ActorUserID:  r.ActorUserID,
		AccountID:    r.AccountID,
		ResourceType: r.ResourceType,
		ResourceID:   r.ResourceID,
		Action:       r.Action,
		From:         r.From,
		To:           r.To,
		Limit:        r.Limit,
		Offset:       r.Offset,
	}
}
