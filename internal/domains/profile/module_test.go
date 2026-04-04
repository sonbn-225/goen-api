package profile

import (
	"context"
	"testing"

	"github.com/sonbn-225/goen-api-v2/internal/domains/auth"
)

type fakeProfileService struct {
	user              auth.User
	getMeErr          error
	updateProfileErr  error
	uploadAvatarErr   error
	changePasswordErr error
}

func (s *fakeProfileService) GetMe(_ context.Context, _ string) (*auth.User, error) {
	if s.getMeErr != nil {
		return nil, s.getMeErr
	}
	cloned := s.user
	return &cloned, nil
}

func (s *fakeProfileService) UpdateMyProfile(_ context.Context, _ string, input auth.UpdateProfileInput) (*auth.User, error) {
	if s.updateProfileErr != nil {
		return nil, s.updateProfileErr
	}
	if input.DisplayName != nil {
		s.user.DisplayName = input.DisplayName
	}
	if input.Email != nil {
		s.user.Email = input.Email
	}
	if input.Phone != nil {
		s.user.Phone = input.Phone
	}
	if input.Username != nil {
		s.user.Username = *input.Username
	}
	cloned := s.user
	return &cloned, nil
}

func (s *fakeProfileService) UploadAvatar(_ context.Context, _ string, _, _ string, _ []byte) (*auth.User, error) {
	if s.uploadAvatarErr != nil {
		return nil, s.uploadAvatarErr
	}
	avatarURL := "/mock/avatar.jpg"
	s.user.AvatarURL = &avatarURL
	cloned := s.user
	return &cloned, nil
}

func (s *fakeProfileService) ChangePassword(_ context.Context, _ string, _, _ string) error {
	return s.changePasswordErr
}

func TestModuleNewWiring(t *testing.T) {
	svc := &fakeProfileService{}
	mod := NewModule(ModuleDeps{Service: svc})

	if mod == nil {
		t.Fatal("expected module not nil")
	}
	if mod.Handler == nil {
		t.Fatal("expected handler not nil")
	}
}
