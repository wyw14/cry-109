package heave

import (
	"sync"
	"time"

	"github.com/wyw14/cry-109/internal/model"
)

type State struct {
	mu        sync.RWMutex
	deckBase  float64
	latest    model.HeaveSample
	hasSample bool
}

func NewState(deckBase float64) *State {
	return &State{deckBase: deckBase}
}

func (s *State) Update(sample model.HeaveSample) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.hasSample && sample.CapturedAt.Before(s.latest.CapturedAt) {
		return
	}
	s.latest = sample
	s.hasSample = true
}

func (s *State) RelativeDeck(at time.Time, maxAge time.Duration) (height, velocity float64, fresh bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.hasSample || at.Sub(s.latest.CapturedAt) > maxAge {
		return s.deckBase, 0, false
	}
	return s.deckBase + s.latest.Meters, s.latest.Velocity, true
}

func (s *State) Snapshot() (model.HeaveSample, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latest, s.hasSample
}
