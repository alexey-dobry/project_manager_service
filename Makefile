SHELL := /bin/bash

SERVICES := auth-service groups-service projects-service api-gateway
BACKENDS := auth-service groups-service projects-service

.PHONY: help
help: ## показать справку
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ───── docker-compose ─────

.PHONY: up down logs ps restart
up: ## поднять всё (БД + сервисы)
	docker-compose up -d --build

up-db: ## поднять только БД (для локальной разработки сервисов)
	docker-compose up -d auth-postgres groups-postgres projects-postgres

down: ## остановить и убрать всё
	docker-compose down

down-volumes: ## ОПАСНО: удалить тома БД (теряем данные)
	docker-compose down -v

logs: ## tail логов всех сервисов
	docker-compose logs -f --tail=200

logs-%: ## tail логов одного сервиса: make logs-auth-service
	docker-compose logs -f --tail=200 $*

ps: ## статус контейнеров
	docker-compose ps

restart-%: ## пересобрать и перезапустить один сервис: make restart-auth-service
	docker-compose up -d --build $*

# ───── миграции ─────

.PHONY: migrate-up migrate-down
migrate-up: ## применить миграции во всех БД (через golang-migrate в каждом сервисе)
	@for s in $(BACKENDS); do \
		echo "→ migrating $$s"; \
		$(MAKE) -C $$s migrate-up || exit 1; \
	done

migrate-down: ## откатить по одной миграции в каждой БД
	@for s in $(BACKENDS); do \
		echo "→ rolling back $$s"; \
		$(MAKE) -C $$s migrate-down || exit 1; \
	done

# ───── тесты ─────

.PHONY: test test-unit test-integration cover
test: ## запустить тесты во всех сервисах
	@for s in $(SERVICES); do \
		echo "→ test $$s"; \
		$(MAKE) -C $$s test || exit 1; \
	done

test-unit:
	@for s in $(BACKENDS); do $(MAKE) -C $$s test-unit || exit 1; done

test-integration:
	@for s in $(SERVICES); do $(MAKE) -C $$s test-integration || exit 1; done

cover: ## coverage usecase-слоя в каждом backend-сервисе
	@for s in $(BACKENDS); do \
		echo "→ cover $$s"; \
		$(MAKE) -C $$s cover || exit 1; \
	done

# ───── разработка ─────

.PHONY: tidy run-% swag
tidy: ## go mod tidy во всех сервисах
	@for s in $(SERVICES); do (cd $$s && go mod tidy); done

run-%: ## запустить один сервис локально: make run-auth-service
	$(MAKE) -C $* run

swag: ## сгенерировать Swagger-доки во всех backend-сервисах
	@for s in $(BACKENDS); do $(MAKE) -C $$s swag || exit 1; done

# ───── smoke ─────

.PHONY: smoke
smoke: ## health-check всех сервисов через gateway
	@echo "gateway:    "      && curl -fs http://localhost:8080/health | head -1
	@echo "auth:       "      && curl -fs http://localhost:8081/health | head -1
	@echo "groups:     "      && curl -fs http://localhost:8082/health | head -1
	@echo "projects:   "      && curl -fs http://localhost:8083/health | head -1

# ───── fuzz ─────
# Покрывает критику: «нет FuzzXxx, нет corpus, нет make-цели».
# Два слоя: native Go fuzz (white-box, рядом с кодом) и Schemathesis (black-box).

FUZZTIME ?= 30s

.PHONY: fuzz fuzz-go fuzz-schemathesis fuzz-corpus

fuzz: fuzz-go ## native Go fuzz по всем сервисам (быстрая регрессия, FUZZTIME=30s по умолчанию)

fuzz-go: ## native Go fuzz: запускает все FuzzXxx по очереди
	@echo "──────── auth-service: FuzzParseAccessToken"
	cd auth-service && go test ./internal/pkg/jwt -run=- -fuzz=FuzzParseAccessToken -fuzztime=$(FUZZTIME)
	@echo "──────── auth-service: FuzzBcryptCompare"
	cd auth-service && go test ./internal/pkg/hasher -run=- -fuzz=FuzzBcryptCompare -fuzztime=$(FUZZTIME)
	@echo "──────── auth-service: FuzzRegisterHandler"
	cd auth-service && go test ./internal/delivery/http -run=- -fuzz=FuzzRegisterHandler -fuzztime=$(FUZZTIME)
	@echo "──────── groups-service: FuzzCreateGroupValidation"
	cd groups-service && go test ./internal/usecase -run=- -fuzz=FuzzCreateGroupValidation -fuzztime=$(FUZZTIME)
	@echo "──────── projects-service: FuzzTaskStateMachine"
	cd projects-service && go test ./internal/domain -run=- -fuzz=FuzzTaskStateMachine -fuzztime=$(FUZZTIME)
	@echo "──────── projects-service: FuzzCreateCommentValidation"
	cd projects-service && go test ./internal/usecase -run=- -fuzz=FuzzCreateCommentValidation -fuzztime=$(FUZZTIME)
	@echo "✓ go fuzz: пройдено за FUZZTIME=$(FUZZTIME) на цель"

fuzz-corpus: ## прогон только по зафиксированному corpus (без новых мутаций, быстро)
	cd auth-service     && go test ./internal/pkg/jwt    -run=FuzzParseAccessToken
	cd auth-service     && go test ./internal/pkg/hasher -run=FuzzBcryptCompare
	cd auth-service     && go test ./internal/delivery/http -run=FuzzRegisterHandler
	cd groups-service   && go test ./internal/usecase   -run=FuzzCreateGroupValidation
	cd projects-service && go test ./internal/domain    -run=FuzzTaskStateMachine
	cd projects-service && go test ./internal/usecase   -run=FuzzCreateCommentValidation
	@echo "✓ corpus regression: ок"

fuzz-schemathesis: ## black-box fuzz через OpenAPI (требует поднятый docker-compose up)
	./fuzz/schemathesis/run.sh
