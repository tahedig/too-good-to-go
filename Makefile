VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)
SWAG_VERSION := v1.16.4

.PHONY: help run build test race cover fmt fmt-check vet tidy wire swagger check docker

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-10s %s\n", $$1, $$2}'

run: ## Run the API locally
	go run ./cmd/api

build: ## Build bin/api with a version stamp
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/api ./cmd/api

test: ## Run all tests
	go test ./...

race: ## Run all tests with the race detector (needs cgo)
	go test -race ./...

cover: ## Run tests and print coverage per function
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

fmt: ## Format all Go files
	gofmt -w cmd internal pkg

fmt-check: ## Fail if any file is not gofmt-formatted
	@test -z "$$(gofmt -l cmd internal pkg)" || (gofmt -l cmd internal pkg; exit 1)

vet: ## Run go vet
	go vet ./...

tidy: ## Sync go.mod and go.sum
	go mod tidy

wire: ## Regenerate internal/injector/wire_gen.go after changing providers
	go generate ./internal/injector

swagger: ## Regenerate docs/ from the handler annotations
	go run github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION) init -g cmd/api/main.go -d ./ -o docs --parseInternal

check: fmt-check vet race ## Everything CI runs

docker: ## Build the Docker image
	docker build --build-arg VERSION=$(VERSION) -t too-good-to-go-backend:$(VERSION) .
