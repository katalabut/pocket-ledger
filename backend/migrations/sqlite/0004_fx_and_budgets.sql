-- +migrate Up
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS fx_rates (
  id TEXT PRIMARY KEY,
  date TEXT NOT NULL,
  base TEXT NOT NULL,
  quote TEXT NOT NULL,
  rate REAL NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(date, base, quote)
);

CREATE INDEX IF NOT EXISTS idx_fx_rates_date ON fx_rates(date);
CREATE INDEX IF NOT EXISTS idx_fx_rates_quote_date ON fx_rates(quote, date);

CREATE TABLE IF NOT EXISTS budgets (
  id TEXT PRIMARY KEY,
  month TEXT NOT NULL,
  category_id TEXT NOT NULL,
  planned_amount_minor_base INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(category_id) REFERENCES categories(id),
  UNIQUE(month, category_id)
);

-- +migrate Down
DROP TABLE IF EXISTS budgets;
DROP INDEX IF EXISTS idx_fx_rates_quote_date;
DROP INDEX IF EXISTS idx_fx_rates_date;
DROP TABLE IF EXISTS fx_rates;
