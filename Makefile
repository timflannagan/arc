.PHONY: build install clean test fmt vet

BINARY := arc
MODULE := github.com/agentregistry-dev/ar
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X $(MODULE)/pkg/cmd.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/arc

install:
	go install $(LDFLAGS) ./cmd/arc

clean:
	rm -rf bin/

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...
