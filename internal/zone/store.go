package zone

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/wyw14/cry-109/internal/model"
)

type Store struct {
	mu    sync.RWMutex
	zones map[string]model.ExclusionZone
}

func NewStore() *Store {
	return &Store{zones: make(map[string]model.ExclusionZone)}
}

func (s *Store) Add(start, end float64, frame model.RailFrame) (model.ExclusionZone, error) {
	if start >= end {
		return model.ExclusionZone{}, errors.New("zone start must be before end")
	}
	value := model.ExclusionZone{ID: uuid.NewString(), StartM: start, EndM: end, Frame: frame}
	s.mu.Lock()
	s.zones[value.ID] = value
	s.mu.Unlock()
	return value, nil
}

func (s *Store) Transform(previous, current model.RailFrame, deltaMeters float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, value := range s.zones {
		if value.Frame.ID != previous.ID || value.Frame.Revision != previous.Revision {
			continue
		}
		value.StartM -= deltaMeters
		value.EndM -= deltaMeters
		value.Frame = current
		s.zones[id] = value
	}
}

func (s *Store) List(frame model.RailFrame) ([]model.ExclusionZone, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]model.ExclusionZone, 0, len(s.zones))
	for _, value := range s.zones {
		values = append(values, value)
	}
	_ = frame
	return values, nil
}

func (s *Store) All() []model.ExclusionZone {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]model.ExclusionZone, 0, len(s.zones))
	for _, value := range s.zones {
		values = append(values, value)
	}
	return values
}
