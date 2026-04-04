package account

import (
	"context"
	"errors"
	"testing"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
)

type fakeAccountRepo struct {
	items           []Account
	createErr       error
	listErr         error
	getErr          error
	defaultCurrency string
	ownerErr        error
	deleteErr       error
	transferErr     error
	owners          map[string]bool
	transferByAcc   map[string]bool
}

func (r *fakeAccountRepo) Create(_ context.Context, acc *Account) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.items = append(r.items, *acc)
	return nil
}

func (r *fakeAccountRepo) ListByUser(_ context.Context, userID string) ([]Account, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	out := make([]Account, 0)
	for _, item := range r.items {
		if item.UserID == userID {
			out = append(out, item)
		}
	}
	return out, nil
}

func (r *fakeAccountRepo) GetByID(_ context.Context, userID, accountID string) (*Account, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	for _, item := range r.items {
		if item.UserID == userID && item.ID == accountID {
			cloned := item
			return &cloned, nil
		}
	}
	return nil, nil
}

func (r *fakeAccountRepo) GetDefaultCurrency(_ context.Context, _ string) (string, error) {
	return r.defaultCurrency, nil
}

func (r *fakeAccountRepo) IsOwner(_ context.Context, userID, accountID string) (bool, error) {
	if r.ownerErr != nil {
		return false, r.ownerErr
	}
	if r.owners != nil {
		if v, ok := r.owners[userID+"|"+accountID]; ok {
			return v, nil
		}
	}
	for _, item := range r.items {
		if item.ID == accountID && item.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeAccountRepo) HasRelatedTransferTransactionsForAccount(_ context.Context, accountID string) (bool, error) {
	if r.transferErr != nil {
		return false, r.transferErr
	}
	if r.transferByAcc != nil {
		return r.transferByAcc[accountID], nil
	}
	return false, nil
}

func (r *fakeAccountRepo) Delete(_ context.Context, _ string, accountID string) (bool, error) {
	if r.deleteErr != nil {
		return false, r.deleteErr
	}
	for i, item := range r.items {
		if item.ID == accountID {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

func TestServiceCreateSuccess(t *testing.T) {
	repo := &fakeAccountRepo{}
	svc := NewService(repo)

	acc, err := svc.Create(context.Background(), "u1", CreateInput{
		Name: "Wallet",
		Type: "Cash",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if acc.Type != "cash" {
		t.Fatalf("expected normalized type cash, got %s", acc.Type)
	}
	if acc.Currency != "VND" {
		t.Fatalf("expected default currency VND, got %s", acc.Currency)
	}
	if len(repo.items) != 1 {
		t.Fatalf("expected 1 account in repo, got %d", len(repo.items))
	}
}

func TestServiceCreateValidationMissingName(t *testing.T) {
	repo := &fakeAccountRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "u1", CreateInput{Type: "cash"})
	assertAccountErrKind(t, err, apperrors.KindValidation)
}

func TestServiceCreateValidationMissingType(t *testing.T) {
	repo := &fakeAccountRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "u1", CreateInput{Name: "Wallet"})
	assertAccountErrKind(t, err, apperrors.KindValidation)
}

func TestServiceCreateValidationInvalidType(t *testing.T) {
	repo := &fakeAccountRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "u1", CreateInput{Name: "Wallet", Type: "invalid"})
	assertAccountErrKind(t, err, apperrors.KindValidation)
}

func TestServiceCreateUnauthorizedMissingUserID(t *testing.T) {
	repo := &fakeAccountRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "", CreateInput{Name: "Wallet", Type: "cash"})
	assertAccountErrKind(t, err, apperrors.KindUnauth)
}

func TestServiceCreateValidationCardRequiresParent(t *testing.T) {
	repo := &fakeAccountRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "u1", CreateInput{Name: "Visa", Type: "card"})
	assertAccountErrKind(t, err, apperrors.KindValidation)
}

func TestServiceCreateValidationParentTypeMismatch(t *testing.T) {
	repo := &fakeAccountRepo{items: []Account{{ID: "p1", UserID: "u1", Type: "wallet"}}}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "u1", CreateInput{Name: "Visa", Type: "card", ParentAccountID: "p1"})
	assertAccountErrKind(t, err, apperrors.KindValidation)
}

