package appstate

import "sync"

// EngineState describes the lifecycle of the aria2 download engine.
type EngineState string

const (
	EngineStarting     EngineState = "starting"
	EngineRunning      EngineState = "running"
	EngineDisconnected EngineState = "disconnected"
	EnginePortConflict EngineState = "port_conflict"
	EngineFailed       EngineState = "failed"
	EngineStopped      EngineState = "stopped"
)

// ExtensionState describes whether the local HTTP listener is reachable.
type ExtensionState string

const (
	ExtensionListening   ExtensionState = "listening"
	ExtensionUnavailable ExtensionState = "unavailable"
)

// Status is the runtime snapshot published to the UI and the extension.
type Status struct {
	Engine           EngineState    `json:"engine"`
	EngineMessage    string         `json:"engine_message"`
	EnginePort       int            `json:"engine_port"`
	Extension        ExtensionState `json:"extension"`
	ExtensionMessage string         `json:"extension_message"`
	ExtensionPort    int            `json:"extension_port"`
}

// Hub stores the current status and publishes changes. The emit callback is
// injected so the package stays free of UI dependencies.
type Hub struct {
	mu     sync.Mutex
	status Status
	emit   func(Status)
}

func NewHub(emit func(Status)) *Hub {
	return &Hub{
		status: Status{Engine: EngineStarting},
		emit:   emit,
	}
}

// Snapshot returns a copy of the current status.
func (h *Hub) Snapshot() Status {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.status
}

// SetEngine records an engine state transition, including terminal failures.
func (h *Hub) SetEngine(state EngineState, port int, message string) {
	h.update(func(s *Status) {
		s.Engine = state
		s.EnginePort = port
		s.EngineMessage = message
	})
}

// SetEngineHealth records a live health transition from the telemetry loop. It
// is ignored when startup already recorded a terminal failure, so a dead engine
// keeps showing the reason instead of degrading to a generic disconnect.
func (h *Hub) SetEngineHealth(state EngineState) {
	h.mu.Lock()
	switch h.status.Engine {
	case EngineStarting, EngineRunning, EngineDisconnected:
	default:
		h.mu.Unlock()
		return
	}
	prev := h.status
	h.status.Engine = state
	next := h.status
	h.mu.Unlock()

	if next != prev && h.emit != nil {
		h.emit(next)
	}
}

// SetExtension records the state of the local HTTP listener.
func (h *Hub) SetExtension(state ExtensionState, port int, message string) {
	h.update(func(s *Status) {
		s.Extension = state
		s.ExtensionPort = port
		s.ExtensionMessage = message
	})
}

func (h *Hub) update(fn func(*Status)) {
	h.mu.Lock()
	prev := h.status
	fn(&h.status)
	next := h.status
	h.mu.Unlock()

	if next != prev && h.emit != nil {
		h.emit(next)
	}
}
