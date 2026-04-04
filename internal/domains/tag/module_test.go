package tag

import "testing"

func TestTagModuleNewWiringWithRepoFallback(t *testing.T) {
	repo := &fakeTagRepo{}
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

func TestTagModuleNewWiringWithInjectedService(t *testing.T) {
	svc := &fakeTagService{}
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
