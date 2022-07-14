# Go Project Template

## Introduction
Find and replace `replaceme` string in all project with your new project name for fast development start.

File structure: 

- cmd -> where the main files live
- config -> package which has the role of loading the starting config
- docs -> swagger generated files
- internal -> clearly defined internal packages and adapters for other services
- schema -> migrations for the used databases
- services -> the servers and it's dependent services. one main file starts one of the services
- 

## Development

### Dependencies

The development setup depends on the following tools:

- go
- docker
- docker-compose
- golangci-lint
- swaggo



### Operations

Make provides a interface to common development operations

```
$ make

Usage:
  make <target>

Targets:
  help        This help.
  test-ci     Runs the tests with coverage checks + lint
  test        Runs all tests normally
  lint        Runs the linter
  build       Builds go binary
  run         Runs main package
  docs        Generate swagger documentation json/yaml
  up          Starts docker containers for dependent services
  stop        Stops docker containers for dependent services
  down        Removes docker containers for dependent services
  deps        Fetches go dependencies
  clean       Removes all docker containers and volumes
  migration   Creates a new migration
  migrate     Runs migrations locally
  revert      Reverts migrations locally
  ```