package app

import (
	"fmt"

	fastapp "github.com/katalabut/fast-app"
	"github.com/katalabut/fast-app/configloader"
	"github.com/katalabut/pocket-ledger/backend/internal/platform/appconfig"
	"github.com/katalabut/pocket-ledger/backend/internal/platform/sqlite"
)

func New() (*fastapp.App, error) {
	cfg, err := configloader.New[appconfig.Config]()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	sqliteSvc, err := sqlite.New(cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("init sqlite: %w", err)
	}

	a := fastapp.New(cfg.App).
		Add(sqliteSvc)

	return a, nil
}
