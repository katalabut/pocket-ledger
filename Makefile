.PHONY: setup dev test lint fmt build backend-build frontend-build docker-up docker-down

setup:
	cd backend && go mod download
	cd frontend && npm ci

dev:
	@echo "Start backend: cd backend && go run ./cmd/api"
	@echo "Start frontend: cd frontend && npm run dev"

test:
	cd backend && go test ./... -v

lint:
	cd backend && go vet ./...

fmt:
	cd backend && gofmt -w .

backend-build:
	cd backend && go build -o bin/pocket-ledger ./cmd/api

frontend-build:
	cd frontend && npm run build

build: backend-build frontend-build

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down
