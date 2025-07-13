.PHONY: all test build clean fmt lint install bench

BINARY_NAME=ccfeedback
GO=go
GOFLAGS=-trimpath
LDFLAGS=-s -w

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
	$(HOME)/go/bin/golangci-lint run ./...

install:
	$(GO) install $(GOFLAGS) -ldflags "$(LDFLAGS)" ./cmd/ccfeedback

clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	$(GO) clean -cache

deps:
	$(GO) mod download
	$(GO) mod tidy

coverage: test
	$(GO) tool cover -html=coverage.out -o coverage.html