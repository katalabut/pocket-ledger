package domain

import "time"

type ImportProfile struct {
	ID             string
	Name           string
	AccountID      string
	Separator      string
	DateFormat     string
	ColumnMapping  map[string]int // "date":0, "amount":1, "currency":2, "description":3, "external_id":4
	AmountSignFlip bool
	SkipHeaderRows int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Import struct {
	ID           string
	ProfileID    string
	AccountID    string
	Filename     string
	Status       string // pending, previewed, committed
	TotalRows    int
	ImportedRows int
	SkippedRows  int
	ErrorRows    int
	CreatedAt    time.Time
	CommittedAt  *time.Time
}

type ImportRow struct {
	ID            string
	ImportID      string
	RowNumber     int
	RawData       string
	Status        string // pending, imported, skipped, error
	ErrorMessage  *string
	TransactionID *string
	DedupeKey     *string
	CreatedAt     time.Time
}
