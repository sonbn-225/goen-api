package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type RotatingSavingsService interface {
	CreateGroup(ctx context.Context, userID string, req CreateRotatingSavingsGroupRequest) (*domain.RotatingSavingsGroup, error)
	GetGroup(ctx context.Context, userID string, groupID string) (*domain.RotatingSavingsGroup, error)
	ListGroups(ctx context.Context, userID string) ([]domain.RotatingSavingsGroup, error)

	CreateContribution(ctx context.Context, userID string, groupID string, req CreateRotatingSavingsContributionRequest) (*domain.RotatingSavingsContribution, error)
	ListContributions(ctx context.Context, userID string, groupID string) ([]domain.RotatingSavingsContribution, error)
}

type CreateRotatingSavingsGroupRequest struct {
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

type CreateRotatingSavingsContributionRequest struct {
	Kind        string  `json:"kind"`
	AccountID   *string `json:"account_id,omitempty"`
	OccurredDate string `json:"occurred_date"`
	OccurredTime *string `json:"occurred_time,omitempty"`
	Amount      string  `json:"amount"`
	CycleNo     *int    `json:"cycle_no,omitempty"`
	DueDate     *string `json:"due_date,omitempty"`
	Note        *string `json:"note,omitempty"`
}

type rotatingSavingsService struct {
	accounts     domain.AccountRepository
	tx           TransactionService
	repo         domain.RotatingSavingsRepository
}

func NewRotatingSavingsService(accounts domain.AccountRepository, tx TransactionService, repo domain.RotatingSavingsRepository) RotatingSavingsService {
	return &rotatingSavingsService{accounts: accounts, tx: tx, repo: repo}
}

func (s *rotatingSavingsService) CreateGroup(ctx context.Context, userID string, req CreateRotatingSavingsGroupRequest) (*domain.RotatingSavingsGroup, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}

	accountIDRaw := strings.TrimSpace(req.AccountID)
	if accountIDRaw == "" {
		return nil, errors.New("account_id is required")
	}
	accountID := accountIDRaw

	if req.MemberCount <= 0 {
		return nil, errors.New("member_count must be > 0")
	}

	contributionAmount := strings.TrimSpace(req.ContributionAmount)
	if contributionAmount == "" {
		return nil, errors.New("contribution_amount is required")
	}
	if !isValidDecimal(contributionAmount) {
		return nil, errors.New("contribution_amount must be a decimal string")
	}

	earlyFee := normalizeOptionalString(req.EarlyPayoutFeeRate)
	if earlyFee != nil && !isValidDecimal(*earlyFee) {
		return nil, errors.New("early_payout_fee_rate must be a decimal string")
	}

	cycleFrequency := strings.TrimSpace(req.CycleFrequency)
	if cycleFrequency != "weekly" && cycleFrequency != "monthly" && cycleFrequency != "custom" {
		return nil, errors.New("cycle_frequency is invalid")
	}

	startDate := strings.TrimSpace(req.StartDate)
	if startDate == "" {
		return nil, errors.New("start_date is required")
	}
	if _, err := time.Parse("2006-01-02", startDate); err != nil {
		return nil, errors.New("start_date is invalid")
	}

	status := "active"
	if req.Status != nil {
		v := strings.TrimSpace(*req.Status)
		if v != "" {
			if v != "active" && v != "completed" && v != "closed" {
				return nil, errors.New("status is invalid")
			}
			status = v
		}
	}

	// Validate user has access.
	if _, err := s.accounts.GetAccountForUser(ctx, userID, accountID); err != nil {
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

func (s *rotatingSavingsService) GetGroup(ctx context.Context, userID string, groupID string) (*domain.RotatingSavingsGroup, error) {
	return s.repo.GetGroup(ctx, userID, groupID)
}

func (s *rotatingSavingsService) ListGroups(ctx context.Context, userID string) ([]domain.RotatingSavingsGroup, error) {
	return s.repo.ListGroups(ctx, userID)
}

func (s *rotatingSavingsService) CreateContribution(ctx context.Context, userID string, groupID string, req CreateRotatingSavingsContributionRequest) (*domain.RotatingSavingsContribution, error) {
	kind := strings.TrimSpace(req.Kind)
	if kind != "contribution" && kind != "payout" {
		return nil, errors.New("kind is invalid")
	}

	amount := strings.TrimSpace(req.Amount)
	if amount == "" {
		return nil, errors.New("amount is required")
	}
	if !isValidDecimal(amount) {
		return nil, errors.New("amount must be a decimal string")
	}

	occurredDate := strings.TrimSpace(req.OccurredDate)
	if occurredDate == "" {
		return nil, errors.New("occurred_date is required")
	}
	if _, err := time.Parse("2006-01-02", occurredDate); err != nil {
		return nil, errors.New("occurred_date is invalid")
	}

	dueDate := normalizeOptionalString(req.DueDate)
	if dueDate != nil {
		if _, err := time.Parse("2006-01-02", *dueDate); err != nil {
			return nil, errors.New("due_date is invalid")
		}
	}

	group, err := s.repo.GetGroup(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}

	accountID := normalizeOptionalString(req.AccountID)
	if accountID == nil {
		accountID = &group.AccountID
	}

	// Validate account access.
	if _, err := s.accounts.GetAccountForUser(ctx, userID, *accountID); err != nil {
		return nil, err
	}

	txType := "expense"
	if kind == "payout" {
		txType = "income"
	}

desc := fmt.Sprintf("RotatingSavings: %s", group.Name)
	txReq := CreateTransactionRequest{
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

func (s *rotatingSavingsService) ListContributions(ctx context.Context, userID string, groupID string) ([]domain.RotatingSavingsContribution, error) {
	// Ensure group belongs to user.
	if _, err := s.repo.GetGroup(ctx, userID, groupID); err != nil {
		return nil, err
	}
	return s.repo.ListContributions(ctx, userID, groupID)
}
