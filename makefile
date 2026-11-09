# Variables
BINARY_NAME=ariadm
FRONTEND_DIR=frontend

.PHONY: all help deps tools fmt fmt-check lint typecheck check test dev build build-windows build-mac clean

all: help

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' Makefile | column -t -s ':' |  sed -e 's/^/ /'

## deps: Download Go modules and install all Bun workspace dependencies
deps:
	@echo "==> Downloading Go dependencies..."
	go mod download
	@echo "==> Installing workspace dependencies with Bun..."
	bun install

## tools: Install the linting and git-hook toolchain
tools:
	@echo "==> Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2
	@echo "==> Installing JavaScript tooling..."
	bun install
	@echo "==> Installing Lefthook git hooks..."
	bunx lefthook install

## fmt: Format Go and frontend sources
fmt:
	wails3 task fmt

## fmt-check: Verify formatting without writing files
fmt-check:
	wails3 task fmt:check

## lint: Run Go and frontend linters
lint:
	wails3 task lint

## typecheck: Type-check the frontend workspaces with Bun
typecheck:
	wails3 task typecheck

## check: Run formatting checks, lint, typecheck, and tests
check:
	wails3 task check

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
	rm -rf web-extension/dist