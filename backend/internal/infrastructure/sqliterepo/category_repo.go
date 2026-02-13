package sqliterepo

import (
	"context"
	"database/sql"
	"time"

	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

type CategoryRepo struct{ db *sql.DB }

func NewCategoryRepo(db *sql.DB) *CategoryRepo { return &CategoryRepo{db: db} }

func (r *CategoryRepo) Create(ctx context.Context, c *domain.Category) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO categories(id, name, parent_id, created_at, updated_at) VALUES(?,?,?,?,?)`,
		c.ID, c.Name, c.ParentID,
		c.CreatedAt.Format(time.RFC3339Nano), c.UpdatedAt.Format(time.RFC3339Nano),
	)
	return err
}

func (r *CategoryRepo) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, parent_id, created_at, updated_at FROM categories WHERE id = ?`, id)
	return scanCategory(row)
}

func (r *CategoryRepo) List(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, parent_id, created_at, updated_at FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Category
	for rows.Next() {
		c, err := scanCategoryRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

func (r *CategoryRepo) Update(ctx context.Context, c *domain.Category) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE categories SET name=?, parent_id=?, updated_at=? WHERE id=?`,
		c.Name, c.ParentID, c.UpdatedAt.Format(time.RFC3339Nano), c.ID,
	)
	if err != nil {
		return err
	}
	return checkAffected(res)
}

func (r *CategoryRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return checkAffected(res)
}

func scanCategory(row *sql.Row) (*domain.Category, error) {
	var c domain.Category
	var parentID sql.NullString
	var createdAt, updatedAt string
	err := row.Scan(&c.ID, &c.Name, &parentID, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if parentID.Valid {
		c.ParentID = &parentID.String
	}
	c.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	c.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return &c, nil
}

func scanCategoryRows(rows *sql.Rows) (*domain.Category, error) {
	var c domain.Category
	var parentID sql.NullString
	var createdAt, updatedAt string
	err := rows.Scan(&c.ID, &c.Name, &parentID, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	if parentID.Valid {
		c.ParentID = &parentID.String
	}
	c.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	c.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return &c, nil
}
