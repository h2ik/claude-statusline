BINARY_NAME := claude-statusline
BUILD_DIR := bin

.PHONY: build test vet lint clean install check

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u '+%Y-%m-%d')
LDFLAGS := -X main.version=$(VERSION) \
           -X main.commit=$(GIT_COMMIT) \
           -X main.date=$(BUILD_DATE)

build:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) .

test:
	go test ./... -v

vet:
	go vet ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. See https://golangci-lint.run/welcome/install/"; exit 1; }
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)

install:
	go install -ldflags "$(LDFLAGS)" .

check: vet lint test
