VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null)
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT)"
BINARY := asylum
BUILD_DIR := build

.PHONY: build build-all clean test test-integration test-e2e

build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/asylum

build-all:
	GOOS=linux  GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-amd64  ./cmd/asylum
	GOOS=linux  GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-arm64  ./cmd/asylum
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 ./cmd/asylum
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 ./cmd/asylum

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./...

test-integration:
	go test -tags integration -v -timeout 30m ./integration/

test-e2e:
	go test -tags e2e -v -timeout 30m ./e2e/
