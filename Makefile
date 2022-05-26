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

test: lint ## Runs the tests
	@for package in ${PKG_LIST} ; do \
		pkgcov=$$(go test -covermode=count -coverprofile="$(COVERAGE_DIR)/$$(basename $${package}).cov" "$${package}"); \
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
	@echo 'mode: count' > "$(COVERAGE_DIR)"/coverage.cov
	@for fcov in "$(COVERAGE_DIR)"/*.cov; do \
		if [ $$fcov != "$(COVERAGE_DIR)/coverage.cov" ]; then \
			tail -q -n +2 $$fcov >> $(COVERAGE_DIR)/coverage.cov ;\
		fi \
	done
	@echo
	@pcoverage=$$(go tool cover -func=$(COVERAGE_DIR)/coverage.cov | grep 'total' | awk '{print substr($$3, 1, length($$3)-1)}');\
	echo "coverage: $$pcoverage% of project" ;\
	if [ $$(echo $${pcoverage%%.*}) -lt $$MINCOVERAGE ] ; then \
      echo "🚨 Test coverage of project is $$pcoverage%" ;\
      echo "FAIL" ;\
      exit 1 ;\
	else \
		echo "🟢 Test coverage of project is $$pcoverage%";\
	fi
lint:
	golangci-lint run ./...
	
build: ## Builds go binary
	go build -o ./build/$(BINARY_NAME) cmd/replaceme/main.go
	go build -o ./build/$(BINARY_NAME)_static cmd/static/main.go

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
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.46.2
	@go mod tidy
	@go mod download		

clean: down ## Removes all docker containers and volumes
	docker system prune --volumes --force
	