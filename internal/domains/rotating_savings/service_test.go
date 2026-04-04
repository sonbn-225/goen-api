package rotatingsavings

import (
	"context"
	"testing"
	"time"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/money"
	"github.com/sonbn-225/goen-api-v2/internal/domains/transaction"
)

type fakeRotatingRepo struct {
	groups        map[string]RotatingSavingsGroup
	contributions map[string]RotatingSavingsContribution
	auditLogs     []RotatingSavingsAuditLog
}

func (r *fakeRotatingRepo) CreateGroup(_ context.Context, group RotatingSavingsGroup) error {
	if r.groups == nil {
		r.groups = make(map[string]RotatingSavingsGroup)
	}
	r.groups[group.ID] = group
	return nil
}

func (r *fakeRotatingRepo) GetGroup(_ context.Context, userID, groupID string) (*RotatingSavingsGroup, error) {
	group, ok := r.groups[groupID]
	if !ok || group.UserID != userID {
		return nil, nil
	}
	copy := group
	return &copy, nil
}

func (r *fakeRotatingRepo) UpdateGroup(_ context.Context, group RotatingSavingsGroup) error {
	if r.groups == nil {
		r.groups = make(map[string]RotatingSavingsGroup)
	}
	r.groups[group.ID] = group
	return nil
}

func (r *fakeRotatingRepo) DeleteGroup(_ context.Context, userID, groupID string) error {
	group, ok := r.groups[groupID]
	if !ok || group.UserID != userID {
		return apperrors.New(apperrors.KindNotFound, "rotating savings group not found")
	}
	delete(r.groups, groupID)
	return nil
}

func (r *fakeRotatingRepo) ListGroups(_ context.Context, userID string) ([]RotatingSavingsGroup, error) {
	items := make([]RotatingSavingsGroup, 0)
	for _, group := range r.groups {
		if group.UserID == userID {
			items = append(items, group)
		}
	}
	return items, nil
}

func (r *fakeRotatingRepo) CreateContribution(_ context.Context, contribution RotatingSavingsContribution) error {
	if r.contributions == nil {
		r.contributions = make(map[string]RotatingSavingsContribution)
	}
	r.contributions[contribution.ID] = contribution
	return nil
}

func (r *fakeRotatingRepo) GetContribution(_ context.Context, userID, contributionID string) (*RotatingSavingsContribution, error) {
	item, ok := r.contributions[contributionID]
	if !ok {
		return nil, nil
	}
	group, ok := r.groups[item.GroupID]
	if !ok || group.UserID != userID {
		return nil, nil
	}
	copy := item
	return &copy, nil
}

func (r *fakeRotatingRepo) ListContributions(_ context.Context, userID, groupID string) ([]RotatingSavingsContribution, error) {
	group, ok := r.groups[groupID]
	if !ok || group.UserID != userID {
		return []RotatingSavingsContribution{}, nil
	}
	items := make([]RotatingSavingsContribution, 0)
	for _, item := range r.contributions {
		if item.GroupID == groupID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *fakeRotatingRepo) DeleteContribution(_ context.Context, _ string, contributionID string) error {
	if _, ok := r.contributions[contributionID]; !ok {
		return apperrors.New(apperrors.KindNotFound, "rotating savings contribution not found")
	}
	delete(r.contributions, contributionID)
	return nil
}

func (r *fakeRotatingRepo) CreateAuditLog(_ context.Context, log RotatingSavingsAuditLog) error {
	r.auditLogs = append(r.auditLogs, log)
	return nil
}

func (r *fakeRotatingRepo) ListAuditLogs(_ context.Context, userID, groupID string) ([]RotatingSavingsAuditLog, error) {
	items := make([]RotatingSavingsAuditLog, 0)
	for _, item := range r.auditLogs {
		if item.UserID == userID && item.GroupID != nil && *item.GroupID == groupID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *fakeRotatingRepo) SoftDeleteTransactionForUser(_ context.Context, _, _ string) error {
	return nil
}

type fakeRotatingTxService struct {
	lastInput transaction.CreateInput
}

func (s *fakeRotatingTxService) Create(_ context.Context, _ string, input transaction.CreateInput) (*transaction.Transaction, error) {
	s.lastInput = input
	return &transaction.Transaction{ID: "tx_1", CreatedAt: time.Now().UTC()}, nil
}

func TestRotatingSavingsServiceCreateGroupSuccess(t *testing.T) {
	repo := &fakeRotatingRepo{}
	svc := NewService(repo, &fakeRotatingTxService{})

	created, err := svc.CreateGroup(context.Background(), "u1", CreateGroupInput{
		AccountID:          "acc_1",
		Name:               "Hui A",
		MemberCount:        10,
		UserSlots:          1,
		ContributionAmount: 500000,
		CycleFrequency:     "monthly",
		StartDate:          "2026-04-01",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created == nil || created.ID == "" {
		t.Fatal("expected created group")
	}
}

func TestRotatingSavingsServiceCreateContributionSuccess(t *testing.T) {
	repo := &fakeRotatingRepo{
		groups: map[string]RotatingSavingsGroup{
			"g1": {
				ID:                 "g1",
				UserID:             "u1",
				AccountID:          "acc_1",
				MemberCount:        10,
				UserSlots:          1,
				ContributionAmount: 500000,
				CycleFrequency:     "monthly",
				StartDate:          "2026-04-01",
				Status:             "active",
			},
		},
	}
	txSvc := &fakeRotatingTxService{}
	svc := NewService(repo, txSvc)

	created, err := svc.CreateContribution(context.Background(), "u1", "g1", CreateContributionInput{
		Kind:                "uncollected",
		OccurredDate:        "2026-04-01",
		Amount:              500000,
		SlotsTaken:          0,
		CollectedFeePerSlot: 0,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created == nil || created.ID == "" {
		t.Fatal("expected created contribution")
	}
	if txSvc.lastInput.Type != "expense" {
		t.Fatalf("expected expense tx type, got %s", txSvc.lastInput.Type)
	}
	if len(txSvc.lastInput.LineItems) != 1 {
		t.Fatalf("expected one line item, got %d", len(txSvc.lastInput.LineItems))
	}
	if txSvc.lastInput.Amount.Cmp(money.MustFromString("500000.00").Decimal) != 0 {
		t.Fatalf("unexpected tx amount %s", txSvc.lastInput.Amount.String())
	}
}

func TestRotatingSavingsServiceRequiresUser(t *testing.T) {
	svc := NewService(&fakeRotatingRepo{}, &fakeRotatingTxService{})
	_, err := svc.ListGroups(context.Background(), "")
	if apperrors.KindOf(err) != apperrors.KindUnauth {
		t.Fatalf("expected unauth kind, got %s", apperrors.KindOf(err))
	}
}
