package budget

import (
	"context"
	"testing"
	"time"

	"github.com/katalabut/pocket-ledger/backend/internal/application/fx"
	"github.com/katalabut/pocket-ledger/backend/internal/application/ledger"
	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

var fixedTime = time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

func clock() time.Time { return fixedTime }

// --- minimal memory implementations ---

type memBudgetRepo struct{ items []domain.Budget }

func (r *memBudgetRepo) Upsert(_ context.Context, b *domain.Budget) error {
	for i, existing := range r.items {
		if existing.Month == b.Month && existing.CategoryID == b.CategoryID {
			r.items[i] = *b
			return nil
		}
	}
	r.items = append(r.items, *b)
	return nil
}
func (r *memBudgetRepo) ListByMonth(_ context.Context, month string) ([]domain.Budget, error) {
	var out []domain.Budget
	for _, b := range r.items {
		if b.Month == month {
			out = append(out, b)
		}
	}
	return out, nil
}

type memAccountRepo struct{ items map[string]*domain.Account }

func (r *memAccountRepo) Create(_ context.Context, a *domain.Account) error {
	r.items[a.ID] = a; return nil
}
func (r *memAccountRepo) GetByID(_ context.Context, id string) (*domain.Account, error) {
	a, ok := r.items[id]; if !ok { return nil, domain.ErrNotFound }; return a, nil
}
func (r *memAccountRepo) List(_ context.Context) ([]domain.Account, error) {
	var out []domain.Account; for _, a := range r.items { out = append(out, *a) }; return out, nil
}
func (r *memAccountRepo) Update(_ context.Context, a *domain.Account) error {
	r.items[a.ID] = a; return nil
}
func (r *memAccountRepo) Delete(_ context.Context, _ string) error { return nil }

type memCategoryRepo struct{ items map[string]*domain.Category }

func (r *memCategoryRepo) Create(_ context.Context, c *domain.Category) error {
	r.items[c.ID] = c; return nil
}
func (r *memCategoryRepo) GetByID(_ context.Context, id string) (*domain.Category, error) {
	c, ok := r.items[id]; if !ok { return nil, domain.ErrNotFound }; return c, nil
}
func (r *memCategoryRepo) List(_ context.Context) ([]domain.Category, error) {
	var out []domain.Category; for _, c := range r.items { out = append(out, *c) }; return out, nil
}
func (r *memCategoryRepo) Update(_ context.Context, c *domain.Category) error {
	r.items[c.ID] = c; return nil
}
func (r *memCategoryRepo) Delete(_ context.Context, _ string) error { return nil }

type memTransactionRepo struct{ items map[string]*domain.Transaction }

func (r *memTransactionRepo) Create(_ context.Context, t *domain.Transaction) error {
	r.items[t.ID] = t; return nil
}
func (r *memTransactionRepo) GetByID(_ context.Context, id string) (*domain.Transaction, error) {
	t, ok := r.items[id]; if !ok { return nil, domain.ErrNotFound }; return t, nil
}
func (r *memTransactionRepo) List(_ context.Context, f ledger.TransactionFilter) ([]domain.Transaction, int, error) {
	var out []domain.Transaction
	for _, t := range r.items {
		if f.From != nil && t.OccurredAt.Format(time.RFC3339) < *f.From {
			continue
		}
		if f.To != nil && t.OccurredAt.Format(time.RFC3339) >= *f.To {
			continue
		}
		out = append(out, *t)
	}
	return out, len(out), nil
}
func (r *memTransactionRepo) Update(_ context.Context, t *domain.Transaction) error {
	r.items[t.ID] = t; return nil
}
func (r *memTransactionRepo) Delete(_ context.Context, _ string) error { return nil }
func (r *memTransactionRepo) ExistsByDedupeKey(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}

type memSplitRepo struct{ items map[string][]domain.Split }

func (r *memSplitRepo) ReplaceSplits(_ context.Context, txID string, splits []domain.Split) error {
	r.items[txID] = splits; return nil
}
func (r *memSplitRepo) GetByTransactionID(_ context.Context, txID string) ([]domain.Split, error) {
	return r.items[txID], nil
}
func (r *memSplitRepo) GetByTransactionIDs(_ context.Context, txIDs []string) (map[string][]domain.Split, error) {
	out := map[string][]domain.Split{}
	for _, id := range txIDs { out[id] = r.items[id] }
	return out, nil
}

type memFXRepo struct{ rates []domain.FXRate }

func (r *memFXRepo) Upsert(_ context.Context, rate *domain.FXRate) error {
	r.rates = append(r.rates, *rate); return nil
}
func (r *memFXRepo) GetRate(_ context.Context, date, base, quote string) (*domain.FXRate, error) {
	for _, rate := range r.rates {
		if rate.Date == date && rate.Base == base && rate.Quote == quote { return &rate, nil }
	}
	return nil, domain.ErrNotFound
}
func (r *memFXRepo) GetLatestRateBefore(_ context.Context, date, base, quote string) (*domain.FXRate, error) {
	return nil, domain.ErrNotFound
}
func (r *memFXRepo) ListRatesByDate(_ context.Context, date string) ([]domain.FXRate, error) {
	return nil, nil
}

type memECB struct{}

func (m *memECB) FetchDailyRates(_ context.Context) ([]domain.FXRate, error) { return nil, nil }

// --- tests ---

func TestBudgetSpentCalculation(t *testing.T) {
	ctx := context.Background()

	catID := "cat-food"
	accountID := "acc-1"
	cats := &memCategoryRepo{items: map[string]*domain.Category{
		catID: {ID: catID, Name: "Food"},
	}}
	accs := &memAccountRepo{items: map[string]*domain.Account{
		accountID: {ID: accountID, Name: "Main", Currency: "EUR", Type: "card"},
	}}
	txRepo := &memTransactionRepo{items: map[string]*domain.Transaction{
		"tx1": {ID: "tx1", AccountID: accountID, OccurredAt: time.Date(2025, 6, 10, 0, 0, 0, 0, time.UTC),
			AmountMinor: -5000, Currency: "EUR", CategoryID: &catID},
		"tx2": {ID: "tx2", AccountID: accountID, OccurredAt: time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
			AmountMinor: -3000, Currency: "EUR", CategoryID: &catID},
	}}
	splitRepo := &memSplitRepo{items: map[string][]domain.Split{}}
	budgetRepo := &memBudgetRepo{}

	ledgerSvc := ledger.NewService(accs, cats, txRepo, splitRepo, clock)
	fxSvc := fx.NewService(&memFXRepo{}, &memECB{}, "EUR", clock)
	svc := NewService(budgetRepo, ledgerSvc, fxSvc, clock)

	// Set budget for food: 10000 (100.00 EUR)
	_, err := svc.Upsert(ctx, UpsertInput{Month: "2025-06", CategoryID: catID, PlannedMinor: 10000})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}

	report, err := svc.GetReport(ctx, "2025-06")
	if err != nil {
		t.Fatalf("report: %v", err)
	}
	if len(report) != 1 {
		t.Fatalf("expected 1 budget row, got %d", len(report))
	}
	r := report[0]
	if r.PlannedMinor != 10000 {
		t.Fatalf("expected planned 10000, got %d", r.PlannedMinor)
	}
	if r.SpentMinor != -8000 {
		t.Fatalf("expected spent -8000, got %d", r.SpentMinor)
	}
	// remaining = 10000 + (-8000) = 2000
	if r.RemainingMinor != 2000 {
		t.Fatalf("expected remaining 2000, got %d", r.RemainingMinor)
	}
}

