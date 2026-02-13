package importer

import (
	"context"

	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

type ImportProfileRepository interface {
	Create(ctx context.Context, p *domain.ImportProfile) error
	GetByID(ctx context.Context, id string) (*domain.ImportProfile, error)
	List(ctx context.Context) ([]domain.ImportProfile, error)
}

type ImportRepository interface {
	CreateImport(ctx context.Context, imp *domain.Import) error
	GetImport(ctx context.Context, id string) (*domain.Import, error)
	UpdateImport(ctx context.Context, imp *domain.Import) error
	CreateRows(ctx context.Context, rows []domain.ImportRow) error
	GetRows(ctx context.Context, importID string) ([]domain.ImportRow, error)
	UpdateRow(ctx context.Context, row *domain.ImportRow) error
}

type TransactionCreator interface {
	Create(ctx context.Context, t *domain.Transaction) error
	ExistsByDedupeKey(ctx context.Context, accountID, dedupeKey string) (bool, error)
}
