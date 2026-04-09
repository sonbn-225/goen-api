package service

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
)

// RotatingSavingsService quản lý các nhóm "Hụi/Họ" (Rotating Savings and Credit Associations - ROSCA).
// Nó xử lý việc tạo nhóm, lịch trình thanh toán và các khoản đóng góp/lĩnh hụi của thành viên.
// Tất cả các biến động tài chính đều được ghi lại dưới dạng giao dịch sổ cái trong TransactionService trung tâm.
type RotatingSavingsService struct {
	repo         interfaces.RotatingSavingsRepository
	txRepo       interfaces.TransactionRepository
	categoryRepo interfaces.CategoryRepository
	accountSvc   interfaces.AccountService
	txSvc        interfaces.TransactionService
	db           *database.Postgres
}

// NewRotatingSavingsService khởi tạo một dịch vụ quản lý hụi/họ mới.
func NewRotatingSavingsService(
	repo interfaces.RotatingSavingsRepository,
	txRepo interfaces.TransactionRepository,
	categoryRepo interfaces.CategoryRepository,
	accountSvc interfaces.AccountService,
	txSvc interfaces.TransactionService,
	db *database.Postgres,
) *RotatingSavingsService {
	return &RotatingSavingsService{
		repo:         repo,
		txRepo:       txRepo,
		categoryRepo: categoryRepo,
		accountSvc:   accountSvc,
		txSvc:        txSvc,
		db:           db,
	}
}

