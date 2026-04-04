package logx

import "testing"

func TestMaskValueSensitiveKey(t *testing.T) {
	got := MaskValue("password", "Password123")
	if got != redacted {
		t.Fatalf("expected %s, got %v", redacted, got)
	}
}

func TestMaskValueBearerToken(t *testing.T) {
	got := MaskValue("authorization", "Bearer abc.def.ghi")
	if got != redacted {
		t.Fatalf("expected %s, got %v", redacted, got)
	}
}

func TestMaskAttrsMap(t *testing.T) {
	attrs := MaskAttrs("payload", map[string]any{"password": "x", "token": "y", "name": "z"})
	payload, ok := attrs[1].(map[string]any)
	if !ok {
		t.Fatalf("expected map payload")
	}
	if payload["password"] != redacted {
		t.Fatalf("expected password masked")
	}
	if payload["token"] != redacted {
		t.Fatalf("expected token masked")
	}
	if payload["name"] != "z" {
		t.Fatalf("expected name unchanged")
	}
}
