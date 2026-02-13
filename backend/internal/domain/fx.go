package domain

import "time"

type FXRate struct {
	ID        string
	Date      string // YYYY-MM-DD
	Base      string // ISO 4217
	Quote     string // ISO 4217
	Rate      float64
	CreatedAt time.Time
}

type Budget struct {
	ID                   string
	Month                string // YYYY-MM
	CategoryID           string
	PlannedAmountMinorBase int64
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
