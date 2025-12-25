# Makefile for pingora-gateway-controller

.PHONY: build test lint test-e2e test-integration clean help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
BINARY_NAME=controller
BINARY_DIR=bin

# Container parameters
CONTAINER_RUNTIME ?= podman
PINGORA_IMAGE ?= pingora-proxy:test

# Default target
.DEFAULT_GOAL := help

## Build

build: ## Build the controller binary
	$(GOBUILD) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/controller

## Testing

test: ## Run unit tests
	$(GOTEST) -v -race ./...

test-envtest: ## Run envtest tests
	$(GOTEST) -v -race -tags=envtest ./...

test-e2e: test-integration ## Alias for test-integration

test-integration: build-proxy-image ## Run integration tests with testcontainers
	TESTCONTAINERS_RYUK_DISABLED=true PINGORA_PROXY_IMAGE=$(PINGORA_IMAGE) $(GOTEST) -v -tags=integration -race -timeout=10m ./test/integration/...

build-proxy-image: ## Build Pingora proxy container image
	$(CONTAINER_RUNTIME) build --tag $(PINGORA_IMAGE) --file proxy/Containerfile proxy/

## Linting

lint: ## Run golangci-lint
	golangci-lint run --timeout=5m

lint-md: ## Run markdownlint
	markdownlint-cli2 '**/*.md'

## Dependencies

deps: ## Download dependencies
	$(GOMOD) download

tidy: ## Run go mod tidy
	$(GOMOD) tidy

vendor: ## Update vendor directory
	$(GOMOD) vendor

## Code generation

generate: ## Generate code (protobuf, etc.)
	buf generate

## Helm

helm-lint: ## Lint Helm chart
	helm lint charts/pingora-gateway-controller

helm-test: ## Run Helm unit tests
	helm unittest charts/pingora-gateway-controller

helm-docs: ## Generate Helm chart README
	helm-docs charts/pingora-gateway-controller

## Clean

clean: ## Clean build artifacts
	rm -rf $(BINARY_DIR)
	rm -rf coverage.out

## Help

help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
