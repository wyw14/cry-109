package vessel

import (
	"errors"
	"sync"
	"time"

	"github.com/wyw14/cry-109/internal/model"
)

type Listener func(previous, current model.VesselProfile)

type Store struct {
	mu        sync.RWMutex
	profile   model.VesselProfile
	listeners []Listener
}

func NewStore(initial model.VesselProfile) *Store {
	return &Store{profile: initial}
}

func (s *Store) Subscribe(listener Listener) {
	s.mu.Lock()
	s.listeners = append(s.listeners, listener)
	s.mu.Unlock()
}

func (s *Store) UpdateDraft(draftMeters, deckMeters, hatchMeters float64, at time.Time) (model.VesselProfile, error) {
	if draftMeters <= 0 || hatchMeters <= deckMeters {
		return model.VesselProfile{}, errors.New("vessel geometry is invalid")
	}
	s.mu.Lock()
	previous := s.profile
	current := previous
	current.DraftMeters = draftMeters
	current.DeckMeters = deckMeters
	current.HatchMeters = hatchMeters
	current.Revision++
	current.UpdatedAt = at
	s.profile = current
	listeners := append([]Listener(nil), s.listeners...)
	s.mu.Unlock()
	_ = previous
	_ = listeners
	return current, nil
}

func (s *Store) Current() model.VesselProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profile
}
