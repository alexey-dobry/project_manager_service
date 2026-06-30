SHELL := /bin/bash

# Папка с миграциями и DSN читаются из переменных окружения
MIGRATIONS_DIR ?= migrations
DB_DSN         ?= postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DBNAME)?sslmode=$(POSTGRES_SSLMODE)

.PHONY: help
help: ## показать справку
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# ───── сборка/запуск ─────

.PHONY: tidy
tidy: ## go mod tidy
	go mod tidy

.PHONY: build
build: ## собрать бинарник
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/auth-service ./cmd

.PHONY: run
run: ## запустить локально (нужен запущенный postgres)
	go run ./cmd

.PHONY: dev
dev: ## hot-reload через air
	air

# ───── тесты ─────

.PHONY: test
test: ## запустить все тесты
	go test ./... -count=1 -race

.PHONY: test-unit
test-unit: ## только unit
	go test ./tests/unit/... -count=1 -race -v

.PHONY: test-integration
test-integration: ## только integration
	go test ./tests/integration/... -count=1 -race -v

.PHONY: cover
cover: ## покрытие usecase-слоя
	go test ./internal/usecase/... ./tests/unit/... -coverpkg=./internal/usecase/... -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -n 1

# ───── миграции ─────

.PHONY: migrate-up
migrate-up: ## применить миграции
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" up

.PHONY: migrate-down
migrate-down: ## откатить одну миграцию
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" down 1

.PHONY: migrate-new
migrate-new: ## создать миграцию: make migrate-new NAME=xxx
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

# ───── swagger ─────

.PHONY: swag
swag: ## сгенерировать docs/ из аннотаций
	swag init -g cmd/main.go -o docs --parseDependency --parseInternal

# ───── docker ─────

.PHONY: docker-build
docker-build: ## собрать образ
	docker build -t auth-service:local .

.PHONY: docker-up
docker-up: ## docker compose up (из корня монорепо)
	cd .. && docker-compose up -d auth-postgres auth-service

# ───── fuzz ─────
FUZZTIME ?= 30s
.PHONY: fuzz fuzz-corpus
fuzz: ## native go fuzz (FUZZTIME=30s по умолчанию)
	go test ./internal/pkg/jwt        -run=- -fuzz=FuzzParseAccessToken  -fuzztime=$(FUZZTIME)
	go test ./internal/pkg/hasher     -run=- -fuzz=FuzzBcryptCompare     -fuzztime=$(FUZZTIME)
	go test ./internal/delivery/http  -run=- -fuzz=FuzzRegisterHandler   -fuzztime=$(FUZZTIME)
fuzz-corpus: ## регрессия по сохранённому corpus, без новых мутаций
	go test ./internal/pkg/jwt        -run=FuzzParseAccessToken
	go test ./internal/pkg/hasher     -run=FuzzBcryptCompare
	go test ./internal/delivery/http  -run=FuzzRegisterHandler
