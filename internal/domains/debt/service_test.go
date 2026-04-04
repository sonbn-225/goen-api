package debt

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/money"
	"github.com/sonbn-225/goen-api-v2/internal/domains/contact"
	"github.com/sonbn-225/goen-api-v2/internal/domains/transaction"
)

type fakeDebtRepo struct {
	items          []Debt
	links          []DebtPaymentLink
	installments   []DebtInstallment
	createErr      error
	getErr         error
	listErr        error
	createLinkErr  error
	listLinkErr    error
	createInstErr  error
	listInstErr    error
	lastLinkUpdate *DebtUpdate
}

func (r *fakeDebtRepo) Create(_ context.Context, _ string, input Debt) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.items = append(r.items, input)
	return nil
}

func (r *fakeDebtRepo) GetByID(_ context.Context, userID, debtID string) (*Debt, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	for _, item := range r.items {
		if item.UserID == userID && item.ID == debtID {
			cloned := item
			return &cloned, nil
		}
	}
	return nil, nil
}

func (r *fakeDebtRepo) ListByUser(_ context.Context, userID string) ([]Debt, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	out := make([]Debt, 0)
	for _, item := range r.items {
		if item.UserID == userID {
			out = append(out, item)
		}
	}
	return out, nil
}

func (r *fakeDebtRepo) CreatePaymentLink(_ context.Context, _ string, input DebtPaymentLink, update DebtUpdate) error {
	if r.createLinkErr != nil {
		return r.createLinkErr
	}
	r.links = append(r.links, input)
	r.lastLinkUpdate = &update
	return nil
}

func (r *fakeDebtRepo) ListPaymentLinks(_ context.Context, _ string, debtID string) ([]DebtPaymentLink, error) {
	if r.listLinkErr != nil {
		return nil, r.listLinkErr
	}
	out := make([]DebtPaymentLink, 0)
	for _, item := range r.links {
		if item.DebtID == debtID {
			out = append(out, item)
		}
	}
	return out, nil
}

func (r *fakeDebtRepo) ListPaymentLinksByTransaction(_ context.Context, _ string, transactionID string) ([]DebtPaymentLink, error) {
	if r.listLinkErr != nil {
		return nil, r.listLinkErr
	}
	out := make([]DebtPaymentLink, 0)
	for _, item := range r.links {
		if item.TransactionID == transactionID {
			out = append(out, item)
		}
	}
	return out, nil
}

func (r *fakeDebtRepo) CreateInstallment(_ context.Context, _ string, input DebtInstallment) error {
	if r.createInstErr != nil {
		return r.createInstErr
	}
	r.installments = append(r.installments, input)
	return nil
}

func (r *fakeDebtRepo) ListInstallments(_ context.Context, _ string, debtID string) ([]DebtInstallment, error) {
	if r.listInstErr != nil {
		return nil, r.listInstErr
	}
	out := make([]DebtInstallment, 0)
	for _, item := range r.installments {
		if item.DebtID == debtID {
			out = append(out, item)
		}
	}
	return out, nil
}

type fakeDebtTxService struct {
	items map[string]*transaction.Transaction
	err   error
}

func (s *fakeDebtTxService) Get(_ context.Context, _ string, transactionID string) (*transaction.Transaction, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.items == nil {
		return nil, apperrors.New(apperrors.KindNotFound, "transaction not found")
	}
	item, ok := s.items[transactionID]
	if !ok {
		return nil, apperrors.New(apperrors.KindNotFound, "transaction not found")
	}
	cloned := *item
	return &cloned, nil
}

type fakeDebtContactService struct {
	items     []contact.Contact
	created   *contact.Contact
	listErr   error
	createErr error
}

func (s *fakeDebtContactService) Create(_ context.Context, _ string, input contact.CreateInput) (*contact.Contact, error) {
	if s.createErr != nil {
		return nil, s.createErr
	}
	if s.created != nil {
		cloned := *s.created
		return &cloned, nil
	}
	created := &contact.Contact{ID: "c-created", Name: input.Name}
	return created, nil
}

func (s *fakeDebtContactService) List(_ context.Context, _ string) ([]contact.Contact, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.items, nil
}

type fakeDebtService struct{}

func (s *fakeDebtService) Create(_ context.Context, _ string, _ CreateInput) (*Debt, error) {
	return &Debt{ID: "d1"}, nil
}

func (s *fakeDebtService) Get(_ context.Context, _ string, _ string) (*Debt, error) {
	return &Debt{ID: "d1"}, nil
}

func (s *fakeDebtService) List(_ context.Context, _ string) ([]Debt, error) {
	return []Debt{{ID: "d1"}}, nil
}

func (s *fakeDebtService) CreatePayment(_ context.Context, _ string, _ string, _ CreatePaymentInput) (*DebtPaymentLink, error) {
	return &DebtPaymentLink{ID: "l1"}, nil
}

func (s *fakeDebtService) ListPayments(_ context.Context, _ string, _ string) ([]DebtPaymentLink, error) {
	return []DebtPaymentLink{{ID: "l1"}}, nil
}

