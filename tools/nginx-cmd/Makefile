# Makefile for nginx-cmd (AIO Engine)

BINARY_NAME=nginx-cmd
VERSION=1.0.0-kaizen
BUILD_DIR=build

.PHONY: all build clean tidy test help

all: tidy build

build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@go build -ldflags "-X main.Version=$(VERSION)" -o $(BINARY_NAME) main.go

tidy:
	@echo "Tidying go modules..."
	@go mod tidy

clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)

test:
	@echo "Running tests..."
	@go test ./...

help:
	@echo "Nginx Control Tool Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build    Build the binary"
	@echo "  make tidy     Sync go modules"
	@echo "  make clean    Remove binary and build artifacts"
	@echo "  make test     Run unit tests"
	@echo "  make all      Tidy and build"
