package lift

import (
	"errors"
	"sync"

	"github.com/wyw14/cry-109/internal/hoist"
	"github.com/wyw14/cry-109/internal/model"
)

type TandemLift struct {
	mu          sync.RWMutex
	coordinator *hoist.TandemCoordinator
	last        model.SpeedTargets
	enabled     bool
}

func NewTandemLift(coordinator *hoist.TandemCoordinator) *TandemLift {
	return &TandemLift{coordinator: coordinator, enabled: true}
}

func (l *TandemLift) Balance(sample model.LoadSample) (model.SpeedTargets, error) {
	l.mu.RLock()
	enabled := l.enabled
	l.mu.RUnlock()
	if !enabled {
		return model.SpeedTargets{}, errors.New("tandem lifting is disabled")
	}
	targets, err := l.coordinator.Apply(sample)
	if err != nil {
		return model.SpeedTargets{}, err
	}
	l.mu.Lock()
	l.last = targets
	l.mu.Unlock()
	return targets, nil
}

func (l *TandemLift) Disable() {
	l.mu.Lock()
	l.enabled = false
	l.mu.Unlock()
}

func (l *TandemLift) Prune(before uint64) {
	l.coordinator.Balancer().Prune(before)
}

func (l *TandemLift) Stats() (cycles int, commits uint64) {
	return l.coordinator.Balancer().Stats()
}

func (l *TandemLift) Status() (bool, model.SpeedTargets) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.enabled, l.last
}
