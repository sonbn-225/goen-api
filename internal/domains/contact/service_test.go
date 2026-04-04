package contact

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
)

type fakeContactRepo struct {
	items     []Contact
	createErr error
	getErr    error
	listErr   error
	updateErr error
	deleteErr error
	user      *LinkedUser
}

func (r *fakeContactRepo) Create(_ context.Context, _ string, input Contact) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.items = append(r.items, input)
	return nil
}

func (r *fakeContactRepo) GetByID(_ context.Context, userID, contactID string) (*Contact, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	for _, item := range r.items {
		if item.UserID == userID && item.ID == contactID {
			cloned := item
			return &cloned, nil
		}
	}
	return nil, nil
}

func (r *fakeContactRepo) ListByUser(_ context.Context, userID string) ([]Contact, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	out := make([]Contact, 0)
	for _, item := range r.items {
		if item.UserID == userID {
			out = append(out, item)
		}
	}
	return out, nil
}

func (r *fakeContactRepo) Update(_ context.Context, userID string, input Contact) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	for i := range r.items {
		if r.items[i].UserID == userID && r.items[i].ID == input.ID {
			r.items[i] = input
			return nil
		}
	}
	return nil
}

func (r *fakeContactRepo) Delete(_ context.Context, _ string, _ string) error {
	return r.deleteErr
}

func (r *fakeContactRepo) FindLinkedUserByEmail(_ context.Context, _ string) (*LinkedUser, error) {
	return r.user, nil
}

func (r *fakeContactRepo) FindLinkedUserByPhone(_ context.Context, _ string) (*LinkedUser, error) {
	return r.user, nil
}

func TestContactServiceCreateSuccess(t *testing.T) {
	repo := &fakeContactRepo{}
	svc := NewService(repo)

	created, err := svc.Create(context.Background(), "u1", CreateInput{Name: "Alice"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created == nil || created.ID == "" {
		t.Fatalf("expected created contact, got %#v", created)
	}
}

func TestContactServiceCreateValidation(t *testing.T) {
	repo := &fakeContactRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "u1", CreateInput{Name: "  "})
	assertContactErrKind(t, err, apperrors.KindValidation)
}

func TestContactServiceGetNotFound(t *testing.T) {
	repo := &fakeContactRepo{}
	svc := NewService(repo)

	_, err := svc.Get(context.Background(), "u1", "missing")
	assertContactErrKind(t, err, apperrors.KindNotFound)
}

func TestContactServiceUpdateSuccess(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeContactRepo{items: []Contact{{ID: "c1", UserID: "u1", Name: "Old", CreatedAt: now, UpdatedAt: now}}}
	svc := NewService(repo)

	updated, err := svc.Update(context.Background(), "u1", "c1", UpdateInput{Name: "New"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Name != "New" {
		t.Fatalf("expected name New, got %s", updated.Name)
	}
}

func TestContactServiceDeleteInternalError(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeContactRepo{items: []Contact{{ID: "c1", UserID: "u1", Name: "Old", CreatedAt: now, UpdatedAt: now}}, deleteErr: errors.New("db error")}
	svc := NewService(repo)

	err := svc.Delete(context.Background(), "u1", "c1")
	assertContactErrKind(t, err, apperrors.KindInternal)
}

func assertContactErrKind(t *testing.T, err error, expected apperrors.Kind) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error kind %s, got nil", expected)
	}
	if got := apperrors.KindOf(err); got != expected {
		t.Fatalf("expected error kind %s, got %s (err=%v)", expected, got, err)
	}
}
