package appconfig

import "github.com/katalabut/fast-app/config"

type Config struct {
	App      config.App `envPrefix:"APP_"`
	HTTP     HTTP       `envPrefix:"HTTP_"`
	Database Database   `envPrefix:"DB_"`
}

type HTTP struct {
	Addr string `env:"ADDR" envDefault:":8080"`
}

type Database struct {
	Path string `env:"PATH" envDefault:"/data/pocket-ledger.db"`
}
