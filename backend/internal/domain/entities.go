package domain

import "time"

type AccountType string

const (
	AccountTypeCard    AccountType = "card"
	AccountTypeCash    AccountType = "cash"
	AccountTypeSavings AccountType = "savings"
	AccountTypeOther   AccountType = "other"
)

func ValidAccountType(s string) bool {
	switch AccountType(s) {
	case AccountTypeCard, AccountTypeCash, AccountTypeSavings, AccountTypeOther:
		return true
	}
	return false
}

type Account struct {
	ID                  string
	Name                string
	Currency            string // ISO 4217
	Type                AccountType
	InitialBalanceMinor int64
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type Category struct {
	ID        string
	Name      string
	ParentID  *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Transaction struct {
	ID          string
	AccountID   string
	OccurredAt  time.Time
	AmountMinor int64
	Currency    string
	Description string
	CategoryID  *string
	DedupeKey   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Split struct {
	ID            string
	TransactionID string
	CategoryID    string
	AmountMinor   int64
	CreatedAt     time.Time
}

// ValidateSplits checks that sum of splits equals the transaction amount.
func ValidateSplits(txAmount int64, splits []Split) error {
	if len(splits) == 0 {
		return nil
	}
	var sum int64
	for _, s := range splits {
		sum += s.AmountMinor
	}
	if sum != txAmount {
		return ErrSplitMismatch
	}
	return nil
}
