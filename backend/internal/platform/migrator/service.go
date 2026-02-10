package migrator

import (
	"context"
	"database/sql"
	"fmt"
)

type Service struct {
	db  *sql.DB
	dir string
}

func New(db *sql.DB, dir string) *Service {
	return &Service{db: db, dir: dir}
}

func (s *Service) Run(ctx context.Context) error {
	m := NewMigrator(s.db, s.dir)
	if err := m.Up(ctx); err != nil {
		return fmt.Errorf("migrations up: %w", err)
	}
	<-ctx.Done()
	return nil
}

func (s *Service) Shutdown(ctx context.Context) error { return nil }
