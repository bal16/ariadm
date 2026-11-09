# Variables
BINARY_NAME=ariadm
FRONTEND_DIR=frontend

.PHONY: all help deps test dev build build-windows build-mac clean

all: help

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' Makefile | column -t -s ':' |  sed -e 's/^/ /'

## deps: Download Go modules and install frontend dependencies
deps:
	@echo "==> Downloading Go dependencies..."
	go mod download
	@echo "==> Installing frontend dependencies with Bun..."
	cd $(FRONTEND_DIR) && bun install

## test: Run all backend unit tests (TDD workflow)
test:
	@echo "==> Running backend unit tests..."
	go test -v -race ./internal/...

test/watch:
	@if ! command -v gotestsum > /dev/null; then \
		echo "==> gotestsum not found. Installing latest version..."; \
		go install gotestsum.org/gotestsum@latest; \
	fi
	@echo "==> Watching backend files for changes..."
	gotestsum --watch --format short-verbose -- ./internal/... -race

## dev: Run the application in Wails development mode (hot-reloading)
dev:
	@echo "==> Starting Wails development server..."
	APP_ENV=development wails3 task dev

## build: Build the production application for the current host OS
build:
	@echo "==> Building production application for host OS..."
	wails3 task build

## build-windows: Cross-compile the production application for Windows
build-windows:
	@echo "==> Building production application for Windows..."
	wails3 task build GOOS=windows ARCH=amd64

## build-mac: Cross-compile the production application for macOS (Universal binary)
build-mac:
	@echo "==> Building production application for macOS..."
	wails3 task darwin:build:universal

## clean: Remove build artifacts and temporary files
clean:
	@echo "==> Cleaning build artifacts..."
	rm -rf bin/*
	rm -rf build/bin/*
	rm -rf $(FRONTEND_DIR)/dist