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
	@echo "==> Installing frontend dependencies..."
	cd $(FRONTEND_DIR) && npm install

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
	wails dev

## build: Build the production application for the current host OS
build:
	@echo "==> Building production application for host OS..."
	wails build -clean

## build-windows: Cross-compile the production application for Windows
build-windows:
	@echo "==> Building production application for Windows..."
	wails build -platform windows/amd64 -clean

## build-mac: Cross-compile the production application for macOS (Universal binary)
build-mac:
	@echo "==> Building production application for macOS..."
	wails build -platform darwin/universal -clean

## clean: Remove build artifacts and temporary files
clean:
	@echo "==> Cleaning build artifacts..."
	rm -rf build/bin/*
	rm -rf $(FRONTEND_DIR)/dist