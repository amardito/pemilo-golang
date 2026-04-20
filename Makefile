.PHONY: run build test migrate-up migrate-down migrate-create docker-up docker-down

DATABASE_URL ?= postgres://pemilo:pemilo_secret@localhost:5432/pemilo?sslmode=disable

run:
	go run cmd/server/main.go

build:
	go build -o pemilo cmd/server/main.go

test:
	go test ./... -v

migrate-up:
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DATABASE_URL)" down

migrate-create:
	goose -dir migrations create $(name) sql

docker-up:
	docker compose up -d

docker-down:
	docker compose down

tidy:
	go mod tidy
