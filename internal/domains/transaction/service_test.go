package transaction

import (
	"context"
	"errors"
	"testing"

	"github.com/sonbn-225/goen-api-v2/internal/core/apperrors"
	"github.com/sonbn-225/goen-api-v2/internal/core/money"
)

type fakeTransactionRepo struct {
	items             []Transaction
	participants      []GroupExpenseParticipant
	createOptions     []CreateOptions
	createErr         error
	updateErr         error
	listErr           error
	getErr            error
	batchPatchErr     error
	batchPatchUpdated []string
}

func (r *fakeTransactionRepo) Create(_ context.Context, tx *Transaction, opts CreateOptions) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.items = append(r.items, *tx)
	r.createOptions = append(r.createOptions, opts)
	return nil
}

func (r *fakeTransactionRepo) Update(_ context.Context, userID, transactionID string, input UpdateInput) (*Transaction, error) {
	if r.updateErr != nil {
		return nil, r.updateErr
	}
	if input.GroupParticipants != nil {
		retained := make([]GroupExpenseParticipant, 0, len(r.participants))
		for _, participant := range r.participants {
			if participant.TransactionID == transactionID && !participant.IsSettled {
				continue
			}
			retained = append(retained, participant)
		}
		r.participants = retained

		for _, participant := range *input.GroupParticipants {
			r.participants = append(r.participants, GroupExpenseParticipant{
				UserID:          userID,
				TransactionID:   transactionID,
				ParticipantName: participant.ParticipantName,
				OriginalAmount:  participant.OriginalAmount.String(),
				ShareAmount:     participant.ShareAmount.String(),
				IsSettled:       false,
			})
		}
	}

	for i, item := range r.items {
		if item.UserID != userID || item.ID != transactionID {
			continue
		}
		if input.LineItems != nil {
			total := money.Zero().Decimal
			lineItems := make([]TransactionLineItem, 0, len(*input.LineItems))
			for _, li := range *input.LineItems {
				total = total.Add(li.Amount.Decimal)
				lineItems = append(lineItems, TransactionLineItem{
					CategoryID: li.CategoryID,
					TagIDs:     li.TagIDs,
					Amount:     li.Amount.String(),
					Note:       li.Note,
				})
			}
			item.Amount = money.Amount{Decimal: total}
			item.LineItems = lineItems
		}
		if input.Note != nil {
			item.Note = *input.Note
			if input.LineItems == nil && len(item.LineItems) > 0 {
				item.LineItems[0].Note = input.Note
			}
		}
		r.items[i] = item
		cloned := item
		return &cloned, nil
	}
	return nil, nil
}

func (r *fakeTransactionRepo) ListByUser(_ context.Context, userID string, filter ListFilter) ([]Transaction, int, error) {
	if r.listErr != nil {
		return nil, 0, r.listErr
	}
	out := make([]Transaction, 0)
	for _, item := range r.items {
		if item.UserID == userID {
			if filter.Type != nil && item.Type != *filter.Type {
				continue
			}
			if filter.Status != nil && item.Status != *filter.Status {
				continue
			}
			out = append(out, item)
		}
	}
	total := len(out)
	if filter.Limit > 0 && len(out) > filter.Limit {
		out = out[:filter.Limit]
	}
	return out, total, nil
}

func (r *fakeTransactionRepo) GetByID(_ context.Context, userID, transactionID string) (*Transaction, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	for _, item := range r.items {
		if item.UserID == userID && item.ID == transactionID {
			cloned := item
			return &cloned, nil
		}
	}
	return nil, nil
}

func (r *fakeTransactionRepo) BatchPatchStatus(_ context.Context, _ string, _ []string, _ string) ([]string, error) {
	if r.batchPatchErr != nil {
		return nil, r.batchPatchErr
	}
	return r.batchPatchUpdated, nil
}

func (r *fakeTransactionRepo) ListGroupParticipantsByTransaction(_ context.Context, _ string, transactionID string) ([]GroupExpenseParticipant, error) {
	out := make([]GroupExpenseParticipant, 0)
	for _, item := range r.participants {
		if item.TransactionID == transactionID {
			out = append(out, item)
		}
	}
	return out, nil
}

