# pocket-ledger

Self-hosted personal/family budget service with CSV import, envelope budgeting, and multi-currency reporting (ECB rates).

## Features

- **Accounts & Transactions** — CRUD for accounts (card/cash/savings), transactions with filters and search
- **Categories & Splits** — Hierarchical categories, split a transaction across multiple categories with sum validation
- **CSV Import** — Configurable import profiles per bank, upload/preview/commit workflow, SHA256-based deduplication
- **Multi-currency (ECB)** — Daily FX rates from ECB, cached in DB, fallback to last available rate for weekends/holidays
- **Budgets** — Monthly budgets per category with planned/spent/remaining, FX-aware spent calculation
- **Reports** — Spending by category (base currency), account balances (native + base equivalent)
- **Auth** — Email magic code login with JWT tokens

## Stack

- **Backend:** Go 1.23+, [fast-app](https://github.com/katalabut/fast-app), SQLite (modernc.org/sqlite)
- **Frontend:** React 19, TypeScript, Vite, Tailwind CSS v4
- **Infra:** Docker Compose

## Quick Start

### Docker Compose (recommended)

```bash
# Clone and configure
cp .env.example .env
# Edit .env with your SMTP and JWT settings

# Start
docker compose up --build -d

# Access: http://localhost:3000
```

### Local Development

```bash
# Backend
cd backend
go mod download

# Option A: config file (recommended for local dev)
cp config.yaml.example config.yaml
# Edit config.yaml — set auth.jwtsecret at minimum.
# email.mode=log prints login codes to stdout (no SMTP needed).
go run ./cmd/api

# Option B: environment variables only
export AUTH_JWT_SECRET=dev-secret
export EMAIL_MODE=log            # print login codes to stdout
export DB_PATH=./dev.db
go run ./cmd/api

# Option C: config file at custom path
export CONFIG_FILE=/path/to/my-config.yaml
go run ./cmd/api

# Env vars always override config file values.

# Frontend (separate terminal)
cd frontend
npm ci
npm run dev
# Access: http://localhost:5173 (proxies API to :8080)
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `CONFIG_FILE` | `config.yaml` | Config file path (yaml/json/toml); env vars override file values |
| `APP_NAME` | pocket-ledger | Application name |
| `HTTP_ADDR` | :8080 | HTTP listen address |
| `DB_PATH` | /data/pocket-ledger.db | SQLite database path |
| `AUTH_JWT_SECRET` | **required** | JWT signing secret |
| `AUTH_CODE_TTL_SECONDS` | 600 | Login code expiry |
| `AUTH_ISSUER` | pocket-ledger | JWT issuer |
| `EMAIL_MODE` | smtp | Legacy switch: `smtp` = send via SMTP; `log` = print login codes to stdout |
| `SMTP_TEST_MODE` | false | Preferred test switch: if true, log outgoing emails instead of sending |
| `SMTP_HOST` | required unless `EMAIL_MODE=log` or `SMTP_TEST_MODE=true` | SMTP server host |
| `SMTP_PORT` | 587 | SMTP server port |
| `SMTP_USERNAME` | required unless `EMAIL_MODE=log` or `SMTP_TEST_MODE=true` | SMTP username |
| `SMTP_PASSWORD` | required unless `EMAIL_MODE=log` or `SMTP_TEST_MODE=true` | SMTP password |
| `SMTP_FROM` | required unless `EMAIL_MODE=log` or `SMTP_TEST_MODE=true` | From email address |
| `FX_BASE_CURRENCY` | EUR | Base currency for reports/budgets |

## API Endpoints

### Auth
- `POST /auth/request_code` — Send login code to email
- `POST /auth/confirm_code` — Verify code and get JWT token

### Accounts
- `GET /api/accounts` — List accounts
- `POST /api/accounts` — Create account
- `GET/PATCH/DELETE /api/accounts/{id}`

### Categories
- `GET /api/categories` — List categories
- `POST /api/categories` — Create category
- `GET/PATCH/DELETE /api/categories/{id}`

### Transactions
- `GET /api/transactions?account_id=&category_id=&from=&to=&q=&limit=&offset=`
- `POST /api/transactions` — Create transaction
- `GET/PATCH/DELETE /api/transactions/{id}`
- `GET/POST /api/transactions/{id}/splits` — View/replace splits

### Import
- `GET/POST /api/import-profiles` — List/create import profiles
- `POST /api/imports/upload` — Upload CSV (multipart, profile_id + file)
- `GET /api/imports/{id}/preview` — Preview parsed rows
- `POST /api/imports/{id}/commit` — Commit import

### FX
- `GET /api/fx/rates?date=` — Get rates for a date
- `POST /api/fx/sync` — Fetch latest ECB rates

### Reports
- `GET /api/reports/spending?from=&to=` — Spending by category (base currency)
- `GET /api/reports/balances` — Account balances

### Budgets
- `GET /api/budgets?month=YYYY-MM` — Budget report with spent/remaining
- `POST /api/budgets` — Upsert budget

## Architecture

DDD-inspired layered architecture. See [docs/architecture.md](docs/architecture.md).

```
backend/
  cmd/api/          — entrypoint
  internal/
    domain/         — entities, value objects, invariants
    application/    — use cases (auth, ledger, importer, fx, budget, reports)
    infrastructure/ — SQLite repos, ECB client, SMTP sender
    interfaces/     — HTTP handlers, JWT middleware
    platform/       — config, migrator, crypto, sqlite
  migrations/sqlite/ — SQL migrations (auto-applied on startup)
```

## Fixtures

Sample CSVs for testing import are in `fixtures/`:
- `sample-revolut.csv` — EUR transactions with comma separator
- `sample-bank-usd.csv` — USD transactions with semicolon separator

## Testing

```bash
cd backend && go test ./... -v
```

## Verify

```bash
# 1) Run tests
cd backend && go test ./... -v

# 2) Start API with log email mode (no SMTP needed)
cd backend
AUTH_JWT_SECRET=dev-secret EMAIL_MODE=log DB_PATH=./dev.db go run ./cmd/api

# 3) In another terminal — health check (watch server logs for request line)
curl -s http://localhost:8080/api/ping

# 4) Trigger auth code request (watch logs for login code)
curl -s -X POST http://localhost:8080/auth/request_code \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com"}'
```
