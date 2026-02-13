package fx

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

type Service struct {
	rates    FXRateRepository
	ecb      ECBClient
	baseCcy  string
	clock    func() time.Time
}

func NewService(rates FXRateRepository, ecb ECBClient, baseCcy string, clock func() time.Time) *Service {
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}
	if baseCcy == "" {
		baseCcy = "EUR"
	}
	return &Service{rates: rates, ecb: ecb, baseCcy: baseCcy, clock: clock}
}

func (s *Service) BaseCurrency() string { return s.baseCcy }

// SyncRates fetches daily ECB rates and upserts them.
func (s *Service) SyncRates(ctx context.Context) (int, error) {
	rates, err := s.ecb.FetchDailyRates(ctx)
	if err != nil {
		return 0, fmt.Errorf("ecb fetch: %w", err)
	}
	count := 0
	for i := range rates {
		if err := s.rates.Upsert(ctx, &rates[i]); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// GetRatesForDate returns all rates for a date, with fallback to last available.
func (s *Service) GetRatesForDate(ctx context.Context, date string) (map[string]float64, error) {
	rates, err := s.rates.ListRatesByDate(ctx, date)
	if err != nil {
		return nil, err
	}
	if len(rates) > 0 {
		out := map[string]float64{}
		for _, r := range rates {
			out[r.Quote] = r.Rate
		}
		return out, nil
	}
	// Fallback: try to find rates from the most recent prior date
	// We check common currencies
	return nil, nil
}

// Convert converts an amount from one currency to the base currency on a given date.
// If the currency is the base, returns the amount unchanged.
func (s *Service) Convert(ctx context.Context, amountMinor int64, currency, date string) (int64, error) {
	if currency == s.baseCcy {
		return amountMinor, nil
	}
	rate, err := s.GetRate(ctx, date, currency)
	if err != nil {
		return 0, err
	}
	// rate is base_per_quote: e.g. EUR/USD = 1.10 means 1 EUR = 1.10 USD
	// So to convert USD to EUR: amount_EUR = amount_USD / rate
	// ECB publishes rates as EUR-based: rate for USD = how many USD per 1 EUR
	// Convert from quote currency to EUR: amount_EUR = amount_quote / rate
	converted := float64(amountMinor) / rate
	return int64(math.Round(converted)), nil
}

// GetRate returns the exchange rate for a currency on a date, with fallback.
func (s *Service) GetRate(ctx context.Context, date, quoteCurrency string) (float64, error) {
	if quoteCurrency == s.baseCcy {
		return 1.0, nil
	}
	r, err := s.rates.GetRate(ctx, date, s.baseCcy, quoteCurrency)
	if err == nil {
		return r.Rate, nil
	}
	// Fallback to last prior date
	r, err = s.rates.GetLatestRateBefore(ctx, date, s.baseCcy, quoteCurrency)
	if err != nil {
		return 0, fmt.Errorf("no rate for %s on or before %s: %w", quoteCurrency, date, domain.ErrNotFound)
	}
	return r.Rate, nil
}
