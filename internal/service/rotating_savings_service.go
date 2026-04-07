package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

type RotatingSavingsService struct {
	repo       interfaces.RotatingSavingsRepository
	accountSvc interfaces.AccountService
	txSvc      interfaces.TransactionService
}

func NewRotatingSavingsService(
	repo interfaces.RotatingSavingsRepository,
	accountSvc interfaces.AccountService,
	txSvc interfaces.TransactionService,
) *RotatingSavingsService {
	return &RotatingSavingsService{
		repo:       repo,
		accountSvc: accountSvc,
		txSvc:      txSvc,
	}
}

func (s *RotatingSavingsService) CreateGroup(ctx context.Context, userID uuid.UUID, req dto.CreateRotatingSavingsGroupRequest) (*dto.RotatingSavingsGroupResponse, error) {
	status := req.Status
	if status == "" {
		status = "active"
	}

	accountID := req.AccountID

	g := entity.RotatingSavingsGroup{
		AuditEntity: entity.AuditEntity{
			BaseEntity: entity.BaseEntity{
				ID: utils.NewID(),
			},
		},
		UserID:              userID,
		AccountID:           accountID,
		Name:                req.Name,
		MemberCount:         req.MemberCount,
		UserSlots:           req.UserSlots,
		ContributionAmount:  req.ContributionAmount,
		FixedInterestAmount: req.FixedInterestAmount,
		CycleFrequency:      req.CycleFrequency,
		StartDate:           req.StartDate,
		Status:              status,
	}

	if g.UserSlots <= 0 {
		g.UserSlots = 1
	}

	if err := s.repo.CreateRotatingGroup(ctx, g); err != nil {
		return nil, err
	}

	created, err := s.repo.GetRotatingGroup(ctx, userID, g.ID)
	if err != nil {
		return nil, err
	}
	if created != nil {
		g = *created
	}

	_ = s.repo.AddAuditLog(ctx, entity.RotatingSavingsAuditLog{
		BaseEntity: entity.BaseEntity{
			ID: utils.NewID(),
		},
		UserID:    userID,
		GroupID:   &g.ID,
		Action:    "group_created",
		Details:   map[string]any{"name": g.Name},
		CreatedAt: time.Now().UTC(),
	})

	resp := dto.NewRotatingSavingsGroupResponse(g)
	return &resp, nil
}

func (s *RotatingSavingsService) GetGroup(ctx context.Context, userID, groupID uuid.UUID) (*dto.RotatingSavingsGroupResponse, error) {
	g, err := s.repo.GetRotatingGroup(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, nil
	}
	resp := dto.NewRotatingSavingsGroupResponse(*g)
	return &resp, nil
}

