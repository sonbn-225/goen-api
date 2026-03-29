package rotatingsavings

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type Service struct {
	accountRepo domain.AccountRepository
	txSvc       TransactionServiceInterface
	repo        domain.RotatingSavingsRepository
}

func NewService(accountRepo domain.AccountRepository, txSvc TransactionServiceInterface, repo domain.RotatingSavingsRepository) *Service {
	return &Service{
		accountRepo: accountRepo,
		txSvc:       txSvc,
		repo:        repo,
	}
}

type TxCreateRequest struct {
	Type         string  `json:"type"`
	OccurredDate *string `json:"occurred_date"`
	OccurredTime *string `json:"occurred_time"`
	Amount       string  `json:"amount"`
	CategoryID   *string `json:"category_id"`
	Description  *string `json:"description"`
	Notes        *string `json:"notes"`
	AccountID    *string `json:"account_id"`
}

type CreateGroupRequest struct {
	AccountID           string   `json:"account_id"`
	Name                string   `json:"name"`
	MemberCount         int      `json:"member_count"`
	UserSlots           int      `json:"user_slots"`
	ContributionAmount  float64  `json:"contribution_amount"`
	FixedInterestAmount *float64 `json:"fixed_interest_amount"`
	CycleFrequency      string   `json:"cycle_frequency"`
	StartDate           string   `json:"start_date"`
	Status              string   `json:"status"`
}

