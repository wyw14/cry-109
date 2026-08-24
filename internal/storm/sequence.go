package storm

import (
	"errors"
	"sync"
	"time"

	"github.com/wyw14/cry-109/internal/brake"
	"github.com/wyw14/cry-109/internal/gantry"
	"github.com/wyw14/cry-109/internal/model"
)

// ErrAnchorAlreadyDescending is returned by Request when the anchor pin is
// already traveling into the socket. It is a sentinel so callers can suppress
// the benign re-trigger race with errors.Is instead of string matching.
var ErrAnchorAlreadyDescending = errors.New("anchor is already descending")

type Sequence struct {
	mu      sync.RWMutex
	state   model.AnchorState
	stop    *gantry.StopProof
	hold    *brake.HoldProof
	started time.Time
	reason  string
}

func NewSequence(stop *gantry.StopProof, hold *brake.HoldProof) *Sequence {
	return &Sequence{state: model.AnchorRaised, stop: stop, hold: hold}
}

func (s *Sequence) Request(at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state == model.AnchorDescending {
		return ErrAnchorAlreadyDescending
	}
	s.state = model.AnchorWaiting
	s.started = at
	s.reason = "waiting for continuous standstill and brake hold"
	return nil
}

func (s *Sequence) Update(at time.Time) model.AnchorState {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Abort on loss of either precondition while the pin is already traveling:
	// sustained standstill must remain proven AND the brake-hold proof must
	// remain proven. A momentary sub-tolerance blip is not enough to keep the
	// crane pinned in place while the pin is between the deck and the socket.
	if s.state == model.AnchorDescending && (!s.stop.Proven(at) || !s.hold.Proven(at)) {
		s.state = model.AnchorAborted
		s.reason = "standstill or brake-hold proof lost during anchor descent"
		return s.state
	}
	// Drop the pin only once continuous standstill AND sustained brake hold are
	// both proven. Building pressure is not instantaneous, so requiring the
	// brake-hold proof alongside the stop proof guarantees the crane cannot be
	// nudged by the next gust while the pin is entering the socket.
	if s.state == model.AnchorWaiting && s.stop.Proven(at) && s.hold.Proven(at) {
		s.state = model.AnchorDescending
		s.reason = ""
	}
	return s.state
}

func (s *Sequence) Secure() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state != model.AnchorDescending {
		return errors.New("anchor cannot be secured before descent")
	}
	s.state = model.AnchorSecured
	return nil
}

func (s *Sequence) Reset() {
	s.mu.Lock()
	s.state = model.AnchorRaised
	s.reason = ""
	s.mu.Unlock()
}

func (s *Sequence) Snapshot() (model.AnchorState, string, time.Time) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state, s.reason, s.started
}
