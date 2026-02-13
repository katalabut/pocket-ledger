package sqliterepo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

type SplitRepo struct{ db *sql.DB }

func NewSplitRepo(db *sql.DB) *SplitRepo { return &SplitRepo{db: db} }

func (r *SplitRepo) ReplaceSplits(ctx context.Context, txID string, splits []domain.Split) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM transaction_splits WHERE transaction_id = ?`, txID); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("delete old splits: %w", err)
	}
	for _, s := range splits {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO transaction_splits(id, transaction_id, category_id, amount_minor, created_at) VALUES(?,?,?,?,?)`,
			s.ID, s.TransactionID, s.CategoryID, s.AmountMinor,
			s.CreatedAt.Format(time.RFC3339Nano),
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert split: %w", err)
		}
	}
	return tx.Commit()
}

func (r *SplitRepo) GetByTransactionID(ctx context.Context, txID string) ([]domain.Split, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, transaction_id, category_id, amount_minor, created_at FROM transaction_splits WHERE transaction_id = ?`, txID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSplitRows(rows)
}

func (r *SplitRepo) GetByTransactionIDs(ctx context.Context, txIDs []string) (map[string][]domain.Split, error) {
	if len(txIDs) == 0 {
		return map[string][]domain.Split{}, nil
	}
	placeholders := strings.Repeat("?,", len(txIDs))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]any, len(txIDs))
	for i, id := range txIDs {
		args[i] = id
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, transaction_id, category_id, amount_minor, created_at FROM transaction_splits WHERE transaction_id IN (`+placeholders+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	splits, err := scanSplitRows(rows)
	if err != nil {
		return nil, err
	}
	out := map[string][]domain.Split{}
	for _, s := range splits {
		out[s.TransactionID] = append(out[s.TransactionID], s)
	}
	return out, nil
}

func scanSplitRows(rows *sql.Rows) ([]domain.Split, error) {
	var out []domain.Split
	for rows.Next() {
		var s domain.Split
		var createdAt string
		if err := rows.Scan(&s.ID, &s.TransactionID, &s.CategoryID, &s.AmountMinor, &createdAt); err != nil {
			return nil, err
		}
		s.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		out = append(out, s)
	}
	return out, rows.Err()
}
