# AriaDM

Aria2 download manager desktop application built with Wails v3, SolidJS, and TypeScript.

## Requirements

- Go 1.26+
- Bun 1.4+
- Wails v3 CLI (`go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0`)
- The native webview dependencies required by Wails for your platform

## Setup

```bash
make deps    # Go modules + all Bun workspace dependencies
make tools   # golangci-lint, JavaScript tooling, and Lefthook git hooks
```

## Development

```bash
make dev      # Wails live development mode
make build    # Production build for the current OS
make test     # Go unit tests
make check    # Format check, lint, typecheck, and tests
```

The underlying build pipeline lives in `Taskfile.yml` and can be run directly with
`wails3 task <name>` (for example `wails3 task build`).

## Workspace layout

- Root Go module: Wails v3 application and backend services.
- `frontend/`: SolidJS UI embedded in the desktop app.
- `web-extension/`: browser extension integration.
- `frontend/` and `web-extension/` are managed as Bun workspaces with a single
  root `bun.lock` and hoisted `node_modules`.

## Tooling

- `oxlint` for JavaScript/TypeScript linting.
- `oxfmt` for JavaScript/TypeScript formatting.
- `golangci-lint` for Go linting (`make lint`).
- `bun check` for fast native TypeScript type checking.
- `lefthook` for pre-commit formatting/linting and pre-push checks.
