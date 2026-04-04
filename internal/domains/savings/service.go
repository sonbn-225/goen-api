package savings

import (
	"context"
	"errors"
	"math/big"
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

func (s *service) Create(ctx context.Context, userID string, input CreateInput) (*SavingsInstrument, error) {
	logger := logx.LoggerFromContext(ctx).With("layer", "service", "domain", "savings", "operation", "create")
	logger.Info("savings_create_started", "user_id", userID)

	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	if s.repo == nil {
		return nil, apperrors.New(apperrors.KindInternal, "savings repository not configured")
	}

	savingsAccountID := normalizeOptionalString(input.SavingsAccountID)
	parentAccountID := normalizeOptionalString(input.ParentAccountID)
	name := normalizeOptionalString(input.Name)
	if savingsAccountID == nil && parentAccountID == nil {
		return nil, apperrors.New(apperrors.KindValidation, "either savings_account_id or parent_account_id is required")
	}

	principal := strings.TrimSpace(input.Principal)
	principalRat, err := parsePositiveDecimal(principal)
	if err != nil {
		return nil, apperrors.New(apperrors.KindValidation, "principal must be a decimal string greater than zero")
	}

	interestRate, err := normalizeOptionalDecimal(input.InterestRate)
	if err != nil {
		return nil, apperrors.New(apperrors.KindValidation, "interest_rate must be a decimal string greater than or equal to zero")
	}

	accrued := "0"
	if input.AccruedInterest != nil {
		v := strings.TrimSpace(*input.AccruedInterest)
		if v != "" {
			if _, err := parseNonNegativeDecimal(v); err != nil {
				return nil, apperrors.New(apperrors.KindValidation, "accrued_interest must be a decimal string greater than or equal to zero")
			}
			accrued = v
		}
	}

	var termMonths *int
	if input.TermMonths != nil {
		if *input.TermMonths < 0 {
			return nil, apperrors.New(apperrors.KindValidation, "term_months must be greater than or equal to zero")
		}
		if *input.TermMonths > 0 {
			v := *input.TermMonths
			termMonths = &v
		}
	}

	startDate, err := normalizeOptionalDate(input.StartDate)
	if err != nil {
		return nil, err
	}
	maturityDate, err := normalizeOptionalDate(input.MaturityDate)
	if err != nil {
		return nil, err
	}

	status := "active"
	if input.Status != nil {
		v := strings.TrimSpace(*input.Status)
		if v != "" {
			if !isValidStatus(v) {
				return nil, apperrors.New(apperrors.KindValidation, "status is invalid")
			}
			status = v
		}
	}

	autoRenew := false
	if input.AutoRenew != nil {
		autoRenew = *input.AutoRenew
	}

	var account *AccountRef
	autoCreatedAccount := false

	if savingsAccountID == nil {
		parent, err := s.repo.GetAccountForUser(ctx, userID, *parentAccountID)
		if err != nil {
			return nil, passThroughOrWrapInternal("failed to read parent account", err)
		}
		if parent == nil {
			return nil, apperrors.New(apperrors.KindNotFound, "parent account not found")
		}
		if parent.Type != "bank" && parent.Type != "wallet" {
			return nil, apperrors.New(apperrors.KindValidation, "parent account must be bank or wallet")
		}

		accountName := "Savings"
		if name != nil {
			accountName = *name
		}
		created, err := s.repo.CreateLinkedSavingsAccount(ctx, userID, parent.ID, accountName, parent.Currency)
		if err != nil {
			return nil, passThroughOrWrapInternal("failed to create savings account", err)
		}
		account = created
		autoCreatedAccount = true
	} else {
		item, err := s.repo.GetAccountForUser(ctx, userID, *savingsAccountID)
		if err != nil {
			return nil, passThroughOrWrapInternal("failed to read savings account", err)
		}
		if item == nil {
			return nil, apperrors.New(apperrors.KindNotFound, "savings account not found")
		}
		account = item
	}

	if account.Type != "savings" {
		return nil, apperrors.New(apperrors.KindValidation, "savings_account_id must reference a savings account")
	}
	if account.ParentAccountID == nil || strings.TrimSpace(*account.ParentAccountID) == "" {
		return nil, apperrors.New(apperrors.KindValidation, "savings account must have parent_account_id")
	}

	now := time.Now().UTC()
	item := SavingsInstrument{
		ID:               uuid.NewString(),
		SavingsAccountID: account.ID,
		ParentAccountID:  *account.ParentAccountID,
		Principal:        principal,
		InterestRate:     interestRate,
		TermMonths:       termMonths,
		StartDate:        startDate,
		MaturityDate:     maturityDate,
		AutoRenew:        autoRenew,
		AccruedInterest:  accrued,
		Status:           status,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if status == "closed" {
		item.ClosedAt = &now
	}

	if err := s.repo.CreateSavingsInstrument(ctx, userID, item); err != nil {
		return nil, passThroughOrWrapInternal("failed to create savings instrument", err)
	}

	if autoCreatedAccount && s.txService != nil && principalRat.Sign() > 0 {
		amount, err := money.NewFromString(item.Principal)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.KindInternal, "invalid principal amount", err)
		}

		fromID := item.ParentAccountID
		toID := item.SavingsAccountID
		note := "Initial savings deposit"
		if account.Name != "" {
			note = "Initial savings deposit: " + account.Name
		}
		_, err = s.txService.Create(ctx, userID, transaction.CreateInput{
			Type:          "transfer",
			FromAccountID: &fromID,
			ToAccountID:   &toID,
			Amount:        amount,
			Note:          note,
		})
		if err != nil {
			_ = s.repo.DeleteSavingsInstrument(ctx, userID, item.ID)
			_ = s.repo.DeleteAccountForUser(ctx, userID, item.SavingsAccountID)
			return nil, err
		}
	}

	created, err := s.repo.GetSavingsInstrument(ctx, userID, item.ID)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to load savings instrument", err)
	}
	if created == nil {
		return nil, apperrors.New(apperrors.KindInternal, "created savings instrument not found")
	}

	logger.Info("savings_create_succeeded", "instrument_id", created.ID)
	return created, nil
}

