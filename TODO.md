Here is your updated **Task / To-Do List** with the new tasks for implementing the port strategies for both the `aria2c` RPC daemon and the Local HTTP API integrated cleanly into Phase 3 and Phase 4.

---

# Todo List

## PROJECT SETUP AND BACKEND PHASE

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

- [ ] **Task 2.4 (TDD - Mass Operations):**
- [ ] **RED**: Write test cases for `TestPauseAll_Success` and `TestResumeAll_Success` to ensure mass status updates validate cleanly against the service layer.
- [ ] **GREEN**: Implement the `PauseAllTasks()` and `ResumeAllTasks()` methods to dispatch execution commands out to the `aria2.pauseAll` and `aria2.unpauseAll` RPC endpoints.

---

### 💻 PHASE 3: Infrastructure Implementation (Technical Delivery)

_Target: Connect the domain interfaces to concrete databases, local files, and system processes._

- [x] **Task 3.1 (Daemon Manager):** Implement `os/exec` logic in `infrastructure/daemon/` to safely _start/stop_ the headless `aria2c` process in the background.
- [ ] **Task 3.1.1 (Port Collision Mitigation):** Implement runtime checks before spawning the daemon process. If the target RPC port is already bound, perform a handshake validation check to determine if it is a lingering instance from our own app or a foreign asset.
- [ ] **Task 3.1.2 (Daemon Orphan Recovery):** Design a hot-recovery routine to automatically re-attach the runtime pipeline to the active daemon process tree if an orphan instance is already running safely, rather than throwing a hard port collision panic.
- [ ] **Task 3.1.3 (Aria2c RPC Port Assignment Strategy):** Implement an isolated, unique port strategy for the `aria2c` daemon (e.g., binding explicitly to `127.0.0.1` on a safe static custom port like `56800` instead of the noisy default `6800`) to guarantee zero interference with external download tools.
- [x] **Task 3.2 (JSON Config Repo):** Implement `JSONConfigRepository` to read and write the `config.json` file inside the OS-specific AppData/Config folder.
- [x] **Task 3.3 (SQLite Repo):** Implement `SQLiteRepository` using a pure Go driver (`modernc.org/sqlite`) to store `Task` records permanently.
- [x] **Task 3.3.1:** Write Integration Tests for SQLiteTaskRepository against a real temporary database file.
- [x] **Task 3.4 (RPC Client):** Implement the actual WebSocket/HTTP client in `infrastructure/rpc/` to dispatch JSON-RPC commands to the `aria2c` port.

---

### 🔌 PHASE 4: Ingress Layer (Entry Points)

_Target: Prepare the backend to intercept commands from both the browser extension and the Wails frontend._

- [x] **Task 4.1 (Local HTTP Server):** Build a lightweight REST API server using Go’s built-in `net/http` package to catch download links forwarded by the browser extension (Chrome/Firefox).
- [ ] **Task 4.1.1 (Local API Port Scanning Fallback Strategy):** Upgrade the server boot sequence to implement a progressive port allocation range (e.g., trying `58300`, falling back sequentially up to `58305` if occupied). Ensure it binds strictly to `127.0.0.1` and handles dynamic health checks so the browser extension can find it reliably.
- [ ] **Task 4.1.2 (CORS & Security Sanitization):** Enforce strict CORS policies on the listener socket to exclusively permit incoming execution targets matching your dedicated extension ID origin (`chrome-extension://...`), blocking unauthenticated cross-origin exploit scripts.

---

## FRONTEND & UI INTEGRATION PHASE

### 🎨 PHASE 5: Frontend Setup & Styling (UI Foundation)

- [x] **Task 5.1:** Sync child compilation workspaces inside the `frontend/` directory.
- [x] **Task 5.2:** Set up stable Tailwind v3 compiler + PostCSS asset processing pipelines.
- [x] **Task 5.3:** Resolve TypeScript solution-style compiler path alias mappings (`~/*`).
- [x] **Task 5.4:** Verify `wails dev` successfully generates types and maps models.

---

### 🎛️ PHASE 6: UI Components & Data Binding (Wiring the Controls)

