package budget

import (
	"context"

	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

type BudgetRepository interface {
	Upsert(ctx context.Context, b *domain.Budget) error
	ListByMonth(ctx context.Context, month string) ([]domain.Budget, error)
}
