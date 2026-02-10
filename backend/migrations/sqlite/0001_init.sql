-- +migrate Up
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS accounts (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  currency TEXT NOT NULL,
  type TEXT NOT NULL,
  initial_balance_minor INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS categories (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  parent_id TEXT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(parent_id) REFERENCES categories(id)
);

CREATE TABLE IF NOT EXISTS transactions (
  id TEXT PRIMARY KEY,
  account_id TEXT NOT NULL,
  occurred_at TEXT NOT NULL,
  amount_minor INTEGER NOT NULL,
  currency TEXT NOT NULL,
  description TEXT NOT NULL,
  category_id TEXT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(account_id) REFERENCES accounts(id),
  FOREIGN KEY(category_id) REFERENCES categories(id)
);

CREATE TABLE IF NOT EXISTS transaction_splits (
  id TEXT PRIMARY KEY,
  transaction_id TEXT NOT NULL,
  category_id TEXT NOT NULL,
  amount_minor INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  FOREIGN KEY(transaction_id) REFERENCES transactions(id) ON DELETE CASCADE,
  FOREIGN KEY(category_id) REFERENCES categories(id)
);

CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_occurred_at ON transactions(occurred_at);
CREATE INDEX IF NOT EXISTS idx_splits_tx ON transaction_splits(transaction_id);

-- +migrate Down
DROP TABLE IF EXISTS transaction_splits;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS accounts;
