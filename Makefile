.PHONY: help docs
.DEFAULT_GOAL := help
.EXPORT_ALL_VARIABLES:

APP:=replaceme
BINARY_NAME:=replaceme
MINCOVERAGE?=70
PKG_LIST=$(shell go list ./... | grep -Ev "vendor|docs|cmd")
GOCACHE=$(shell pwd)/.build
COVERAGE_DIR?=.coverage
LOG_LEVEL?=debug
LOG_PRETTY?=true
REST_PRETTY?=true


help: ## This help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

test-ci: lint ## Runs the tests with coverage checks + lint
	@if [ -d ${COVERAGE_DIR} ]; then rm -rf ${COVERAGE_DIR}/*; else mkdir ${COVERAGE_DIR}; fi;
	@for package in ${PKG_LIST} ; do \
		pkgcov=$$(go test -covermode=atomic -test.count=1  -race -coverprofile="$(COVERAGE_DIR)/$$(basename $${package}).cov" "$${package}"); \
		retVal=$$? ;\
		if [ $$retVal -ne 0 ]; then \
			echo "$$pkgcov"; \
			echo " 🚨 TEST FAIL" ;\
			exit $$retVal; \
		fi;\
		pcoverage=$$(echo $$pkgcov| grep "coverage" | sed -E "s/.*coverage: ([0-9]*\.[0-9]+)\% of statements/\1/g") ;\
		if [ ! -z "$$pcoverage" ]; then \
			if [ $$(echo $${pcoverage%%.*}) -lt $(MINCOVERAGE) ] ; then \
				echo " 🚨 COVERAGE FAIL";\
				echo " 🚨 Test coverage of $$package is $$pcoverage%";\
				echo "FAIL" ;\
				echo ;\
				exit 1 ;\
			else \
				echo " ✅  Test coverage of $$package is $$pcoverage%" ;\
			fi \
		else \
			echo " ❗ No tests for $$package" ;\
		fi \
	done 
	@echo 'mode: atomic' > "$(COVERAGE_DIR)"/coverage.cov ;\
	for fcov in "$(COVERAGE_DIR)"/*.cov; do \
		if [ $$fcov != "$(COVERAGE_DIR)/coverage.cov" ]; then \
			tail -q -n +2 $$fcov >> $(COVERAGE_DIR)/coverage.cov ;\
		fi \
	done
	@echo 
	@pcoverage=$$(go tool cover -func=$(COVERAGE_DIR)/coverage.cov | grep 'total' | awk '{print substr($$3, 1, length($$3)-1)}');\
	echo "coverage: $$pcoverage% of project" ;\
	if [ $$(echo $${pcoverage%%.*}) -lt $$MINCOVERAGE ] ; then \
      echo ">> 🚨 Test coverage of project is $$pcoverage%" ;\
      echo "FAIL" ;\
      exit 1 ;\
	else \
		echo ">>  ✅  Test coverage of project is $$pcoverage%";\
	fi

test: ## Runs all tests normally
	go test -covermode=atomic -test.count=1 -race ./...

lint: ## Runs the linter
	golangci-lint run ./...

build: ## Builds go binary
	go build -o ./build/$(BINARY_NAME) main.go

run: docs ## Runs main package
	go run main.go

docs swag: ## Generate swagger documentation json/yaml
	@swag --version
	@swag init -p camelcase -g ../main.go -o docs/swagger -d ./config,./services/,./internal --md docs

up: ## Starts docker containers for dependent services
	@docker-compose up -d --build --remove-orphans

stop: ## Stops docker containers for dependent services
	@docker-compose stop

down: ## Removes docker containers for dependent services
	@docker-compose down --remove-orphans

deps: ## Fetches go dependencies
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin
	@go mod tidy
	@go mod download

clean: down ## Removes all docker containers and volumes
	docker system prune --volumes --force

migration: ## Creates a new migration
	@echo "Enter the name of the migration: e.g. 'JIRA-001-create-users-table' :" ;\
	read name;\
	migrate -database postgres create -ext sql -dir ./schema/sqldb/ -digits 3 -seq $$name

migrate: ## Runs migrations locally
	@migrate -path ./schema/sqldb -database "postgres://root:root@localhost:5432/replaceme?sslmode=disable" up

revert: ## Reverts migrations locally
	@migrate -path ./schema/sqldb -database "postgres://root:root@localhost:5432/replaceme?sslmode=disable" down
