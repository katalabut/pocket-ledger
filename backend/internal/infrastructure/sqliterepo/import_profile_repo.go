package sqliterepo

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

type ImportProfileRepo struct{ db *sql.DB }

func NewImportProfileRepo(db *sql.DB) *ImportProfileRepo { return &ImportProfileRepo{db: db} }

func (r *ImportProfileRepo) Create(ctx context.Context, p *domain.ImportProfile) error {
	mappingJSON, _ := json.Marshal(p.ColumnMapping)
	flip := 0
	if p.AmountSignFlip {
		flip = 1
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO import_profiles(id, name, account_id, separator, date_format, column_mapping, amount_sign_flip, skip_header_rows, created_at, updated_at)
		 VALUES(?,?,?,?,?,?,?,?,?,?)`,
		p.ID, p.Name, p.AccountID, p.Separator, p.DateFormat, string(mappingJSON), flip, p.SkipHeaderRows,
		p.CreatedAt.Format(time.RFC3339Nano), p.UpdatedAt.Format(time.RFC3339Nano),
	)
	return err
}

func (r *ImportProfileRepo) GetByID(ctx context.Context, id string) (*domain.ImportProfile, error) {
	var p domain.ImportProfile
	var mappingJSON, createdAt, updatedAt string
	var flip int
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, account_id, separator, date_format, column_mapping, amount_sign_flip, skip_header_rows, created_at, updated_at
		 FROM import_profiles WHERE id = ?`, id).
		Scan(&p.ID, &p.Name, &p.AccountID, &p.Separator, &p.DateFormat, &mappingJSON, &flip, &p.SkipHeaderRows, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(mappingJSON), &p.ColumnMapping)
	p.AmountSignFlip = flip != 0
	p.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	p.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return &p, nil
}

func (r *ImportProfileRepo) List(ctx context.Context) ([]domain.ImportProfile, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, account_id, separator, date_format, column_mapping, amount_sign_flip, skip_header_rows, created_at, updated_at
		 FROM import_profiles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ImportProfile
	for rows.Next() {
		var p domain.ImportProfile
		var mappingJSON, createdAt, updatedAt string
		var flip int
		if err := rows.Scan(&p.ID, &p.Name, &p.AccountID, &p.Separator, &p.DateFormat, &mappingJSON, &flip, &p.SkipHeaderRows, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(mappingJSON), &p.ColumnMapping)
		p.AmountSignFlip = flip != 0
		p.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		p.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
		out = append(out, p)
	}
	return out, rows.Err()
}
