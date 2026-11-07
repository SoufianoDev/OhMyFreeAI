package runtime

import (
	"errors"
	"time"
)

var ErrStaleGeneration = errors.New("runtime session generation is stale")

type Session struct {
	ID         string
	Generation uint64
	State      SessionState
	UpdatedAt  time.Time
}

func NewSession(id string, generation uint64, now time.Time) Session {
	return Session{ID: id, Generation: generation, State: StateUnknown, UpdatedAt: now.UTC()}
}

func (s Session) Transition(next SessionState, generation uint64, now time.Time) (Session, error) {
	if generation != s.Generation {
		return s, ErrStaleGeneration
	}
	if err := ValidateTransition(s.State, next); err != nil {
		return s, err
	}
	s.State = next
	s.UpdatedAt = now.UTC()
	return s, nil
}
