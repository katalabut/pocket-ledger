package ledger

import (
	"context"
	"testing"
	"time"

	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

var fixedTime = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

func clock() time.Time { return fixedTime }

func newTestService() *Service {
	return NewService(
		&memAccountRepo{items: map[string]*domain.Account{}},
		&memCategoryRepo{items: map[string]*domain.Category{}},
		&memTransactionRepo{items: map[string]*domain.Transaction{}},
		&memSplitRepo{items: map[string][]domain.Split{}},
		clock,
	)
}

// --- memory repos ---

type memAccountRepo struct{ items map[string]*domain.Account }

func (r *memAccountRepo) Create(_ context.Context, a *domain.Account) error {
	r.items[a.ID] = a
	return nil
}
func (r *memAccountRepo) GetByID(_ context.Context, id string) (*domain.Account, error) {
	a, ok := r.items[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return a, nil
}
func (r *memAccountRepo) List(_ context.Context) ([]domain.Account, error) {
	out := make([]domain.Account, 0, len(r.items))
	for _, a := range r.items {
		out = append(out, *a)
	}
	return out, nil
}
func (r *memAccountRepo) Update(_ context.Context, a *domain.Account) error {
	if _, ok := r.items[a.ID]; !ok {
		return domain.ErrNotFound
	}
	r.items[a.ID] = a
	return nil
}
func (r *memAccountRepo) Delete(_ context.Context, id string) error {
	delete(r.items, id)
	return nil
}

type memCategoryRepo struct{ items map[string]*domain.Category }

func (r *memCategoryRepo) Create(_ context.Context, c *domain.Category) error {
	r.items[c.ID] = c
	return nil
}
func (r *memCategoryRepo) GetByID(_ context.Context, id string) (*domain.Category, error) {
	c, ok := r.items[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return c, nil
}
func (r *memCategoryRepo) List(_ context.Context) ([]domain.Category, error) {
	out := make([]domain.Category, 0, len(r.items))
	for _, c := range r.items {
		out = append(out, *c)
	}
	return out, nil
}
func (r *memCategoryRepo) Update(_ context.Context, c *domain.Category) error {
	if _, ok := r.items[c.ID]; !ok {
		return domain.ErrNotFound
	}
	r.items[c.ID] = c
	return nil
}
func (r *memCategoryRepo) Delete(_ context.Context, id string) error {
	delete(r.items, id)
	return nil
}

type memTransactionRepo struct{ items map[string]*domain.Transaction }

func (r *memTransactionRepo) Create(_ context.Context, t *domain.Transaction) error {
	r.items[t.ID] = t
	return nil
}
func (r *memTransactionRepo) GetByID(_ context.Context, id string) (*domain.Transaction, error) {
	t, ok := r.items[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return t, nil
}
func (r *memTransactionRepo) List(_ context.Context, _ TransactionFilter) ([]domain.Transaction, int, error) {
	out := make([]domain.Transaction, 0, len(r.items))
	for _, t := range r.items {
		out = append(out, *t)
	}
	return out, len(out), nil
}
func (r *memTransactionRepo) Update(_ context.Context, t *domain.Transaction) error {
	if _, ok := r.items[t.ID]; !ok {
		return domain.ErrNotFound
	}
	r.items[t.ID] = t
	return nil
}
func (r *memTransactionRepo) Delete(_ context.Context, id string) error {
	delete(r.items, id)
	return nil
}
func (r *memTransactionRepo) ExistsByDedupeKey(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

type memSplitRepo struct{ items map[string][]domain.Split }

func (r *memSplitRepo) ReplaceSplits(_ context.Context, txID string, splits []domain.Split) error {
	r.items[txID] = splits
	return nil
}
func (r *memSplitRepo) GetByTransactionID(_ context.Context, txID string) ([]domain.Split, error) {
	return r.items[txID], nil
}
func (r *memSplitRepo) GetByTransactionIDs(_ context.Context, txIDs []string) (map[string][]domain.Split, error) {
	out := map[string][]domain.Split{}
	for _, id := range txIDs {
		out[id] = r.items[id]
	}
	return out, nil
}

// --- tests ---

func TestCreateAccount_Valid(t *testing.T) {
	svc := newTestService()
	a, err := svc.CreateAccount(context.Background(), CreateAccountInput{
		Name: "Test", Currency: "EUR", Type: "card",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Name != "Test" || a.Currency != "EUR" {
		t.Fatalf("unexpected account: %+v", a)
	}
}

func TestCreateAccount_InvalidCurrency(t *testing.T) {
	svc := newTestService()
	_, err := svc.CreateAccount(context.Background(), CreateAccountInput{
		Name: "Test", Currency: "XX", Type: "card",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateTransaction_Valid(t *testing.T) {
	svc := newTestService()
	a, _ := svc.CreateAccount(context.Background(), CreateAccountInput{
		Name: "Test", Currency: "EUR", Type: "card",
	})
	tx, err := svc.CreateTransaction(context.Background(), CreateTransactionInput{
		AccountID: a.ID, OccurredAt: "2025-01-01T00:00:00Z",
		AmountMinor: -1000, Currency: "EUR", Description: "Grocery",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.AmountMinor != -1000 {
		t.Fatalf("unexpected amount: %d", tx.AmountMinor)
	}
}

func TestReplaceSplits_Valid(t *testing.T) {
	svc := newTestService()
	a, _ := svc.CreateAccount(context.Background(), CreateAccountInput{
		Name: "Test", Currency: "EUR", Type: "card",
	})
	c1, _ := svc.CreateCategory(context.Background(), CreateCategoryInput{Name: "Food"})
	c2, _ := svc.CreateCategory(context.Background(), CreateCategoryInput{Name: "Drink"})
	tx, _ := svc.CreateTransaction(context.Background(), CreateTransactionInput{
		AccountID: a.ID, OccurredAt: "2025-01-01T00:00:00Z",
		AmountMinor: -1000, Currency: "EUR", Description: "Grocery",
	})
	splits, err := svc.ReplaceSplits(context.Background(), tx.ID, []SplitInput{
		{CategoryID: c1.ID, AmountMinor: -600},
		{CategoryID: c2.ID, AmountMinor: -400},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(splits) != 2 {
		t.Fatalf("expected 2 splits, got %d", len(splits))
	}
}

func TestReplaceSplits_Mismatch(t *testing.T) {
	svc := newTestService()
	a, _ := svc.CreateAccount(context.Background(), CreateAccountInput{
		Name: "Test", Currency: "EUR", Type: "card",
	})
	c1, _ := svc.CreateCategory(context.Background(), CreateCategoryInput{Name: "Food"})
	tx, _ := svc.CreateTransaction(context.Background(), CreateTransactionInput{
		AccountID: a.ID, OccurredAt: "2025-01-01T00:00:00Z",
		AmountMinor: -1000, Currency: "EUR", Description: "Grocery",
	})
	_, err := svc.ReplaceSplits(context.Background(), tx.ID, []SplitInput{
		{CategoryID: c1.ID, AmountMinor: -500},
	})
	if err != domain.ErrSplitMismatch {
		t.Fatalf("expected ErrSplitMismatch, got %v", err)
	}
}
