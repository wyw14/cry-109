package trolley

import (
	"errors"
	"sync"

	"github.com/wyw14/cry-109/internal/antisway"
	"github.com/wyw14/cry-109/internal/model"
)

type TrajectoryService struct {
	mu      sync.RWMutex
	planner *antisway.Planner
	pending map[string]model.Trajectory
}

func NewTrajectoryService(planner *antisway.Planner) *TrajectoryService {
	return &TrajectoryService{planner: planner, pending: make(map[string]model.Trajectory)}
}

func (s *TrajectoryService) Plan(liftID string, sway model.AntiswayModel, vesselRevision uint64, distance, maxSpeed float64) (model.Trajectory, error) {
	trajectory, err := s.planner.Plan(liftID, sway, distance, maxSpeed)
	if err != nil {
		return model.Trajectory{}, err
	}
	trajectory.VesselRevision = vesselRevision
	s.mu.Lock()
	s.pending[liftID] = trajectory
	s.mu.Unlock()
	return trajectory, nil
}

func (s *TrajectoryService) Pending(liftID string) (model.Trajectory, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	trajectory, ok := s.pending[liftID]
	return trajectory, ok
}

func (s *TrajectoryService) Replace(liftID string, trajectory model.Trajectory) error {
	if !trajectory.Valid || trajectory.LiftID != liftID {
		return errors.New("replacement trajectory is invalid")
	}
	s.mu.Lock()
	s.pending[liftID] = trajectory
	s.mu.Unlock()
	return nil
}

func (s *TrajectoryService) InvalidateGeometry(previous, current model.Geometry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, trajectory := range s.pending {
		if trajectory.GeometryRevision == previous.Revision && previous.Revision != current.Revision {
			trajectory.Valid = false
			trajectory.Reason = "spreader geometry changed before trolley traversal"
			s.pending[key] = trajectory
		}
	}
}

func (s *TrajectoryService) InvalidateVessel(revision uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, trajectory := range s.pending {
		if trajectory.VesselRevision != revision {
			trajectory.Valid = false
			trajectory.Reason = "vessel geometry changed before trolley traversal"
			s.pending[key] = trajectory
		}
	}
}
