# AriaDM

A desktop download manager for Aria2. The desktop app is built with Wails v3 and SolidJS, and a browser extension sends downloads from your browser to the app.

![Go 1.26](https://shieldcn.dev/badge/Go-1.26-00ADD8.svg?logo=go&logoColor=white)
![Wails v3](https://shieldcn.dev/badge/Wails-v3-E34F26.svg?logo=wails&logoColor=white)
![SolidJS 1.9](https://shieldcn.dev/badge/SolidJS-1.9-2C4F7C.svg?logo=solid&logoColor=white)
![TypeScript 6.0](https://shieldcn.dev/badge/TypeScript-6.0-3178C6.svg?logo=typescript&logoColor=white)
![Bun 1.4](https://shieldcn.dev/badge/Bun-1.4-000000.svg?logo=bun&logoColor=white)
![Vite 8](https://shieldcn.dev/badge/Vite-8-646CFF.svg?logo=vite&logoColor=white)
![Tailwind CSS v4](https://shieldcn.dev/badge/Tailwind_CSS-v4-06B6D4.svg?logo=tailwindcss&logoColor=white)
![Aria2](https://shieldcn.dev/badge/Aria2-downloads-009688.svg?logo=aria2&logoColor=white)

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
