package sqliterepo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/katalabut/pocket-ledger/backend/internal/domain"
)

type AccountRepo struct{ db *sql.DB }

func NewAccountRepo(db *sql.DB) *AccountRepo { return &AccountRepo{db: db} }

func (r *AccountRepo) Create(ctx context.Context, a *domain.Account) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO accounts(id, name, currency, type, initial_balance_minor, created_at, updated_at) VALUES(?,?,?,?,?,?,?)`,
		a.ID, a.Name, a.Currency, string(a.Type), a.InitialBalanceMinor,
		a.CreatedAt.Format(time.RFC3339Nano), a.UpdatedAt.Format(time.RFC3339Nano),
	)
	return err
}

func (r *AccountRepo) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, currency, type, initial_balance_minor, created_at, updated_at FROM accounts WHERE id = ?`, id)
	return scanAccount(row)
}

func (r *AccountRepo) List(ctx context.Context) ([]domain.Account, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, currency, type, initial_balance_minor, created_at, updated_at FROM accounts ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Account
	for rows.Next() {
		a, err := scanAccountRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	return out, rows.Err()
}

func (r *AccountRepo) Update(ctx context.Context, a *domain.Account) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE accounts SET name=?, currency=?, type=?, initial_balance_minor=?, updated_at=? WHERE id=?`,
		a.Name, a.Currency, string(a.Type), a.InitialBalanceMinor,
		a.UpdatedAt.Format(time.RFC3339Nano), a.ID,
	)
	if err != nil {
		return err
	}
	return checkAffected(res)
}

func (r *AccountRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM accounts WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return checkAffected(res)
}

func scanAccount(row *sql.Row) (*domain.Account, error) {
	var a domain.Account
	var typ, createdAt, updatedAt string
	err := row.Scan(&a.ID, &a.Name, &a.Currency, &typ, &a.InitialBalanceMinor, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	a.Type = domain.AccountType(typ)
	a.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	a.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return &a, nil
}

func scanAccountRows(rows *sql.Rows) (*domain.Account, error) {
	var a domain.Account
	var typ, createdAt, updatedAt string
	err := rows.Scan(&a.ID, &a.Name, &a.Currency, &typ, &a.InitialBalanceMinor, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	a.Type = domain.AccountType(typ)
	a.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	a.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return &a, nil
}

func checkAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
