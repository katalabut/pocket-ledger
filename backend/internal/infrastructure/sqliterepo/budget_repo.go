package sqliterepo

import (
	"context"
	"database/sql"
	"time"

	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

type BudgetRepo struct{ db *sql.DB }

func NewBudgetRepo(db *sql.DB) *BudgetRepo { return &BudgetRepo{db: db} }

func (r *BudgetRepo) Upsert(ctx context.Context, b *domain.Budget) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO budgets(id, month, category_id, planned_amount_minor_base, created_at, updated_at) VALUES(?,?,?,?,?,?)
		 ON CONFLICT(month, category_id) DO UPDATE SET planned_amount_minor_base=excluded.planned_amount_minor_base, updated_at=excluded.updated_at`,
		b.ID, b.Month, b.CategoryID, b.PlannedAmountMinorBase,
		b.CreatedAt.Format(time.RFC3339Nano), b.UpdatedAt.Format(time.RFC3339Nano),
	)
	return err
}

func (r *BudgetRepo) ListByMonth(ctx context.Context, month string) ([]domain.Budget, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, month, category_id, planned_amount_minor_base, created_at, updated_at FROM budgets WHERE month = ?`, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Budget
	for rows.Next() {
		var b domain.Budget
		var createdAt, updatedAt string
		if err := rows.Scan(&b.ID, &b.Month, &b.CategoryID, &b.PlannedAmountMinorBase, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		b.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		b.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
		out = append(out, b)
	}
	return out, rows.Err()
}
