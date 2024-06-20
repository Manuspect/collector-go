# Load env variales from .env if it exist
ifneq (,$(wildcard .env))
  ENV := $(PWD)/.env
  include $(ENV)
endif

SHELL = /bin/bash

ifeq ($(uname -m),$(filter $( uname -m), Darwin x86_64))
  GOARCH ?= amd64
else
  GOARCH ?= $(shell uname -m)
endif

GO111MODULE := on
CGO_ENABLED ?= 0
GOOS ?= linux
TARGET := collectorgo
COMMIT_TAG := $(shell git describe --tags)
COMMIT_SHORT_SHA := $(shell git rev-parse --short HEAD)
BUILD_VERSION:=$(COMMIT_SHORT_SHA)
BUILD_DATE:=$(shell date "+%Y.%m.%d_%H:%M:%S")

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n\033[36m\033[0m"} /^[$$()% a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: lint
lint: ## Run linter
	golangci-lint version
	golangci-lint run

.PHONY: tidy
tidy: ## Run go mod tidy
	go mod tidy

.PHONY: download
download: ## Run go mod download
	go mod download

.PHONY: clean
clean: ## Remove temporary files
	go clean -i all
	@rm -rf ./bin/$(TARGET)

.PHONY: build
build: clean tidy ## Build service
	GOOS=${GOOS} GOARCH=${GOARCH} GO111MODULE=${GO111MODULE} CGO_ENABLED=${CGO_ENABLED} \
		go build -ldflags "-s -w -X main.Version=$(BUILD_VERSION) -X main.BuildDate=$(BUILD_DATE)" -o bin/$(TARGET) ./cmd/

.PHONY: vendors
vendors: ## Run go mod vendor
	go mod vendor

.PHONY: up
up: ## Run docker compose up -d
	mkdir -p -m 777 ./_data/pgadmin
	docker compose -f docker-compose.local.yml down && docker compose -f docker-compose.local.yml up -d

.PHONY: down
down: ## Run docker compose down
	docker compose -f docker-compose.local.yml down

.PHONY: dev
dev: tidy ## Run go run
	go run ./cmd/

.PHONY: goose-up
goose-up: ## Run gooose up
	goose -dir ./sql/migrations postgres "postgres://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}/${DB_NAME}" up

.PHONY: goose-down
goose-down: ## Run gooose down
	goose -dir ./sql/migrations postgres "postgres://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}/${DB_NAME}" down

.PHONY: swager
swager: ## Run swager init
	swag init -g ./cmd/main.go

.PHONY: swager-fmt
swager-fmt: ## Run swager fmt
	swag fmt -g ./internal/service/

.PHONY: sqlc-gen
sqlc-gen: ## Run sqlc generate
	sqlc generate
