package lift

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-109/internal/journal"
	"github.com/wyw14/cry-109/internal/model"
)

type Service struct {
	repository *journal.Repository
}

func NewService(repository *journal.Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Start(containerID, sessionID string, geometryRevision uint64, at time.Time) (model.Lift, error) {
	if containerID == "" || sessionID == "" {
		return model.Lift{}, errors.New("container and engage session are required")
	}
	lift := model.Lift{
		ID:               uuid.NewString(),
		ContainerID:      containerID,
		Phase:            model.LiftApproaching,
		SessionID:        sessionID,
		GeometryRevision: geometryRevision,
		StartedAt:        at,
		UpdatedAt:        at,
	}
	event, err := model.NewEvent(uuid.NewString(), lift.ID, "lift.started", at, lift)
	if err != nil {
		return model.Lift{}, err
	}
	if err := s.repository.PutLift(lift, event); err != nil {
		return model.Lift{}, err
	}
	return lift, nil
}

func (s *Service) Advance(id string, phase model.LiftPhase, at time.Time) (model.Lift, error) {
	current, ok := s.repository.Lift(id)
	if !ok {
		return model.Lift{}, errors.New("lift not found")
	}
	next, err := current.Advance(phase, at)
	if err != nil {
		return model.Lift{}, err
	}
	event, err := model.NewEvent(uuid.NewString(), id, "lift.phase", at, next)
	if err != nil {
		return model.Lift{}, err
	}
	if err := s.repository.PutLift(next, event); err != nil {
		return model.Lift{}, err
	}
	return next, nil
}

func (s *Service) List() []model.Lift {
	return s.repository.Lifts()
}

func (s *Service) Get(id string) (model.Lift, bool) {
	return s.repository.Lift(id)
}