func TestServiceCreateExpenseSuccess(t *testing.T) {
	repo := &fakeTransactionRepo{}
	svc := NewService(repo)

	tx, err := svc.Create(context.Background(), "u1", CreateInput{
		AccountID: strPtr("a1"),
		Type:      "expense",
		Amount:    money.MustFromString("10"),
		LineItems: []CreateTransactionLineItemInput{{
			CategoryID: strPtr("c1"),
			Amount:     money.MustFromString("10"),
		}},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.Type != "expense" {
		t.Fatalf("expected type expense, got %s", tx.Type)
	}
	if len(repo.items) != 1 {
		t.Fatalf("expected 1 transaction in repo, got %d", len(repo.items))
	}
	if len(repo.createOptions) != 1 || len(repo.createOptions[0].LineItems) != 1 {
		t.Fatalf("expected 1 line item in create options")
	}
}

func TestServiceCreateUnauthorizedMissingUserID(t *testing.T) {
	repo := &fakeTransactionRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "", CreateInput{AccountID: strPtr("a1"), Type: "expense", Amount: money.MustFromString("10"), LineItems: []CreateTransactionLineItemInput{{CategoryID: strPtr("c1"), Amount: money.MustFromString("10")}}})
	assertTransactionErrKind(t, err, apperrors.KindUnauth)
}

func TestServiceCreateValidationUnsupportedType(t *testing.T) {
	repo := &fakeTransactionRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "u1", CreateInput{AccountID: strPtr("a1"), Type: "borrow", Amount: money.MustFromString("10"), LineItems: []CreateTransactionLineItemInput{{CategoryID: strPtr("c1"), Amount: money.MustFromString("10")}}})
	assertTransactionErrKind(t, err, apperrors.KindValidation)
}

func TestServiceCreateValidationZeroAmount(t *testing.T) {
	repo := &fakeTransactionRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "u1", CreateInput{AccountID: strPtr("a1"), Type: "expense", Amount: money.Zero(), LineItems: []CreateTransactionLineItemInput{{CategoryID: strPtr("c1"), Amount: money.MustFromString("10")}}})
	assertTransactionErrKind(t, err, apperrors.KindValidation)
}

func TestServiceCreateValidationMissingAccountIDForExpense(t *testing.T) {
	repo := &fakeTransactionRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "u1", CreateInput{Type: "expense", Amount: money.MustFromString("10"), LineItems: []CreateTransactionLineItemInput{{CategoryID: strPtr("c1"), Amount: money.MustFromString("10")}}})
	assertTransactionErrKind(t, err, apperrors.KindValidation)
}

func TestServiceCreateValidationTransferSameAccount(t *testing.T) {
	repo := &fakeTransactionRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "u1", CreateInput{
		Type:          "transfer",
		FromAccountID: strPtr("a1"),
		ToAccountID:   strPtr("a1"),
		Amount:        money.MustFromString("10"),
	})
	assertTransactionErrKind(t, err, apperrors.KindValidation)
}

func TestServiceCreateValidationLineItemsRequiredForExpense(t *testing.T) {
	repo := &fakeTransactionRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "u1", CreateInput{
		AccountID: strPtr("a1"),
		Type:      "expense",
		Amount:    money.MustFromString("10"),
	})
	assertTransactionErrKind(t, err, apperrors.KindValidation)
}

func TestServiceCreateValidationLineItemsMustBeEmptyForTransfer(t *testing.T) {
	repo := &fakeTransactionRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "u1", CreateInput{
		Type:          "transfer",
		FromAccountID: strPtr("a1"),
		ToAccountID:   strPtr("a2"),
		Amount:        money.MustFromString("10"),
		LineItems: []CreateTransactionLineItemInput{{
			CategoryID: strPtr("c1"),
			Amount:     money.MustFromString("10"),
		}},
	})
	assertTransactionErrKind(t, err, apperrors.KindValidation)
}

