.PHONY: all build test lint dev clean docker-up docker-down

all: build

# Go Backend
build:
	go build -o bin/erp-api.exe ./cmd/erp-api
	go build -o bin/erp-worker.exe ./cmd/erp-worker
	go build -o bin/erp-cli.exe ./cmd/erp-cli
	go build -o bin/erp-migrate.exe ./cmd/erp-migrate

test:
	@test -n "$(TEST_DATABASE_URL)" || (echo "Set TEST_DATABASE_URL to a disposable PostgreSQL database; integration tests rebuild its application schemas." >&2; exit 1)
	go test -v -race -p 1 -timeout 30m ./cmd/... ./internal/... ./tests/...

# Frontend
web-install:
	cd apps/web && npm ci

web-build:
	cd apps/web && npm run build

web-lint:
	cd apps/web && npm run lint

web-dev:
	cd apps/web && npm run dev

# Docker Compose
docker-up:
	docker compose up -d

docker-down:
	docker compose down

clean:
	rm -rf bin/
