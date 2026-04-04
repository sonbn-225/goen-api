package debt

import "testing"

func TestDebtModuleNewWiringWithRepoFallback(t *testing.T) {
	repo := &fakeDebtRepo{}
	txSvc := &fakeDebtTxService{}
	contactSvc := &fakeDebtContactService{}
	mod := NewModule(ModuleDeps{Repo: repo, TxService: txSvc, ContactService: contactSvc})

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

func TestDebtModuleNewWiringWithInjectedService(t *testing.T) {
	svc := &fakeDebtService{}
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
