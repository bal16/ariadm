# Todo List

Here is the updated **Task / To-Do List** for your download manager project, translated into English and tailored to your clean architecture + TDD + SolidJS workflow.

---

### 📋 PHASE 0: Project Setup & Environment (Foundation)

_Target: Project structure initialized and ready to compile test suites._

- [x] **Task 0.1:** Initialize a blank Wails project using the TypeScript vanilla template:

```bash
wails init -n go-aria2-dm -template vanilla-ts

```

- [x] **Task 0.2 (Official SolidJS Setup):** Delete the default contents of the `frontend/` directory, then re-generate it using the official Vite initializer:

```bash
npm create vite@latest frontend -- --template solid-ts

```

- [x] **Task 0.3:** Install the `testify` testing library in your Go project root:

```bash
go get github.com/stretchr/testify

```

- [x] **Task 0.4:** Create the internal backend directory tree in the project root:

```text
internal/
├── domain/
│   ├── config/
│   └── task/
├── infrastructure/
│   ├── daemon/
│   ├── rpc/
│   └── database/
└── ingress/

```

---

### ⚙️ PHASE 1: TDD ConfigService (First Module)

_Target: Manage application preferences (`config.json`) and support live synchronization with aria2c._

- [x] **Task 1.1:** Define the `AppConfig` struct in `internal/domain/config/entity.go` (fields: `DefaultDownloadPath`, `SpeedLimit`, etc.).
- [x] **Task 1.2:** Define the `ConfigRepository` interface in `internal/domain/config/repository.go`.
- [x] **Task 1.3 (TDD - GetConfig):**
- [x] **RED**: Create `service_test.go`, write a manual `ConfigRepositoryMock`, and draft the `TestGetConfig_Success` case.
- [x] **GREEN**: Create `service.go`, implement the `ConfigService` struct and its `GetConfig()` method. Run `go test ./...` until it passes.

- [x] **Task 1.4 (TDD - UpdateSettings):**
- [x] **RED**: Write the `TestUpdateSettings_Success` case, expecting data persistence to a JSON file _and_ an RPC call to aria2c.
- [x] **GREEN**: Implement the `UpdateSettings()` method in your service.

---

### 📥 PHASE 2: TDD TaskService (Core Download Logic)

_Target: Handle the download queue logic, link validation, and task control commands._

- [x] **Task 2.1:** Create the `Task` struct (Entity) and `TaskRepository` interface in `internal/domain/task/`.
- [x] **Task 2.2 (TDD - DownloadFile):**
- [x] **RED**: Write the `TestDownloadFile_Success` case. Manually mock `TaskRepository` and `Aria2Client`. Map out the expected sequence: Get path from config ➡️ Fire RPC ➡️ Save to SQLite.
- [x] **GREEN**: Implement the `DownloadFile(url string)` method in `TaskService`.
- [x] **Refactor**: Add URL format validation (Regex) before triggering the RPC layer.

- [x] **Task 2.3 (TDD - TogglePause):**
- [x] **RED**: Write test scenarios for both pausing and resuming tasks based on the current GID status in the database.
- [x] **GREEN**: Implement the `TogglePauseTask(taskID string)` method.

---

### 💻 PHASE 3: Infrastructure Implementation (Technical Delivery)

_Target: Connect the domain interfaces to concrete databases, local files, and system processes._

- [x] **Task 3.1 (Daemon Manager):** Implement `os/exec` logic in `infrastructure/daemon/` to safely _start/stop_ the headless `aria2c` process in the background.
- [x] **Task 3.2 (JSON Config Repo):** Implement `JSONConfigRepository` to read and write the `config.json` file inside the OS-specific AppData/Config folder.
- [x] **Task 3.3 (SQLite Repo):** Implement `SQLiteRepository` using a pure Go driver (`modernc.org/sqlite`) to store `Task` records permanently.
- [x] **Task 3.3.1:** Write Integration Tests for SQLiteTaskRepository against a real temporary database file.
- [x] **Task 3.4 (RPC Client):** Implement the actual WebSocket client in `infrastructure/rpc/` to dispatch JSON-RPC commands to the `aria2c` port (6800).

---

### 🔌 PHASE 4: Ingress Layer (Entry Points)

_Target: Prepare the backend to intercept commands from both the browser extension and the Wails frontend._

- [x] **Task 4.1 (Local HTTP Server):** Build a lightweight REST API server using Go’s built-in `net/http` package to catch download links forwarded by the browser extension (Chrome/Firefox).
- [x] **Task 4.2 (Wails Bridge Controller):** Create `ingress/wailsbridge/controller.go`. This file will expose clean, decoupled methods wrapping `TaskService` and `ConfigService` so Wails can bind them directly to the frontend.
