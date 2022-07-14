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

### Operations

Make provides a interface to common development operations

 - test        Runs the tests
 - build       Builds go binary
 - run         Runs main package
 - swag        Generate swagger documentation json/yaml
 - up          Starts docker containers for dependent services
 - down        Removes docker containers for dependent services
 - deps        Fetches go mod dependencies
 - clean       Removes all docker containers and volumes