func TestServiceDeleteSuccess(t *testing.T) {
	repo := &fakeAccountRepo{items: []Account{{ID: "a1", UserID: "u1", Name: "Wallet", Type: "wallet"}}}
	svc := NewService(repo)

	err := svc.Delete(context.Background(), "u1", "a1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repo.items) != 0 {
		t.Fatalf("expected account to be deleted")
	}
}

func TestServiceDeleteValidationMissingAccountID(t *testing.T) {
	repo := &fakeAccountRepo{}
	svc := NewService(repo)

	err := svc.Delete(context.Background(), "u1", "")
	assertAccountErrKind(t, err, apperrors.KindValidation)
}

func TestServiceDeleteNotFound(t *testing.T) {
	repo := &fakeAccountRepo{}
	svc := NewService(repo)

	err := svc.Delete(context.Background(), "u1", "missing")
	assertAccountErrKind(t, err, apperrors.KindNotFound)
}

func TestServiceDeleteForbiddenNotOwner(t *testing.T) {
	repo := &fakeAccountRepo{
		items:  []Account{{ID: "a1", UserID: "u1", Name: "Wallet", Type: "wallet"}},
		owners: map[string]bool{"u1|a1": false},
	}
	svc := NewService(repo)

	err := svc.Delete(context.Background(), "u1", "a1")
	assertAccountErrKind(t, err, apperrors.KindForbidden)
}

func TestServiceDeleteValidationCashAccount(t *testing.T) {
	repo := &fakeAccountRepo{items: []Account{{ID: "a1", UserID: "u1", Name: "Cash", Type: "cash"}}}
	svc := NewService(repo)

	err := svc.Delete(context.Background(), "u1", "a1")
	assertAccountErrKind(t, err, apperrors.KindValidation)
}

func TestServiceDeleteValidationHasTransfers(t *testing.T) {
	repo := &fakeAccountRepo{
		items:         []Account{{ID: "a1", UserID: "u1", Name: "Wallet", Type: "wallet"}},
		transferByAcc: map[string]bool{"a1": true},
	}
	svc := NewService(repo)

	err := svc.Delete(context.Background(), "u1", "a1")
	assertAccountErrKind(t, err, apperrors.KindValidation)
}

func TestServiceListUnauthorizedMissingUserID(t *testing.T) {
	repo := &fakeAccountRepo{}
	svc := NewService(repo)

	_, err := svc.List(context.Background(), "")
	assertAccountErrKind(t, err, apperrors.KindUnauth)
}

func TestServiceListInternalError(t *testing.T) {
	repo := &fakeAccountRepo{listErr: errors.New("db down")}
	svc := NewService(repo)

	_, err := svc.List(context.Background(), "u1")
	assertAccountErrKind(t, err, apperrors.KindInternal)
}

func TestServiceGetSuccess(t *testing.T) {
	repo := &fakeAccountRepo{items: []Account{{ID: "a1", UserID: "u1", Name: "Wallet", Type: "cash"}}}
	svc := NewService(repo)

	acc, err := svc.Get(context.Background(), "u1", "a1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if acc.ID != "a1" {
		t.Fatalf("expected account a1, got %s", acc.ID)
	}
}

func TestServiceGetValidationMissingAccountID(t *testing.T) {
	repo := &fakeAccountRepo{}
	svc := NewService(repo)

	_, err := svc.Get(context.Background(), "u1", "")
	assertAccountErrKind(t, err, apperrors.KindValidation)
}

func TestServiceGetNotFound(t *testing.T) {
	repo := &fakeAccountRepo{}
	svc := NewService(repo)

	_, err := svc.Get(context.Background(), "u1", "missing")
	assertAccountErrKind(t, err, apperrors.KindNotFound)
}

func assertAccountErrKind(t *testing.T, err error, expected apperrors.Kind) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error kind %s, got nil", expected)
	}
	if got := apperrors.KindOf(err); got != expected {
		t.Fatalf("expected error kind %s, got %s (err=%v)", expected, got, err)
	}
}
