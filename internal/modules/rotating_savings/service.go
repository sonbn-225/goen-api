package rotatingsavings

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/domain"
	"github.com/sonbn-225/goen-api/internal/httpapi"
	"github.com/sonbn-225/goen-api/internal/i18n"
)

// TxCreateRequest is a local representation of transaction create request.
type TxCreateRequest struct {
	Type         string  `json:"type"`
	OccurredDate *string `json:"occurred_date,omitempty"`
	OccurredTime *string `json:"occurred_time,omitempty"`
	Amount       string  `json:"amount"`
	Description  *string `json:"description,omitempty"`
	AccountID    *string `json:"account_id,omitempty"`
	Notes        *string `json:"notes,omitempty"`
}

// CreateGroupRequest contains group create parameters.
type CreateGroupRequest struct {
	SelfLabel          *string `json:"self_label,omitempty"`
	AccountID          string  `json:"account_id"`
	Name               string  `json:"name"`
	MemberCount        int     `json:"member_count"`
	ContributionAmount string  `json:"contribution_amount"`
	EarlyPayoutFeeRate *string `json:"early_payout_fee_rate,omitempty"`
	CycleFrequency     string  `json:"cycle_frequency"`
	StartDate          string  `json:"start_date"`
	Status             *string `json:"status,omitempty"`
}

// CreateContributionRequest contains contribution create parameters.
type CreateContributionRequest struct {
	Kind         string  `json:"kind"`
	AccountID    *string `json:"account_id,omitempty"`
	OccurredDate string  `json:"occurred_date"`
	OccurredTime *string `json:"occurred_time,omitempty"`
	Amount       string  `json:"amount"`
	CycleNo      *int    `json:"cycle_no,omitempty"`
	DueDate      *string `json:"due_date,omitempty"`
	Note         *string `json:"note,omitempty"`
}

// Service handles rotating savings business logic.
type Service struct {
	accounts domain.AccountRepository
	tx       TransactionServiceInterface
	repo     domain.RotatingSavingsRepository
}

// NewService creates a new rotating savings service.
func NewService(accounts domain.AccountRepository, tx TransactionServiceInterface, repo domain.RotatingSavingsRepository) *Service {
	return &Service{accounts: accounts, tx: tx, repo: repo}
}

