package migrator

import (
	"bufio"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Migrator struct {
	db  *sql.DB
	dir string
}

func NewMigrator(db *sql.DB, dir string) *Migrator {
	return &Migrator{db: db, dir: dir}
}

func (m *Migrator) Up(ctx context.Context) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	files, err := filepath.Glob(filepath.Join(m.dir, "*.sql"))
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(files)
	for _, f := range files {
		name := filepath.Base(f)
		applied, err := m.isApplied(ctx, name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		upSQL, hash, err := readUpSection(f)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if strings.TrimSpace(upSQL) == "" {
			return fmt.Errorf("migration %s: empty Up section", name)
		}

		tx, err := m.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		if _, err := tx.ExecContext(ctx, upSQL); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("exec migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO schema_migrations(name, sha256, applied_at) VALUES(?, ?, ?)`,
			name, hash, time.Now().UTC().Format(time.RFC3339Nano),
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}
	return nil
}

func (m *Migrator) ensureTable(ctx context.Context) error {
	_, err := m.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations(
		name TEXT PRIMARY KEY,
		sha256 TEXT NOT NULL,
		applied_at TEXT NOT NULL
	)`)
	return err
}

func (m *Migrator) isApplied(ctx context.Context, name string) (bool, error) {
	var cnt int
	if err := m.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM schema_migrations WHERE name = ?`, name).Scan(&cnt); err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func readUpSection(path string) (string, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	up, err := extractSection(f, "-- +migrate Up", "-- +migrate Down")
	if err != nil {
		return "", "", err
	}
	h := sha256.Sum256([]byte(up))
	return up, hex.EncodeToString(h[:]), nil
}

func extractSection(r io.Reader, start, end string) (string, error) {
	s := bufio.NewScanner(r)
	var b strings.Builder
	in := false
	for s.Scan() {
		line := s.Text()
		if strings.TrimSpace(line) == start {
			in = true
			continue
		}
		if strings.TrimSpace(line) == end {
			break
		}
		if in {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	return b.String(), s.Err()
}