func (s *Service) CreateGroup(ctx context.Context, userID string, req CreateGroupRequest) (*domain.RotatingSavingsGroup, error) {
	if req.AccountID == "" {
		return nil, apperrors.ErrAccountIDRequired
	}
	if req.Name == "" {
		return nil, apperrors.ErrRotatingSavingsNameRequired
	}

	status := req.Status
	if status == "" || status == "closed" {
		status = "active"
	}

	g := domain.RotatingSavingsGroup{
		ID:                  uuid.New().String(),
		UserID:              userID,
		AccountID:           req.AccountID,
		Name:                req.Name,
		MemberCount:         req.MemberCount,
		UserSlots:           req.UserSlots,
		ContributionAmount:  req.ContributionAmount,
		FixedInterestAmount: req.FixedInterestAmount,
		CycleFrequency:      req.CycleFrequency,
		StartDate:           req.StartDate,
		Status:              status,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if g.UserSlots <= 0 {
		g.UserSlots = 1
	}

	if err := s.repo.CreateGroup(ctx, g); err != nil {
		return nil, err
	}

	_ = s.repo.CreateAuditLog(ctx, domain.RotatingSavingsAuditLog{
		ID:        uuid.New().String(),
		UserID:    userID,
		GroupID:   &g.ID,
		Action:    "group_created",
		Details:   map[string]any{"name": g.Name},
		CreatedAt: time.Now(),
	})

	return &g, nil
}

type UpdateGroupRequest struct {
	AccountID           *string  `json:"account_id"`
	Name                *string  `json:"name"`
	ContributionAmount  *float64 `json:"contribution_amount"`
	FixedInterestAmount *float64 `json:"fixed_interest_amount"`
	PayoutCycleNo       *int     `json:"payout_cycle_no"`
	Status              *string  `json:"status"`
}

func (s *Service) UpdateGroup(ctx context.Context, userID string, groupID string, req UpdateGroupRequest) (*domain.RotatingSavingsGroup, error) {
	g, err := s.repo.GetGroup(ctx, userID, groupID)
	if err != nil {
		return nil, err
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

	g.UpdatedAt = time.Now()

	if err := s.repo.UpdateGroup(ctx, *g); err != nil {
		return nil, err
	}

	_ = s.repo.CreateAuditLog(ctx, domain.RotatingSavingsAuditLog{
		ID:        uuid.New().String(),
		UserID:    userID,
		GroupID:   &g.ID,
		Action:    "group_updated",
		Details:   map[string]any{"status": g.Status},
		CreatedAt: time.Now(),
	})

	return g, nil
}

func (s *Service) GetGroup(ctx context.Context, userID string, groupID string) (*domain.RotatingSavingsGroup, error) {
	return s.repo.GetGroup(ctx, userID, groupID)
}

type ScheduleCycle struct {
	CycleNo           int      `json:"cycle_no"`
	DueDate           string   `json:"due_date"`
	ExpectedAmount    float64  `json:"expected_amount"`
	Kind              string   `json:"kind"` // 'living' or 'dead'
	IsPaid            bool     `json:"is_paid"`
	ContributionID    *string  `json:"contribution_id"`
	IsPayout          bool     `json:"is_payout"`
	PayoutID          *string  `json:"payout_id"`
	PayoutAmount      float64  `json:"payout_amount"`
	PayoutSlots       int      `json:"payout_slots"`
	GroupCollectedSlots int    `json:"group_collected_slots"`
	UserCollectedSlots  int    `json:"user_collected_slots"`
	AccruedInterest     float64 `json:"accrued_interest"`
}

type GroupSummary struct {
	Group           domain.RotatingSavingsGroup `json:"group"`
	TotalPaid       float64                     `json:"total_paid"`
	TotalReceived   float64                     `json:"total_received"`
	RemainingAmount float64                     `json:"remaining_amount"`
	CompletedCycles int                         `json:"completed_cycles"`
	TotalCycles     int                         `json:"total_cycles"`
	NextDueDate     *string                     `json:"next_due_date"`
}

type GroupDetailResponse struct {
	Group              domain.RotatingSavingsGroup          `json:"group"`
	Schedule           []ScheduleCycle                       `json:"schedule"`
	CollectedSlotsCount int                                  `json:"collected_slots_count"`
	CurrentPayoutValue float64                               `json:"current_payout_value"`
	CurrentAccruedInterest float64                          `json:"current_accrued_interest"`
	Contributions      []domain.RotatingSavingsContribution `json:"contributions"`
	AuditLogs          []domain.RotatingSavingsAuditLog     `json:"audit_logs"`
	TotalPaid          float64                               `json:"total_paid"`
	TotalReceived      float64                               `json:"total_received"`
	NextPayment        float64                               `json:"next_payment"`
	RemainingAmount    float64                               `json:"remaining_amount"`
}

func (s *Service) GetGroupDetail(ctx context.Context, userID string, groupID string) (*GroupDetailResponse, error) {
	g, err := s.repo.GetGroup(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}

	contributions, err := s.repo.ListContributions(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}

	auditLogs, _ := s.repo.ListAuditLogs(ctx, userID, groupID)

	schedule := s.generateSchedule(*g, contributions)

	collectedSlotsCount := 0
	for _, c := range contributions {
		if c.Kind == "payout" {
			collectedSlotsCount += c.SlotsTaken
		}
	}

	interest := 0.0
	if g.FixedInterestAmount != nil {
		interest = *g.FixedInterestAmount
	}

	// Calculate payout value for the NEXT payout 
	lastUserPayoutCycle := 0
	for _, c := range contributions {
		if c.Kind == "payout" && c.CycleNo != nil && *c.CycleNo > lastUserPayoutCycle {
			lastUserPayoutCycle = *c.CycleNo
		}
	}

	// For suggestion, use the current active cycle (first cycle without a contribution)
	nextPayoutCycle := 1
	confirmedCycles := make(map[int]bool)
	for _, c := range contributions {
		if c.Kind != "payout" && c.CycleNo != nil {
			confirmedCycles[*c.CycleNo] = true
		}
	}
	for i := 1; i <= g.MemberCount; i++ {
		if !confirmedCycles[i] {
			nextPayoutCycle = i
			break
		}
	}

	userLivingSlotsBeforeNext := g.UserSlots - collectedSlotsCount
	numAccCycles := nextPayoutCycle - lastUserPayoutCycle
	if lastUserPayoutCycle == 0 {
		numAccCycles = nextPayoutCycle - 2
	}
	if numAccCycles < 0 {
		numAccCycles = 0
	}

	// Accrued interest is collected for ALL living slots
	accruedInterest := float64(userLivingSlotsBeforeNext) * float64(numAccCycles) * interest
	
	// Suggested payout amount (PRINCIPAL ONLY for one slot)
	payoutValue := float64(g.MemberCount)*g.ContributionAmount

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
	for _, sc := range schedule {
		if !sc.IsPaid {
			nextPayment = sc.ExpectedAmount
			break
		}
	}

	totalExpected := 0.0
	for _, sc := range schedule {
		totalExpected += sc.ExpectedAmount
	}

	return &GroupDetailResponse{
		Group:               *g,
		Schedule:            schedule,
		CollectedSlotsCount: collectedSlotsCount,
		CurrentPayoutValue:  payoutValue,
		CurrentAccruedInterest: accruedInterest,
		Contributions:       contributions,
		AuditLogs:           auditLogs,
		TotalPaid:           totalPaid,
		TotalReceived:       totalReceived,
		NextPayment:         nextPayment,
		RemainingAmount:     totalExpected - totalPaid,
	}, nil
}

func (s *Service) generateSchedule(g domain.RotatingSavingsGroup, history []domain.RotatingSavingsContribution) []ScheduleCycle {
	count := g.MemberCount
	if count <= 0 {
		count = 10
	}

	startDate, _ := time.Parse("2006-01-02", g.StartDate)
	schedule := make([]ScheduleCycle, count)

	// Build history map for quick lookup
	type cycleHistory struct {
		Payout       *domain.RotatingSavingsContribution
		Contribution *domain.RotatingSavingsContribution
	}
	histMap := make(map[int]cycleHistory)
	for _, c := range history {
		if c.CycleNo != nil {
			ch := histMap[*c.CycleNo]
			if c.Kind == "payout" {
				ch.Payout = &c
			} else {
				ch.Contribution = &c
			}
			histMap[*c.CycleNo] = ch
		}
	}

	interest := 0.0
	if g.FixedInterestAmount != nil {
		interest = *g.FixedInterestAmount
	}

	// Sort payouts to track interest collection cycles
	sortedUserPayouts := make([]domain.RotatingSavingsContribution, 0)
	for _, c := range history {
		if c.Kind == "payout" && c.CycleNo != nil {
			sortedUserPayouts = append(sortedUserPayouts, c)
		}
	}
	sort.Slice(sortedUserPayouts, func(i, j int) bool {
		return *sortedUserPayouts[i].CycleNo < *sortedUserPayouts[j].CycleNo
	})

	for i := 1; i <= count; i++ {
		dueDate := startDate
		switch g.CycleFrequency {
		case "weekly":
			dueDate = startDate.AddDate(0, 0, (i-1)*7)
		case "monthly":
			dueDate = startDate.AddDate(0, (i-1), 0)
		}

		// 1. Calculate user's internal state strictly BEFORE i
		userCollectedSlotsStrictlyBeforeI := 0
		lastUserPayoutCycleBeforeI := 0
		for _, c := range sortedUserPayouts {
			if *c.CycleNo < i {
				userCollectedSlotsStrictlyBeforeI += c.SlotsTaken
				if *c.CycleNo > lastUserPayoutCycleBeforeI {
					lastUserPayoutCycleBeforeI = *c.CycleNo
				}
			}
		}

		userLivingSlotsBeforeI := g.UserSlots - userCollectedSlotsStrictlyBeforeI

		// 2. Suggested payout amount using clearance logic
		numAccCycles := i - lastUserPayoutCycleBeforeI
		if lastUserPayoutCycleBeforeI == 0 {
			numAccCycles = i - 2
		}
		if numAccCycles < 0 {
			numAccCycles = 0
		}
		
		accruedInterestForAllLiving := float64(userLivingSlotsBeforeI) * float64(numAccCycles) * interest
		suggestedPayoutAmount := float64(g.MemberCount)*g.ContributionAmount + accruedInterestForAllLiving

		// 3. Expected Contribution amount
		expectedContribution := float64(userCollectedSlotsStrictlyBeforeI)*(g.ContributionAmount+interest) +
			float64(userLivingSlotsBeforeI)*g.ContributionAmount

		ch := histMap[i]
		
		isPaid := ch.Contribution != nil
		var contribID *string
		kind := "uncollected"
		if isPaid {
			contribID = &ch.Contribution.ID
			expectedContribution = ch.Contribution.Amount
			kind = ch.Contribution.Kind
		}

		isPayout := ch.Payout != nil
		var payoutID *string
		payoutAmount := suggestedPayoutAmount
		payoutSlots := 0
		payoutAccruedInterest := accruedInterestForAllLiving
		if isPayout {
			payoutID = &ch.Payout.ID
			payoutAmount = ch.Payout.Amount
			payoutSlots = ch.Payout.SlotsTaken
			// If already paid out, we don't have recorded accrued interest separately,
			// but for planned cycles it's useful.
		}

		// 4. Update state for end of cycle (for Badge status)
		userCollectedSlotsAtEndOfI := userCollectedSlotsStrictlyBeforeI + payoutSlots
		if userCollectedSlotsAtEndOfI > 0 {
			if userCollectedSlotsAtEndOfI < g.UserSlots {
				kind = "partial_collected"
			} else {
				kind = "collected"
			}
		}

		schedule[i-1] = ScheduleCycle{
			CycleNo:           i,
			DueDate:           dueDate.Format("2006-01-02"),
			ExpectedAmount:    expectedContribution,
			Kind:              kind,
			IsPaid:            isPaid,
			ContributionID:    contribID,
			IsPayout:          isPayout,
			PayoutID:          payoutID,
			PayoutAmount:      payoutAmount,
			PayoutSlots:       payoutSlots,
			UserCollectedSlots: userCollectedSlotsAtEndOfI,
			AccruedInterest:    payoutAccruedInterest,
		}
	}

	return schedule
}

func (s *Service) DeleteContribution(ctx context.Context, userID string, groupID string, contributionID string) error {
	c, err := s.repo.GetContribution(ctx, userID, contributionID)
	if err != nil {
		return err
	}
	if c.GroupID != groupID {
		return apperrors.ErrRotatingSavingsContributionNotFound
	}

	// Delete linked transaction if exists
	if c.TransactionID != "" {
		if err := s.txSvc.Delete(ctx, userID, c.TransactionID); err != nil {
			// Nếu giao dịch không tồn tại (đã bị xóa trước đó), vẫn tiếp tục xóa bản ghi hụi
			if !errors.Is(err, apperrors.ErrTransactionNotFound) {
				return err
			}
		}
	}

	if err := s.repo.DeleteContribution(ctx, userID, contributionID); err != nil {
		return err
	}

	_ = s.repo.CreateAuditLog(ctx, domain.RotatingSavingsAuditLog{
		ID:      uuid.New().String(),
		UserID:  userID,
		GroupID: &groupID,
		Action:  "contribution_deleted",
		Details: map[string]any{
			"cycle_no": c.CycleNo,
			"kind":     c.Kind,
			"amount":   c.Amount,
			"note":     c.Note,
		},
		CreatedAt: time.Now(),
	})

	// Re-check status
	history, err := s.repo.ListContributions(ctx, userID, groupID)
	if err == nil {
		group, err := s.repo.GetGroup(ctx, userID, groupID)
		if err == nil {
			completedCycles := make(map[int]bool)
			for _, item := range history {
				if item.CycleNo != nil {
					completedCycles[*item.CycleNo] = true
				}
			}
			if len(completedCycles) < group.MemberCount && group.Status == "completed" {
				group.Status = "active"
				group.UpdatedAt = time.Now()
				_ = s.repo.UpdateGroup(ctx, *group)
			}
		}
	}

	return nil
}

func (s *Service) DeleteGroup(ctx context.Context, userID string, groupID string) error {
	// 1. Lấy tất cả contributions để dọn dẹp transactions
	history, err := s.repo.ListContributions(ctx, userID, groupID)
	if err != nil {
		return err
	}

	// 2. Xóa từng transaction liên kết
	for _, c := range history {
		if c.TransactionID != "" {
			if err := s.txSvc.Delete(ctx, userID, c.TransactionID); err != nil {
				// Bỏ qua lỗi nếu transaction không tồn tại
				if !errors.Is(err, apperrors.ErrTransactionNotFound) {
					return err
				}
			}
		}
	}

	// 3. Xóa dây hụi (Contributions sẽ tự CASCADE DELETE ở DB)
	return s.repo.DeleteGroup(ctx, userID, groupID)
}

func (s *Service) ListGroups(ctx context.Context, userID string) ([]GroupSummary, error) {
	groups, err := s.repo.ListGroups(ctx, userID)
	if err != nil {
		return nil, err
	}

	summaries := make([]GroupSummary, 0, len(groups))
	for _, g := range groups {
		contributions, err := s.repo.ListContributions(ctx, userID, g.ID)
		if err != nil {
			return nil, err
		}

		totalPaid := 0.0
		totalReceived := 0.0
		completedCyclesMap := make(map[int]bool)
		
		for _, c := range contributions {
			if c.Kind == "payout" {
				totalReceived += c.Amount
			} else {
				totalPaid += c.Amount
			}
			if c.CycleNo != nil {
				completedCyclesMap[*c.CycleNo] = true
			}
		}

		schedule := s.generateSchedule(g, contributions)
		var nextDueDate *string
		for _, sc := range schedule {
			if !sc.IsPaid {
				d := sc.DueDate
				nextDueDate = &d
				break
			}
		}

		totalExpected := 0.0
		for _, sc := range schedule {
			totalExpected += sc.ExpectedAmount
		}

		summaries = append(summaries, GroupSummary{
			Group:           g,
			TotalPaid:       totalPaid,
			TotalReceived:   totalReceived,
			RemainingAmount: totalExpected - totalPaid,
			CompletedCycles: len(completedCyclesMap),
			TotalCycles:     g.MemberCount,
			NextDueDate:     nextDueDate,
		})
	}

	return summaries, nil
}

type CreateContributionRequest struct {
	Kind           string   `json:"kind"`
	AccountID      *string  `json:"account_id"`
	OccurredDate   string   `json:"occurred_date"`
	OccurredTime   *string  `json:"occurred_time"`
	Amount           string   `json:"amount"`
	SlotsTaken       int      `json:"slots_taken"`
	CollectedFeePerSlot float64 `json:"collected_fee_per_slot"`
	CycleNo          *int     `json:"cycle_no"`
	DueDate          *string  `json:"due_date"`
	Note             *string  `json:"note"`
}

func (s *Service) CreateContribution(ctx context.Context, userID string, groupID string, req CreateContributionRequest) (*domain.RotatingSavingsContribution, error) {
	if _, err := s.repo.GetGroup(ctx, userID, groupID); err != nil {
		return nil, err
	}

	occurredAt, _ := time.Parse("2006-01-02", req.OccurredDate)
	if req.OccurredTime != nil && *req.OccurredTime != "" {
		t, err := time.Parse("15:04", *req.OccurredTime)
		if err == nil {
			occurredAt = time.Date(occurredAt.Year(), occurredAt.Month(), occurredAt.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)
		}
	}

	// Create real transaction
	txType := "expense"
	if req.Kind == "payout" {
		txType = "income"
	}
	
	desc := "Đóng hụi"
	if req.Kind == "payout" {
		desc = "Lĩnh hụi"
	} else if req.Kind == "collected" {
		desc = "Đóng hụi đã lĩnh"
	}
	
	catID := "cat_sys_rotating_savings_contribution"
	if req.Kind == "payout" {
		catID = "cat_sys_rotating_savings_payout"
	}
	
	txReq := TxCreateRequest{
		Type:         txType,
		OccurredDate: &req.OccurredDate,
		OccurredTime: req.OccurredTime,
		Amount:       req.Amount,
		AccountID:    req.AccountID,
		CategoryID:   &catID,
		Description:  &desc,
		Notes:        req.Note,
	}
	
	tx, err := s.txSvc.Create(ctx, userID, txReq)
	if err != nil {
		return nil, err
	}

	amount, _ := strconv.ParseFloat(req.Amount, 64)

	c := domain.RotatingSavingsContribution{
		ID:             uuid.New().String(),
		GroupID:        groupID,
		TransactionID:  tx.ID,
		Kind:           req.Kind,
		CycleNo:        req.CycleNo,
		DueDate:        req.DueDate,
		Amount:         amount,
		SlotsTaken:     req.SlotsTaken,
		CollectedFeePerSlot: req.CollectedFeePerSlot,
		OccurredAt:     occurredAt,
		Note:           req.Note,
		CreatedAt:      time.Now(),
	}

	if err := s.repo.CreateContribution(ctx, c); err != nil {
		return nil, err
	}

	_ = s.repo.CreateAuditLog(ctx, domain.RotatingSavingsAuditLog{
		ID:      uuid.New().String(),
		UserID:  userID,
		GroupID: &groupID,
		Action:  "contribution_created",
		Details: map[string]any{
			"cycle_no": c.CycleNo,
			"kind":     c.Kind,
			"amount":   c.Amount,
			"note":     c.Note,
		},
		CreatedAt: time.Now(),
	})

	// 1. Get all contributions to check completions
	history, err := s.repo.ListContributions(ctx, userID, groupID)
	if err == nil {
		group, err := s.repo.GetGroup(ctx, userID, groupID)
		if err == nil {
			completedCycles := make(map[int]bool)
			for _, item := range history {
				if item.CycleNo != nil {
					completedCycles[*item.CycleNo] = true
				}
			}
			if len(completedCycles) >= group.MemberCount {
				group.Status = "completed"
				group.UpdatedAt = time.Now()
				_ = s.repo.UpdateGroup(ctx, *group)
			}
		}
	}

	return &c, nil
}

func (s *Service) ListContributions(ctx context.Context, userID string, groupID string) ([]domain.RotatingSavingsContribution, error) {
	return s.repo.ListContributions(ctx, userID, groupID)
}