- [x] **Task 6.1 (Settings Screen):** Build a dedicated settings panel/modal. Bind its inputs to `GetApplicationConfig` and `SaveApplicationConfig` exposed by the `WailsBridge`.
- [x] **Task 6.2 (Add Task Dialog):** Design a clean URL input modal to accept new downloads. Link the submission handler to your backend `TriggerNewDownload` method.
- [x] **Task 6.3 (Download List Component):** Create a robust task dashboard row displaying the file name, sizes, a progress bar, and control buttons that fire off `ToggleTaskPauseState`.
- [ ] **Task 6.4 (Toolbar Global Action Wiring):** Wire the action buttons for "Pause All" and "Resume All" inside the top-level application navigation toolbar of `App.tsx` to call your new mass-operation bridge methods.

---

### ⚡ PHASE 7: Real-Time Progress Synchronization (State Update Loop)

- [x] **Task 7.1 (Backend Ticker):** Setup a background goroutine loop in Go (using `time.Ticker` firing every 500ms) that queries `aria2.tellActive` and broadcasts active progress map blocks down to the window using **Wails Events** (`wails.EventsEmit`).
- [x] **Task 7.2 (Frontend Listener):** Wire up a global event hook inside SolidJS (`wails.EventsOn`) to catch the incoming stream payloads and feed them straight into reactive SolidJS _Signals_ for fine-grained DOM manipulation.

---

### 🗑️ PHASE 8: Task Management (Delete & Cleanup)

- [x] **Task 8.1 (Delete Task):** Implement full delete pipeline — `TaskRepository.Delete`, `aria2.remove` for active/paused tasks, `aria2.removeDownloadResult` for completed/errored entries, `WailsBridge.DeleteTask` exposed to Wails, and frontend `handleDelete` wired to the real backend call.

---

### ✨ PHASE 9: UI Polish & UX Features

- [x] **Task 9.1 (Delete Confirmation Dialog):** Add a robust confirmation dialog before dropping tasks.
- [x] **Task 9.2 (Global Keybinds):** Wire up `Ctrl+N` (New Task) and `Ctrl+P` (Preferences) at the global `App.tsx` layer.
- [x] **Task 9.3 (Close/Quit Confirmation):** Catch the window close event via `OnBeforeClose` to show a prompt dialog.
- [x] **Task 9.4 (Background Functionality):** Implement seamless running in the background. Hiding the window keeps the active daemon alive, and configuring Wails `SingleInstanceLock` ensures that reopening the app summons the existing process back to the foreground seamlessly.
- [ ] **Task 9.5 (Sleeper Mode - Go Memory Trimming):** Intercept the `OnWindowHide` / `OnWindowMinimize` lifecycle event hooks in the Wails backend framework. Execute a routine to explicitly invoke the Go garbage collector via `runtime/debug.FreeOSMemory()` and `runtime.GC()` to drop unused virtual heap from the system memory instantly.
- [ ] **Task 9.6 (Sleeper Mode - WebKit Cache Drop):** Dispatch an optimization signal down to the active Webview layer during background transitions to dump GPU rendering texture assets, and throttle down or freeze the SolidJS polling interval entirely to keep RAM usage near flat while running hidden.

---

### 🧩 PHASE 10: Browser Web Extension Development

_Target: Create a lightweight, secure Manifest V3 browser extension to capture download links natively from web interactions._

- [x] **Task 10.1 (Extension Boilerplate):** Build a browser extension directory path incorporating a compliant `manifest.json`, a background service worker (`background.js`), and a standard configuration popup panel asset.
- [ ] **Task 10.2 (Context Menu Interceptor):** Register a dedicated custom entry in the browser right-click context menu titled "Download with AriaDM" to intercept destination parameters like `linkUrl` or `srcUrl`.
- [x] **Task 10.3 (Secure Handshake API Linker):** Implement structural JSON payload transfers using standard asynchronous `fetch` API methods inside the background service worker pointing straight to the local AriaDM HTTP REST port selection range (`127.0.0.1:58300-58305`).
- [ ] **Task 10.4 (Auth Token Exchange):** Secure the cross-process pipe by injecting an ephemeral, static authorization token within the request headers to ensure the Local API rejects unauthenticated third-party command execution strings.
