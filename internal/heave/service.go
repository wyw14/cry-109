package heave

import (
	"time"

	"github.com/wyw14/cry-109/internal/model"
)

type Service struct {
	fusion *Fusion
	state  *State
}

func NewService(deckBase float64) *Service {
	return &Service{
		fusion: NewFusion(50*time.Millisecond, 150*time.Millisecond),
		state:  NewState(deckBase),
	}
}

func (s *Service) IngestHeave(sample model.HeaveSample, now time.Time) (model.Compensation, bool) {
	s.state.Update(sample)
	return s.fusion.AddHeave(sample, now)
}

func (s *Service) IngestHoist(sample model.HoistSample, now time.Time) (model.Compensation, bool) {
	return s.fusion.AddHoist(sample, now)
}

func (s *Service) Fusion() *Fusion { return s.fusion }
func (s *Service) State() *State   { return s.state }
