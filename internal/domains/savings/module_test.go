package savings

import "testing"

func TestSavingsModuleNewWiringWithRepoFallback(t *testing.T) {
	repo := &fakeSavingsRepo{}
	mod := NewModule(ModuleDeps{Repo: repo, TxService: &fakeSavingsTxService{}})

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

func TestSavingsModuleNewWiringWithInjectedService(t *testing.T) {
	svc := &fakeSavingsService{}
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
