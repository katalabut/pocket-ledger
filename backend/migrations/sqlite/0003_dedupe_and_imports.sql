-- +migrate Up
PRAGMA foreign_keys = ON;

ALTER TABLE transactions ADD COLUMN dedupe_key TEXT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_dedupe ON transactions(account_id, dedupe_key) WHERE dedupe_key IS NOT NULL;

CREATE TABLE IF NOT EXISTS import_profiles (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  account_id TEXT NOT NULL,
  separator TEXT NOT NULL DEFAULT ',',
  date_format TEXT NOT NULL DEFAULT '2006-01-02',
  column_mapping TEXT NOT NULL, -- JSON: {"date":0,"amount":1,"currency":2,"description":3,"external_id":4}
  amount_sign_flip INTEGER NOT NULL DEFAULT 0,
  skip_header_rows INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(account_id) REFERENCES accounts(id)
);

CREATE TABLE IF NOT EXISTS imports (
  id TEXT PRIMARY KEY,
  profile_id TEXT NOT NULL,
  account_id TEXT NOT NULL,
  filename TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending', -- pending, previewed, committed
  total_rows INTEGER NOT NULL DEFAULT 0,
  imported_rows INTEGER NOT NULL DEFAULT 0,
  skipped_rows INTEGER NOT NULL DEFAULT 0,
  error_rows INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  committed_at TEXT NULL,
  FOREIGN KEY(profile_id) REFERENCES import_profiles(id),
  FOREIGN KEY(account_id) REFERENCES accounts(id)
);

CREATE TABLE IF NOT EXISTS import_rows (
  id TEXT PRIMARY KEY,
  import_id TEXT NOT NULL,
  row_number INTEGER NOT NULL,
  raw_data TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending', -- pending, imported, skipped, error
  error_message TEXT NULL,
  transaction_id TEXT NULL,
  dedupe_key TEXT NULL,
  created_at TEXT NOT NULL,
  FOREIGN KEY(import_id) REFERENCES imports(id) ON DELETE CASCADE,
  FOREIGN KEY(transaction_id) REFERENCES transactions(id)
);

CREATE INDEX IF NOT EXISTS idx_import_rows_import ON import_rows(import_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_import_rows_import;
DROP TABLE IF EXISTS import_rows;
DROP TABLE IF EXISTS imports;
DROP TABLE IF EXISTS import_profiles;
DROP INDEX IF EXISTS idx_transactions_dedupe;
