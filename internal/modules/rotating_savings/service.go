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
		Status:              req.Status,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if g.UserSlots <= 0 {
		g.UserSlots = 1
	}

	if err := s.repo.CreateGroup(ctx, g); err != nil {
		return nil, err
	}

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
}

type GroupSummary struct {
	Group           domain.RotatingSavingsGroup `json:"group"`
	TotalPaid       float64                     `json:"total_paid"`
	TotalReceived   float64                     `json:"total_received"`
	NetPosition     float64                     `json:"net_position"`
	CompletedCycles int                         `json:"completed_cycles"`
	TotalCycles     int                         `json:"total_cycles"`
	NextDueDate     *string                     `json:"next_due_date"`
}

type GroupDetailResponse struct {
	Group              domain.RotatingSavingsGroup          `json:"group"`
	Schedule           []ScheduleCycle                       `json:"schedule"`
	CollectedSlotsCount int                                  `json:"collected_slots_count"`
	CurrentPayoutValue float64                               `json:"current_payout_value"`
	Contributions      []domain.RotatingSavingsContribution `json:"contributions"`
	TotalPaid          float64                               `json:"total_paid"`
	TotalReceived      float64                               `json:"total_received"`
	NextPayment        float64                               `json:"next_payment"`
	NetPosition        float64                               `json:"net_position"`
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

	// Calculate payout value for the NEXT payout based on historical dead slots
	// Logic: payout at cycle N = Base * TotalSlots + (Dead slots from 1 to N-1) * Interest
	var nextPayoutCycle int
	lastPayoutCycle := 0
	for _, c := range contributions {
		if c.Kind == "payout" && c.CycleNo != nil && *c.CycleNo > lastPayoutCycle {
			lastPayoutCycle = *c.CycleNo
		}
	}
	nextPayoutCycle = lastPayoutCycle + 1
	if nextPayoutCycle > g.MemberCount {
		nextPayoutCycle = g.MemberCount
	}

	deadSlotsBeforeNext := 0
	for _, c := range contributions {
		if c.Kind == "payout" && c.CycleNo != nil && *c.CycleNo < nextPayoutCycle {
			deadSlotsBeforeNext += c.SlotsTaken
		}
	}

	payoutValue := float64(g.MemberCount)*g.ContributionAmount + float64(deadSlotsBeforeNext)*interest

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

	return &GroupDetailResponse{
		Group:              *g,
		Schedule:           schedule,
		CollectedSlotsCount: collectedSlotsCount,
		CurrentPayoutValue: payoutValue,
		Contributions:      contributions,
		TotalPaid:          totalPaid,
		TotalReceived:      totalReceived,
		NextPayment:        nextPayment,
		NetPosition:        totalReceived - totalPaid,
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
	histMap := make(map[int]domain.RotatingSavingsContribution)
	for _, c := range history {
		if c.CycleNo != nil {
			histMap[*c.CycleNo] = c
		}
	}

	// Sort history by cycle number to calculate cumulative dead slots/fees
	sortedHist := make([]domain.RotatingSavingsContribution, 0)
	for _, c := range history {
		if c.CycleNo != nil {
			sortedHist = append(sortedHist, c)
		}
	}
	sort.Slice(sortedHist, func(i, j int) bool {
		return *sortedHist[i].CycleNo < *sortedHist[j].CycleNo
	})

	for i := 1; i <= count; i++ {
		dueDate := startDate
		switch g.CycleFrequency {
		case "weekly":
			dueDate = startDate.AddDate(0, 0, (i-1)*7)
		case "monthly":
			dueDate = startDate.AddDate(0, (i-1), 0)
		}

		// Calculate cumulative collected slots up to this cycle
		groupCollectedSlots := 0
		for _, h := range sortedHist {
			if h.CycleNo != nil && *h.CycleNo <= i && h.Kind == "payout" {
				groupCollectedSlots += h.SlotsTaken
			}
		}

		interest := 0.0
		if g.FixedInterestAmount != nil {
			interest = *g.FixedInterestAmount
		}

		kind := "uncollected"
		// Logic: if user's slots taken in history includes this cycle, it's a payout.
		// If user has collected slots from earlier cycles, they pay Base + Interest.
		// If user is uncollected, they pay Base.
		
		// Sửa lại cách tính số suất đã lĩnh: Số suất đã lĩnh cho kỳ i là tổng số suất đã lĩnh TRƯỚC kỳ i.
		// (Nếu lĩnh ở kỳ i, suất đó vẫn được coi là chưa lĩnh khi CHƯA lĩnh xong kỳ đó).
		userCollectedSlots := 0
		for _, h := range sortedHist {
			if h.CycleNo != nil && *h.CycleNo < i && h.Kind == "payout" {
				userCollectedSlots += h.SlotsTaken
			}
		}

		expectedAmount := float64(g.UserSlots-userCollectedSlots)*g.ContributionAmount + float64(userCollectedSlots)*(g.ContributionAmount+interest)

		if userCollectedSlots > 0 {
			if userCollectedSlots < g.UserSlots {
				kind = "partial_collected"
			} else {
				kind = "collected"
			}
		}

		// Group history by kind for this cycle
		var contrib *domain.RotatingSavingsContribution
		var payout *domain.RotatingSavingsContribution
		
		for _, h := range history {
			if h.CycleNo != nil && *h.CycleNo == i {
				if h.Kind == "payout" {
					payout = &h
				} else {
					// Lấy bản ghi xác nhận đóng hụi (living hoặc dead)
					contrib = &h
				}
			}
		}

		isPaid := contrib != nil
		var contribID *string
		if isPaid {
			id := contrib.ID
			contribID = &id
			expectedAmount = contrib.Amount
			kind = contrib.Kind
		}

		isPayout := payout != nil
		var payoutID *string
		payoutAmount := 0.0
		payoutSlots := 0
		if isPayout {
			id := payout.ID
			payoutID = &id
			payoutAmount = payout.Amount
			payoutSlots = payout.SlotsTaken
		}

		// displayUserCollectedSlots includes payouts in the current cycle for audit/status display
		displayUserCollectedSlots := 0
		for _, h := range sortedHist {
			if h.CycleNo != nil && *h.CycleNo <= i && h.Kind == "payout" {
				displayUserCollectedSlots += h.SlotsTaken
			}
		}

		schedule[i-1] = ScheduleCycle{
			CycleNo:           i,
			DueDate:           dueDate.Format("2006-01-02"),
			ExpectedAmount:    expectedAmount,
			Kind:              kind,
			IsPaid:            isPaid,
			ContributionID:    contribID,
			IsPayout:          isPayout,
			PayoutID:          payoutID,
			PayoutAmount:      payoutAmount,
			PayoutSlots:       payoutSlots,
			GroupCollectedSlots: groupCollectedSlots,
			UserCollectedSlots:  displayUserCollectedSlots,
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

	return s.repo.DeleteContribution(ctx, userID, contributionID)
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

		summaries = append(summaries, GroupSummary{
			Group:           g,
			TotalPaid:       totalPaid,
			TotalReceived:   totalReceived,
			NetPosition:     totalReceived - totalPaid,
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

	return &c, nil
}

func (s *Service) ListContributions(ctx context.Context, userID string, groupID string) ([]domain.RotatingSavingsContribution, error) {
	return s.repo.ListContributions(ctx, userID, groupID)
}
