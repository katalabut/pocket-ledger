package sqliterepo

import (
	"context"
	"database/sql"
	"time"

	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

type ImportRepo struct{ db *sql.DB }

func NewImportRepo(db *sql.DB) *ImportRepo { return &ImportRepo{db: db} }

func (r *ImportRepo) CreateImport(ctx context.Context, imp *domain.Import) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO imports(id, profile_id, account_id, filename, status, total_rows, imported_rows, skipped_rows, error_rows, created_at, committed_at)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
		imp.ID, imp.ProfileID, imp.AccountID, imp.Filename, imp.Status,
		imp.TotalRows, imp.ImportedRows, imp.SkippedRows, imp.ErrorRows,
		imp.CreatedAt.Format(time.RFC3339Nano), nil,
	)
	return err
}

func (r *ImportRepo) GetImport(ctx context.Context, id string) (*domain.Import, error) {
	var imp domain.Import
	var createdAt string
	var committedAt sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id, profile_id, account_id, filename, status, total_rows, imported_rows, skipped_rows, error_rows, created_at, committed_at
		 FROM imports WHERE id = ?`, id).
		Scan(&imp.ID, &imp.ProfileID, &imp.AccountID, &imp.Filename, &imp.Status,
			&imp.TotalRows, &imp.ImportedRows, &imp.SkippedRows, &imp.ErrorRows,
			&createdAt, &committedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	imp.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	if committedAt.Valid {
		t, _ := time.Parse(time.RFC3339Nano, committedAt.String)
		imp.CommittedAt = &t
	}
	return &imp, nil
}

func (r *ImportRepo) UpdateImport(ctx context.Context, imp *domain.Import) error {
	var committedAtStr *string
	if imp.CommittedAt != nil {
		s := imp.CommittedAt.Format(time.RFC3339Nano)
		committedAtStr = &s
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE imports SET status=?, total_rows=?, imported_rows=?, skipped_rows=?, error_rows=?, committed_at=? WHERE id=?`,
		imp.Status, imp.TotalRows, imp.ImportedRows, imp.SkippedRows, imp.ErrorRows, committedAtStr, imp.ID)
	return err
}

func (r *ImportRepo) CreateRows(ctx context.Context, rows []domain.ImportRow) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO import_rows(id, import_id, row_number, raw_data, status, error_message, transaction_id, dedupe_key, created_at) VALUES(?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, row := range rows {
		_, err := stmt.ExecContext(ctx,
			row.ID, row.ImportID, row.RowNumber, row.RawData, row.Status,
			row.ErrorMessage, row.TransactionID, row.DedupeKey,
			row.CreatedAt.Format(time.RFC3339Nano))
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (r *ImportRepo) GetRows(ctx context.Context, importID string) ([]domain.ImportRow, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, import_id, row_number, raw_data, status, error_message, transaction_id, dedupe_key, created_at
		 FROM import_rows WHERE import_id = ? ORDER BY row_number`, importID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ImportRow
	for rows.Next() {
		var row domain.ImportRow
		var errMsg, txID, dk sql.NullString
		var createdAt string
		if err := rows.Scan(&row.ID, &row.ImportID, &row.RowNumber, &row.RawData, &row.Status, &errMsg, &txID, &dk, &createdAt); err != nil {
			return nil, err
		}
		if errMsg.Valid {
			row.ErrorMessage = &errMsg.String
		}
		if txID.Valid {
			row.TransactionID = &txID.String
		}
		if dk.Valid {
			row.DedupeKey = &dk.String
		}
		row.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *ImportRepo) UpdateRow(ctx context.Context, row *domain.ImportRow) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE import_rows SET status=?, error_message=?, transaction_id=? WHERE id=?`,
		row.Status, row.ErrorMessage, row.TransactionID, row.ID)
	return err
}
