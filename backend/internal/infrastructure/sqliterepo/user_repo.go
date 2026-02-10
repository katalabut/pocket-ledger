package sqliterepo

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type UserRepo struct{ db *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) EnsureUser(ctx context.Context, email string) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `SELECT id FROM users WHERE email = ?`, email).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}
	id = uuid.NewString()
	_, err = r.db.ExecContext(ctx, `INSERT INTO users(id, email, created_at) VALUES(?, ?, ?)` , id, email, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		// Race: user created concurrently
		err2 := r.db.QueryRowContext(ctx, `SELECT id FROM users WHERE email = ?`, email).Scan(&id)
		if err2 == nil {
			return id, nil
		}
		return "", err
	}
	return id, nil
}
