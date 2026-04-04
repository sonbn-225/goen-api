package rotatingsavings

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/logx"
	"github.com/sonbn-225/goen-api-v2/internal/core/money"
	"github.com/sonbn-225/goen-api-v2/internal/domains/transaction"
)

type service struct {
	repo      Repository
	txService TransactionService
}

var _ Service = (*service)(nil)

func NewService(repo Repository, txService TransactionService) Service {
	return &service{repo: repo, txService: txService}
}

func (s *service) CreateGroup(ctx context.Context, userID string, input CreateGroupInput) (*RotatingSavingsGroup, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	if s.repo == nil {
		return nil, apperrors.New(apperrors.KindInternal, "rotating savings repository not configured")
	}
	if strings.TrimSpace(input.AccountID) == "" {
		return nil, apperrors.New(apperrors.KindValidation, "account_id is required")
	}
	if strings.TrimSpace(input.Name) == "" {
		return nil, apperrors.New(apperrors.KindValidation, "name is required")
	}
	if input.MemberCount <= 0 {
		return nil, apperrors.New(apperrors.KindValidation, "member_count must be greater than zero")
	}
	if input.UserSlots <= 0 {
		input.UserSlots = 1
	}
	if input.ContributionAmount <= 0 {
		return nil, apperrors.New(apperrors.KindValidation, "contribution_amount must be greater than zero")
	}
	if input.FixedInterestAmount != nil && *input.FixedInterestAmount < 0 {
		return nil, apperrors.New(apperrors.KindValidation, "fixed_interest_amount must be greater than or equal to zero")
	}
	if _, err := time.Parse("2006-01-02", strings.TrimSpace(input.StartDate)); err != nil {
		return nil, apperrors.New(apperrors.KindValidation, "start_date must be YYYY-MM-DD")
	}

	freq := strings.TrimSpace(strings.ToLower(input.CycleFrequency))
	if freq != "weekly" && freq != "monthly" && freq != "custom" {
		return nil, apperrors.New(apperrors.KindValidation, "cycle_frequency must be one of: weekly, monthly, custom")
	}

	status := strings.TrimSpace(strings.ToLower(input.Status))
	if status == "" || status == "closed" {
		status = "active"
	}
	if status != "active" && status != "completed" {
		return nil, apperrors.New(apperrors.KindValidation, "status must be one of: active, completed")
	}

	now := time.Now().UTC()
	group := RotatingSavingsGroup{
		ID:                  uuid.NewString(),
		UserID:              userID,
		AccountID:           strings.TrimSpace(input.AccountID),
		Name:                strings.TrimSpace(input.Name),
		MemberCount:         input.MemberCount,
		UserSlots:           input.UserSlots,
		ContributionAmount:  input.ContributionAmount,
		FixedInterestAmount: input.FixedInterestAmount,
		CycleFrequency:      freq,
		StartDate:           strings.TrimSpace(input.StartDate),
		Status:              status,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := s.repo.CreateGroup(ctx, group); err != nil {
		return nil, passThroughOrWrapInternal("failed to create rotating savings group", err)
	}

	_ = s.repo.CreateAuditLog(ctx, RotatingSavingsAuditLog{
		ID:        uuid.NewString(),
		UserID:    userID,
		GroupID:   &group.ID,
		Action:    "group_created",
		Details:   map[string]any{"name": group.Name},
		CreatedAt: now,
	})

	return &group, nil
}

func (s *service) UpdateGroup(ctx context.Context, userID, groupID string, input UpdateGroupInput) (*RotatingSavingsGroup, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return nil, apperrors.New(apperrors.KindValidation, "groupId is required")
	}

	group, err := s.repo.GetGroup(ctx, userID, groupID)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to get rotating savings group", err)
	}
	if group == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "rotating savings group not found")
	}

	if input.AccountID != nil {
		accountID := strings.TrimSpace(*input.AccountID)
		if accountID == "" {
			return nil, apperrors.New(apperrors.KindValidation, "account_id is required")
		}
		group.AccountID = accountID
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, apperrors.New(apperrors.KindValidation, "name is required")
		}
		group.Name = name
	}
	if input.ContributionAmount != nil {
		if *input.ContributionAmount <= 0 {
			return nil, apperrors.New(apperrors.KindValidation, "contribution_amount must be greater than zero")
		}
		group.ContributionAmount = *input.ContributionAmount
	}
	if input.FixedInterestAmount != nil {
		if *input.FixedInterestAmount < 0 {
			return nil, apperrors.New(apperrors.KindValidation, "fixed_interest_amount must be greater than or equal to zero")
		}
		group.FixedInterestAmount = input.FixedInterestAmount
	}
	if input.PayoutCycleNo != nil {
		if *input.PayoutCycleNo <= 0 {
			group.PayoutCycleNo = nil
		} else {
			group.PayoutCycleNo = input.PayoutCycleNo
		}
	}
	if input.Status != nil {
		status := strings.TrimSpace(strings.ToLower(*input.Status))
		if status != "active" && status != "completed" {
			return nil, apperrors.New(apperrors.KindValidation, "status must be one of: active, completed")
		}
		group.Status = status
	}

	group.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateGroup(ctx, *group); err != nil {
		return nil, passThroughOrWrapInternal("failed to update rotating savings group", err)
	}

	_ = s.repo.CreateAuditLog(ctx, RotatingSavingsAuditLog{
		ID:        uuid.NewString(),
		UserID:    userID,
		GroupID:   &group.ID,
		Action:    "group_updated",
		Details:   map[string]any{"status": group.Status},
		CreatedAt: time.Now().UTC(),
	})

	return group, nil
}

