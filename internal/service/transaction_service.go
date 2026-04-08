package service

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sonbn-225/goen-api/internal/domain/dto"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
	"github.com/sonbn-225/goen-api/internal/pkg/database"
	"github.com/sonbn-225/goen-api/internal/pkg/utils"
)

// TransactionService xử lý logic cốt lõi của sổ cái ứng dụng.
// Nó đóng vai trò là "Sổ cái trung tâm", ghi lại mọi biến động tiền tệ
// qua các tài khoản, hạng mục và nhãn. Các dịch vụ khác (Nợ, Đầu tư, Tiết kiệm)
// phụ thuộc vào dịch vụ này để phản ánh trạng thái tài chính của họ trong số dư của người dùng.
type TransactionService struct {
	repo        interfaces.TransactionRepository
	tagSvc      interfaces.TagService
	debtSvc     interfaces.DebtService
	accountRepo interfaces.AccountRepository
	db          *database.Postgres
}

// NewTransactionService khởi tạo một thực thể mới của dịch vụ sổ cái trung tâm.
func NewTransactionService(repo interfaces.TransactionRepository, tagSvc interfaces.TagService, accountRepo interfaces.AccountRepository, db *database.Postgres) *TransactionService {
	return &TransactionService{repo: repo, tagSvc: tagSvc, accountRepo: accountRepo, db: db}
}

// SetDebtService được sử dụng để tiêm phụ thuộc nhằm giải quyết vòng lặp phụ thuộc với DebtService.
func (s *TransactionService) SetDebtService(ds interfaces.DebtService) {
	s.debtSvc = ds
}

// List trả về danh sách phân trang các giao dịch được lọc theo nhiều tiêu chí khác nhau.
func (s *TransactionService) List(ctx context.Context, userID uuid.UUID, req dto.ListTransactionsRequest) ([]dto.TransactionResponse, *string, int, error) {
	filter := entity.TransactionListFilter{
		Page:  req.Page,
		Limit: req.Limit,
	}

	if req.AccountID != nil {
		filter.AccountID = req.AccountID
	}
	if req.CategoryID != nil {
		filter.CategoryID = req.CategoryID
	}
	filter.Type = req.Type
	filter.Search = req.Search

	if req.From != nil {
		t, err := utils.ParseTimeOrDate(*req.From)
		if err == nil {
			filter.From = &t
		}
	}
	if req.To != nil {
		t, err := utils.ParseTimeOrDate(*req.To)
		if err == nil {
			filter.To = &t
		}
	}

	items, cursor, total, err := s.repo.ListTransactions(ctx, userID, filter)
	if err != nil {
		return nil, nil, 0, err
	}
	return dto.NewTransactionResponses(items), cursor, total, nil
}

// Get lấy thông tin một giao dịch duy nhất theo ID, đảm bảo nó thuộc về người dùng được chỉ định.
func (s *TransactionService) Get(ctx context.Context, userID, transactionID uuid.UUID) (*dto.TransactionResponse, error) {
	it, err := s.repo.GetTransactionTx(ctx, nil, userID, transactionID)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, nil
	}
	resp := dto.NewTransactionResponse(*it)
	return &resp, nil
}

