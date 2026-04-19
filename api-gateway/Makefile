SHELL := /bin/bash

.PHONY: help tidy build run dev test cover docker-build
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

tidy: ## go mod tidy
	go mod tidy
build: ## собрать бинарник
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/api-gateway ./cmd
run: ## запустить
	go run ./cmd
dev: ## hot-reload
	air
test: ## все тесты
	go test ./... -count=1 -race
cover: ## coverage
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -n 1
docker-build:
	docker build -t api-gateway:local .