func (s *service) DeleteGroup(ctx context.Context, userID, groupID string) error {
	if strings.TrimSpace(userID) == "" {
		return apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return apperrors.New(apperrors.KindValidation, "groupId is required")
	}

	history, err := s.repo.ListContributions(ctx, userID, groupID)
	if err != nil {
		return passThroughOrWrapInternal("failed to list rotating savings contributions", err)
	}
	for _, item := range history {
		if strings.TrimSpace(item.TransactionID) == "" {
			continue
		}
		if err := s.repo.SoftDeleteTransactionForUser(ctx, userID, item.TransactionID); err != nil {
			return passThroughOrWrapInternal("failed to cleanup rotating savings transaction", err)
		}
	}

	if err := s.repo.DeleteGroup(ctx, userID, groupID); err != nil {
		return passThroughOrWrapInternal("failed to delete rotating savings group", err)
	}
	return nil
}

func (s *service) ListGroups(ctx context.Context, userID string) ([]GroupSummary, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	groups, err := s.repo.ListGroups(ctx, userID)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to list rotating savings groups", err)
	}

	summaries := make([]GroupSummary, 0, len(groups))
	for _, group := range groups {
		contributions, err := s.repo.ListContributions(ctx, userID, group.ID)
		if err != nil {
			return nil, passThroughOrWrapInternal("failed to list rotating savings contributions", err)
		}

		totalPaid := 0.0
		totalReceived := 0.0
		completedCycles := make(map[int]struct{})
		for _, c := range contributions {
			if c.Kind == "payout" {
				totalReceived += c.Amount
			} else {
				totalPaid += c.Amount
			}
			if c.CycleNo != nil {
				completedCycles[*c.CycleNo] = struct{}{}
			}
		}

		schedule := s.generateSchedule(group, contributions)
		var nextDueDate *string
		totalExpected := 0.0
		for _, cycle := range schedule {
			totalExpected += cycle.ExpectedAmount
			if nextDueDate == nil && !cycle.IsPaid {
				d := cycle.DueDate
				nextDueDate = &d
			}
		}

		summaries = append(summaries, GroupSummary{
			Group:           group,
			TotalPaid:       totalPaid,
			TotalReceived:   totalReceived,
			RemainingAmount: totalExpected - totalPaid,
			CompletedCycles: len(completedCycles),
			TotalCycles:     group.MemberCount,
			NextDueDate:     nextDueDate,
		})
	}

	return summaries, nil
}

