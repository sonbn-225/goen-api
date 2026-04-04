package category

import "testing"

func TestModuleNewWiringWithRepoFallback(t *testing.T) {
	repo := &fakeCategoryRepo{}
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

func TestModuleNewWiringWithInjectedService(t *testing.T) {
	svc := &fakeCategoryService{}
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
