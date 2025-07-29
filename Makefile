# Azure Searcher Makefile

.PHONY: help test test-unit test-integration test-coverage clean build lint format deps check-deps

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Dependencies
deps: ## Download and verify dependencies
	go mod download
	go mod verify
	go mod tidy

check-deps: ## Check for dependency updates
	go list -u -m all

# Testing
test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	go test -v -race -coverprofile=coverage.out ./src/...

test-integration: ## Run integration tests
	go test -v -race -tags=integration ./...

test-coverage: test-unit ## Run tests and generate coverage report
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	go tool cover -func=coverage.out

test-short: ## Run tests in short mode
	go test -short -race ./src/...

test-verbose: ## Run tests with verbose output
	go test -v -race ./src/... ./...

bench: ## Run benchmarks
	go test -bench=. -benchmem -run=^$ ./src/...

# Code quality
lint: ## Run linter
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run --timeout=5m

format: ## Format code
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	go vet ./...

security: ## Run security checks
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
	gosec ./...
	@which govulncheck > /dev/null || (echo "Installing govulncheck..." && go install golang.org/x/vuln/cmd/govulncheck@latest)
	govulncheck ./...

# Build
build: ## Build the binary
	go build -v -o azure-searcher .

build-all: ## Build for multiple platforms
	GOOS=linux GOARCH=amd64 go build -o dist/azure-searcher-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o dist/azure-searcher-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o dist/azure-searcher-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o dist/azure-searcher-windows-amd64.exe .

install: ## Install the binary
	go install .

# Development
dev: ## Run in development mode
	go run . || true

dev-watch: ## Watch for changes and rebuild (requires entr)
	find . -name "*.go" | entr -r make build

# Clean
clean: ## Clean build artifacts and test cache
	go clean
	go clean -testcache
	rm -f coverage.out coverage.html
	rm -rf dist/

clean-all: clean ## Clean everything including module cache
	go clean -modcache

# Pre-commit checks
check: deps format vet lint test ## Run pre-commit checks

ci: deps vet lint test build ## Run CI pipeline locally

# Documentation
docs: ## Generate and serve documentation
	@echo "Generating documentation..."
	godoc -http=:6060
	@echo "Documentation available at http://localhost:6060"

# Release helpers
tag: ## Tag a new version (VERSION=v1.0.0 make tag)
	@test -n "$(VERSION)" || (echo "VERSION is required. Usage: VERSION=v1.0.0 make tag" && exit 1)
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

changelog: ## Generate changelog (requires git-chglog)
	@which git-chglog > /dev/null || (echo "Please install git-chglog: https://github.com/git-chglog/git-chglog" && exit 1)
	git-chglog -o CHANGELOG.md

# Quick commands
fast-test: ## Run fast tests only (no race detection, short mode)
	go test -short ./src/...

smoke-test: ## Quick smoke test
	go build -o /tmp/azure-searcher . && /tmp/azure-searcher --help

# Variables for make
COVERAGE_THRESHOLD = 80

coverage-check: test-unit ## Check if coverage meets threshold
	@coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Current coverage: $$coverage%"; \
	if [ $$(echo "$$coverage < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "Coverage $$coverage% is below threshold of $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	else \
		echo "Coverage $$coverage% meets threshold of $(COVERAGE_THRESHOLD)%"; \
	fi

# Help for common workflows
setup: deps ## Initial project setup
	@echo "Setting up development environment..."
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "Setup complete! Run 'make help' to see available commands."

# Container support
docker-build: ## Build Docker image
	docker build -t azure-searcher .

docker-test: ## Run tests in Docker
	docker run --rm -v $(PWD):/app -w /app golang:1.23 make test