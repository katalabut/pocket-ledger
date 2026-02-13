package fx

import (
	"context"
	"testing"
	"time"

	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

type memFXRepo struct {
	rates []domain.FXRate
}

func (r *memFXRepo) Upsert(_ context.Context, rate *domain.FXRate) error {
	for i, existing := range r.rates {
		if existing.Date == rate.Date && existing.Base == rate.Base && existing.Quote == rate.Quote {
			r.rates[i] = *rate
			return nil
		}
	}
	r.rates = append(r.rates, *rate)
	return nil
}

func (r *memFXRepo) GetRate(_ context.Context, date, base, quote string) (*domain.FXRate, error) {
	for _, rate := range r.rates {
		if rate.Date == date && rate.Base == base && rate.Quote == quote {
			return &rate, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *memFXRepo) GetLatestRateBefore(_ context.Context, date, base, quote string) (*domain.FXRate, error) {
	var best *domain.FXRate
	for i, rate := range r.rates {
		if rate.Base == base && rate.Quote == quote && rate.Date < date {
			if best == nil || rate.Date > best.Date {
				best = &r.rates[i]
			}
		}
	}
	if best == nil {
		return nil, domain.ErrNotFound
	}
	return best, nil
}

func (r *memFXRepo) ListRatesByDate(_ context.Context, date string) ([]domain.FXRate, error) {
	var out []domain.FXRate
	for _, rate := range r.rates {
		if rate.Date == date {
			out = append(out, rate)
		}
	}
	return out, nil
}

type memECB struct {
	rates []domain.FXRate
}

func (m *memECB) FetchDailyRates(_ context.Context) ([]domain.FXRate, error) {
	return m.rates, nil
}

func TestConvert_SameCurrency(t *testing.T) {
	svc := NewService(&memFXRepo{}, &memECB{}, "EUR", nil)
	got, err := svc.Convert(context.Background(), -1000, "EUR", "2025-01-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != -1000 {
		t.Fatalf("expected -1000, got %d", got)
	}
}

func TestConvert_WithRate(t *testing.T) {
	repo := &memFXRepo{rates: []domain.FXRate{
		{Date: "2025-01-15", Base: "EUR", Quote: "USD", Rate: 1.10},
	}}
	svc := NewService(repo, &memECB{}, "EUR", nil)
	// 1100 USD cents → EUR: 1100 / 1.10 = 1000 EUR cents
	got, err := svc.Convert(context.Background(), 1100, "USD", "2025-01-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 1000 {
		t.Fatalf("expected 1000, got %d", got)
	}
}

func TestConvert_FallbackToPriorDate(t *testing.T) {
	repo := &memFXRepo{rates: []domain.FXRate{
		{Date: "2025-01-13", Base: "EUR", Quote: "USD", Rate: 1.10}, // Friday
		// No rate on 2025-01-14 (Sat), 2025-01-15 (Sun)
	}}
	svc := NewService(repo, &memECB{}, "EUR", nil)
	// Should fallback to 2025-01-13 rate
	got, err := svc.Convert(context.Background(), 1100, "USD", "2025-01-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 1000 {
		t.Fatalf("expected 1000, got %d", got)
	}
}

func TestConvert_NoRate(t *testing.T) {
	svc := NewService(&memFXRepo{}, &memECB{}, "EUR", nil)
	_, err := svc.Convert(context.Background(), 1000, "USD", "2025-01-15")
	if err == nil {
		t.Fatal("expected error for missing rate")
	}
}

func TestSyncRates(t *testing.T) {
	now := time.Date(2025, 1, 15, 14, 0, 0, 0, time.UTC)
	repo := &memFXRepo{}
	ecb := &memECB{rates: []domain.FXRate{
		{Date: "2025-01-15", Base: "EUR", Quote: "USD", Rate: 1.10},
		{Date: "2025-01-15", Base: "EUR", Quote: "GBP", Rate: 0.85},
	}}
	svc := NewService(repo, ecb, "EUR", func() time.Time { return now })
	count, err := svc.SyncRates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 synced, got %d", count)
	}
	if len(repo.rates) != 2 {
		t.Fatalf("expected 2 rates in repo, got %d", len(repo.rates))
	}
}
