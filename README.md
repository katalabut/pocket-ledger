# pocket-ledger

Self-hosted personal/family budget service with CSV import, envelope budgeting, and multi-currency reporting (ECB rates).

## Stack
- Backend: Go
- Frontend: React + TypeScript
- DB: SQLite
- Infra: Docker Compose

## Status
Bootstrap phase (public repository foundation).

## Roadmap
- Iteration 1: accounts/categories/transactions + minimal transactions UI
- Iteration 2: CSV import profiles + preview/commit + dedupe
- Iteration 3: FX (ECB) + reports
- Iteration 4: budgets screen + budget reports

## Development (planned)
```bash
make setup
make dev
make test
```

## Architecture
See [docs/architecture.md](docs/architecture.md)
