package dto

import (
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

func NewAccountShareResponse(it entity.AccountShare) AccountShareResponse {
	return AccountShareResponse{
		ID:              it.ID,
		AccountID:       it.AccountID,
		UserID:          it.UserID,
		Permission:      it.Permission,
		Status:          it.Status,
		RevokedAt:       it.RevokedAt,
		CreatedAt:       it.CreatedAt,
		UpdatedAt:       it.UpdatedAt,
		UserEmail:       it.UserEmail,
		UserPhone:       it.UserPhone,
		UserDisplayName: it.UserDisplayName,
	}
}

func NewAccountShareResponses(items []entity.AccountShare) []AccountShareResponse {
	out := make([]AccountShareResponse, len(items))
	for i, it := range items {
		out[i] = NewAccountShareResponse(it)
	}
	return out
}
