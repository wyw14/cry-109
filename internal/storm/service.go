package storm

import (
	"errors"
	"sync"
	"time"

	"github.com/wyw14/cry-109/internal/brake"
	"github.com/wyw14/cry-109/internal/gantry"
	"github.com/wyw14/cry-109/internal/model"
)

type Service struct {
	mu       sync.RWMutex
	windMPS  float64
	sequence *Sequence
	gantry   *gantry.Controller
	brake    *brake.Controller
}

func NewService(sequence *Sequence, controller *gantry.Controller, brakeController *brake.Controller) *Service {
	return &Service{sequence: sequence, gantry: controller, brake: brakeController}
}

func (s *Service) ObserveWind(speedMPS float64, at time.Time) (model.AnchorState, error) {
	s.mu.Lock()
	s.windMPS = speedMPS
	s.mu.Unlock()
	if speedMPS < 18 {
		state, _, _ := s.sequence.Snapshot()
		return state, nil
	}
	if err := s.gantry.CommandVelocity(0, at); err != nil {
		return model.AnchorRaised, err
	}
	if _, err := s.brake.SetMode(model.BrakeFriction, 90, "storm stop"); err != nil {
		return model.AnchorRaised, err
	}
	if err := s.sequence.Request(at); err != nil && !errors.Is(err, errors.New("anchor is already descending")) {
		return model.AnchorRaised, err
	}
	state, _, _ := s.sequence.Snapshot()
	return state, nil
}

func (s *Service) Wind() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.windMPS
}

func (s *Service) Sequence() *Sequence { return s.sequence }
