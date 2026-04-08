package dto

import (
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

func NewAccountAuditEventResponse(it entity.AccountAuditEvent) AccountAuditEventResponse {
	return AccountAuditEventResponse{
		ID:          it.ID,
		AccountID:   it.AccountID,
		ActorUserID: it.ActorUserID,
		Action:      it.Action,
		EntityType:  it.EntityType,
		EntityID:    it.EntityID,
		OccurredAt:  it.OccurredAt,
		Diff:        it.Diff,
	}
}

func NewAccountAuditEventResponses(items []entity.AccountAuditEvent) []AccountAuditEventResponse {
	out := make([]AccountAuditEventResponse, len(items))
	for i, it := range items {
		out[i] = NewAccountAuditEventResponse(it)
	}
	return out
}
