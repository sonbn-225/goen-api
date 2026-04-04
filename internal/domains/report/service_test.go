package report

import (
	"context"
	"errors"
	"testing"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
)

type fakeReportRepo struct {
	balances    []AccountBalance
	cashflow    []CashflowStat
	topExpenses []CategoryExpenseStat

	balancesErr error
	cashflowErr error
	topErr      error
}

func (r *fakeReportRepo) ListAccountBalances(_ context.Context, _ string) ([]AccountBalance, error) {
	if r.balancesErr != nil {
		return nil, r.balancesErr
	}
	return r.balances, nil
}

func (r *fakeReportRepo) GetCashflow(_ context.Context, _ string, _ int) ([]CashflowStat, error) {
	if r.cashflowErr != nil {
		return nil, r.cashflowErr
	}
	return r.cashflow, nil
}

func (r *fakeReportRepo) GetTopExpenses(_ context.Context, _ string, _, _, _ int) ([]CategoryExpenseStat, error) {
	if r.topErr != nil {
		return nil, r.topErr
	}
	return r.topExpenses, nil
}

func TestReportServiceGetDashboardSuccess(t *testing.T) {
	repo := &fakeReportRepo{
		balances:    []AccountBalance{{AccountID: "a1", Currency: "VND", Balance: "1000"}},
		cashflow:    []CashflowStat{{Month: "2026-04", Income: "1500", Expense: "300"}},
		topExpenses: []CategoryExpenseStat{{CategoryID: "cat_food", Amount: "300"}},
	}
	svc := NewService(repo)

	out, err := svc.GetDashboardReport(context.Background(), "u1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out == nil {
		t.Fatal("expected report not nil")
	}
	if len(out.TotalBalances) != 1 || len(out.Cashflow6Months) != 1 || len(out.TopExpensesMonth) != 1 {
		t.Fatalf("unexpected report payload: %#v", out)
	}
}

func TestReportServiceGetDashboardRequiresUser(t *testing.T) {
	repo := &fakeReportRepo{}
	svc := NewService(repo)

	_, err := svc.GetDashboardReport(context.Background(), "")
	if apperrors.KindOf(err) != apperrors.KindUnauth {
		t.Fatalf("expected unauth kind, got %s", apperrors.KindOf(err))
	}
}

func TestReportServiceGetDashboardInternalError(t *testing.T) {
	repo := &fakeReportRepo{balancesErr: errors.New("db down")}
	svc := NewService(repo)

	_, err := svc.GetDashboardReport(context.Background(), "u1")
	if apperrors.KindOf(err) != apperrors.KindInternal {
		t.Fatalf("expected internal kind, got %s", apperrors.KindOf(err))
	}
}

func TestReportServiceGetDashboardNormalizesNilSlices(t *testing.T) {
	repo := &fakeReportRepo{}
	svc := NewService(repo)

	out, err := svc.GetDashboardReport(context.Background(), "u1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out.TotalBalances == nil || out.Cashflow6Months == nil || out.TopExpensesMonth == nil {
		t.Fatalf("expected slices to be non-nil, got %#v", out)
	}
}
