package sqliterepo

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

type FXRateRepo struct{ db *sql.DB }

func NewFXRateRepo(db *sql.DB) *FXRateRepo { return &FXRateRepo{db: db} }

func (r *FXRateRepo) Upsert(ctx context.Context, rate *domain.FXRate) error {
	if rate.ID == "" {
		rate.ID = uuid.NewString()
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO fx_rates(id, date, base, quote, rate, created_at) VALUES(?,?,?,?,?,?)
		 ON CONFLICT(date, base, quote) DO UPDATE SET rate=excluded.rate`,
		rate.ID, rate.Date, rate.Base, rate.Quote, rate.Rate,
		rate.CreatedAt.Format(time.RFC3339Nano),
	)
	return err
}

func (r *FXRateRepo) GetRate(ctx context.Context, date, base, quote string) (*domain.FXRate, error) {
	var rate domain.FXRate
	var createdAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, date, base, quote, rate, created_at FROM fx_rates WHERE date=? AND base=? AND quote=?`,
		date, base, quote).
		Scan(&rate.ID, &rate.Date, &rate.Base, &rate.Quote, &rate.Rate, &createdAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	rate.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	return &rate, nil
}

func (r *FXRateRepo) GetLatestRateBefore(ctx context.Context, date, base, quote string) (*domain.FXRate, error) {
	var rate domain.FXRate
	var createdAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, date, base, quote, rate, created_at FROM fx_rates WHERE date < ? AND base=? AND quote=? ORDER BY date DESC LIMIT 1`,
		date, base, quote).
		Scan(&rate.ID, &rate.Date, &rate.Base, &rate.Quote, &rate.Rate, &createdAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	rate.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	return &rate, nil
}

func (r *FXRateRepo) ListRatesByDate(ctx context.Context, date string) ([]domain.FXRate, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, date, base, quote, rate, created_at FROM fx_rates WHERE date=?`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.FXRate
	for rows.Next() {
		var rate domain.FXRate
		var createdAt string
		if err := rows.Scan(&rate.ID, &rate.Date, &rate.Base, &rate.Quote, &rate.Rate, &createdAt); err != nil {
			return nil, err
		}
		rate.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		out = append(out, rate)
	}
	return out, rows.Err()
}
