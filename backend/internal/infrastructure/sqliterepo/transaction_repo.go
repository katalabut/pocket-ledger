package sqliterepo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/katalabut/pocket-ledger/backend/internal/application/ledger"
	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

type TransactionRepo struct{ db *sql.DB }

func NewTransactionRepo(db *sql.DB) *TransactionRepo { return &TransactionRepo{db: db} }

func (r *TransactionRepo) Create(ctx context.Context, t *domain.Transaction) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO transactions(id, account_id, occurred_at, amount_minor, currency, description, category_id, dedupe_key, created_at, updated_at)
		 VALUES(?,?,?,?,?,?,?,?,?,?)`,
		t.ID, t.AccountID, t.OccurredAt.Format(time.RFC3339Nano), t.AmountMinor,
		t.Currency, t.Description, t.CategoryID, t.DedupeKey,
		t.CreatedAt.Format(time.RFC3339Nano), t.UpdatedAt.Format(time.RFC3339Nano),
	)
	return err
}

func (r *TransactionRepo) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, account_id, occurred_at, amount_minor, currency, description, category_id, dedupe_key, created_at, updated_at
		 FROM transactions WHERE id = ?`, id)
	return scanTransaction(row)
}

func (r *TransactionRepo) List(ctx context.Context, f ledger.TransactionFilter) ([]domain.Transaction, int, error) {
	where := []string{"1=1"}
	args := []any{}

	if f.AccountID != nil && *f.AccountID != "" {
		where = append(where, "account_id = ?")
		args = append(args, *f.AccountID)
	}
	if f.CategoryID != nil && *f.CategoryID != "" {
		where = append(where, "(category_id = ? OR id IN (SELECT transaction_id FROM transaction_splits WHERE category_id = ?))")
		args = append(args, *f.CategoryID, *f.CategoryID)
	}
	if f.From != nil && *f.From != "" {
		where = append(where, "occurred_at >= ?")
		args = append(args, *f.From)
	}
	if f.To != nil && *f.To != "" {
		where = append(where, "occurred_at <= ?")
		args = append(args, *f.To)
	}
	if f.Query != nil && *f.Query != "" {
		where = append(where, "description LIKE ?")
		args = append(args, "%"+*f.Query+"%")
	}

	w := strings.Join(where, " AND ")

	// count
	var total int
	countQ := "SELECT COUNT(1) FROM transactions WHERE " + w
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count: %w", err)
	}

	q := "SELECT id, account_id, occurred_at, amount_minor, currency, description, category_id, dedupe_key, created_at, updated_at FROM transactions WHERE " + w + " ORDER BY occurred_at DESC LIMIT ? OFFSET ?"
	args = append(args, f.Limit, f.Offset)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []domain.Transaction
	for rows.Next() {
		t, err := scanTransactionRows(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *t)
	}
	return out, total, rows.Err()
}

func (r *TransactionRepo) Update(ctx context.Context, t *domain.Transaction) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE transactions SET account_id=?, occurred_at=?, amount_minor=?, currency=?, description=?, category_id=?, updated_at=? WHERE id=?`,
		t.AccountID, t.OccurredAt.Format(time.RFC3339Nano), t.AmountMinor,
		t.Currency, t.Description, t.CategoryID,
		t.UpdatedAt.Format(time.RFC3339Nano), t.ID,
	)
	if err != nil {
		return err
	}
	return checkAffected(res)
}

func (r *TransactionRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM transactions WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return checkAffected(res)
}

func (r *TransactionRepo) ExistsByDedupeKey(ctx context.Context, accountID, dedupeKey string) (bool, error) {
	var cnt int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM transactions WHERE account_id = ? AND dedupe_key = ?`,
		accountID, dedupeKey).Scan(&cnt)
	return cnt > 0, err
}

func scanTransaction(row *sql.Row) (*domain.Transaction, error) {
	var t domain.Transaction
	var catID, dedupeKey sql.NullString
	var occurredAt, createdAt, updatedAt string
	err := row.Scan(&t.ID, &t.AccountID, &occurredAt, &t.AmountMinor, &t.Currency, &t.Description, &catID, &dedupeKey, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if catID.Valid {
		t.CategoryID = &catID.String
	}
	if dedupeKey.Valid {
		t.DedupeKey = &dedupeKey.String
	}
	t.OccurredAt, _ = time.Parse(time.RFC3339Nano, occurredAt)
	t.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	t.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return &t, nil
}

func scanTransactionRows(rows *sql.Rows) (*domain.Transaction, error) {
	var t domain.Transaction
	var catID, dedupeKey sql.NullString
	var occurredAt, createdAt, updatedAt string
	err := rows.Scan(&t.ID, &t.AccountID, &occurredAt, &t.AmountMinor, &t.Currency, &t.Description, &catID, &dedupeKey, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	if catID.Valid {
		t.CategoryID = &catID.String
	}
	if dedupeKey.Valid {
		t.DedupeKey = &dedupeKey.String
	}
	t.OccurredAt, _ = time.Parse(time.RFC3339Nano, occurredAt)
	t.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	t.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return &t, nil
}
