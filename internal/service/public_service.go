package service
 
import (
	"context"
	"strings"
 
	"github.com/google/uuid"
	"github.com/sonbn-225/goen-api/internal/domain/entity"
	"github.com/sonbn-225/goen-api/internal/domain/interfaces"
	"github.com/sonbn-225/goen-api/internal/pkg/apperr"
)
 
type PublicService struct {
	userRepo         interfaces.UserRepository
	accountRepo      interfaces.AccountRepository
	debtRepo         interfaces.DebtRepository
}
 
func NewPublicService(
	userRepo interfaces.UserRepository,
	accountRepo interfaces.AccountRepository,
	debtRepo interfaces.DebtRepository,
) *PublicService {
	return &PublicService{
		userRepo:         userRepo,
		accountRepo:      accountRepo,
		debtRepo:         debtRepo,
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
		return nil, apperr.NotFound("public payment settings not found")
	}
	accIDStr, ok := settings["default_account_id"].(string)
	if !ok || accIDStr == "" {
		return nil, apperr.BadRequest("missing_account", "default payment account not configured")
	}
 
	accID, err := uuid.Parse(accIDStr)
	if err != nil {
		return nil, apperr.BadRequest("invalid_account_id", "invalid default account id format")
	}
 
	acc, err := s.accountRepo.GetAccountForUserTx(ctx, nil, u.ID, accID)
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
 
func (s *PublicService) GetParticipants(ctx context.Context, userRef string) ([]string, error) {
	u, err := s.resolvePublicUser(ctx, userRef)
	if err != nil {
		return nil, err
	}
 
	return s.debtRepo.ListPublicParticipantsTx(ctx, nil, u.ID)
}
 
func (s *PublicService) GetDebts(ctx context.Context, userRef string, participantName string) ([]entity.PublicDebt, error) {
	u, err := s.resolvePublicUser(ctx, userRef)
	if err != nil {
		return nil, err
	}
 
	return s.debtRepo.ListPublicDebtsByParticipantTx(ctx, nil, u.ID, participantName)
}
 
func (s *PublicService) resolvePublicUser(ctx context.Context, userRef string) (*entity.User, error) {
	ref := strings.TrimSpace(userRef)
	if ref == "" {
		return nil, apperr.BadRequest("missing_reference", "user reference is required")
	}
 
	// Resolve user by username, email or phone
	byRef, err := ResolveUserByLoginTx(ctx, nil, s.userRepo, ref)
	var u *entity.User
	if err == nil && byRef != nil {
		u = &byRef.User
	} else {
		// Fallback to lookup by UUID string if direct resolution fails
		uid, err := uuid.Parse(ref)
		if err != nil {
			return nil, apperr.NotFound("user not found")
		}
		u, err = s.userRepo.FindUserByIDTx(ctx, nil, uid)
		if err != nil {
			return nil, err
		}
	}

	// Check if public sharing is enabled
	settings, _ := u.Settings.(map[string]any)
	enabled, _ := settings["public_sharing_enabled"].(bool)
	if !enabled {
		return nil, apperr.Forbidden("sharing_disabled", "user has disabled public sharing")
	}

	return u, nil
}
