package hoist

import (
	"errors"
	"sync"
	"time"

	"github.com/wyw14/cry-109/internal/model"
)

type ReevingListener func(previous, current model.ReevingConfig)

type ReevingService struct {
	mu        sync.RWMutex
	config    model.ReevingConfig
	listeners []ReevingListener
}

func NewReevingService(factor int, at time.Time) *ReevingService {
	return &ReevingService{config: model.ReevingConfig{Factor: factor, Revision: 1, UpdatedAt: at}}
}

func (s *ReevingService) Subscribe(listener ReevingListener) {
	s.mu.Lock()
	s.listeners = append(s.listeners, listener)
	s.mu.Unlock()
}

func (s *ReevingService) Apply(factor int, at time.Time) (model.ReevingConfig, error) {
	if factor != 4 && factor != 8 {
		return model.ReevingConfig{}, errors.New("reeving factor must be four or eight")
	}
	s.mu.Lock()
	previous := s.config
	if factor == previous.Factor {
		s.mu.Unlock()
		return previous, nil
	}
	current := model.ReevingConfig{Factor: factor, Revision: previous.Revision + 1, UpdatedAt: at}
	s.config = current
	listeners := append([]ReevingListener(nil), s.listeners...)
	s.mu.Unlock()
	for _, listener := range listeners {
		listener(previous, current)
	}
	return current, nil
}

func (s *ReevingService) Current() model.ReevingConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *ReevingService) HookTravel(drumMeters float64) float64 {
	config := s.Current()
	return drumMeters / float64(config.Factor)
}