func (s *RotatingSavingsService) UpdateGroup(ctx context.Context, userID, groupID uuid.UUID, req dto.UpdateRotatingSavingsGroupRequest) (*dto.RotatingSavingsGroupResponse, error) {
	g, err := s.repo.GetRotatingGroup(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, errors.New("group not found")
	}

	if req.AccountID != nil {
		g.AccountID = *req.AccountID
	}
	if req.Name != nil {
		g.Name = *req.Name
	}
	if req.ContributionAmount != nil {
		g.ContributionAmount = *req.ContributionAmount
	}
	if req.FixedInterestAmount != nil {
		g.FixedInterestAmount = req.FixedInterestAmount
	}
	if req.PayoutCycleNo != nil {
		g.PayoutCycleNo = req.PayoutCycleNo
	}
	if req.Status != nil {
		g.Status = *req.Status
	}

	if err := s.repo.UpdateRotatingGroup(ctx, *g); err != nil {
		return nil, err
	}

	updated, err := s.repo.GetRotatingGroup(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	if updated != nil {
		g = updated
	}

	_ = s.repo.AddAuditLog(ctx, entity.RotatingSavingsAuditLog{
		BaseEntity: entity.BaseEntity{
			ID: utils.NewID(),
		},
		UserID:    userID,
		GroupID:   &g.ID,
		Action:    "group_updated",
		Details:   map[string]any{"status": g.Status},
		CreatedAt: time.Now().UTC(),
	})

	resp := dto.NewRotatingSavingsGroupResponse(*g)
	return &resp, nil
}

func (s *RotatingSavingsService) GetGroupDetail(ctx context.Context, userID, groupID uuid.UUID) (*dto.RotatingSavingsGroupDetailResponse, error) {
	g, err := s.repo.GetRotatingGroup(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, errors.New("group not found")
	}

	contributions, err := s.repo.GetContributions(ctx, groupID)
	if err != nil {
		return nil, err
	}

	auditLogs, _ := s.repo.GetAuditLogs(ctx, groupID)

	schedule := s.generateSchedule(*g, contributions)

	collectedSlotsCount := 0
	totalPaid := 0.0
	totalReceived := 0.0
	for _, c := range contributions {
		if c.Kind == "payout" {
			collectedSlotsCount += c.SlotsTaken
			totalReceived += c.Amount
		} else {
			totalPaid += c.Amount
		}
	}

	totalExpected := 0.0
	nextPayment := 0.0
	for _, sc := range schedule {
		totalExpected += sc.ExpectedAmount
		if !sc.IsPaid && nextPayment == 0 {
			nextPayment = sc.ExpectedAmount
		}
	}

	// Suggestion logic (simplified)
	payoutValue := float64(g.MemberCount) * g.ContributionAmount

	// Accrued interest logic could be more complex, keeping legacy parity
	accruedInterest := 0.0

	return &dto.RotatingSavingsGroupDetailResponse{
		Group:                  dto.NewRotatingSavingsGroupResponse(*g),
		Schedule:               schedule,
		CollectedSlotsCount:    collectedSlotsCount,
		CurrentPayoutValue:     payoutValue,
		CurrentAccruedInterest: accruedInterest,
		Contributions:          dto.NewRotatingSavingsContributionResponses(contributions),
		AuditLogs:              dto.NewRotatingSavingsAuditLogResponses(auditLogs),
		TotalPaid:              totalPaid,
		TotalReceived:          totalReceived,
		NextPayment:            nextPayment,
		RemainingAmount:        totalExpected - totalPaid,
	}, nil
}

func (s *RotatingSavingsService) generateSchedule(g entity.RotatingSavingsGroup, history []entity.RotatingSavingsContribution) []dto.RotatingSavingsScheduleCycle {
	count := g.MemberCount
	if count <= 0 {
		count = 10
	}

	startDate, _ := time.Parse("2006-01-02", g.StartDate)
	schedule := make([]dto.RotatingSavingsScheduleCycle, count)

	histMap := make(map[int]struct {
		P *entity.RotatingSavingsContribution
		C *entity.RotatingSavingsContribution
	})
	for i := range history {
		c := &history[i]
		if c.CycleNo != nil {
			ch := histMap[*c.CycleNo]
			if c.Kind == "payout" {
				ch.P = c
			} else {
				ch.C = c
			}
			histMap[*c.CycleNo] = ch
		}
	}

	interest := 0.0
	if g.FixedInterestAmount != nil {
		interest = *g.FixedInterestAmount
	}

	for i := 1; i <= count; i++ {
		dueDate := startDate
		if g.CycleFrequency == "weekly" {
			dueDate = startDate.AddDate(0, 0, (i-1)*7)
		} else {
			dueDate = startDate.AddDate(0, (i - 1), 0)
		}

		userCollectedBeforeI := 0
		lastPayoutBeforeI := 0
		for _, c := range history {
			if c.Kind == "payout" && c.CycleNo != nil && *c.CycleNo < i {
				userCollectedBeforeI += c.SlotsTaken
				if *c.CycleNo > lastPayoutBeforeI {
					lastPayoutBeforeI = *c.CycleNo
				}
			}
		}

		userLivingBeforeI := g.UserSlots - userCollectedBeforeI

		numAccCycles := i - lastPayoutBeforeI
		if lastPayoutBeforeI == 0 {
			numAccCycles = i - 2
		}
		if numAccCycles < 0 {
			numAccCycles = 0
		}

		accruedInterest := float64(userLivingBeforeI) * float64(numAccCycles) * interest
		payoutAmount := float64(g.MemberCount)*g.ContributionAmount + accruedInterest

		expectedContrib := float64(userCollectedBeforeI)*(g.ContributionAmount+interest) + float64(userLivingBeforeI)*g.ContributionAmount

		ch := histMap[i]
		isPaid := ch.C != nil
		var contribID *uuid.UUID
		kind := "uncollected"
		if isPaid {
			contribID = &ch.C.ID
			expectedContrib = ch.C.Amount
			kind = "collected"
		}

		isPayout := ch.P != nil
		var payoutID *uuid.UUID
		if isPayout {
			payoutID = &ch.P.ID
			payoutAmount = ch.P.Amount
		}

		schedule[i-1] = dto.RotatingSavingsScheduleCycle{
			CycleNo: i, DueDate: dueDate.Format("2006-01-02"), ExpectedAmount: expectedContrib,
			Kind: kind, IsPaid: isPaid, ContributionID: contribID, IsPayout: isPayout, PayoutID: payoutID,
			PayoutAmount: payoutAmount, PayoutSlots: 0, UserCollectedSlots: userCollectedBeforeI, AccruedInterest: accruedInterest,
		}
	}
	return schedule
}

func (s *RotatingSavingsService) ListGroups(ctx context.Context, userID uuid.UUID) ([]dto.RotatingSavingsGroupSummary, error) {
	groups, err := s.repo.ListRotatingGroups(ctx, userID)
	if err != nil {
		return nil, err
	}

	summaries := make([]dto.RotatingSavingsGroupSummary, 0, len(groups))
	for _, g := range groups {
		contributions, _ := s.repo.GetContributions(ctx, g.ID)
		totalPaid := 0.0
		totalReceived := 0.0
		completedCycles := make(map[int]bool)
		for _, c := range contributions {
			if c.Kind == "payout" {
				totalReceived += c.Amount
			} else {
				totalPaid += c.Amount
			}
			if c.CycleNo != nil {
				completedCycles[*c.CycleNo] = true
			}
		}

		schedule := s.generateSchedule(g, contributions)
		var nextDate *string
		totalExpected := 0.0
		for _, sc := range schedule {
			totalExpected += sc.ExpectedAmount
			if !sc.IsPaid && nextDate == nil {
				d := sc.DueDate
				nextDate = &d
			}
		}

		summaries = append(summaries, dto.RotatingSavingsGroupSummary{
			Group: dto.NewRotatingSavingsGroupResponse(g), TotalPaid: totalPaid, TotalReceived: totalReceived,
			RemainingAmount: totalExpected - totalPaid, CompletedCycles: len(completedCycles),
			TotalCycles: g.MemberCount, NextDueDate: nextDate,
		})
	}
	return summaries, nil
}

func (s *RotatingSavingsService) CreateContribution(ctx context.Context, userID, groupID uuid.UUID, req dto.RotatingSavingsContributionRequest) (*dto.RotatingSavingsContributionResponse, error) {
	g, err := s.repo.GetRotatingGroup(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, errors.New("group not found")
	}

	txType := "expense"
	if req.Kind == "payout" {
		txType = "income"
	}

	desc := "Đóng hụi"
	if req.Kind == "payout" {
		desc = "Lĩnh hụi"
	}

	var catID uuid.UUID
	if req.Kind == "payout" {
		catID, _ = uuid.Parse("00000000-0000-0000-0000-000000000005")
	} else {
		catID, _ = uuid.Parse("00000000-0000-0000-0000-000000000006")
	}

	if req.Note != nil && *req.Note != "" {
		desc = desc + " - " + *req.Note
	}

	tx, err := s.txSvc.Create(ctx, userID, dto.CreateTransactionRequest{
		Type: txType, OccurredDate: &req.OccurredDate, Amount: req.Amount, CategoryID: &catID, Description: &desc, AccountID: &g.AccountID,
	})
	if err != nil {
		return nil, err
	}

	parsedAmount, _ := strconv.ParseFloat(req.Amount, 64)

	c := entity.RotatingSavingsContribution{
		BaseEntity: entity.BaseEntity{
			ID: utils.NewID(),
		},
		GroupID:             groupID,
		TransactionID:       tx.ID,
		Kind:                req.Kind,
		CycleNo:             req.CycleNo,
		DueDate:             req.DueDate,
		Amount:              parsedAmount,
		SlotsTaken:          req.SlotsTaken,
		CollectedFeePerSlot: req.CollectedFeePerSlot,
		OccurredAt:          time.Now().UTC(),
		Note:                req.Note,
		CreatedAt:           time.Now().UTC(),
	}

	if err := s.repo.CreateContribution(ctx, c); err != nil {
		return nil, err
	}

	_ = s.repo.AddAuditLog(ctx, entity.RotatingSavingsAuditLog{
		BaseEntity: entity.BaseEntity{
			ID: utils.NewID(),
		},
		UserID:    userID,
		GroupID:   &groupID,
		Action:    "contribution_created",
		Details:   map[string]any{"kind": c.Kind, "amount": req.Amount},
		CreatedAt: time.Now().UTC(),
	})

	resp := dto.NewRotatingSavingsContributionResponse(c)
	return &resp, nil
}

func (s *RotatingSavingsService) DeleteContribution(ctx context.Context, userID, groupID, id uuid.UUID) error {
	// Finding contribution in group because GetContribution was removed from interface for consolidation
	conts, err := s.repo.GetContributions(ctx, groupID)
	if err != nil {
		return err
	}

	var target *entity.RotatingSavingsContribution
	for i := range conts {
		if conts[i].ID == id {
			target = &conts[i]
			break
		}
	}

	if target == nil {
		return errors.New("contribution not found")
	}

	if target.TransactionID != uuid.Nil {
		_ = s.txSvc.Delete(ctx, userID, target.TransactionID)
	}

	return s.repo.DeleteContribution(ctx, id)
}

func (s *RotatingSavingsService) DeleteGroup(ctx context.Context, userID, groupID uuid.UUID) error {
	conts, _ := s.repo.GetContributions(ctx, groupID)
	for _, c := range conts {
		if c.TransactionID != uuid.Nil {
			_ = s.txSvc.Delete(ctx, userID, c.TransactionID)
		}
	}
	return s.repo.DeleteRotatingGroup(ctx, userID, groupID)
}

