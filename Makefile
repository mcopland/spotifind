export PATH := $(shell go env GOPATH)/bin:$(PATH)

.PHONY: dev up down logs migrate migrate-host migrate-docker \
        dev-server dev-web \
        db-up db-down db-test-up db-test-down db-backup db-restore \
        prod-up prod-down \
        build test test-cover test-integration test-ci tools

TEST_DATABASE_URL ?= postgres://spotifind:spotifind@localhost:5433/spotifind_test

# Full Docker dev stack (default workflow)
dev: up

up:
	docker compose up -d --build
	@$(MAKE) migrate
	docker compose logs -f server web

down:
	docker compose down

logs:
	docker compose logs -f

# Run migrations inside Docker (default)
migrate: migrate-docker

migrate-docker:
	docker compose --profile tools run --rm migrate

# Run migrations on the host (requires Go and a running db)
migrate-host:
	cd server && go run ./cmd/migrate

# Host-only escape hatches (no Docker)
dev-server:
	cd server && air

dev-web:
	cd web && npm run dev

# Dev database shortcuts (for host-based workflow)
db-up:
	docker compose up -d db
	@echo "Waiting for database..."
	@until docker compose exec db pg_isready -U spotifind > /dev/null 2>&1; do sleep 1; done

db-down:
	docker compose down

# Test database
db-test-up:
	docker compose -f docker-compose.test.yml up -d db-test
	@echo "Waiting for test database..."
	@until docker compose -f docker-compose.test.yml exec db-test pg_isready -U spotifind > /dev/null 2>&1; do sleep 1; done

db-test-down:
	docker compose -f docker-compose.test.yml down

# Backups
db-backup:
	./scripts/db-backup.sh

db-restore:
	@test -n "$(FILE)" || (echo "usage: make db-restore FILE=backups/spotifind-<timestamp>.dump CONFIRM=1"; exit 1)
	CONFIRM=$(CONFIRM) ./scripts/db-restore.sh "$(FILE)"

# Production stack
prod-up:
	docker compose -f docker-compose.prod.yml up -d --build

prod-migrate:
	docker compose -f docker-compose.prod.yml --profile tools run --rm migrate

prod-down:
	docker compose -f docker-compose.prod.yml down

# Build binaries and frontend locally
build:
	cd server && go build -o bin/spotifind ./cmd/spotifind
	cd web && npm run build

# Tests
test:
	cd server && go test -coverpkg=./internal/... -coverprofile=unit.out ./internal/...

test-cover: test
	cd server && go tool cover -func=unit.out

test-integration: db-test-up
	cd server && TEST_DATABASE_URL=$(TEST_DATABASE_URL) go test -tags integration \
		-coverpkg=./internal/... -coverprofile=integration.out ./internal/...

tools:
	go install github.com/wadey/gocovmerge@latest
	go install github.com/vladopajic/go-test-coverage/v2@latest

test-ci: db-test-up
	cd server && go test -coverpkg=./internal/... -coverprofile=unit.out ./internal/...
	cd server && TEST_DATABASE_URL=$(TEST_DATABASE_URL) go test -tags integration \
		-coverpkg=./internal/... -coverprofile=integration.out ./internal/repository/...
	cd server && gocovmerge unit.out integration.out > combined.out
	cd server && go-test-coverage --config=.testcoverage.yml