// CreateGroup khởi tạo một nhóm hụi mới.
func (s *RotatingSavingsService) CreateGroup(ctx context.Context, userID uuid.UUID, req dto.CreateRotatingSavingsGroupRequest) (*dto.RotatingSavingsGroupResponse, error) {
	status := req.Status
	if status == "" {
		status = entity.RotatingSavingsStatusActive
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

	err := s.db.WithTx(ctx, func(tx pgx.Tx) error {
		if err := s.repo.CreateRotatingGroupTx(ctx, tx, g); err != nil {
			return err
		}

		return s.repo.AddAuditLogTx(ctx, tx, entity.RotatingSavingsAuditLog{
			BaseEntity: entity.BaseEntity{
				ID: utils.NewID(),
			},
			UserID:    userID,
			GroupID:   &g.ID,
			Action:    entity.RotatingSavingsAuditActionGroupCreated,
			Details:   map[string]any{"name": g.Name},
			CreatedAt: utils.Now(),
			UpdatedAt: utils.Now(),
		})
	})

	if err != nil {
		return nil, err
	}

	created, err := s.repo.GetRotatingGroupTx(ctx, nil, userID, g.ID)
	if err != nil {
		return nil, err
	}
	if created != nil {
		g = *created
	}

	resp := dto.NewRotatingSavingsGroupResponse(g)
	return &resp, nil
}

// GetGroup lấy thông tin cơ bản về một nhóm.
func (s *RotatingSavingsService) GetGroup(ctx context.Context, userID, groupID uuid.UUID) (*dto.RotatingSavingsGroupResponse, error) {
	g, err := s.repo.GetRotatingGroupTx(ctx, nil, userID, groupID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, nil
	}
	resp := dto.NewRotatingSavingsGroupResponse(*g)
	return &resp, nil
}

// UpdateGroup cập nhật thông tin cấu hình nhóm.
func (s *RotatingSavingsService) UpdateGroup(ctx context.Context, userID, groupID uuid.UUID, req dto.UpdateRotatingSavingsGroupRequest) (*dto.RotatingSavingsGroupResponse, error) {
	g, err := s.repo.GetRotatingGroupTx(ctx, nil, userID, groupID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, apperr.NotFound("group not found")
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

	err = s.db.WithTx(ctx, func(tx pgx.Tx) error {
		if err := s.repo.UpdateRotatingGroupTx(ctx, tx, *g); err != nil {
			return err
		}

		return s.repo.AddAuditLogTx(ctx, tx, entity.RotatingSavingsAuditLog{
			BaseEntity: entity.BaseEntity{
				ID: utils.NewID(),
			},
			UserID:    userID,
			GroupID:   &g.ID,
			Action:    entity.RotatingSavingsAuditActionGroupUpdated,
			Details:   map[string]any{"status": g.Status},
			CreatedAt: utils.Now(),
			UpdatedAt: utils.Now(),
		})
	})

	if err != nil {
		return nil, err
	}

	updated, err := s.repo.GetRotatingGroupTx(ctx, nil, userID, groupID)
	if err != nil {
		return nil, err
	}
	if updated != nil {
		g = updated
	}

	resp := dto.NewRotatingSavingsGroupResponse(*g)
	return &resp, nil
}

// GetGroupDetail lấy đầy đủ thông tin chi tiết nhóm bao gồm lịch trình và lịch sử.
func (s *RotatingSavingsService) GetGroupDetail(ctx context.Context, userID, groupID uuid.UUID) (*dto.RotatingSavingsGroupDetailResponse, error) {
	g, err := s.repo.GetRotatingGroupTx(ctx, nil, userID, groupID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, apperr.NotFound("group not found")
	}

	contributions, err := s.repo.ListContributionsTx(ctx, nil, groupID)
	if err != nil {
		return nil, err
	}

	auditLogs, _ := s.repo.ListAuditLogsTx(ctx, nil, groupID)

	schedule := s.generateSchedule(*g, contributions)

	collectedSlotsCount := 0
	totalPaid := 0.0
	totalReceived := 0.0
	for _, c := range contributions {
		if c.Kind == entity.RotatingSavingsContributionKindPayout {
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

	// Logic gợi ý (đã đơn giản hóa)
	payoutValue := float64(g.MemberCount) * g.ContributionAmount

	// Logic lãi tích lũy có thể phức tạp hơn, giữ nguyên tính năng cũ
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
			if c.Kind == entity.RotatingSavingsContributionKindPayout {
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
			if c.Kind == entity.RotatingSavingsContributionKindPayout && c.CycleNo != nil && *c.CycleNo < i {
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

// ListGroups liệt kê tất cả các nhóm kèm theo tóm tắt tiến độ.
func (s *RotatingSavingsService) ListGroups(ctx context.Context, userID uuid.UUID) ([]dto.RotatingSavingsGroupSummary, error) {
	groups, err := s.repo.ListRotatingGroupsTx(ctx, nil, userID)
	if err != nil {
		return nil, err
	}

	summaries := make([]dto.RotatingSavingsGroupSummary, 0, len(groups))
	for _, g := range groups {
		contributions, _ := s.repo.ListContributionsTx(ctx, nil, g.ID)
		totalPaid := 0.0
		totalReceived := 0.0
		completedCycles := make(map[int]bool)
		for _, c := range contributions {
			if c.Kind == entity.RotatingSavingsContributionKindPayout {
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

// CreateContribution ghi lại việc đóng hụi hoặc lĩnh hụi của thành viên.
// Nó tạo ra một giao dịch sổ cái một cách nguyên tử trong tài khoản được chỉ định và
// liên kết nó với bản ghi của nhóm.
func (s *RotatingSavingsService) CreateContribution(ctx context.Context, userID, groupID uuid.UUID, req dto.RotatingSavingsContributionRequest) (*dto.RotatingSavingsContributionResponse, error) {
	g, err := s.repo.GetRotatingGroupTx(ctx, nil, userID, groupID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, apperr.NotFound("group not found")
	}

	txType := entity.TransactionTypeExpense
	if req.Kind == entity.RotatingSavingsContributionKindPayout {
		txType = entity.TransactionTypeIncome
	}

	desc := "Đóng hụi"
	if req.Kind == entity.RotatingSavingsContributionKindPayout {
		desc = "Lĩnh hụi"
	}

	var catID uuid.UUID
	catKey := "cat_sys_rotating_savings_contribution"
	if req.Kind == entity.RotatingSavingsContributionKindPayout {
		catKey = "cat_sys_rotating_savings_payout"
	}

	cat, err := s.categoryRepo.GetCategoryByKeyTx(ctx, nil, catKey)
	if err != nil {
		return nil, err
	}
	catID = cat.ID

	if req.Note != nil && *req.Note != "" {
		desc = desc + " - " + *req.Note
	}

	parsedAmount, _ := strconv.ParseFloat(req.Amount, 64)
	occAt, _ := time.Parse("2006-01-02", req.OccurredDate)

	var resp *dto.RotatingSavingsContributionResponse

	err = s.db.WithTx(ctx, func(tx pgx.Tx) error {
		// 1. Create Ledger Transaction
		ledgerTx := entity.Transaction{
			AuditEntity:  entity.AuditEntity{BaseEntity: entity.BaseEntity{ID: utils.NewID()}},
			Type:         txType,
			OccurredAt:   occAt.UTC(),
			OccurredDate: req.OccurredDate,
			Amount:       req.Amount,
			Description:  &desc,
			AccountID:    &g.AccountID,
			Status:       entity.TransactionStatusPosted,
		}
		ledgerLine := []entity.TransactionLineItem{
			{BaseEntity: entity.BaseEntity{ID: utils.NewID()}, Amount: req.Amount, CategoryID: &catID, Note: &desc},
		}

		if err := s.txRepo.CreateTransactionTx(ctx, tx, userID, ledgerTx, ledgerLine, nil); err != nil {
			return err
		}

		// 2. Create Contribution Record
		c := entity.RotatingSavingsContribution{
			BaseEntity: entity.BaseEntity{
				ID: utils.NewID(),
			},
			GroupID:             groupID,
			TransactionID:       ledgerTx.ID,
			Kind:                req.Kind,
			CycleNo:             req.CycleNo,
			DueDate:             req.DueDate,
			Amount:              parsedAmount,
			SlotsTaken:          req.SlotsTaken,
			CollectedFeePerSlot: req.CollectedFeePerSlot,
			OccurredAt:          utils.Now(),
			Note:                req.Note,
			CreatedAt:           utils.Now(),
			UpdatedAt:           utils.Now(),
		}

		if err := s.repo.CreateContributionTx(ctx, tx, c); err != nil {
			return err
		}

		// 3. Add Audit Log
		_ = s.repo.AddAuditLogTx(ctx, tx, entity.RotatingSavingsAuditLog{
			BaseEntity: entity.BaseEntity{
				ID: utils.NewID(),
			},
			UserID:    userID,
			GroupID:   &groupID,
			Action:    entity.RotatingSavingsAuditActionContributionCreated,
			Details:   map[string]any{"kind": c.Kind, "amount": req.Amount},
			CreatedAt: utils.Now(),
			UpdatedAt: utils.Now(),
		})

		tr := dto.NewRotatingSavingsContributionResponse(c)
		resp = &tr
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DeleteContribution xóa một bản ghi đóng/lĩnh hụi.
func (s *RotatingSavingsService) DeleteContribution(ctx context.Context, userID, groupID, id uuid.UUID) error {
	// Tìm kiếm đóng góp trong nhóm vì GetContribution đã bị gỡ bỏ khỏi interface để hợp nhất
	conts, err := s.repo.ListContributionsTx(ctx, nil, groupID)
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
		return apperr.NotFound("contribution not found")
	}

	return s.db.WithTx(ctx, func(tx pgx.Tx) error {
		if target.TransactionID != uuid.Nil {
			_ = s.txSvc.Delete(ctx, userID, target.TransactionID)
		}

		return s.repo.DeleteContributionTx(ctx, tx, id)
	})
}

// DeleteGroup xóa toàn bộ nhóm và các giao dịch liên quan.
func (s *RotatingSavingsService) DeleteGroup(ctx context.Context, userID, groupID uuid.UUID) error {
	return s.db.WithTx(ctx, func(tx pgx.Tx) error {
		conts, _ := s.repo.ListContributionsTx(ctx, tx, groupID)
		for _, c := range conts {
			if c.TransactionID != uuid.Nil {
				_ = s.txSvc.Delete(ctx, userID, c.TransactionID)
			}
		}
		return s.repo.DeleteRotatingGroupTx(ctx, tx, userID, groupID)
	})
}

func (s *RotatingSavingsService) CleanupTransactionLinksTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, transactionID uuid.UUID) error {
	return s.repo.DeleteContributionByTransactionTx(ctx, tx, transactionID)
}
