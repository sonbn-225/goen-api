package investment

import "testing"

func TestInvestmentModuleNewWiringWithRepoFallback(t *testing.T) {
	repo := &fakeInvestmentRepo{}
	mod := NewModule(ModuleDeps{Repo: repo})

	if mod == nil {
		t.Fatal("expected module not nil")
	}
	if mod.Service == nil {
		t.Fatal("expected service not nil")
	}
	if mod.Handler == nil {
		t.Fatal("expected handler not nil")
	}
}

func TestInvestmentModuleNewWiringWithInjectedService(t *testing.T) {
	svc := &fakeInvestmentService{}
	mod := NewModule(ModuleDeps{Service: svc})

	if mod == nil {
		t.Fatal("expected module not nil")
	}
	if mod.Service == nil {
		t.Fatal("expected service not nil")
	}
	if mod.Handler == nil {
		t.Fatal("expected handler not nil")
	}
}
