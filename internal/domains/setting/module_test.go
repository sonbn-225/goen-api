package setting

import (
	"context"
	"testing"

	"github.com/sonbn-225/goen-api-v2/internal/domains/auth"
)

type fakeSettingService struct {
	user      auth.User
	updateErr error
}

func (s *fakeSettingService) UpdateMySettings(_ context.Context, _ string, patch map[string]any) (*auth.User, error) {
	if s.updateErr != nil {
		return nil, s.updateErr
	}
	if s.user.Settings == nil {
		s.user.Settings = map[string]any{}
	}
	for k, v := range patch {
		s.user.Settings[k] = v
	}
	cloned := s.user
	return &cloned, nil
}

func TestModuleNewWiring(t *testing.T) {
	svc := &fakeSettingService{}
	mod := NewModule(ModuleDeps{Service: svc})

	if mod == nil {
		t.Fatal("expected module not nil")
	}
	if mod.Handler == nil {
		t.Fatal("expected handler not nil")
	}
}
