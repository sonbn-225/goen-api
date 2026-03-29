package public

import (
	"context"
	"errors"
	"strings"

	"github.com/sonbn-225/goen-api/internal/apperrors"
	"github.com/sonbn-225/goen-api/internal/domain"
)

type Service struct {
	userRepo         domain.UserRepository
	accountRepo      domain.AccountRepository
	groupExpenseRepo domain.GroupExpenseRepository
}

func NewService(userRepo domain.UserRepository, accountRepo domain.AccountRepository, groupExpenseRepo domain.GroupExpenseRepository) *Service {
	return &Service{
		userRepo:         userRepo,
		accountRepo:      accountRepo,
		groupExpenseRepo: groupExpenseRepo,
	}
}

type PublicProfile struct {
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
	Username    string  `json:"username"`
	ID          string  `json:"id"`
}

type PaymentInfo struct {
	AccountNumber string `json:"account_number"`
	BankName      string `json:"bank_name"`
}

func (s *Service) GetPublicProfile(ctx context.Context, userRef string) (*PublicProfile, error) {
	u, err := s.resolvePublicUser(ctx, userRef)
	if err != nil {
		return nil, err
	}

	displayName := ""
	if u.DisplayName != nil {
		displayName = *u.DisplayName
	}
	if displayName == "" {
		displayName = u.Username
	}

	return &PublicProfile{
		DisplayName: displayName,
		AvatarURL:   u.AvatarURL,
		Username:    u.Username,
		ID:          u.ID,
	}, nil
}

func (s *Service) GetPaymentInfo(ctx context.Context, userRef string) (*PaymentInfo, error) {
	u, err := s.resolvePublicUser(ctx, userRef)
	if err != nil {
		return nil, err
	}

	// Extract public_payment settings
	uSettings, _ := u.Settings.(map[string]any)
	settings, _ := uSettings["public_payment"].(map[string]any)
	if settings == nil {
		return nil, apperrors.ErrAccountNotFound
	}
	accID, ok := settings["default_account_id"].(string)
	if !ok || accID == "" {
		return nil, apperrors.ErrAccountNotFound
	}

	acc, err := s.accountRepo.GetAccountForUser(ctx, u.ID, accID)
	if err != nil {
		return nil, err
	}

	accNum := ""
	if acc.AccountNumber != nil {
		accNum = *acc.AccountNumber
	}

	return &PaymentInfo{
		AccountNumber: accNum,
		BankName:      acc.Name,
	}, nil
}

func (s *Service) ListDebtsByName(ctx context.Context, userRef string, participantName string) ([]domain.GroupExpenseParticipant, error) {
	u, err := s.resolvePublicUser(ctx, userRef)
	if err != nil {
		return nil, err
	}

	return s.groupExpenseRepo.ListUnsettledParticipantsByName(ctx, u.ID, participantName)
}

func (s *Service) ListParticipants(ctx context.Context, userRef string) ([]string, error) {
	u, err := s.resolvePublicUser(ctx, userRef)
	if err != nil {
		return nil, err
	}

	return s.groupExpenseRepo.ListUniqueParticipantNames(ctx, u.ID, 100)
}

func (s *Service) resolvePublicUser(ctx context.Context, userRef string) (*domain.User, error) {
	ref := strings.TrimSpace(userRef)
	if ref == "" {
		return nil, apperrors.Validation("user reference is required", map[string]any{"field": "userId"})
	}

	byUsername, err := s.userRepo.FindUserByUsername(ctx, strings.ToLower(ref))
	if err == nil {
		u := byUsername.User
		return &u, nil
	}
	if !errors.Is(err, apperrors.ErrUserNotFound) {
		return nil, err
	}

	return s.userRepo.FindUserByID(ctx, ref)
}
