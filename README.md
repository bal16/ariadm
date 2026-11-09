# AriaDM

A desktop download manager for Aria2. The desktop app is built with Wails v3 and SolidJS, and a browser extension sends downloads from your browser to the app.

## Requirements

- Go 1.26 or newer
- Bun 1.4 or newer
- The Wails v3 CLI
- The webview libraries Wails needs for your operating system

Install the Wails CLI with:

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0
```

## Run it locally

```bash
make deps
make tools
make dev
```

`make deps` installs the Go and JavaScript dependencies. `make tools` adds the linters and git hooks. `make dev` starts the app with live reload.

## Build

```bash
make build
```

The binary is written to `bin/`.

## Other commands

| Command | Purpose |
| --- | --- |
| `make test` | Run the Go tests |
| `make check` | Check formatting, lint, types, and tests |
| `make fmt` | Format Go and frontend code |
| `make clean` | Remove build output |

## Project layout

- `main.go`, `app.go`, `internal/`: desktop app and Aria2 integration
- `frontend/`: SolidJS interface used by the desktop app
- `web-extension/`: browser extension
- `build/`: packaging files and platform assets

The build pipeline lives in `Taskfile.yml`. You can run the tasks directly with `wails3 task <name>`; `make` is the shorter path.

## Before opening a pull request

Run `make check`. It covers formatting, linting, type checking, and the Go test suite.
