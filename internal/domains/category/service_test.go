package category

import (
	"context"
	"errors"
	"testing"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
)

type fakeCategoryRepo struct {
	items   []Category
	getErr  error
	listErr error
}

func (r *fakeCategoryRepo) GetByID(_ context.Context, _ string, categoryID string) (*Category, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	for _, item := range r.items {
		if item.ID == categoryID {
			cloned := item
			return &cloned, nil
		}
	}
	return nil, nil
}

func (r *fakeCategoryRepo) ListByUser(_ context.Context, _ string) ([]Category, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.items, nil
}

func TestCategoryServiceGetSuccess(t *testing.T) {
	repo := &fakeCategoryRepo{items: []Category{{ID: "cat_food"}}}
	svc := NewService(repo)

	item, err := svc.Get(context.Background(), "u1", "cat_food")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if item == nil || item.ID != "cat_food" {
		t.Fatalf("expected category cat_food, got %#v", item)
	}
}

func TestCategoryServiceGetValidationMissingID(t *testing.T) {
	repo := &fakeCategoryRepo{}
	svc := NewService(repo)

	_, err := svc.Get(context.Background(), "u1", "")
	assertCategoryErrKind(t, err, apperrors.KindValidation)
}

func TestCategoryServiceGetNotFound(t *testing.T) {
	repo := &fakeCategoryRepo{}
	svc := NewService(repo)

	_, err := svc.Get(context.Background(), "u1", "missing")
	assertCategoryErrKind(t, err, apperrors.KindNotFound)
}

func TestCategoryServiceGetInternalError(t *testing.T) {
	repo := &fakeCategoryRepo{getErr: errors.New("db error")}
	svc := NewService(repo)

	_, err := svc.Get(context.Background(), "u1", "cat_food")
	assertCategoryErrKind(t, err, apperrors.KindInternal)
}

func TestCategoryServiceListFilterIncome(t *testing.T) {
	income := "income"
	both := "both"
	expense := "expense"
	repo := &fakeCategoryRepo{items: []Category{
		{ID: "cat_income", Type: &income},
		{ID: "cat_both", Type: &both},
		{ID: "cat_expense", Type: &expense},
		{ID: "cat_nil"},
	}}
	svc := NewService(repo)

	items, err := svc.List(context.Background(), "u1", "income")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 categories after income filter, got %d", len(items))
	}
}

func TestCategoryServiceListInvalidType(t *testing.T) {
	repo := &fakeCategoryRepo{}
	svc := NewService(repo)

	_, err := svc.List(context.Background(), "u1", "invalid")
	assertCategoryErrKind(t, err, apperrors.KindValidation)
}

func TestCategoryServiceListInternalError(t *testing.T) {
	repo := &fakeCategoryRepo{listErr: errors.New("db error")}
	svc := NewService(repo)

	_, err := svc.List(context.Background(), "u1", "")
	assertCategoryErrKind(t, err, apperrors.KindInternal)
}

func assertCategoryErrKind(t *testing.T, err error, expected apperrors.Kind) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error kind %s, got nil", expected)
	}
	if got := apperrors.KindOf(err); got != expected {
		t.Fatalf("expected error kind %s, got %s (err=%v)", expected, got, err)
	}
}
