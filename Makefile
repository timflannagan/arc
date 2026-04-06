.PHONY: help build install clean test fmt vet

BINARY := arc
MODULE := github.com/agentregistry-dev/ar
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X $(MODULE)/pkg/cmd.Version=$(VERSION)"

# Help
# `make help` self-documents targets annotated with `##`.
help: NAME_COLUMN_WIDTH=35
help: LINE_COLUMN_WIDTH=5
help: ## Output the self-documenting make targets
	@grep -hnE '^[%a-zA-Z0-9_./-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk '{line=$$0; lineno=line; sub(/:.*/, "", lineno); sub(/^[^:]*:/, "", line); target=line; sub(/:.*/, "", target); desc=line; sub(/^.*##[[:space:]]*/, "", desc); printf "\033[36mL%-$(LINE_COLUMN_WIDTH)s%-$(NAME_COLUMN_WIDTH)s\033[0m %s\n", lineno, target, desc}'

build: ## Build the CLI binary
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/arc

install: ## Install the CLI to GOPATH/bin
	go install $(LDFLAGS) ./cmd/arc

clean: ## Remove build artifacts
	rm -rf bin/

test: ## Run the test suite
	go test ./...

fmt: ## Run go fmt
	go fmt ./...

vet: ## Run go vet
	go vet ./...