func (s *service) GetGroupDetail(ctx context.Context, userID, groupID string) (*GroupDetailResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return nil, apperrors.New(apperrors.KindValidation, "groupId is required")
	}

	group, err := s.repo.GetGroup(ctx, userID, groupID)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to get rotating savings group", err)
	}
	if group == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "rotating savings group not found")
	}

	contributions, err := s.repo.ListContributions(ctx, userID, groupID)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to list rotating savings contributions", err)
	}

	auditLogs, err := s.repo.ListAuditLogs(ctx, userID, groupID)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to list rotating savings audit logs", err)
	}

	schedule := s.generateSchedule(*group, contributions)
	collectedSlotsCount := 0
	for _, c := range contributions {
		if c.Kind == "payout" {
			collectedSlotsCount += c.SlotsTaken
		}
	}

	interest := 0.0
	if group.FixedInterestAmount != nil {
		interest = *group.FixedInterestAmount
	}

	lastUserPayoutCycle := 0
	for _, c := range contributions {
		if c.Kind == "payout" && c.CycleNo != nil && *c.CycleNo > lastUserPayoutCycle {
			lastUserPayoutCycle = *c.CycleNo
		}
	}

	nextPayoutCycle := 1
	confirmed := make(map[int]struct{})
	for _, c := range contributions {
		if c.Kind != "payout" && c.CycleNo != nil {
			confirmed[*c.CycleNo] = struct{}{}
		}
	}
	for i := 1; i <= group.MemberCount; i++ {
		if _, ok := confirmed[i]; !ok {
			nextPayoutCycle = i
			break
		}
	}

	userLivingSlotsBeforeNext := group.UserSlots - collectedSlotsCount
	numAccCycles := nextPayoutCycle - lastUserPayoutCycle
	if lastUserPayoutCycle == 0 {
		numAccCycles = nextPayoutCycle - 2
	}
	if numAccCycles < 0 {
		numAccCycles = 0
	}
	accruedInterest := float64(userLivingSlotsBeforeNext) * float64(numAccCycles) * interest
	payoutValue := float64(group.MemberCount) * group.ContributionAmount

	totalPaid := 0.0
	totalReceived := 0.0
	for _, c := range contributions {
		if c.Kind == "payout" {
			totalReceived += c.Amount
		} else {
			totalPaid += c.Amount
		}
	}

	nextPayment := 0.0
	totalExpected := 0.0
	for _, sc := range schedule {
		totalExpected += sc.ExpectedAmount
		if nextPayment == 0 && !sc.IsPaid {
			nextPayment = sc.ExpectedAmount
		}
	}

	return &GroupDetailResponse{
		Group:                  *group,
		Schedule:               schedule,
		CollectedSlotsCount:    collectedSlotsCount,
		CurrentPayoutValue:     payoutValue,
		CurrentAccruedInterest: accruedInterest,
		Contributions:          contributions,
		AuditLogs:              auditLogs,
		TotalPaid:              totalPaid,
		TotalReceived:          totalReceived,
		NextPayment:            nextPayment,
		RemainingAmount:        totalExpected - totalPaid,
	}, nil
}

