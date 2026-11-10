package appstate_test

import (
	"testing"

	"github.com/bal16/ariadm/internal/appstate"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHub_StartsAsStarting(t *testing.T) {
	hub := appstate.NewHub(nil)

	assert.Equal(t, appstate.EngineStarting, hub.Snapshot().Engine)
}

func TestHub_EmitsOnChange(t *testing.T) {
	var emitted []appstate.Status
	hub := appstate.NewHub(func(status appstate.Status) {
		emitted = append(emitted, status)
	})

	hub.SetEngine(appstate.EngineRunning, 6800, "")

	require.Len(t, emitted, 1)
	assert.Equal(t, appstate.EngineRunning, emitted[0].Engine)
	assert.Equal(t, 6800, emitted[0].EnginePort)
}

func TestHub_DoesNotEmitWhenUnchanged(t *testing.T) {
	count := 0
	hub := appstate.NewHub(func(appstate.Status) {
		count++
	})

	hub.SetEngine(appstate.EngineRunning, 6800, "")
	hub.SetEngine(appstate.EngineRunning, 6800, "")

	assert.Equal(t, 1, count)
}

func TestHub_HealthKeepsTerminalFailure(t *testing.T) {
	hub := appstate.NewHub(nil)

	hub.SetEngine(appstate.EnginePortConflict, 0, "all ports busy")
	hub.SetEngineHealth(appstate.EngineDisconnected)

	snapshot := hub.Snapshot()
	assert.Equal(t, appstate.EnginePortConflict, snapshot.Engine)
	assert.Equal(t, "all ports busy", snapshot.EngineMessage)
}

func TestHub_HealthUpdatesRunningEngine(t *testing.T) {
	hub := appstate.NewHub(nil)

	hub.SetEngine(appstate.EngineRunning, 6801, "")
	hub.SetEngineHealth(appstate.EngineDisconnected)

	snapshot := hub.Snapshot()
	assert.Equal(t, appstate.EngineDisconnected, snapshot.Engine)
	assert.Equal(t, 6801, snapshot.EnginePort)
}

func TestHub_SetExtension(t *testing.T) {
	hub := appstate.NewHub(nil)

	hub.SetExtension(appstate.ExtensionUnavailable, 0, "port range exhausted")

	snapshot := hub.Snapshot()
	assert.Equal(t, appstate.ExtensionUnavailable, snapshot.Extension)
	assert.Equal(t, "port range exhausted", snapshot.ExtensionMessage)

	hub.SetExtension(appstate.ExtensionListening, 10002, "")

	snapshot = hub.Snapshot()
	assert.Equal(t, appstate.ExtensionListening, snapshot.Extension)
	assert.Equal(t, 10002, snapshot.ExtensionPort)
	assert.Empty(t, snapshot.ExtensionMessage)
}
