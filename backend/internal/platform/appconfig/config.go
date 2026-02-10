package appconfig

import "github.com/katalabut/fast-app/config"

type Config struct {
	App      config.App `envPrefix:"APP_"`
	Database Database   `envPrefix:"DB_"`
}

type Database struct {
	Path string `env:"PATH" envDefault:"/data/pocket-ledger.db"`
}
