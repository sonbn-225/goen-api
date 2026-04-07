package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
)

type CreateAccountRequest struct {
	Name            string  `json:"name"`
	AccountNumber   *string `json:"account_number,omitempty"`
	Color           *string `json:"color,omitempty"`
	AccountType     string  `json:"account_type"`
	Currency        string  `json:"currency"`
	ParentAccountID *string `json:"parent_account_id,omitempty"`
}

type UpsertShareRequest struct {
	Login      string `json:"login"`
	Permission string `json:"permission"`
}

type PatchAccountRequest struct {
	Name   *string `json:"name,omitempty"`
	Color  *string `json:"color,omitempty"`
	Status *string `json:"status,omitempty"`
}

type AccountResponse struct {
	ID                  uuid.UUID  `json:"id"`
	Name                string     `json:"name"`
	AccountNumber       *string    `json:"account_number,omitempty"`
	Color               *string    `json:"color,omitempty"`
	AccountType         string     `json:"account_type"`
	Currency            string     `json:"currency"`
	ParentAccountID     *uuid.UUID `json:"parent_account_id,omitempty"`
	Status              string     `json:"status"`
	Balance             string     `json:"balance"`
	InvestmentAccountID *uuid.UUID `json:"investment_account_id,omitempty"`
}

func NewAccountResponse(it entity.Account) AccountResponse {
	return AccountResponse{
		ID:                  it.ID,
		Name:                it.Name,
		AccountNumber:       it.AccountNumber,
		Color:               it.Color,
		AccountType:         it.AccountType,
		Currency:            it.Currency,
		ParentAccountID:     it.ParentAccountID,
		Status:              it.Status,
		Balance:             it.Balance,
		InvestmentAccountID: it.InvestmentAccountID,
	}
}

func NewAccountResponses(items []entity.Account) []AccountResponse {
	out := make([]AccountResponse, len(items))
	for i, it := range items {
		out[i] = NewAccountResponse(it)
	}
	return out
}

type AccountBalanceResponse struct {
	AccountID uuid.UUID `json:"account_id"`
	Currency  string    `json:"currency"`
	Balance   string    `json:"balance"`
}

func NewAccountBalanceResponse(it entity.AccountBalance) AccountBalanceResponse {
	return AccountBalanceResponse{
		AccountID: it.AccountID,
		Currency:  it.Currency,
		Balance:   it.Balance,
	}
}

func NewAccountBalanceResponses(items []entity.AccountBalance) []AccountBalanceResponse {
	out := make([]AccountBalanceResponse, len(items))
	for i, it := range items {
		out[i] = NewAccountBalanceResponse(it)
	}
	return out
}

type AccountShareResponse struct {
	ID              uuid.UUID  `json:"id"`
	AccountID       uuid.UUID  `json:"account_id"`
	UserID          uuid.UUID  `json:"user_id"`
	Permission      string     `json:"permission"`
	Status          string     `json:"status"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	UserEmail       *string    `json:"user_email,omitempty"`
	UserPhone       *string    `json:"user_phone,omitempty"`
	UserDisplayName *string    `json:"user_display_name,omitempty"`
}

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

type AccountAuditEventResponse struct {
	ID          uuid.UUID      `json:"id"`
	AccountID   uuid.UUID      `json:"account_id"`
	ActorUserID uuid.UUID      `json:"actor_user_id"`
	Action      string         `json:"action"`
	EntityType  string         `json:"entity_type"`
	EntityID    uuid.UUID      `json:"entity_id"`
	OccurredAt  time.Time      `json:"occurred_at"`
	Diff        map[string]any `json:"diff,omitempty"`
}

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
