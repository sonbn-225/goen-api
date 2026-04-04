package savings

import (
	"context"
	"errors"
	"testing"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/money"
	"github.com/sonbn-225/goen-api-v2/internal/domains/transaction"
)

type fakeSavingsRepo struct {
	accounts    map[string]*AccountRef
	instruments map[string]SavingsInstrument

	createErr error
	getErr    error
	listErr   error
	updateErr error
	deleteErr error
}

func (r *fakeSavingsRepo) GetAccountForUser(_ context.Context, _ string, accountID string) (*AccountRef, error) {
	if r.accounts == nil {
		return nil, nil
	}
	item, ok := r.accounts[accountID]
	if !ok {
		return nil, nil
	}
	copy := *item
	return &copy, nil
}

func (r *fakeSavingsRepo) CreateLinkedSavingsAccount(_ context.Context, _ string, parentAccountID, accountName, currency string) (*AccountRef, error) {
	if r.accounts == nil {
		r.accounts = make(map[string]*AccountRef)
	}
	id := "acc_savings_auto"
	item := &AccountRef{ID: id, Name: accountName, Type: "savings", Currency: currency, ParentAccountID: &parentAccountID}
	r.accounts[id] = item
	copy := *item
	return &copy, nil
}

func (r *fakeSavingsRepo) DeleteAccountForUser(_ context.Context, _ string, accountID string) error {
	if r.accounts != nil {
		delete(r.accounts, accountID)
	}
	return nil
}

func (r *fakeSavingsRepo) CreateSavingsInstrument(_ context.Context, _ string, item SavingsInstrument) error {
	if r.createErr != nil {
		return r.createErr
	}
	if r.instruments == nil {
		r.instruments = make(map[string]SavingsInstrument)
	}
	r.instruments[item.ID] = item
	return nil
}

func (r *fakeSavingsRepo) GetSavingsInstrument(_ context.Context, _ string, instrumentID string) (*SavingsInstrument, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	if r.instruments == nil {
		return nil, nil
	}
	item, ok := r.instruments[instrumentID]
	if !ok {
		return nil, nil
	}
	copy := item
	return &copy, nil
}

func (r *fakeSavingsRepo) ListSavingsInstruments(_ context.Context, _ string) ([]SavingsInstrument, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	items := make([]SavingsInstrument, 0, len(r.instruments))
	for _, v := range r.instruments {
		items = append(items, v)
	}
	return items, nil
}

func (r *fakeSavingsRepo) UpdateSavingsInstrument(_ context.Context, _ string, item SavingsInstrument) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	if r.instruments == nil {
		r.instruments = make(map[string]SavingsInstrument)
	}
	r.instruments[item.ID] = item
	return nil
}

func (r *fakeSavingsRepo) DeleteSavingsInstrument(_ context.Context, _ string, instrumentID string) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	if r.instruments != nil {
		delete(r.instruments, instrumentID)
	}
	return nil
}

type fakeSavingsTxService struct {
	called bool
	input  transaction.CreateInput
}

func (s *fakeSavingsTxService) Create(_ context.Context, _ string, input transaction.CreateInput) (*transaction.Transaction, error) {
	s.called = true
	s.input = input
	return &transaction.Transaction{ID: "tx_1"}, nil
}

func TestSavingsServiceCreateWithAutoLinkedAccount(t *testing.T) {
	repo := &fakeSavingsRepo{
		accounts: map[string]*AccountRef{
			"acc_parent": {ID: "acc_parent", Name: "Main", Type: "bank", Currency: "VND"},
		},
	}
	txSvc := &fakeSavingsTxService{}
	svc := NewService(repo, txSvc)

	name := "Term Deposit"
	parentID := "acc_parent"
	created, err := svc.Create(context.Background(), "u1", CreateInput{
		Name:            &name,
		ParentAccountID: &parentID,
		Principal:       "1500000.00",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created == nil {
		t.Fatal("expected created instrument")
	}
	if created.SavingsAccountID == "" {
		t.Fatal("expected savings account id")
	}
	if !txSvc.called {
		t.Fatal("expected initial transfer transaction to be created")
	}
	if txSvc.input.Type != "transfer" {
		t.Fatalf("expected transfer tx, got %s", txSvc.input.Type)
	}
	if txSvc.input.Amount.Cmp(money.MustFromString("1500000.00").Decimal) != 0 {
		t.Fatalf("unexpected transfer amount: %s", txSvc.input.Amount.String())
	}
}

func TestSavingsServiceCreateRequiresUser(t *testing.T) {
	svc := NewService(&fakeSavingsRepo{}, &fakeSavingsTxService{})

	_, err := svc.Create(context.Background(), "", CreateInput{Principal: "1000"})
	if apperrors.KindOf(err) != apperrors.KindUnauth {
		t.Fatalf("expected unauth kind, got %s", apperrors.KindOf(err))
	}
}

func TestSavingsServiceListWrapsInternalError(t *testing.T) {
	svc := NewService(&fakeSavingsRepo{listErr: errors.New("db down")}, &fakeSavingsTxService{})

	_, err := svc.List(context.Background(), "u1")
	if apperrors.KindOf(err) != apperrors.KindInternal {
		t.Fatalf("expected internal kind, got %s", apperrors.KindOf(err))
	}
}
