-- +migrate Up
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS auth_codes (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL,
  code_sha256 TEXT NOT NULL,
  expires_at TEXT NOT NULL,
  consumed_at TEXT NULL,
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_auth_codes_email_created ON auth_codes(email, created_at);

-- +migrate Down
DROP TABLE IF EXISTS auth_codes;
DROP TABLE IF EXISTS users;
