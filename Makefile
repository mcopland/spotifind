export PATH := $(shell go env GOPATH)/bin:$(PATH)

.PHONY: dev dev-server dev-web db-up db-down db-test-up db-test-down migrate build test test-cover test-integration test-ci tools

TEST_DATABASE_URL ?= postgres://spotifind:spotifind@localhost:5433/spotifind_test

dev: db-up migrate
	@echo "Starting development servers..."
	@$(MAKE) -j2 dev-server dev-web

dev-server:
	cd server && air

dev-web:
	cd web && npm run dev

db-up:
	docker compose up -d db
	@echo "Waiting for database..."
	@until docker compose exec db pg_isready -U spotifind > /dev/null 2>&1; do sleep 1; done

db-down:
	docker compose down

db-test-up:
	docker compose -f docker-compose.test.yml up -d db-test
	@echo "Waiting for test database..."
	@until docker compose -f docker-compose.test.yml exec db-test pg_isready -U spotifind > /dev/null 2>&1; do sleep 1; done

db-test-down:
	docker compose -f docker-compose.test.yml down

migrate:
	cd server && go run ./cmd/migrate

build:
	cd server && go build -o bin/spotifind ./cmd/spotifind
	cd web && npm run build

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
