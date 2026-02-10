package sqliterepo

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AuthCodeRepo struct{ db *sql.DB }

func NewAuthCodeRepo(db *sql.DB) *AuthCodeRepo { return &AuthCodeRepo{db: db} }

func (r *AuthCodeRepo) CreateCode(ctx context.Context, email, codeSha256 string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO auth_codes(id, email, code_sha256, expires_at, consumed_at, created_at) VALUES(?, ?, ?, ?, NULL, ?)` ,
		uuid.NewString(), email, codeSha256, expiresAt.UTC().Format(time.RFC3339Nano), time.Now().UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (r *AuthCodeRepo) ConsumeValidCode(ctx context.Context, email, codeSha256 string, now time.Time) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE auth_codes
		 SET consumed_at = ?
		 WHERE id = (
			SELECT id FROM auth_codes
			 WHERE email = ? AND code_sha256 = ? AND consumed_at IS NULL AND expires_at > ?
			 ORDER BY created_at DESC
			 LIMIT 1
		 )`,
		now.UTC().Format(time.RFC3339Nano), email, codeSha256, now.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return false, fmt.Errorf("consume code: %w", err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("rows affected: %w", err)
	}
	return aff > 0, nil
}
