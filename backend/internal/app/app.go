package app

import (
	"context"

	fastapp "github.com/katalabut/fast-app"
	"github.com/katalabut/fast-app/config"
	"github.com/katalabut/fast-app/configloader"
)

type AppConfig struct {
	App config.App
}

type APIService struct{}

func (s *APIService) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (s *APIService) Shutdown(ctx context.Context) error {
	return nil
}

func New() (*fastapp.App, error) {
	cfg, err := configloader.New[AppConfig]()
	if err != nil {
		return nil, err
	}

	a := fastapp.New(cfg.App).
		Add(&APIService{})

	return a, nil
}
