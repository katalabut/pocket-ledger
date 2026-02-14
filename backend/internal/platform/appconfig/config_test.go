package appconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FromConfigFileEnv(t *testing.T) {
	t.Setenv("CONFIG_FILE", filepath.Join(t.TempDir(), "config.yaml"))
	configPath := os.Getenv("CONFIG_FILE")

	content := `http:
  addr: ":9191"
database:
  path: "./from-file.db"
auth:
  jwtsecret: "file-secret"
smtp:
  testmode: true
fx:
  basecurrency: "USD"
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.HTTP.Addr != ":9191" {
		t.Fatalf("HTTP.Addr = %q, want :9191", cfg.HTTP.Addr)
	}
	if cfg.Database.Path != "./from-file.db" {
		t.Fatalf("Database.Path = %q, want ./from-file.db", cfg.Database.Path)
	}
	if cfg.Auth.JWTSecret != "file-secret" {
		t.Fatalf("Auth.JWTSecret = %q, want file-secret", cfg.Auth.JWTSecret)
	}
	if !cfg.SMTP.TestMode {
		t.Fatal("SMTP.TestMode = false, want true")
	}
	if cfg.FX.BaseCurrency != "USD" {
		t.Fatalf("FX.BaseCurrency = %q, want USD", cfg.FX.BaseCurrency)
	}
}
