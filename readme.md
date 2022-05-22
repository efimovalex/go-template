# Go Project Template

## Introduction
Find and replace `replaceme` string in all project with your new project name.
## Development

### Dependencies

The development setup depends on the following tools:

- go
- docker
- docker-compose
- golint

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
