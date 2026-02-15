BINARY_NAME := claude-statusline
BUILD_DIR := bin

.PHONY: build test vet lint clean install check

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

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
	go install .

check: vet lint test
