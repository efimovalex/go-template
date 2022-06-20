.PHONY: help
.DEFAULT_GOAL := help
.EXPORT_ALL_VARIABLES:

APP:=replaceme
BINARY_NAME:=replaceme
MINCOVERAGE:=70
PKG_LIST:=$(shell  go list ./... | grep -v /vendor/)
GOCACHE=$(shell pwd)/.build
LOG_LEVEL:=debug
LOG_DEV:=true

help: ## This help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

script-test test: # Runs the tests
	@./scripts/test.sh

lint: ## Runs the linter
	golangci-lint run ./...

build: ## Builds go binary
	go build -o ./build/$(BINARY_NAME) cmd/replaceme/main.go

run: ## Runs main package
	go run cmd/replaceme/main.go;

swag:  ## Generate swagger documentation json/yaml
	@swag --version
	@swag init -p pascalcase -g ../cmd/replaceme/main.go -o docs/swagger -d ./services/,./internal,./pkg --md docs

up: ## Starts docker containers for dependent services
	@docker-compose up -d --build --remove-orphans

stop: ## Stops docker containers for dependent services
	@docker-compose stop

down: ## Removes docker containers for dependent services
	@docker-compose down --remove-orphans

deps: ## Fetches go mod dependencies
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.46.2
	@go mod tidy
	@go mod download

clean: down ## Removes all docker containers and volumes
	docker system prune --volumes --force

migration: ## Creates a new migration
	@echo "Enter the name of the migration: e.g. 'HVO-001-create-users-table' :" ;\
	read name;\
	migrate -database postgres create -ext sql -dir ./schema/sqldb/ -digits 3 -seq $$name
migrate: ## Runs migrations locally
	@migrate -path ./schema/sqldb -database "postgres://root:root@localhost:5432/replaceme?sslmode=disable" up
revert: ## Reverts migrations locally
	@migrate -path ./schema/sqldb -database "postgres://root:root@localhost:5432/replaceme?sslmode=disable" down