// Create ghi nhận một biến động tài chính mới. Nó xử lý:
// 1. Chi phí/Thu nhập thông thường (với tùy chọn chia nhỏ các hạng mục chi tiết)
// 2. Chuyển khoản giữa các tài khoản (từ tài khoản gốc -> tài khoản đích)
// 3. Chi phí dùng chung (tự động tạo nợ cho những người tham gia)
// Tất cả các thao tác được thực hiện trong một giao dịch cơ sở dữ liệu duy nhất để đảm bảo tính nguyên tử.
func (s *TransactionService) Create(ctx context.Context, userID uuid.UUID, req dto.CreateTransactionRequest) (*dto.TransactionResponse, error) {
	kind := strings.TrimSpace(string(req.Type))
	if kind != string(entity.TransactionTypeExpense) && kind != string(entity.TransactionTypeIncome) && kind != string(entity.TransactionTypeTransfer) {
		return nil, apperr.BadRequest("invalid_type", "invalid transaction type")
	}

	amount := strings.TrimSpace(req.Amount)
	if !utils.IsValidDecimal(amount) {
		return nil, apperr.BadRequest("invalid_amount", "invalid decimal amount")
	}

	fromAmount := utils.NormalizeOptionalString(req.FromAmount)
	toAmount := utils.NormalizeOptionalString(req.ToAmount)
	if (fromAmount != nil) != (toAmount != nil) {
		return nil, apperr.BadRequest("invalid_fx", "from_amount and to_amount must be provided together for FX transfers")
	}

	occurredAt, occurredDate, err := utils.NormalizeOccurredAt(req.OccurredAt, req.OccurredDate, req.OccurredTime)
	if err != nil {
		return nil, err
	}

	lineItems := make([]entity.TransactionLineItem, 0, len(req.LineItems))
	if kind == "income" && len(req.LineItems) > 1 {
		return nil, apperr.BadRequest("invalid_line_items", "income transactions support a single line item only")
	}
	if kind == "income" && len(req.GroupParticipants) > 0 {
		return nil, apperr.BadRequest("invalid_participants", "group participants are only supported for expense transactions")
	}

	// If CategoryID is top-level and no line items, create a default one.
	if len(req.LineItems) == 0 && req.CategoryID != nil && *req.CategoryID != uuid.Nil {
		lineItems = append(lineItems, entity.TransactionLineItem{
			BaseEntity: entity.BaseEntity{
				ID: utils.NewID(),
			},
			CategoryID: req.CategoryID,
			Amount:     amount,
		})
	}

	// Resolve tags for transaction level if TagService is available
	tagIDs, _ := s.ensureTags(ctx, userID, req.TagIDs, req.Lang)

	// Sum line items if present (except transfer)
	if kind != "transfer" && len(req.LineItems) > 0 {
		sum := big.NewRat(0, 1)
		for _, li := range req.LineItems {
			if !utils.IsValidDecimal(li.Amount) {
				return nil, apperr.BadRequest("invalid_amount", "invalid line item amount")
			}
			r, _ := new(big.Rat).SetString(li.Amount)
			sum.Add(sum, r)

			// Resolve tags for line item if TagService is available
			liTags, _ := s.ensureTags(ctx, userID, li.TagIDs, req.Lang)

			var lCatID *uuid.UUID = li.CategoryID

			lineItems = append(lineItems, entity.TransactionLineItem{
				BaseEntity: entity.BaseEntity{
					ID: utils.NewID(),
				},
				CategoryID: lCatID,
				Amount:     li.Amount,
				Note:       utils.NormalizeOptionalString(li.Note),
				TagIDs:     liTags,
			})
		}
		amount = sum.FloatString(2)
	}

	description := utils.NormalizeOptionalString(req.Description)
	if description != nil && len(lineItems) > 0 {
		if lineItems[0].Note == nil || strings.TrimSpace(*lineItems[0].Note) == "" {
			lineItems[0].Note = description
		}
	}

	id := utils.NewID()

	tx := entity.Transaction{
		AuditEntity: entity.AuditEntity{
			BaseEntity: entity.BaseEntity{
				ID: id,
			},
		},
		ExternalRef:   utils.NormalizeOptionalString(req.ExternalRef),
		Type:          entity.TransactionType(kind),
		OccurredAt:    occurredAt,
		OccurredDate:  occurredDate,
		Amount:        amount,
		FromAmount:    fromAmount,
		ToAmount:      toAmount,
		Description:   description,
		AccountID:     req.AccountID,
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		ExchangeRate:  utils.NormalizeOptionalString(req.ExchangeRate),
		Status:        entity.TransactionStatusPosted,
	}

	var resp *dto.TransactionResponse
	err = s.db.WithTx(ctx, func(txConn pgx.Tx) error {
		// 1. Create Transaction
		if err := s.repo.CreateTransactionTx(ctx, txConn, userID, tx, lineItems, tagIDs); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == "fk_tli_category" {
				return apperr.BadRequest("invalid_category", "invalid category_id")
			}
			return err
		}

		// 2. Create Debts for Shared Expenses
		if len(req.GroupParticipants) > 0 {
			shares := s.allocateGroupParticipants(userID, id, tx.Amount, req.OwnerOriginalAmount, req.GroupParticipants)
			if s.debtSvc != nil && tx.AccountID != nil {
				for _, share := range shares {
					debtName := share.Name
					if description != nil {
						debtName = *description + " (" + share.Name + ")"
					}
					originTxId := id.String()
					_, err := s.debtSvc.CreateTx(ctx, txConn, userID, dto.CreateDebtRequest{
						AccountID:                tx.AccountID.String(),
						OriginatingTransactionID: &originTxId,
						Direction:                "lent",
						Name:                     &debtName,
						ContactName:              &share.Name,
						Principal:                share.Amount,
						StartDate:                tx.OccurredDate,
						DueDate:                  "2099-12-31", // Default far-future due date
					})
					if err != nil {
						return err
					}
				}
			}
		}

		// Fetch back for response
		it, err := s.repo.GetTransactionTx(ctx, txConn, userID, id)
		if err != nil {
			return err
		}
		if it == nil {
			return errors.New("failed to retrieve created transaction")
		}
		tr := dto.NewTransactionResponse(*it)
		resp = &tr

		// Audit Logging
		if it.AccountID != nil {
			_ = s.accountRepo.RecordAccountAuditEventTx(ctx, txConn, entity.AccountAuditEvent{
				BaseEntity:  entity.BaseEntity{ID: utils.NewID()},
				AccountID:   *it.AccountID,
				ActorUserID: userID,
				Action:      "transaction_created",
				EntityType:  "transaction",
				EntityID:    it.ID,
				OccurredAt:  utils.Now(),
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Patch áp dụng các bản cập nhật một phần cho một giao dịch hiện có.
// Nó hỗ trợ cập nhật mô tả, số tiền, trạng thái, hạng mục và nhãn.
func (s *TransactionService) Patch(ctx context.Context, userID, transactionID uuid.UUID, req dto.TransactionPatchRequest) (*dto.TransactionResponse, error) {
	tagIDs, _ := s.ensureTags(ctx, userID, req.TagIDs, req.Lang)

	patch := entity.TransactionPatch{
		Description: utils.NormalizeOptionalString(req.Description),
		CategoryIDs: req.CategoryIDs,
		TagIDs:      tagIDs,
		Amount:      utils.NormalizeOptionalString(req.Amount),
		Status:      req.Status,
	}
	if req.OccurredAt != nil {
		t, err := utils.ParseTimeOrDate(*req.OccurredAt)
		if err == nil {
			patch.OccurredAt = &t
		}
	}
	if req.LineItems != nil {
		lis := make([]entity.TransactionLineItem, len(*req.LineItems))
		for i, li := range *req.LineItems {
			lis[i] = entity.TransactionLineItem{
				CategoryID: li.CategoryID,
				TagIDs:     li.TagIDs,
				Amount:     li.Amount,
				Note:       li.Note,
			}
		}
		patch.LineItems = &lis
	}

	it, err := s.repo.PatchTransaction(ctx, userID, transactionID, patch)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, nil
	}
	resp := dto.NewTransactionResponse(*it)

	// Audit Logging
	if it.AccountID != nil {
		var diff map[string]any
		if patchBytes, err := json.Marshal(patch); err == nil {
			_ = json.Unmarshal(patchBytes, &diff)
		}

		_ = s.accountRepo.RecordAccountAuditEventTx(ctx, nil, entity.AccountAuditEvent{
			BaseEntity:  entity.BaseEntity{ID: utils.NewID()},
			AccountID:   *it.AccountID,
			ActorUserID: userID,
			Action:      "transaction_updated",
			EntityType:  "transaction",
			EntityID:    it.ID,
			OccurredAt:  utils.Now(),
			Diff:        diff,
		})
	}

	return &resp, nil
}

func (s *TransactionService) BatchPatch(ctx context.Context, userID uuid.UUID, req dto.BatchPatchRequest) (*dto.BatchPatchResult, error) {
	mode := "atomic"
	if req.Mode != nil {
		mode = *req.Mode
	}

	tagIDs, _ := s.ensureTags(ctx, userID, req.Patch.TagIDs, req.Patch.Lang)

	patches := make(map[uuid.UUID]entity.TransactionPatch, len(req.TransactionIDs))
	for _, id := range req.TransactionIDs {
		// Just reuse same patch payload for all (legacy behavior)
		p := entity.TransactionPatch{
			Description: utils.NormalizeOptionalString(req.Patch.Description),
			CategoryIDs: req.Patch.CategoryIDs,
			TagIDs:      tagIDs,
			Amount:      utils.NormalizeOptionalString(req.Patch.Amount),
			Status:      req.Patch.Status,
		}
		if req.Patch.LineItems != nil {
			lis := make([]entity.TransactionLineItem, len(*req.Patch.LineItems))
			for i, li := range *req.Patch.LineItems {
				lis[i] = entity.TransactionLineItem{
					CategoryID: li.CategoryID,
					TagIDs:     li.TagIDs,
					Amount:     li.Amount,
					Note:       li.Note,
				}
			}
			p.LineItems = &lis
		}
		patches[id] = p
	}

	updated, failed, err := s.repo.BatchPatchTransactions(ctx, userID, req.TransactionIDs, patches, mode)
	if err != nil {
		return nil, err
	}

	return &dto.BatchPatchResult{
		Mode:         mode,
		UpdatedCount: len(updated),
		FailedCount:  len(failed),
		UpdatedIDs:   updated,
		FailedIDs:    failed,
	}, nil
}

// Delete xóa một giao dịch và xử lý các ràng buộc liên quan.
func (s *TransactionService) Delete(ctx context.Context, userID, transactionID uuid.UUID) error {
	return s.db.WithTx(ctx, func(txConn pgx.Tx) error {
		// 1. Lấy thông tin giao dịch để biết tài khoản phục vụ mục đích kiểm toán
		tx, _ := s.repo.GetTransactionTx(ctx, txConn, userID, transactionID)

		// 2. Dọn dẹp các khoản nợ/thanh toán liên kết
		if s.debtSvc != nil {
			if err := s.debtSvc.CleanupTransactionLinksTx(ctx, txConn, userID, transactionID); err != nil {
				return err
			}
		}

		// 3. Xóa chính giao dịch đó
		if err := s.repo.DeleteTransactionTx(ctx, txConn, userID, transactionID); err != nil {
			return err
		}

		// 4. Ghi nhật ký kiểm toán
		if tx != nil && tx.AccountID != nil {
			_ = s.accountRepo.RecordAccountAuditEventTx(ctx, txConn, entity.AccountAuditEvent{
				BaseEntity:  entity.BaseEntity{ID: utils.NewID()},
				AccountID:   *tx.AccountID,
				ActorUserID: userID,
				Action:      "transaction_deleted",
				EntityType:  "transaction",
				EntityID:    transactionID,
				OccurredAt:  utils.Now(),
			})
		}
		return nil
	})
}

func (s *TransactionService) ListForExport(ctx context.Context, userID uuid.UUID, filter entity.TransactionListFilter) ([]entity.ExportTransactionRow, error) {
	// Reuse existing search/list logic but with a larger limit for direct export
	filter.Limit = 10000
	transactions, _, _, err := s.repo.ListTransactions(ctx, userID, filter)
	if err != nil {
		return nil, err
	}

	rows := make([]entity.ExportTransactionRow, len(transactions))
	for i, t := range transactions {
		var catName *string
		if len(t.CategoryNames) > 0 {
			catName = &t.CategoryNames[0]
		}
		var tagName *string
		if len(t.TagNames) > 0 {
			tagName = &t.TagNames[0]
		}

		rows[i] = entity.ExportTransactionRow{
			ID:           t.ID,
			Description:  t.Description,
			Amount:       t.Amount,
			Type:         string(t.Type),
			OccurredDate: t.OccurredDate,
			AccountName:  t.AccountName,
			CategoryName: catName,
			TagName:      tagName,
			ExternalRef:  t.ExternalRef,
		}
	}

	return rows, nil
}

// Import xử lý việc ghi nhận hàng loạt giao dịch từ dữ liệu thô.
func (s *TransactionService) Import(ctx context.Context, userID uuid.UUID, reqs []dto.CreateTransactionRequest) ([]dto.TransactionResponse, error) {
	resps := make([]dto.TransactionResponse, 0, len(reqs))
	for _, req := range reqs {
		resp, err := s.Create(ctx, userID, req)
		if err != nil {
			return nil, err
		}
		if resp != nil {
			resps = append(resps, *resp)
		}
	}
	return resps, nil
}

// Helpers
func (s *TransactionService) ensureTags(ctx context.Context, userID uuid.UUID, inputs []uuid.UUID, lang string) ([]uuid.UUID, error) {

	if s.tagSvc == nil || len(inputs) == 0 {
		return nil, nil
	}
	// Note: now inputs are already UUIDs from DTO.
	// But ensureTags was also handling names.
	// I'll update it to accept the new type but keep logic for creating tags if needed?
	// Actually, if they are already UUIDs, we just return them.
	// Wait, if the handler receives strings, it should call a different method if it wants to resolve names.
	// Currently the DTO uses uuid.UUID, which means the JSON must contain valid UUID strings.
	// If the user wants to create a tag by name, they should probably do it separately or we need a special DTO field.
	// For now, I'll just return the UUIDs.
	return inputs, nil
}

type ComputedShare struct {
	Name   string
	Amount string
}

func (s *TransactionService) allocateGroupParticipants(userID uuid.UUID, txID uuid.UUID, txAmt string, ownerAmt *string, inputs []dto.GroupParticipantInput) []ComputedShare {
	totalPaid, ok := new(big.Rat).SetString(txAmt)
	if !ok {
		return nil
	}

	type person struct {
		name     string
		original *big.Rat
		origStr  string
	}
	involved := []person{}
	if ownerAmt != nil && *ownerAmt != "" {
		if r, ok := new(big.Rat).SetString(*ownerAmt); ok && r.Sign() > 0 {
			involved = append(involved, person{name: "owner", original: r, origStr: *ownerAmt})
		}
	}
	for _, p := range inputs {
		if r, ok := new(big.Rat).SetString(p.OriginalAmount); ok && r.Sign() > 0 {
			involved = append(involved, person{name: p.ParticipantName, original: r, origStr: p.OriginalAmount})
		}
	}

	if len(involved) == 0 {
		return nil
	}

	sumOriginal := new(big.Rat)
	for _, p := range involved {
		sumOriginal.Add(sumOriginal, p.original)
	}

	shares := make([]*big.Rat, 0, len(involved))
	allocated := new(big.Rat)
	for i, p := range involved {
		if i < len(involved)-1 {
			raw := new(big.Rat).Mul(totalPaid, p.original)
			raw.Quo(raw, sumOriginal)
			rounded := s.roundRat(raw, 2)
			shares = append(shares, rounded)
			allocated.Add(allocated, rounded)
		} else {
			last := new(big.Rat).Sub(totalPaid, allocated) // remainder
			shares = append(shares, s.roundRat(last, 2))
		}
	}

	out := make([]ComputedShare, 0, len(involved))
	for i, p := range involved {
		if p.name == "owner" {
			continue
		}
		out = append(out, ComputedShare{
			Name:   p.name,
			Amount: shares[i].FloatString(2),
		})
	}
	return out
}

func (s *TransactionService) roundRat(r *big.Rat, scale int) *big.Rat {
	if r == nil {
		return big.NewRat(0, 1)
	}
	if scale < 0 {
		scale = 0
	}
	factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
	num := new(big.Int).Mul(r.Num(), factor)
	den := new(big.Int).Set(r.Denom())
	q, rem := new(big.Int).QuoRem(num, den, new(big.Int))
	if rem.Sign() >= 0 {
		twoRem := new(big.Int).Mul(rem, big.NewInt(2))
		if twoRem.Cmp(den) >= 0 {
			q.Add(q, big.NewInt(1))
		}
	}
	return new(big.Rat).SetFrac(q, factor)
}
