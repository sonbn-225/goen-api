package rotatingsavings

import (
	"context"
	"testing"
)

type fakeRotatingService struct{}

func (s *fakeRotatingService) ListGroups(_ context.Context, _ string) ([]GroupSummary, error) {
	return []GroupSummary{}, nil
}

func (s *fakeRotatingService) CreateGroup(_ context.Context, _ string, _ CreateGroupInput) (*RotatingSavingsGroup, error) {
	return &RotatingSavingsGroup{ID: "g1"}, nil
}

func (s *fakeRotatingService) GetGroupDetail(_ context.Context, _, _ string) (*GroupDetailResponse, error) {
	return &GroupDetailResponse{}, nil
}

func (s *fakeRotatingService) UpdateGroup(_ context.Context, _, _ string, _ UpdateGroupInput) (*RotatingSavingsGroup, error) {
	return &RotatingSavingsGroup{ID: "g1"}, nil
}

func (s *fakeRotatingService) DeleteGroup(_ context.Context, _, _ string) error {
	return nil
}

func (s *fakeRotatingService) ListContributions(_ context.Context, _, _ string) ([]RotatingSavingsContribution, error) {
	return []RotatingSavingsContribution{}, nil
}

func (s *fakeRotatingService) CreateContribution(_ context.Context, _, _ string, _ CreateContributionInput) (*RotatingSavingsContribution, error) {
	return &RotatingSavingsContribution{ID: "c1"}, nil
}

func (s *fakeRotatingService) DeleteContribution(_ context.Context, _, _, _ string) error {
	return nil
}

func TestRotatingSavingsModuleWiringWithRepoFallback(t *testing.T) {
	mod := NewModule(ModuleDeps{Repo: &fakeRotatingRepo{}, TxService: &fakeRotatingTxService{}})
	if mod == nil {
		t.Fatal("expected module")
	}
	if mod.Service == nil {
		t.Fatal("expected service")
	}
	if mod.Handler == nil {
		t.Fatal("expected handler")
	}
}

func TestRotatingSavingsModuleWiringWithInjectedService(t *testing.T) {
	mod := NewModule(ModuleDeps{Service: &fakeRotatingService{}})
	if mod == nil {
		t.Fatal("expected module")
	}
	if mod.Service == nil {
		t.Fatal("expected service")
	}
	if mod.Handler == nil {
		t.Fatal("expected handler")
	}
}
