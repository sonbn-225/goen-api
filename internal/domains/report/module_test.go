package report

import "testing"

func TestReportModuleNewWiringWithRepoFallback(t *testing.T) {
	repo := &fakeReportRepo{}
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

func TestReportModuleNewWiringWithInjectedService(t *testing.T) {
	svc := &fakeReportService{}
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
