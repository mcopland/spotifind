.PHONY: dev dev-server dev-web db-up db-down migrate build test

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

migrate:
	cd server && go run ./cmd/migrate

build:
	cd server && go build -o bin/spotifind ./cmd/spotifind
	cd web && npm run build

test:
	cd server && go test ./...
