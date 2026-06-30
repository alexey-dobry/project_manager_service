SHELL := /bin/bash

MIGRATIONS_DIR ?= migrations
DB_DSN         ?= postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DBNAME)?sslmode=$(POSTGRES_SSLMODE)

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

.PHONY: tidy build run dev test test-unit test-integration cover
tidy:
	go mod tidy
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/projects-service ./cmd
run:
	go run ./cmd
dev:
	air
test:
	go test ./... -count=1 -race
test-unit:
	go test ./tests/unit/... -count=1 -race -v
test-integration:
	go test ./tests/integration/... -count=1 -race -v
cover:
	go test ./internal/usecase/... ./tests/unit/... -coverpkg=./internal/usecase/...,./internal/domain/... -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -n 1

.PHONY: migrate-up migrate-down migrate-new
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" up
migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" down 1
migrate-new:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

.PHONY: swag docker-build
swag:
	swag init -g cmd/main.go -o docs --parseDependency --parseInternal
docker-build:
	docker build -t projects-service:local .

# ───── fuzz ─────
FUZZTIME ?= 30s
.PHONY: fuzz fuzz-corpus
fuzz:
	go test ./internal/domain  -run=- -fuzz=FuzzTaskStateMachine          -fuzztime=$(FUZZTIME)
	go test ./internal/usecase -run=- -fuzz=FuzzCreateCommentValidation   -fuzztime=$(FUZZTIME)
fuzz-corpus:
	go test ./internal/domain  -run=FuzzTaskStateMachine
	go test ./internal/usecase -run=FuzzCreateCommentValidation
