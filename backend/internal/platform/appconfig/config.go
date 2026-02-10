package appconfig

import "github.com/katalabut/fast-app/config"

type Config struct {
	App      config.App `envPrefix:"APP_"`
	HTTP     HTTP       `envPrefix:"HTTP_"`
	Database Database   `envPrefix:"DB_"`
	Auth     Auth       `envPrefix:"AUTH_"`
	SMTP     SMTP       `envPrefix:"SMTP_"`
}

type HTTP struct {
	Addr string `env:"ADDR" envDefault:":8080"`
}

type Database struct {
	Path string `env:"PATH" envDefault:"/data/pocket-ledger.db"`
}

type Auth struct {
	JWTSecret        string `env:"JWT_SECRET"`
	CodeTTLSeconds   int    `env:"CODE_TTL_SECONDS" envDefault:"600"`
	CodeLengthDigits int    `env:"CODE_LENGTH" envDefault:"6"`
	Issuer           string `env:"ISSUER" envDefault:"pocket-ledger"`
}

type SMTP struct {
	Host     string `env:"HOST"`
	Port     int    `env:"PORT" envDefault:"587"`
	Username string `env:"USERNAME"`
	Password string `env:"PASSWORD"`
	From     string `env:"FROM"`
}
