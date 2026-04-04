package budget

import "testing"

func TestBudgetModuleNewWiringWithRepoFallback(t *testing.T) {
	repo := &fakeBudgetRepo{}
	categoryRepo := &fakeBudgetCategoryRepo{}
	mod := NewModule(ModuleDeps{Repo: repo, CategoryRepo: categoryRepo})

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

func TestBudgetModuleNewWiringWithInjectedService(t *testing.T) {
	svc := &fakeBudgetService{}
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
