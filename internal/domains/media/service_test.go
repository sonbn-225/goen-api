package media

import (
	"context"
	"errors"
	"testing"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
)

func TestServiceGetObjectSuccess(t *testing.T) {
	svc := NewService(fakeStorage{content: []byte("x"), contentType: "image/png"})

	obj, info, err := svc.GetObject(context.Background(), "goen", "avatars/u1/a.png")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if obj == nil {
		t.Fatal("expected object not nil")
	}
	_ = obj.Close()
	if info.ContentType != "image/png" {
		t.Fatalf("expected image/png, got %s", info.ContentType)
	}
}

func TestServiceGetObjectStorageNotConfigured(t *testing.T) {
	svc := NewService(nil)

	_, _, err := svc.GetObject(context.Background(), "goen", "avatars/u1/a.png")
	if apperrors.KindOf(err) != apperrors.KindInternal {
		t.Fatalf("expected internal kind, got %v", apperrors.KindOf(err))
	}
}

func TestServiceGetObjectInvalidPath(t *testing.T) {
	svc := NewService(fakeStorage{})

	_, _, err := svc.GetObject(context.Background(), "", "")
	if apperrors.KindOf(err) != apperrors.KindValidation {
		t.Fatalf("expected validation kind, got %v", apperrors.KindOf(err))
	}
}

func TestServiceGetObjectNotFound(t *testing.T) {
	svc := NewService(fakeStorage{err: errors.New("not found")})

	_, _, err := svc.GetObject(context.Background(), "goen", "avatars/u1/missing.png")
	if apperrors.KindOf(err) != apperrors.KindNotFound {
		t.Fatalf("expected not_found kind, got %v", apperrors.KindOf(err))
	}
}