func (s *service) Get(ctx context.Context, userID, instrumentID string) (*SavingsInstrument, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	instrumentID = strings.TrimSpace(instrumentID)
	if instrumentID == "" {
		return nil, apperrors.New(apperrors.KindValidation, "instrumentId is required")
	}

	item, err := s.repo.GetSavingsInstrument(ctx, userID, instrumentID)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to get savings instrument", err)
	}
	if item == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "savings instrument not found")
	}
	return item, nil
}

func (s *service) List(ctx context.Context, userID string) ([]SavingsInstrument, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}

	items, err := s.repo.ListSavingsInstruments(ctx, userID)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to list savings instruments", err)
	}
	return items, nil
}

func (s *service) Patch(ctx context.Context, userID, instrumentID string, input PatchInput) (*SavingsInstrument, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	instrumentID = strings.TrimSpace(instrumentID)
	if instrumentID == "" {
		return nil, apperrors.New(apperrors.KindValidation, "instrumentId is required")
	}

	item, err := s.repo.GetSavingsInstrument(ctx, userID, instrumentID)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to get savings instrument", err)
	}
	if item == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "savings instrument not found")
	}

	changed := false
	if input.Principal != nil {
		v := strings.TrimSpace(*input.Principal)
		if _, err := parsePositiveDecimal(v); err != nil {
			return nil, apperrors.New(apperrors.KindValidation, "principal must be a decimal string greater than zero")
		}
		item.Principal = v
		changed = true
	}

	if input.InterestRate != nil {
		v := strings.TrimSpace(*input.InterestRate)
		if v == "" {
			item.InterestRate = nil
		} else {
			if _, err := parseNonNegativeDecimal(v); err != nil {
				return nil, apperrors.New(apperrors.KindValidation, "interest_rate must be a decimal string greater than or equal to zero")
			}
			item.InterestRate = &v
		}
		changed = true
	}

	if input.AccruedInterest != nil {
		v := strings.TrimSpace(*input.AccruedInterest)
		if v == "" {
			item.AccruedInterest = "0"
		} else {
			if _, err := parseNonNegativeDecimal(v); err != nil {
				return nil, apperrors.New(apperrors.KindValidation, "accrued_interest must be a decimal string greater than or equal to zero")
			}
			item.AccruedInterest = v
		}
		changed = true
	}

	if input.TermMonths != nil {
		if *input.TermMonths < 0 {
			return nil, apperrors.New(apperrors.KindValidation, "term_months must be greater than or equal to zero")
		}
		if *input.TermMonths == 0 {
			item.TermMonths = nil
		} else {
			v := *input.TermMonths
			item.TermMonths = &v
		}
		changed = true
	}

	if input.StartDate != nil {
		v := strings.TrimSpace(*input.StartDate)
		if v == "" {
			item.StartDate = nil
		} else {
			if _, err := time.Parse("2006-01-02", v); err != nil {
				return nil, apperrors.New(apperrors.KindValidation, "start_date must be YYYY-MM-DD")
			}
			item.StartDate = &v
		}
		changed = true
	}

	if input.MaturityDate != nil {
		v := strings.TrimSpace(*input.MaturityDate)
		if v == "" {
			item.MaturityDate = nil
		} else {
			if _, err := time.Parse("2006-01-02", v); err != nil {
				return nil, apperrors.New(apperrors.KindValidation, "maturity_date must be YYYY-MM-DD")
			}
			item.MaturityDate = &v
		}
		changed = true
	}

	if input.AutoRenew != nil {
		item.AutoRenew = *input.AutoRenew
		changed = true
	}

	if input.Status != nil {
		v := strings.TrimSpace(*input.Status)
		if v != "" {
			if !isValidStatus(v) {
				return nil, apperrors.New(apperrors.KindValidation, "status is invalid")
			}
			item.Status = v
			if v == "closed" {
				now := time.Now().UTC()
				item.ClosedAt = &now
			} else {
				item.ClosedAt = nil
			}
			changed = true
		}
	}

	if !changed {
		return nil, apperrors.New(apperrors.KindValidation, "no fields to update")
	}

	item.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateSavingsInstrument(ctx, userID, *item); err != nil {
		return nil, passThroughOrWrapInternal("failed to update savings instrument", err)
	}

	updated, err := s.repo.GetSavingsInstrument(ctx, userID, instrumentID)
	if err != nil {
		return nil, passThroughOrWrapInternal("failed to get savings instrument", err)
	}
	if updated == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "savings instrument not found")
	}
	return updated, nil
}

