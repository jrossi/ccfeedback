.PHONY: all test build clean fmt lint install bench snapshot release

# Build information
BINARY_NAME=ccfeedback
GO=go
GOFLAGS=-trimpath

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Build flags
LDFLAGS=-s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE) \
	-X main.builtBy=make

all: fmt lint test build

build:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/ccfeedback

test:
	$(GO) test -v -race -coverprofile=coverage.out ./...

bench:
	$(GO) test -bench=. -benchmem ./...

fmt:
	$(GO) fmt ./...
	gofmt -s -w .

lint:
	@which golangci-lint > /dev/null 2>&1 || (echo "golangci-lint not found, installing..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

install:
	$(GO) install $(GOFLAGS) -ldflags "$(LDFLAGS)" ./cmd/ccfeedback

clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	rm -rf dist/
	$(GO) clean -cache

deps:
	$(GO) mod download
	$(GO) mod tidy

coverage: test
	$(GO) tool cover -html=coverage.out -o coverage.html

# GoReleaser targets
snapshot:
	@command -v goreleaser > /dev/null || (echo "goreleaser not found. Install from https://goreleaser.com" && exit 1)
	goreleaser release --snapshot --clean

release:
	@command -v goreleaser > /dev/null || (echo "goreleaser not found. Install from https://goreleaser.com" && exit 1)
	goreleaser release --clean