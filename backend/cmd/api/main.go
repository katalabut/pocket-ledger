package main

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

func main() {
	cfg, err := configloader.New[AppConfig]()
	if err != nil {
		panic(err)
	}

	fastapp.New(cfg.App).
		Add(&APIService{}).
		Start()
}