func (s *service) CreateContribution(ctx context.Context, userID, groupID string, input CreateContributionInput) (*RotatingSavingsContribution, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "rotating_savings", "operation", "create_contribution")
	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return nil, apperrors.New(apperrors.KindValidation, "groupId is required")
	}

	group, err := s.repo.GetGroup(ctx, userID, groupID)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to get rotating savings group", err)
	}
	if group == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "rotating savings group not found")
	}

	kind := strings.TrimSpace(strings.ToLower(input.Kind))
	switch kind {
	case "uncollected", "payout", "collected", "partial_collected":
	default:
		return nil, apperrors.New(apperrors.KindValidation, "kind is invalid")
	}

	if input.Amount <= 0 {
		return nil, apperrors.New(apperrors.KindValidation, "amount must be greater than zero")
	}
	if input.SlotsTaken < 0 {
		return nil, apperrors.New(apperrors.KindValidation, "slots_taken must be greater than or equal to zero")
	}
	if input.CollectedFeePerSlot < 0 {
		return nil, apperrors.New(apperrors.KindValidation, "collected_fee_per_slot must be greater than or equal to zero")
	}
	if strings.TrimSpace(input.OccurredDate) == "" {
		return nil, apperrors.New(apperrors.KindValidation, "occurred_date is required")
	}

	occurredAt, err := parseOccurredAt(input.OccurredDate, input.OccurredTime)
	if err != nil {
		return nil, err
	}

	var dueDate *string
	if input.DueDate != nil {
		raw := strings.TrimSpace(*input.DueDate)
		if raw != "" {
			if _, err := time.Parse("2006-01-02", raw); err != nil {
				return nil, apperrors.New(apperrors.KindValidation, "due_date must be YYYY-MM-DD")
			}
			dueDate = &raw
		}
	}

	accountID := strings.TrimSpace(group.AccountID)
	if input.AccountID != nil && strings.TrimSpace(*input.AccountID) != "" {
		accountID = strings.TrimSpace(*input.AccountID)
	}
	if accountID == "" {
		return nil, apperrors.New(apperrors.KindValidation, "account_id is required")
	}

	txType := "expense"
	desc := "Rotating savings contribution"
	catID := "cat_sys_rotating_savings_contribution"
	if kind == "payout" {
		txType = "income"
		desc = "Rotating savings payout"
		catID = "cat_sys_rotating_savings_payout"
	}
	if kind == "collected" {
		desc = "Rotating savings collected"
	}

	amountMoney, err := money.NewFromString(fmt.Sprintf("%.2f", input.Amount))
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KindValidation, "amount is invalid", err)
	}
	categoryID := catID
	note := desc
	if input.Note != nil && strings.TrimSpace(*input.Note) != "" {
		note = strings.TrimSpace(*input.Note)
	}

	if s.txService == nil {
		return nil, apperrors.New(apperrors.KindInternal, "transaction service not configured")
	}

	txEntity, err := s.txService.Create(ctx, userID, transaction.CreateInput{
		AccountID: &accountID,
		Type:      txType,
		Amount:    amountMoney,
		Note:      note,
		LineItems: []transaction.CreateTransactionLineItemInput{{
			CategoryID: &categoryID,
			Amount:     amountMoney,
		}},
	})
	if err != nil {
		logger.Error("rotating_savings_create_contribution_failed", "error", err)
		return nil, err
	}

	now := time.Now().UTC()
	item := RotatingSavingsContribution{
		ID:                  uuid.NewString(),
		GroupID:             groupID,
		TransactionID:       txEntity.ID,
		Kind:                kind,
		CycleNo:             input.CycleNo,
		DueDate:             dueDate,
		Amount:              input.Amount,
		SlotsTaken:          input.SlotsTaken,
		CollectedFeePerSlot: input.CollectedFeePerSlot,
		OccurredAt:          occurredAt,
		Note:                normalizeOptionalString(input.Note),
		CreatedAt:           now,
	}
	if err := s.repo.CreateContribution(ctx, item); err != nil {
		_ = s.repo.SoftDeleteTransactionForUser(ctx, userID, txEntity.ID)
		return nil, passThroughOrWrapInternal("failed to create rotating savings contribution", err)
	}

	_ = s.repo.CreateAuditLog(ctx, RotatingSavingsAuditLog{
		ID:      uuid.NewString(),
		UserID:  userID,
		GroupID: &groupID,
		Action:  "contribution_created",
		Details: map[string]any{
			"cycle_no": item.CycleNo,
			"kind":     item.Kind,
			"amount":   item.Amount,
			"note":     item.Note,
		},
		CreatedAt: now,
	})

	if err := s.recomputeGroupStatus(ctx, userID, groupID); err != nil {
		logger.Warn("rotating_savings_recompute_group_status_failed", "error", err)
	}

	return &item, nil
}

func (s *service) ListContributions(ctx context.Context, userID, groupID string) ([]RotatingSavingsContribution, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return nil, apperrors.New(apperrors.KindValidation, "groupId is required")
	}
	items, err := s.repo.ListContributions(ctx, userID, groupID)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to list rotating savings contributions", err)
	}
	return items, nil
}

