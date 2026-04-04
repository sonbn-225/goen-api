package tag

import (
	"context"
	"errors"
	"testing"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
)

type fakeTagRepo struct {
	items     []Tag
	createErr error
	getErr    error
	listErr   error
}

func (r *fakeTagRepo) Create(_ context.Context, _ string, input Tag) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.items = append(r.items, input)
	return nil
}

func (r *fakeTagRepo) GetByID(_ context.Context, userID, tagID string) (*Tag, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	for _, item := range r.items {
		if item.UserID == userID && item.ID == tagID {
			cloned := item
			return &cloned, nil
		}
	}
	return nil, nil
}

func (r *fakeTagRepo) ListByUser(_ context.Context, userID string) ([]Tag, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	out := make([]Tag, 0)
	for _, item := range r.items {
		if item.UserID == userID {
			out = append(out, item)
		}
	}
	return out, nil
}

func TestTagServiceCreateSuccess(t *testing.T) {
	repo := &fakeTagRepo{}
	svc := NewService(repo)

	name := "Food"
	created, err := svc.Create(context.Background(), "u1", CreateInput{NameEN: &name})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created == nil || created.ID == "" {
		t.Fatalf("expected created tag with id, got %#v", created)
	}
}

func TestTagServiceCreateValidation(t *testing.T) {
	repo := &fakeTagRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "u1", CreateInput{})
	assertTagErrKind(t, err, apperrors.KindValidation)
}

func TestTagServiceGetNotFound(t *testing.T) {
	repo := &fakeTagRepo{}
	svc := NewService(repo)

	_, err := svc.Get(context.Background(), "u1", "missing")
	assertTagErrKind(t, err, apperrors.KindNotFound)
}

func TestTagServiceListInternalError(t *testing.T) {
	repo := &fakeTagRepo{listErr: errors.New("db error")}
	svc := NewService(repo)

	_, err := svc.List(context.Background(), "u1")
	assertTagErrKind(t, err, apperrors.KindInternal)
}

func TestTagServiceGetOrCreateByName(t *testing.T) {
	repo := &fakeTagRepo{}
	svc := NewService(repo)

	id, err := svc.GetOrCreateByName(context.Background(), "u1", "Coffee", "en")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id == "" {
		t.Fatal("expected tag id not empty")
	}
}

func assertTagErrKind(t *testing.T, err error, expected apperrors.Kind) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error kind %s, got nil", expected)
	}
	if got := apperrors.KindOf(err); got != expected {
		t.Fatalf("expected error kind %s, got %s (err=%v)", expected, got, err)
	}
}