func TestBudgetWithSplitsAndFX(t *testing.T) {
	ctx := context.Background()

	catFood := "cat-food"
	catDrink := "cat-drink"
	accountID := "acc-1"
	cats := &memCategoryRepo{items: map[string]*domain.Category{
		catFood:  {ID: catFood, Name: "Food"},
		catDrink: {ID: catDrink, Name: "Drink"},
	}}
	accs := &memAccountRepo{items: map[string]*domain.Account{
		accountID: {ID: accountID, Name: "USD Card", Currency: "USD", Type: "card"},
	}}
	// Transaction in USD with splits
	txRepo := &memTransactionRepo{items: map[string]*domain.Transaction{
		"tx1": {ID: "tx1", AccountID: accountID, OccurredAt: time.Date(2025, 6, 10, 0, 0, 0, 0, time.UTC),
			AmountMinor: -1100, Currency: "USD"},
	}}
	splitRepo := &memSplitRepo{items: map[string][]domain.Split{
		"tx1": {
			{ID: "s1", TransactionID: "tx1", CategoryID: catFood, AmountMinor: -660},
			{ID: "s2", TransactionID: "tx1", CategoryID: catDrink, AmountMinor: -440},
		},
	}}
	budgetRepo := &memBudgetRepo{}

	// FX: 1 EUR = 1.10 USD
	fxRepo := &memFXRepo{rates: []domain.FXRate{
		{Date: "2025-06-10", Base: "EUR", Quote: "USD", Rate: 1.10},
	}}

	ledgerSvc := ledger.NewService(accs, cats, txRepo, splitRepo, clock)
	fxSvc := fx.NewService(fxRepo, &memECB{}, "EUR", clock)
	svc := NewService(budgetRepo, ledgerSvc, fxSvc, clock)

	// Set budgets
	svc.Upsert(ctx, UpsertInput{Month: "2025-06", CategoryID: catFood, PlannedMinor: 10000})
	svc.Upsert(ctx, UpsertInput{Month: "2025-06", CategoryID: catDrink, PlannedMinor: 5000})

	report, err := svc.GetReport(ctx, "2025-06")
	if err != nil {
		t.Fatalf("report: %v", err)
	}

	for _, r := range report {
		if r.CategoryID == catFood {
			// -660 USD / 1.10 = -600 EUR cents
			if r.SpentMinor != -600 {
				t.Fatalf("food spent: expected -600, got %d", r.SpentMinor)
			}
			if r.RemainingMinor != 9400 {
				t.Fatalf("food remaining: expected 9400, got %d", r.RemainingMinor)
			}
		}
		if r.CategoryID == catDrink {
			// -440 USD / 1.10 = -400 EUR cents
			if r.SpentMinor != -400 {
				t.Fatalf("drink spent: expected -400, got %d", r.SpentMinor)
			}
			if r.RemainingMinor != 4600 {
				t.Fatalf("drink remaining: expected 4600, got %d", r.RemainingMinor)
			}
		}
	}
}
