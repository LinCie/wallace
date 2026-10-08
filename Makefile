.PHONY: dev start build postgres-up postgres-down migrate-up migrate-down migrate-status migrate-version migrate-reset migration seed

-include .env

DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/wallace?sslmode=disable
MIGRATIONS_DIR ?= migrations
export DATABASE_URL MIGRATIONS_DIR
GOOSE := go run github.com/pressly/goose/v3/cmd/goose@v3.27.3

dev:
	air

start:
	./bin/web

build:
	go build -o ./bin/web ./cmd/web

postgres-up:
	docker compose up -d postgres

postgres-down:
	docker compose down

migrate-up:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate up

migrate-down:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate down

migrate-status:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate status

migrate-version:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate version

migrate-reset:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/migrate reset

seed:
	DATABASE_URL="$(DATABASE_URL)" go run ./cmd/seed

migration:
	@test -n "$(name)" || (echo "usage: make migration name=create_users" && exit 1)
	$(GOOSE) -dir "$(MIGRATIONS_DIR)" postgres "$(DATABASE_URL)" create "$(name)" sql
