package money

import (
	"encoding/json"
	"testing"
)

func TestAmountNewFromString(t *testing.T) {
	a, err := NewFromString(" 123.45 ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if a.Decimal.String() != "123.45" {
		t.Fatalf("expected 123.45, got %s", a.Decimal.String())
	}
}

func TestAmountUnmarshalJSON(t *testing.T) {
	t.Run("string input", func(t *testing.T) {
		var a Amount
		if err := json.Unmarshal([]byte(`"10.50"`), &a); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if a.Decimal.String() != "10.5" {
			t.Fatalf("expected 10.5, got %s", a.Decimal.String())
		}
	})

	t.Run("number input", func(t *testing.T) {
		var a Amount
		if err := json.Unmarshal([]byte(`10.75`), &a); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if a.Decimal.String() != "10.75" {
			t.Fatalf("expected 10.75, got %s", a.Decimal.String())
		}
	})

	t.Run("null input", func(t *testing.T) {
		var a Amount
		if err := json.Unmarshal([]byte(`null`), &a); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !a.Decimal.IsZero() {
			t.Fatalf("expected zero amount, got %s", a.Decimal.String())
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		var a Amount
		if err := json.Unmarshal([]byte(`"abc"`), &a); err == nil {
			t.Fatal("expected error for invalid decimal string")
		}
	})
}

func TestAmountMarshalJSON(t *testing.T) {
	a := MustFromString("99.90")
	b, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(b) != `"99.9"` {
		t.Fatalf("expected \"99.9\", got %s", string(b))
	}
}