func (s *service) DeleteContribution(ctx context.Context, userID, groupID, contributionID string) error {
	if strings.TrimSpace(userID) == "" {
		return apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	groupID = strings.TrimSpace(groupID)
	contributionID = strings.TrimSpace(contributionID)
	if groupID == "" {
		return apperrors.New(apperrors.KindValidation, "groupId is required")
	}
	if contributionID == "" {
		return apperrors.New(apperrors.KindValidation, "contributionId is required")
	}

	item, err := s.repo.GetContribution(ctx, userID, contributionID)
	if err != nil {
		return passThroughOrWrapInternal("failed to get rotating savings contribution", err)
	}
	if item == nil || item.GroupID != groupID {
		return apperrors.New(apperrors.KindNotFound, "rotating savings contribution not found")
	}

	if strings.TrimSpace(item.TransactionID) != "" {
		if err := s.repo.SoftDeleteTransactionForUser(ctx, userID, item.TransactionID); err != nil {
			return passThroughOrWrapInternal("failed to cleanup rotating savings transaction", err)
		}
	}

	if err := s.repo.DeleteContribution(ctx, userID, contributionID); err != nil {
		return passThroughOrWrapInternal("failed to delete rotating savings contribution", err)
	}

	_ = s.repo.CreateAuditLog(ctx, RotatingSavingsAuditLog{
		ID:      uuid.NewString(),
		UserID:  userID,
		GroupID: &groupID,
		Action:  "contribution_deleted",
		Details: map[string]any{
			"cycle_no": item.CycleNo,
			"kind":     item.Kind,
			"amount":   item.Amount,
			"note":     item.Note,
		},
		CreatedAt: time.Now().UTC(),
	})

	return s.recomputeGroupStatus(ctx, userID, groupID)
}

func (s *service) recomputeGroupStatus(ctx context.Context, userID, groupID string) error {
	group, err := s.repo.GetGroup(ctx, userID, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return apperrors.New(apperrors.KindNotFound, "rotating savings group not found")
	}
	contributions, err := s.repo.ListContributions(ctx, userID, groupID)
	if err != nil {
		return err
	}

	completedCycles := make(map[int]struct{})
	for _, item := range contributions {
		if item.CycleNo != nil {
			completedCycles[*item.CycleNo] = struct{}{}
		}
	}

	status := "active"
	if len(completedCycles) >= group.MemberCount {
		status = "completed"
	}
	if group.Status == status {
		return nil
	}

	group.Status = status
	group.UpdatedAt = time.Now().UTC()
	return s.repo.UpdateGroup(ctx, *group)
}

func (s *service) generateSchedule(group RotatingSavingsGroup, history []RotatingSavingsContribution) []ScheduleCycle {
	count := group.MemberCount
	if count <= 0 {
		count = 10
	}

	startDate, err := time.Parse("2006-01-02", group.StartDate)
	if err != nil {
		startDate = time.Now().UTC()
	}

	type cycleHistory struct {
		Payout       *RotatingSavingsContribution
		Contribution *RotatingSavingsContribution
	}
	historyMap := make(map[int]cycleHistory)
	for i := range history {
		if history[i].CycleNo == nil {
			continue
		}
		cycleNo := *history[i].CycleNo
		ch := historyMap[cycleNo]
		if history[i].Kind == "payout" {
			ch.Payout = &history[i]
		} else {
			ch.Contribution = &history[i]
		}
		historyMap[cycleNo] = ch
	}

	interest := 0.0
	if group.FixedInterestAmount != nil {
		interest = *group.FixedInterestAmount
	}

	sortedPayouts := make([]RotatingSavingsContribution, 0)
	for _, c := range history {
		if c.Kind == "payout" && c.CycleNo != nil {
			sortedPayouts = append(sortedPayouts, c)
		}
	}
	sort.Slice(sortedPayouts, func(i, j int) bool {
		return *sortedPayouts[i].CycleNo < *sortedPayouts[j].CycleNo
	})

	schedule := make([]ScheduleCycle, 0, count)
	for i := 1; i <= count; i++ {
		dueDate := startDate
		switch group.CycleFrequency {
		case "weekly":
			dueDate = startDate.AddDate(0, 0, (i-1)*7)
		case "monthly":
			dueDate = startDate.AddDate(0, i-1, 0)
		}

		userCollectedSlotsBefore := 0
		lastUserPayoutCycleBefore := 0
		for _, p := range sortedPayouts {
			if *p.CycleNo < i {
				userCollectedSlotsBefore += p.SlotsTaken
				if *p.CycleNo > lastUserPayoutCycleBefore {
					lastUserPayoutCycleBefore = *p.CycleNo
				}
			}
		}

		userLivingSlotsBefore := group.UserSlots - userCollectedSlotsBefore
		numAccCycles := i - lastUserPayoutCycleBefore
		if lastUserPayoutCycleBefore == 0 {
			numAccCycles = i - 2
		}
		if numAccCycles < 0 {
			numAccCycles = 0
		}

		accruedInterest := float64(userLivingSlotsBefore) * float64(numAccCycles) * interest
		suggestedPayoutAmount := float64(group.MemberCount)*group.ContributionAmount + accruedInterest
		expectedContribution := float64(userCollectedSlotsBefore)*(group.ContributionAmount+interest) + float64(userLivingSlotsBefore)*group.ContributionAmount

		ch := historyMap[i]
		isPaid := ch.Contribution != nil
		kind := "uncollected"
		var contributionID *string
		if isPaid {
			kind = ch.Contribution.Kind
			contributionID = &ch.Contribution.ID
			expectedContribution = ch.Contribution.Amount
		}

		isPayout := ch.Payout != nil
		var payoutID *string
		payoutAmount := suggestedPayoutAmount
		payoutSlots := 0
		if isPayout {
			payoutID = &ch.Payout.ID
			payoutAmount = ch.Payout.Amount
			payoutSlots = ch.Payout.SlotsTaken
		}

		userCollectedSlotsEnd := userCollectedSlotsBefore + payoutSlots
		if userCollectedSlotsEnd > 0 {
			if userCollectedSlotsEnd < group.UserSlots {
				kind = "partial_collected"
			} else {
				kind = "collected"
			}
		}

		schedule = append(schedule, ScheduleCycle{
			CycleNo:            i,
			DueDate:            dueDate.Format("2006-01-02"),
			ExpectedAmount:     expectedContribution,
			Kind:               kind,
			IsPaid:             isPaid,
			ContributionID:     contributionID,
			IsPayout:           isPayout,
			PayoutID:           payoutID,
			PayoutAmount:       payoutAmount,
			PayoutSlots:        payoutSlots,
			UserCollectedSlots: userCollectedSlotsEnd,
			AccruedInterest:    accruedInterest,
		})
	}

	return schedule
}

func parseOccurredAt(occurredDate string, occurredTime *string) (time.Time, error) {
	base, err := time.Parse("2006-01-02", strings.TrimSpace(occurredDate))
	if err != nil {
		return time.Time{}, apperrors.New(apperrors.KindValidation, "occurred_date must be YYYY-MM-DD")
	}
	if occurredTime == nil || strings.TrimSpace(*occurredTime) == "" {
		return time.Date(base.Year(), base.Month(), base.Day(), 0, 0, 0, 0, time.UTC), nil
	}

	timeRaw := strings.TrimSpace(*occurredTime)
	if hm, err := time.Parse("15:04", timeRaw); err == nil {
		return time.Date(base.Year(), base.Month(), base.Day(), hm.Hour(), hm.Minute(), 0, 0, time.UTC), nil
	}
	if hms, err := time.Parse("15:04:05", timeRaw); err == nil {
		return time.Date(base.Year(), base.Month(), base.Day(), hms.Hour(), hms.Minute(), hms.Second(), 0, time.UTC), nil
	}

	return time.Time{}, apperrors.New(apperrors.KindValidation, "occurred_time must be HH:mm or HH:mm:ss")
}

func normalizeOptionalString(v *string) *string {
	if v == nil {
		return nil
	}
	clean := strings.TrimSpace(*v)
	if clean == "" {
		return nil
	}
	return &clean
}

func passThroughOrWrapInternal(message string, err error) error {
	if err == nil {
		return nil
	}
	var appErr *apperrors.Error
	if errors.As(err, &appErr) {
		return err
	}
	return apperrors.Wrap(apperrors.KindInternal, message, err)
}