func (s *fakeDebtService) ListPaymentsByTransaction(_ context.Context, _ string, _ string) ([]DebtPaymentLink, error) {
	return []DebtPaymentLink{{ID: "l1"}}, nil
}

func (s *fakeDebtService) CreateInstallment(_ context.Context, _ string, _ string, _ CreateInstallmentInput) (*DebtInstallment, error) {
	return &DebtInstallment{ID: "i1"}, nil
}

func (s *fakeDebtService) ListInstallments(_ context.Context, _ string, _ string) ([]DebtInstallment, error) {
	return []DebtInstallment{{ID: "i1"}}, nil
}

func TestDebtServiceCreateSuccessAutoCreateContact(t *testing.T) {
	repo := &fakeDebtRepo{}
	txSvc := &fakeDebtTxService{}
	contactSvc := &fakeDebtContactService{created: &contact.Contact{ID: "c1", Name: "Alice"}}
	svc := NewService(repo, txSvc, contactSvc)

	created, err := svc.Create(context.Background(), "u1", CreateInput{
		AccountID: "a1",
		Direction: "borrowed",
		Name:      strPtr("Alice"),
		Principal: "100.00",
		StartDate: "2026-04-01",
		DueDate:   "2026-05-01",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created == nil || created.ID == "" {
		t.Fatalf("expected created debt with id, got %#v", created)
	}
	if created.ContactID == nil || *created.ContactID != "c1" {
		t.Fatalf("expected auto-created contact id c1, got %#v", created.ContactID)
	}
}

func TestDebtServiceCreateValidation(t *testing.T) {
	repo := &fakeDebtRepo{}
	svc := NewService(repo, &fakeDebtTxService{}, &fakeDebtContactService{})

	_, err := svc.Create(context.Background(), "u1", CreateInput{
		AccountID: "a1",
		Direction: "bad",
		Principal: "100.00",
		StartDate: "2026-04-01",
		DueDate:   "2026-05-01",
	})
	assertDebtErrKind(t, err, apperrors.KindValidation)
}

func TestDebtServiceGetNotFound(t *testing.T) {
	repo := &fakeDebtRepo{}
	svc := NewService(repo, &fakeDebtTxService{}, &fakeDebtContactService{})

	_, err := svc.Get(context.Background(), "u1", "missing")
	assertDebtErrKind(t, err, apperrors.KindNotFound)
}

func TestDebtServiceCreatePaymentSuccess(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeDebtRepo{items: []Debt{{
		ID:                   "d1",
		UserID:               "u1",
		Direction:            "borrowed",
		Principal:            "100.00",
		OutstandingPrincipal: "100.00",
		AccruedInterest:      "5.00",
		InterestRule:         strPtr("interest_first"),
		Status:               "active",
		CreatedAt:            now,
		UpdatedAt:            now,
	}}}
	txSvc := &fakeDebtTxService{items: map[string]*transaction.Transaction{
		"t1": {ID: "t1", UserID: "u1", Type: "expense", Amount: money.MustFromString("10.00")},
	}}
	svc := NewService(repo, txSvc, &fakeDebtContactService{})

	item, err := svc.CreatePayment(context.Background(), "u1", "d1", CreatePaymentInput{TransactionID: "t1"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if item == nil || item.ID == "" {
		t.Fatalf("expected created payment link, got %#v", item)
	}
	if repo.lastLinkUpdate == nil {
		t.Fatal("expected payment link update to be recorded")
	}
	if repo.lastLinkUpdate.OutstandingPrincipal != "95.00" {
		t.Fatalf("expected outstanding_principal=95.00, got %s", repo.lastLinkUpdate.OutstandingPrincipal)
	}
	if repo.lastLinkUpdate.AccruedInterest != "0.00" {
		t.Fatalf("expected accrued_interest=0.00, got %s", repo.lastLinkUpdate.AccruedInterest)
	}
}

func TestDebtServiceListInternalError(t *testing.T) {
	repo := &fakeDebtRepo{listErr: errors.New("db down")}
	svc := NewService(repo, &fakeDebtTxService{}, &fakeDebtContactService{})

	_, err := svc.List(context.Background(), "u1")
	assertDebtErrKind(t, err, apperrors.KindInternal)
}

func TestDebtServiceCreateInstallmentValidation(t *testing.T) {
	repo := &fakeDebtRepo{}
	svc := NewService(repo, &fakeDebtTxService{}, &fakeDebtContactService{})

	_, err := svc.CreateInstallment(context.Background(), "u1", "d1", CreateInstallmentInput{
		InstallmentNo: 0,
		DueDate:       "2026-04-30",
		AmountDue:     "10.00",
	})
	assertDebtErrKind(t, err, apperrors.KindValidation)
}

func assertDebtErrKind(t *testing.T, err error, expected apperrors.Kind) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error kind %s, got nil", expected)
	}
	if got := apperrors.KindOf(err); got != expected {
		t.Fatalf("expected error kind %s, got %s (err=%v)", expected, got, err)
	}
}

func strPtr(v string) *string {
	return &v
}
