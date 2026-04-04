package contact

import "testing"

func TestContactModuleNewWiringWithRepoFallback(t *testing.T) {
	repo := &fakeContactRepo{}
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

func TestContactModuleNewWiringWithInjectedService(t *testing.T) {
	svc := &fakeContactService{}
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
