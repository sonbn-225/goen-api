package budget

import (
	"context"
	"errors"
	"testing"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/domains/category"
)

type fakeBudgetRepo struct {
	items      []Budget
	createErr  error
	getErr     error
	listErr    error
	computeErr error
	spent      string
}

func (r *fakeBudgetRepo) Create(_ context.Context, _ string, input Budget) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.items = append(r.items, input)
	return nil
}

func (r *fakeBudgetRepo) GetByID(_ context.Context, userID, budgetID string) (*Budget, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	for _, item := range r.items {
		if item.UserID == userID && item.ID == budgetID {
			cloned := item
			return &cloned, nil
		}
	}
	return nil, nil
}

func (r *fakeBudgetRepo) ListByUser(_ context.Context, userID string) ([]Budget, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	out := make([]Budget, 0)
	for _, item := range r.items {
		if item.UserID == userID {
			out = append(out, item)
		}
	}
	return out, nil
}

func (r *fakeBudgetRepo) ComputeSpent(_ context.Context, _ string, _ string, _, _ string) (string, error) {
	if r.computeErr != nil {
		return "", r.computeErr
	}
	if r.spent == "" {
		return "0", nil
	}
	return r.spent, nil
}

type fakeBudgetCategoryRepo struct {
	item *category.Category
	err  error
}

func (r *fakeBudgetCategoryRepo) GetByID(_ context.Context, _ string, _ string) (*category.Category, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.item == nil {
		return nil, nil
	}
	cloned := *r.item
	return &cloned, nil
}

func TestBudgetServiceCreateSuccess(t *testing.T) {
	repo := &fakeBudgetRepo{}
	categoryID := "cat_def_food"
	categoryRepo := &fakeBudgetCategoryRepo{item: &category.Category{ID: categoryID, IsActive: true}}
	svc := NewService(repo, categoryRepo)

	start := "2026-04-01"
	created, err := svc.Create(context.Background(), "u1", CreateInput{
		Period:      "month",
		PeriodStart: &start,
		Amount:      "1000.00",
		CategoryID:  &categoryID,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created == nil || created.ID == "" {
		t.Fatalf("expected created budget, got %#v", created)
	}
}

func TestBudgetServiceCreateValidationPeriod(t *testing.T) {
	repo := &fakeBudgetRepo{}
	categoryID := "cat_def_food"
	categoryRepo := &fakeBudgetCategoryRepo{item: &category.Category{ID: categoryID, IsActive: true}}
	svc := NewService(repo, categoryRepo)

	_, err := svc.Create(context.Background(), "u1", CreateInput{Period: "year", Amount: "100", CategoryID: &categoryID})
	assertBudgetErrKind(t, err, apperrors.KindValidation)
}

func TestBudgetServiceCreateValidationCategory(t *testing.T) {
	repo := &fakeBudgetRepo{}
	categoryID := "cat_missing"
	categoryRepo := &fakeBudgetCategoryRepo{item: nil}
	svc := NewService(repo, categoryRepo)

	start := "2026-04-01"
	_, err := svc.Create(context.Background(), "u1", CreateInput{Period: "month", PeriodStart: &start, Amount: "100", CategoryID: &categoryID})
	assertBudgetErrKind(t, err, apperrors.KindValidation)
}

func TestBudgetServiceGetNotFound(t *testing.T) {
	repo := &fakeBudgetRepo{}
	svc := NewService(repo, &fakeBudgetCategoryRepo{})

	_, err := svc.Get(context.Background(), "u1", "missing")
	assertBudgetErrKind(t, err, apperrors.KindNotFound)
}

func TestBudgetServiceListInternalError(t *testing.T) {
	repo := &fakeBudgetRepo{listErr: errors.New("db error")}
	svc := NewService(repo, &fakeBudgetCategoryRepo{})

	_, err := svc.List(context.Background(), "u1")
	assertBudgetErrKind(t, err, apperrors.KindInternal)
}

func assertBudgetErrKind(t *testing.T, err error, expected apperrors.Kind) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error kind %s, got nil", expected)
	}
	if got := apperrors.KindOf(err); got != expected {
		t.Fatalf("expected error kind %s, got %s (err=%v)", expected, got, err)
	}
}
