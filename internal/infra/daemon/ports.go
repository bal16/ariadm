package daemon

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// PortKind is the result of classifying a candidate RPC port.
type PortKind int

const (
	PortFree PortKind = iota
	PortOurs
	PortForeign
)

// PortStatus pairs a candidate port with its classification.
type PortStatus struct {
	Port int
	Kind PortKind
}

func (s PortStatus) String() string {
	switch s.Kind {
	case PortOurs:
		return fmt.Sprintf("%d:ours", s.Port)
	case PortForeign:
		return fmt.Sprintf("%d:foreign", s.Port)
	default:
		return fmt.Sprintf("%d:free", s.Port)
	}
}

// Prober classifies ports and shuts down instances that belong to this app.
type Prober interface {
	Identify(port int) PortKind
	Reclaim(port int) error
}

// ErrNoPortAvailable reports that every candidate was occupied by a process we
// do not own or could not bind.
var ErrNoPortAvailable = errors.New("no usable aria2 RPC port in range")

// SelectEnginePort reclaims every daemon that is provably ours across the range
// and returns the first port a new daemon can use.
func SelectEnginePort(ports []int, prober Prober, canBind func(int) bool) (int, []PortStatus, error) {
	statuses := make([]PortStatus, len(ports))

	var wg sync.WaitGroup
	for i, port := range ports {
		wg.Add(1)
		go func(i, port int) {
			defer wg.Done()
			statuses[i] = PortStatus{Port: port, Kind: prober.Identify(port)}
		}(i, port)
	}
	wg.Wait()

	// Reclaim every orphan, even one left on a fallback port by an earlier run.
	// Leaving it alive would let two engines coexist.
	reclaimed := make(map[int]bool)
	for _, status := range statuses {
		if status.Kind == PortOurs {
			reclaimed[status.Port] = true
			_ = prober.Reclaim(status.Port)
		}
	}

	for _, status := range statuses {
		if status.Kind == PortForeign {
			continue
		}
		// Reclaimed ports need a moment to be bindable again, and a port may
		// have been taken since the classification pass.
		if waitForBind(status.Port, canBind, reclaimed[status.Port]) {
			return status.Port, statuses, nil
		}
	}

	return 0, statuses, fmt.Errorf("%w: %s", ErrNoPortAvailable, formatPortStatuses(statuses))
}

// waitForBind reports whether port can be bound. A reclaimed port is polled for
// a while because the previous process may not have released it yet.
func waitForBind(port int, canBind func(int) bool, reclaimed bool) bool {
	if canBind(port) {
		return true
	}
	if !reclaimed {
		return false
	}

	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(250 * time.Millisecond)
		if canBind(port) {
			return true
		}
	}
	return false
}

func formatPortStatuses(statuses []PortStatus) string {
	if len(statuses) == 0 {
		return "empty range"
	}

	formatted := ""
	for i, status := range statuses {
		if i > 0 {
			formatted += ", "
		}
		formatted += status.String()
	}
	return formatted
}
