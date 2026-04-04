package auth_test

import (
	"testing"

	"github.com/sonbn-225/goen-api-v2/internal/domains/auth"
)

func TestModuleNewWiring(t *testing.T) {
	repo := newFakeUserRepo()
	mod := auth.NewModule(auth.ModuleDeps{
		UserRepo:         repo,
		Hasher:           fakeHasher{},
		Issuer:           fakeIssuer{},
		AccessTTLMinutes: 60,
	})
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
