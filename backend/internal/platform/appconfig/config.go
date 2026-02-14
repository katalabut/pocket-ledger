package appconfig

import (
	"github.com/katalabut/fast-app/config"
	"github.com/katalabut/fast-app/configloader"
)

type Config struct {
	App      config.App `envPrefix:"APP_"`
	HTTP     HTTP       `envPrefix:"HTTP_"`
	Database Database   `envPrefix:"DB_"`
	Auth     Auth       `envPrefix:"AUTH_"`
	Email    Email      `envPrefix:"EMAIL_"`
	SMTP     SMTP       `envPrefix:"SMTP_"`
	FX       FX         `envPrefix:"FX_"`
}

type FX struct {
	BaseCurrency string `env:"BASE_CURRENCY" envDefault:"EUR"`
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

// Email controls delivery mode: "smtp" (default) or "log".
// In "log" mode login codes are written to stdout instead of sending via SMTP.
type Email struct {
	Mode string `env:"MODE" envDefault:"smtp"`
}

type SMTP struct {
	Host     string `env:"HOST"`
	Port     int    `env:"PORT" envDefault:"587"`
	Username string `env:"USERNAME"`
	Password string `env:"PASSWORD"`
	From     string `env:"FROM"`
	TestMode bool   `env:"TEST_MODE" envDefault:"false"`
}

func Load() (*Config, error) {
	return configloader.New[Config](configloader.WithFileFromEnv("config.yaml"))
}
