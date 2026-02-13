package ledger

import (
	"context"

	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

type AccountRepository interface {
	Create(ctx context.Context, a *domain.Account) error
	GetByID(ctx context.Context, id string) (*domain.Account, error)
	List(ctx context.Context) ([]domain.Account, error)
	Update(ctx context.Context, a *domain.Account) error
	Delete(ctx context.Context, id string) error
}

type CategoryRepository interface {
	Create(ctx context.Context, c *domain.Category) error
	GetByID(ctx context.Context, id string) (*domain.Category, error)
	List(ctx context.Context) ([]domain.Category, error)
	Update(ctx context.Context, c *domain.Category) error
	Delete(ctx context.Context, id string) error
}

type TransactionFilter struct {
	AccountID  *string
	CategoryID *string
	From       *string // ISO date
	To         *string // ISO date
	Query      *string // search description
	Limit      int
	Offset     int
}

type TransactionRepository interface {
	Create(ctx context.Context, t *domain.Transaction) error
	GetByID(ctx context.Context, id string) (*domain.Transaction, error)
	List(ctx context.Context, f TransactionFilter) ([]domain.Transaction, int, error) // items, total, err
	Update(ctx context.Context, t *domain.Transaction) error
	Delete(ctx context.Context, id string) error
	ExistsByDedupeKey(ctx context.Context, accountID, dedupeKey string) (bool, error)
}

type SplitRepository interface {
	ReplaceSplits(ctx context.Context, txID string, splits []domain.Split) error
	GetByTransactionID(ctx context.Context, txID string) ([]domain.Split, error)
	GetByTransactionIDs(ctx context.Context, txIDs []string) (map[string][]domain.Split, error)
}
