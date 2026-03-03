# Include variables from the .envrc file
include .envrc

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N]' && read ans && [ $${ans:-N} = y ]

## run/server: run the cmd/server application
.PHONY: run/server
run/server:
	@go run ./cmd/server -port=${CC_PORT} -env=${CC_ENVIRONMENT} -redis-addr=${CC_REDIS_ADDR} -redis-password=${CC_REDIS_PASSWORD}

## run/race: run the cmd/api application with the Go race detector
.PHONY: run/race
run/race:
	@go run -race ./cmd/server -port=${CC_PORT} -env=${CC_ENVIRONMENT} -redis-addr=${CC_REDIS_ADDR} -redis-password=${CC_REDIS_PASSWORD}

## tidy: format all .go files and tidy module dependencies
.PHONY: tidy
tidy:
	@echo 'Formatting .go files...'
	go fmt ./...
	@echo 'Tidying module dependencies...'
	go mod tidy

## audit: run quality control checks
.PHONY: audit
audit:
	@echo 'Checking module dependencies...'
	go mod tidy -diff
	go mod verify
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...
