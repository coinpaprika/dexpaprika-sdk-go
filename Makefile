.PHONY: build run-example test tidy check vuln help
.DEFAULT_GOAL: all

all: check test build ## Default target: check, test, build

build: ## Build all artifacts
	@go build -trimpath -o ./bin/production_usage examples/production_usage.go

run-example: ## Run real world example
	@go run examples/production_usage.go
test:
	@go test -shuffle=on -race ./...

tidy: ## Run go mod tidy
	@go mod tidy

check: ## Linting and static analysis
# binary will be $(go env GOPATH)/bin/golangci-lint
	@if test ! -e ./bin/golangci-lint; then \
		curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/v2.1.2/install.sh| sh -s v2.1.2; \
	fi

	@./bin/golangci-lint run -c .golangci.yml

vuln: ## Run vulnerability checks (requires Go 1.24)
	@echo "Checking Go version for vulnerability scanning..."
	@go version
	@if go version | grep -q "go1.24"; then \
		echo "Go 1.24 detected, running vulnerability check..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
		$(shell go env GOPATH)/bin/govulncheck ./...; \
	else \
		echo "Error: Vulnerability check requires Go 1.24+"; \
		exit 1; \
	fi

format: ## Format go code with goimports
	@go install golang.org/x/tools/cmd/goimports@latest
	@goimports -l -w .

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
