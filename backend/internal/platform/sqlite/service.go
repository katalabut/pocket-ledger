package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/katalabut/fast-app/health"
	_ "modernc.org/sqlite"
)

type Service struct {
	db *sql.DB
}

func New(path string) (*Service, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(time.Hour)

	return &Service{db: db}, nil
}

func (s *Service) DB() *sql.DB {
	return s.db
}

func (s *Service) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (s *Service) Shutdown(ctx context.Context) error {
	return s.db.Close()
}

func (s *Service) HealthChecks() []health.HealthChecker {
	return []health.HealthChecker{
		health.NewCustomCheck("sqlite", func(ctx context.Context) health.HealthResult {
			if err := s.db.PingContext(ctx); err != nil {
				return health.NewUnhealthyResult("sqlite ping failed").WithDetails("error", err.Error())
			}
			return health.NewHealthyResult("sqlite ok")
		}),
	}
}
