.PHONY: all build test lint dev clean docker-up docker-down

all: build

# Go Backend
build:
	go build -o bin/erp-api.exe ./cmd/erp-api
	go build -o bin/erp-worker.exe ./cmd/erp-worker
	go build -o bin/erp-cli.exe ./cmd/erp-cli

test:
	go test -v -race ./...

# Frontend
web-install:
	cd apps/web && npm install

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
