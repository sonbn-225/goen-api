package service

import (
	"context"
	"errors"
	"strings"

	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
)

type PublicService struct {
	userRepo         interfaces.UserRepository
	accountRepo      interfaces.AccountRepository
	groupExpenseRepo interfaces.GroupExpenseRepository
}

func NewPublicService(
	userRepo interfaces.UserRepository,
	accountRepo interfaces.AccountRepository,
	groupExpenseRepo interfaces.GroupExpenseRepository,
) *PublicService {
	return &PublicService{
		userRepo:         userRepo,
		accountRepo:      accountRepo,
		groupExpenseRepo: groupExpenseRepo,
	}
}

func (s *PublicService) GetPublicProfile(ctx context.Context, userRef string) (*entity.PublicProfile, error) {
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

	return &entity.PublicProfile{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: displayName,
		AvatarURL:   u.AvatarURL,
	}, nil
}

func (s *PublicService) GetPaymentInfo(ctx context.Context, userRef string) (*entity.PaymentInfo, error) {
	u, err := s.resolvePublicUser(ctx, userRef)
	if err != nil {
		return nil, err
	}

	// Extract public_payment settings from user metadata
	uSettings, _ := u.Settings.(map[string]any)
	settings, _ := uSettings["public_payment"].(map[string]any)
	if settings == nil {
		return nil, errors.New("public payment settings not found")
	}
	accID, ok := settings["default_account_id"].(string)
	if !ok || accID == "" {
		return nil, errors.New("default payment account not configured")
	}

	acc, err := s.accountRepo.GetAccountForUser(ctx, u.ID, accID)
	if err != nil {
		return nil, err
	}

	accNum := ""
	if acc.AccountNumber != nil {
		accNum = *acc.AccountNumber
	}

	return &entity.PaymentInfo{
		AccountNumber: accNum,
		BankName:      acc.Name,
	}, nil
}

func (s *PublicService) resolvePublicUser(ctx context.Context, userRef string) (*entity.User, error) {
	ref := strings.TrimSpace(userRef)
	if ref == "" {
		return nil, errors.New("user reference is required")
	}

	// Try lookup by username first
	byUsername, err := s.userRepo.FindUserByUsername(ctx, strings.ToLower(ref))
	if err == nil {
		return &byUsername.User, nil
	}

	// Then try lookup by ID
	return s.userRepo.FindUserByID(ctx, ref)
}