// CreateGroup creates a new rotating savings group.
func (s *Service) CreateGroup(ctx context.Context, userID string, req CreateGroupRequest) (*domain.RotatingSavingsGroup, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperrors.Validation("name is required", map[string]any{"field": "name"})
	}

	accountIDRaw := strings.TrimSpace(req.AccountID)
	if accountIDRaw == "" {
		return nil, apperrors.Validation("account_id is required", map[string]any{"field": "account_id"})
	}
	accountID := accountIDRaw

	if req.MemberCount <= 0 {
		return nil, apperrors.Validation("member_count must be > 0", map[string]any{"field": "member_count"})
	}

	contributionAmount := strings.TrimSpace(req.ContributionAmount)
	if contributionAmount == "" {
		return nil, apperrors.Validation("contribution_amount is required", map[string]any{"field": "contribution_amount"})
	}
	if !isValidDecimal(contributionAmount) {
		return nil, apperrors.Validation("contribution_amount must be a decimal string", map[string]any{"field": "contribution_amount"})
	}

	earlyFee := normalizeOptionalString(req.EarlyPayoutFeeRate)
	if earlyFee != nil && !isValidDecimal(*earlyFee) {
		return nil, apperrors.Validation("early_payout_fee_rate must be a decimal string", map[string]any{"field": "early_payout_fee_rate"})
	}

	cycleFrequency := strings.TrimSpace(req.CycleFrequency)
	if cycleFrequency != "weekly" && cycleFrequency != "monthly" && cycleFrequency != "custom" {
		return nil, apperrors.Validation("cycle_frequency is invalid", map[string]any{"field": "cycle_frequency"})
	}

	startDate := strings.TrimSpace(req.StartDate)
	if startDate == "" {
		return nil, apperrors.Validation("start_date is required", map[string]any{"field": "start_date"})
	}
	if _, err := time.Parse("2006-01-02", startDate); err != nil {
		return nil, apperrors.Validation("start_date is invalid", map[string]any{"field": "start_date"})
	}

	status := "active"
	if req.Status != nil {
		v := strings.TrimSpace(*req.Status)
		if v != "" {
			if v != "active" && v != "completed" && v != "closed" {
				return nil, apperrors.Validation("status is invalid", map[string]any{"field": "status"})
			}
			status = v
		}
	}

	if _, err := s.accounts.GetAccountForUser(ctx, userID, accountID); err != nil {
		if errors.Is(err, apperrors.ErrAccountNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "account not found", err)
		}
		if errors.Is(err, apperrors.ErrAccountForbidden) {
			return nil, apperrors.Wrap(apperrors.KindForbidden, "account forbidden", err)
		}
		return nil, err
	}

	now := time.Now().UTC()
	id := uuid.NewString()

	g := domain.RotatingSavingsGroup{
		ID:                 id,
		UserID:             userID,
		SelfLabel:          normalizeOptionalString(req.SelfLabel),
		AccountID:          accountID,
		Name:               name,
		MemberCount:        req.MemberCount,
		ContributionAmount: contributionAmount,
		EarlyPayoutFeeRate: earlyFee,
		CycleFrequency:     cycleFrequency,
		StartDate:          startDate,
		Status:             status,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := s.repo.CreateGroup(ctx, userID, g); err != nil {
		return nil, err
	}

	created, err := s.repo.GetGroup(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// GetGroup retrieves a rotating savings group by ID.
func (s *Service) GetGroup(ctx context.Context, userID, groupID string) (*domain.RotatingSavingsGroup, error) {
	g, err := s.repo.GetGroup(ctx, userID, groupID)
	if err != nil {
		if errors.Is(err, apperrors.ErrRotatingSavingsGroupNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "group not found", err)
		}
		return nil, err
	}
	return g, nil
}

// ListGroups returns all rotating savings groups for a user.
func (s *Service) ListGroups(ctx context.Context, userID string) ([]domain.RotatingSavingsGroup, error) {
	return s.repo.ListGroups(ctx, userID)
}

// CreateContribution creates a contribution or payout for a group.
func (s *Service) CreateContribution(ctx context.Context, userID, groupID string, req CreateContributionRequest) (*domain.RotatingSavingsContribution, error) {
	kind := strings.TrimSpace(req.Kind)
	if kind != "contribution" && kind != "payout" {
		return nil, apperrors.Validation("kind is invalid", map[string]any{"field": "kind"})
	}

	amount := strings.TrimSpace(req.Amount)
	if amount == "" {
		return nil, apperrors.Validation("amount is required", map[string]any{"field": "amount"})
	}
	if !isValidDecimal(amount) {
		return nil, apperrors.Validation("amount must be a decimal string", map[string]any{"field": "amount"})
	}

	occurredDate := strings.TrimSpace(req.OccurredDate)
	if occurredDate == "" {
		return nil, apperrors.Validation("occurred_date is required", map[string]any{"field": "occurred_date"})
	}
	if _, err := time.Parse("2006-01-02", occurredDate); err != nil {
		return nil, apperrors.Validation("occurred_date is invalid", map[string]any{"field": "occurred_date"})
	}

	dueDate := normalizeOptionalString(req.DueDate)
	if dueDate != nil {
		if _, err := time.Parse("2006-01-02", *dueDate); err != nil {
			return nil, apperrors.Validation("due_date is invalid", map[string]any{"field": "due_date"})
		}
	}

	group, err := s.repo.GetGroup(ctx, userID, groupID)
	if err != nil {
		if errors.Is(err, apperrors.ErrRotatingSavingsGroupNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "group not found", err)
		}
		return nil, err
	}

	accountID := normalizeOptionalString(req.AccountID)
	if accountID == nil {
		accountID = &group.AccountID
	}

	if _, err := s.accounts.GetAccountForUser(ctx, userID, *accountID); err != nil {
		if errors.Is(err, apperrors.ErrAccountNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "account not found", err)
		}
		if errors.Is(err, apperrors.ErrAccountForbidden) {
			return nil, apperrors.Wrap(apperrors.KindForbidden, "account forbidden", err)
		}
		return nil, err
	}

	txType := "expense"
	if kind == "payout" {
		txType = "income"
	}

	lang := httpapi.LangFromContext(ctx)
	desc := fmt.Sprintf("%s: %s", i18n.T(lang, "rotating_savings_prefix"), group.Name)
	txReq := TxCreateRequest{
		Type:         txType,
		OccurredDate: &occurredDate,
		OccurredTime: normalizeOptionalString(req.OccurredTime),
		Amount:       amount,
		Description:  &desc,
		AccountID:    accountID,
		Notes:        normalizeOptionalString(req.Note),
	}

	tx, err := s.tx.Create(ctx, userID, txReq)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	cid := uuid.NewString()

	c := domain.RotatingSavingsContribution{
		ID:            cid,
		GroupID:       groupID,
		TransactionID: tx.ID,
		Kind:          kind,
		CycleNo:       req.CycleNo,
		DueDate:       dueDate,
		Amount:        amount,
		OccurredAt:    tx.OccurredAt,
		Note:          normalizeOptionalString(req.Note),
		CreatedAt:     now,
	}

	if err := s.repo.CreateContribution(ctx, userID, c); err != nil {
		return nil, err
	}

	items, err := s.repo.ListContributions(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].ID == cid {
			return &items[i], nil
		}
	}
	return &c, nil
}

// ListContributions returns all contributions for a group.
func (s *Service) ListContributions(ctx context.Context, userID, groupID string) ([]domain.RotatingSavingsContribution, error) {
	if _, err := s.repo.GetGroup(ctx, userID, groupID); err != nil {
		if errors.Is(err, apperrors.ErrRotatingSavingsGroupNotFound) {
			return nil, apperrors.Wrap(apperrors.KindNotFound, "group not found", err)
		}
		return nil, err
	}
	return s.repo.ListContributions(ctx, userID, groupID)
}

func normalizeOptionalString(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}

func isValidDecimal(s string) bool {
	_, ok := new(big.Rat).SetString(s)
	return ok
}