func (s *service) Delete(ctx context.Context, userID, instrumentID string) error {
	if strings.TrimSpace(userID) == "" {
		return apperrors.New(apperrors.KindUnauth, "missing user context")
	}
	instrumentID = strings.TrimSpace(instrumentID)
	if instrumentID == "" {
		return apperrors.New(apperrors.KindValidation, "instrumentId is required")
	}

	if err := s.repo.DeleteSavingsInstrument(ctx, userID, instrumentID); err != nil {
		return passThroughOrWrapInternal("failed to delete savings instrument", err)
	}
	return nil
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

func normalizeOptionalDate(v *string) (*string, error) {
	if v == nil {
		return nil, nil
	}
	clean := strings.TrimSpace(*v)
	if clean == "" {
		return nil, nil
	}
	if _, err := time.Parse("2006-01-02", clean); err != nil {
		return nil, apperrors.New(apperrors.KindValidation, "date must be YYYY-MM-DD")
	}
	return &clean, nil
}

func normalizeOptionalDecimal(v *string) (*string, error) {
	if v == nil {
		return nil, nil
	}
	clean := strings.TrimSpace(*v)
	if clean == "" {
		return nil, nil
	}
	if _, err := parseNonNegativeDecimal(clean); err != nil {
		return nil, err
	}
	return &clean, nil
}

func parsePositiveDecimal(v string) (*big.Rat, error) {
	rat, ok := new(big.Rat).SetString(strings.TrimSpace(v))
	if !ok || rat.Sign() <= 0 {
		return nil, errors.New("invalid")
	}
	return rat, nil
}

func parseNonNegativeDecimal(v string) (*big.Rat, error) {
	rat, ok := new(big.Rat).SetString(strings.TrimSpace(v))
	if !ok || rat.Sign() < 0 {
		return nil, errors.New("invalid")
	}
	return rat, nil
}

func isValidStatus(v string) bool {
	switch strings.TrimSpace(v) {
	case "active", "matured", "closed":
		return true
	default:
		return false
	}
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
