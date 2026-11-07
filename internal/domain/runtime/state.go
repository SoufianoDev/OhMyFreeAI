package runtime

import "fmt"

type SessionState string

const (
	StateUnknown     SessionState = "unknown"
	StateDiscovering SessionState = "discovering"
	StateAvailable   SessionState = "available"
	StateConnecting  SessionState = "connecting"
	StateConnected   SessionState = "connected"
	StateStarting    SessionState = "starting"
	StateReady       SessionState = "ready"
	StateDegraded    SessionState = "degraded"
	StateRecovering  SessionState = "recovering"
	StatePaused      SessionState = "paused"
	StateTerminated  SessionState = "terminated"
	StateUnavailable SessionState = "unavailable"
	StateError       SessionState = "error"
)

var transitions = map[SessionState]map[SessionState]struct{}{
	StateUnknown:      set(StateDiscovering, StateUnavailable, StateError),
	StateDiscovering:  set(StateAvailable, StateUnavailable, StateError),
	StateAvailable:    set(StateConnecting, StatePaused, StateUnavailable, StateError),
	StateConnecting:   set(StateConnected, StateRecovering, StateUnavailable, StateError),
	StateConnected:    set(StateStarting, StateRecovering, StatePaused, StateTerminated, StateError),
	StateStarting:     set(StateReady, StateDegraded, StateRecovering, StateTerminated, StateError),
	StateReady:        set(StateDegraded, StateRecovering, StatePaused, StateTerminated, StateError),
	StateDegraded:     set(StateReady, StateRecovering, StatePaused, StateTerminated, StateError),
	StateRecovering:   set(StateConnecting, StateConnected, StateStarting, StateReady, StateDegraded, StateTerminated, StateUnavailable, StateError),
	StatePaused:       set(StateDiscovering, StateConnecting, StateTerminated, StateUnavailable, StateError),
	StateUnavailable:  set(StateDiscovering, StateTerminated),
	StateError:        set(StateDiscovering, StateTerminated),
	StateTerminated:   {},
}

func set(states ...SessionState) map[SessionState]struct{} {
	result := make(map[SessionState]struct{}, len(states))
	for _, state := range states {
		result[state] = struct{}{}
	}
	return result
}

func CanTransition(from, to SessionState) bool {
	_, ok := transitions[from][to]
	return ok
}

func ValidateTransition(from, to SessionState) error {
	if !CanTransition(from, to) {
		return fmt.Errorf("invalid runtime transition %q -> %q", from, to)
	}
	return nil
}
