package interlock

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-109/internal/model"
)

type Service struct {
	mu        sync.RWMutex
	incidents []model.SafetyIncident
}

func NewService() *Service { return &Service{} }

func (s *Service) Trip(kind, component, message string, at time.Time) model.SafetyIncident {
	incident := model.SafetyIncident{
		ID: uuid.NewString(), Kind: kind, Component: component,
		Message: message, RaisedAt: at,
	}
	s.mu.Lock()
	s.incidents = append(s.incidents, incident)
	s.mu.Unlock()
	return incident
}

func (s *Service) Resolve(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for index := range s.incidents {
		if s.incidents[index].ID == id {
			s.incidents[index].Resolved = true
			return true
		}
	}
	return false
}

func (s *Service) Active() []model.SafetyIncident {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]model.SafetyIncident, 0)
	for _, incident := range s.incidents {
		if !incident.Resolved {
			values = append(values, incident)
		}
	}
	return values
}
