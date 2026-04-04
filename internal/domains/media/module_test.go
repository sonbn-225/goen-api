package media

import "testing"

func TestModuleNewWiring(t *testing.T) {
	mod := NewModule(ModuleDeps{Storage: fakeStorage{}})
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
