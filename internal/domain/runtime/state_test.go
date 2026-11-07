package runtime

import (
	"errors"
	"testing"
	"time"
)

func TestSessionTransitionAndGeneration(t *testing.T) {
	now := time.Date(2026, 11, 7, 15, 0, 0, 0, time.FixedZone("UTC+1", 3600))
	s := NewSession("session-1", 4, now)
	if s.State != StateUnknown || s.UpdatedAt.Location() != time.UTC {
		t.Fatal("new session must start unknown with UTC timestamp")
	}
	s, err := s.Transition(StateDiscovering, 4, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if s.State != StateDiscovering {
		t.Fatalf("state = %q", s.State)
	}
	if _, err := s.Transition(StateAvailable, 3, now); !errors.Is(err, ErrStaleGeneration) {
		t.Fatalf("expected stale generation error, got %v", err)
	}
}

func TestRejectsInvalidTransition(t *testing.T) {
	if CanTransition(StateTerminated, StateReady) {
		t.Fatal("terminated session must not become ready")
	}
	if err := ValidateTransition(StateTerminated, StateReady); err == nil {
		t.Fatal("expected invalid transition error")
	}
}
