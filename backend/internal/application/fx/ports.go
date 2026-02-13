package fx

import (
	"context"

	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

type FXRateRepository interface {
	Upsert(ctx context.Context, rate *domain.FXRate) error
	GetRate(ctx context.Context, date, base, quote string) (*domain.FXRate, error)
	GetLatestRateBefore(ctx context.Context, date, base, quote string) (*domain.FXRate, error)
	ListRatesByDate(ctx context.Context, date string) ([]domain.FXRate, error)
}

type ECBClient interface {
	FetchDailyRates(ctx context.Context) ([]domain.FXRate, error)
}
