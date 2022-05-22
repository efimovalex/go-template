.PHONY: help
.DEFAULT_GOAL := help
.EXPORT_ALL_VARIABLES:

APP:=replaceme
BINARY_NAME:=replaceme
CONFIG_FILE:=config/config_dev.toml
COVERAGE_DIR:="./.coverage"
MINCOVERAGE:=70
PKG_LIST:=$(shell  go list ./... | grep -v /vendor/)

help: ## This help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

test: ## Runs the tests
	@for package in ${PKG_LIST} ; do \
		pkgcov=$$(go test -covermode=atomic -coverprofile="$(COVERAGE_DIR)/$$(basename $${package}).cover" "$${package}"); \
		pcoverage=$$(echo $$pkgcov| grep "coverage" | sed -E "s/.*coverage: ([0-9]*\.[0-9]+)\% of statements/\1/g") ;\
		if [ ! -z "$$pcoverage" ]; then \
			if [ $$(echo $${pcoverage%%.*}) -lt $(MINCOVERAGE) ] ; then \
				echo "🚨 Test coverage of $$package is $$pcoverage%";\
				echo "FAIL" ;\
				exit 1 ;\
			else \
				echo "🟢 Test coverage of $$package is $$pcoverage%" ;\
			fi \
		else \
			echo "➖ No tests for $$package" ;\
		fi \
	done

	
build: ## Builds go binary
	go build -o ./$(BINARY_NAME) cmd/main.go

run: ## Runs main package
	go run cmd/main.go;

swag:  ## Generate swagger documentation json/yaml
	@swag --version
	@swag init -p pascalcase -g ../../cmd/main.go -o docs/swagger -d ./services/,./lib,./pkg --md docs
	
up: ## Starts docker containers for dependent services
	@docker-compose up -d --build --remove-orphans

down: ## Removes docker containers for dependent services
	@docker-compose down --remove-orphans

deps: ## Fetches go mod dependencies
	@go mod tidy
	@go mod download

clean: down ## Removes all docker containers and volumes
	docker system prune --volumes --force
	