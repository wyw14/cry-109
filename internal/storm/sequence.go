package storm

import (
	"errors"
	"sync"
	"time"

	"github.com/wyw14/cry-109/internal/brake"
	"github.com/wyw14/cry-109/internal/gantry"
	"github.com/wyw14/cry-109/internal/model"
)

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
		return errors.New("anchor is already descending")
	}
	s.state = model.AnchorWaiting
	s.started = at
	s.reason = "waiting for continuous standstill and brake hold"
	return nil
}

func (s *Sequence) Update(at time.Time) model.AnchorState {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state == model.AnchorDescending && (s.stop.Moving() || !s.hold.Proven(at)) {
		s.state = model.AnchorAborted
		s.reason = "motion or brake-hold loss during anchor descent"
		return s.state
	}
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
