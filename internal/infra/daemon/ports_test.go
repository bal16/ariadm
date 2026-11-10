package daemon

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeProber struct {
	mu         sync.Mutex
	kinds      map[int]PortKind
	reclaimed  []int
	reclaimErr error
}

func (f *fakeProber) Identify(port int) PortKind {
	f.mu.Lock()
	defer f.mu.Unlock()
	if kind, ok := f.kinds[port]; ok {
		return kind
	}
	return PortFree
}

func (f *fakeProber) Reclaim(port int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reclaimed = append(f.reclaimed, port)
	return f.reclaimErr
}

func (f *fakeProber) reclaimedPorts() []int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]int(nil), f.reclaimed...)
}

func alwaysBindable(int) bool { return true }

func TestSelectEnginePort_PrefersFirstFree(t *testing.T) {
	prober := &fakeProber{}

	port, statuses, err := SelectEnginePort([]int{6800, 6801}, prober, alwaysBindable)

	require.NoError(t, err)
	assert.Equal(t, 6800, port)
	assert.Len(t, statuses, 2)
	assert.Empty(t, prober.reclaimedPorts())
}

func TestSelectEnginePort_ReclaimsOursAndReusesPort(t *testing.T) {
	prober := &fakeProber{kinds: map[int]PortKind{6800: PortOurs}}

	port, _, err := SelectEnginePort([]int{6800, 6801}, prober, alwaysBindable)

	require.NoError(t, err)
	assert.Equal(t, 6800, port)
	assert.Equal(t, []int{6800}, prober.reclaimedPorts())
}

func TestSelectEnginePort_ReclaimsOrphanOnFallbackPort(t *testing.T) {
	prober := &fakeProber{kinds: map[int]PortKind{6803: PortOurs}}

	port, _, err := SelectEnginePort([]int{6800, 6803}, prober, alwaysBindable)

	require.NoError(t, err)
	assert.Equal(t, 6800, port)
	assert.Equal(t, []int{6803}, prober.reclaimedPorts())
}

func TestSelectEnginePort_SkipsForeignWithoutReclaiming(t *testing.T) {
	prober := &fakeProber{kinds: map[int]PortKind{6800: PortForeign}}

	port, _, err := SelectEnginePort([]int{6800, 6801}, prober, alwaysBindable)

	require.NoError(t, err)
	assert.Equal(t, 6801, port)
	assert.Empty(t, prober.reclaimedPorts())
}

func TestSelectEnginePort_UsesNextPortWhenBindFails(t *testing.T) {
	prober := &fakeProber{}
	canBind := func(port int) bool { return port != 6800 }

	port, _, err := SelectEnginePort([]int{6800, 6801}, prober, canBind)

	require.NoError(t, err)
	assert.Equal(t, 6801, port)
}

func TestSelectEnginePort_ReturnsErrorWhenAllForeign(t *testing.T) {
	prober := &fakeProber{kinds: map[int]PortKind{6800: PortForeign, 6801: PortForeign}}

	port, statuses, err := SelectEnginePort([]int{6800, 6801}, prober, alwaysBindable)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoPortAvailable))
	assert.Equal(t, 0, port)
	require.Len(t, statuses, 2)
	assert.Contains(t, err.Error(), "6800:foreign")
	assert.Empty(t, prober.reclaimedPorts())
}