func TestServiceCreateExpenseWithGroupParticipantsAutoCalculatesShare(t *testing.T) {
	repo := &fakeTransactionRepo{}
	svc := NewService(repo)

	tx, err := svc.Create(context.Background(), "u1", CreateInput{
		AccountID: strPtr("a1"),
		Type:      "expense",
		Amount:    money.MustFromString("100"),
		GroupParticipants: []CreateGroupExpenseParticipantInput{
			{ParticipantName: "Alice", OriginalAmount: money.MustFromString("20")},
			{ParticipantName: "Bob", OriginalAmount: money.MustFromString("30")},
		},
		LineItems: []CreateTransactionLineItemInput{{
			CategoryID: strPtr("c1"),
			Amount:     money.MustFromString("100"),
		}},
		OwnerOriginalAmount: amountPtr("50"),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx == nil {
		t.Fatalf("expected transaction, got nil")
	}
	if len(repo.createOptions) != 1 {
		t.Fatalf("expected 1 create options item, got %d", len(repo.createOptions))
	}
	if len(repo.createOptions[0].GroupParticipants) != 2 {
		t.Fatalf("expected 2 group participants, got %d", len(repo.createOptions[0].GroupParticipants))
	}
	if len(repo.createOptions[0].LineItems) != 1 {
		t.Fatalf("expected 1 line item, got %d", len(repo.createOptions[0].LineItems))
	}
	if got := repo.createOptions[0].GroupParticipants[0].ShareAmount.String(); got != "20" {
		t.Fatalf("expected Alice share 20, got %s", got)
	}
	if got := repo.createOptions[0].GroupParticipants[1].ShareAmount.String(); got != "30" {
		t.Fatalf("expected Bob share 30, got %s", got)
	}
}

func TestServiceCreateTransferRejectsGroupParticipants(t *testing.T) {
	repo := &fakeTransactionRepo{}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "u1", CreateInput{
		Type:          "transfer",
		FromAccountID: strPtr("a1"),
		ToAccountID:   strPtr("a2"),
		Amount:        money.MustFromString("10"),
		GroupParticipants: []CreateGroupExpenseParticipantInput{{
			ParticipantName: "Alice",
			OriginalAmount:  money.MustFromString("10"),
		}},
	})
	assertTransactionErrKind(t, err, apperrors.KindValidation)
}

func TestServiceListUnauthorizedMissingUserID(t *testing.T) {
	repo := &fakeTransactionRepo{}
	svc := NewService(repo)

	_, _, err := svc.List(context.Background(), "", ListFilter{})
	assertTransactionErrKind(t, err, apperrors.KindUnauth)
}

func TestServiceListInternalError(t *testing.T) {
	repo := &fakeTransactionRepo{listErr: errors.New("db down")}
	svc := NewService(repo)

	_, _, err := svc.List(context.Background(), "u1", ListFilter{})
	assertTransactionErrKind(t, err, apperrors.KindInternal)
}

func TestServiceGetSuccess(t *testing.T) {
	repo := &fakeTransactionRepo{items: []Transaction{{ID: "t1", UserID: "u1", Type: "expense", Status: "pending", Amount: money.MustFromString("10")}}}
	svc := NewService(repo)

	item, err := svc.Get(context.Background(), "u1", "t1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if item.ID != "t1" {
		t.Fatalf("expected transaction t1, got %s", item.ID)
	}
}

func TestServiceBatchPatchStatusSuccess(t *testing.T) {
	repo := &fakeTransactionRepo{batchPatchUpdated: []string{"t1"}}
	svc := NewService(repo)

	result, err := svc.BatchPatchStatus(context.Background(), "u1", BatchPatchRequest{
		TransactionIDs: []string{"t1", "t2"},
		Patch:          BatchPatchData{Status: "posted"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.UpdatedCount != 1 || result.FailedCount != 1 {
		t.Fatalf("expected updated=1 failed=1, got updated=%d failed=%d", result.UpdatedCount, result.FailedCount)
	}
}

func TestServiceUpdateExpenseWithLineItemsSuccess(t *testing.T) {
	repo := &fakeTransactionRepo{items: []Transaction{{ID: "t1", UserID: "u1", Type: "expense", Amount: money.MustFromString("10")}}}
	svc := NewService(repo)

	updated, err := svc.Update(context.Background(), "u1", "t1", UpdateInput{
		Note: strPtr("updated note"),
		LineItems: &[]UpdateTransactionLineItemInput{{
			CategoryID: strPtr("c1"),
			Amount:     money.MustFromString("30"),
		}},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated == nil {
		t.Fatalf("expected updated transaction, got nil")
	}
	if updated.Amount.String() != "30" {
		t.Fatalf("expected amount 30, got %s", updated.Amount.String())
	}
	if len(updated.LineItems) != 1 {
		t.Fatalf("expected 1 line item, got %d", len(updated.LineItems))
	}
}

func TestServiceUpdateTransferRejectsNonEmptyLineItems(t *testing.T) {
	repo := &fakeTransactionRepo{items: []Transaction{{ID: "t1", UserID: "u1", Type: "transfer", Amount: money.MustFromString("10")}}}
	svc := NewService(repo)

	_, err := svc.Update(context.Background(), "u1", "t1", UpdateInput{
		LineItems: &[]UpdateTransactionLineItemInput{{
			CategoryID: strPtr("c1"),
			Amount:     money.MustFromString("10"),
		}},
	})
	assertTransactionErrKind(t, err, apperrors.KindValidation)
}

func TestServiceUpdateExpenseRequiresLineItemsWhenProvided(t *testing.T) {
	repo := &fakeTransactionRepo{items: []Transaction{{ID: "t1", UserID: "u1", Type: "expense", Amount: money.MustFromString("10")}}}
	svc := NewService(repo)

	_, err := svc.Update(context.Background(), "u1", "t1", UpdateInput{LineItems: &[]UpdateTransactionLineItemInput{}})
	assertTransactionErrKind(t, err, apperrors.KindValidation)
}

func TestServiceUpdateTransferRejectsGroupParticipants(t *testing.T) {
	repo := &fakeTransactionRepo{items: []Transaction{{ID: "t1", UserID: "u1", Type: "transfer", Amount: money.MustFromString("10")}}}
	svc := NewService(repo)

	_, err := svc.Update(context.Background(), "u1", "t1", UpdateInput{
		GroupParticipants: &[]UpdateGroupExpenseParticipantInput{{
			ParticipantName: "Alice",
			OriginalAmount:  money.MustFromString("5"),
			ShareAmount:     money.MustFromString("5"),
		}},
	})
	assertTransactionErrKind(t, err, apperrors.KindValidation)
}

func TestServiceUpdateExpenseReplacesOnlyUnsettledGroupParticipants(t *testing.T) {
	repo := &fakeTransactionRepo{
		items: []Transaction{{ID: "t1", UserID: "u1", Type: "expense", Amount: money.MustFromString("10")}},
		participants: []GroupExpenseParticipant{
			{TransactionID: "t1", ParticipantName: "OldUnsettled", OriginalAmount: "4", ShareAmount: "4", IsSettled: false},
			{TransactionID: "t1", ParticipantName: "SettledKeep", OriginalAmount: "6", ShareAmount: "6", IsSettled: true},
		},
	}
	svc := NewService(repo)

	updated, err := svc.Update(context.Background(), "u1", "t1", UpdateInput{
		GroupParticipants: &[]UpdateGroupExpenseParticipantInput{{
			ParticipantName: "NewParticipant",
			OriginalAmount:  money.MustFromString("8"),
			ShareAmount:     money.MustFromString("8"),
		}},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated == nil {
		t.Fatalf("expected updated transaction, got nil")
	}

	if len(repo.participants) != 2 {
		t.Fatalf("expected 2 participants after replacement, got %d", len(repo.participants))
	}

	containsSettled := false
	containsNew := false
	containsOldUnsettled := false
	for _, participant := range repo.participants {
		switch participant.ParticipantName {
		case "SettledKeep":
			containsSettled = true
		case "NewParticipant":
			containsNew = true
		case "OldUnsettled":
			containsOldUnsettled = true
		}
	}

	if !containsSettled {
		t.Fatalf("expected settled participant to be preserved")
	}
	if !containsNew {
		t.Fatalf("expected new participant to be inserted")
	}
	if containsOldUnsettled {
		t.Fatalf("expected old unsettled participant to be removed")
	}
}

func assertTransactionErrKind(t *testing.T, err error, expected apperrors.Kind) {
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

func amountPtr(v string) *money.Amount {
	a := money.MustFromString(v)
	return &a
}
