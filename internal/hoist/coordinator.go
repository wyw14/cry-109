package hoist

import (
	"sync"

	"github.com/wyw14/cry-109/internal/model"
	"github.com/wyw14/cry-109/internal/tandem"
)

type TandemCoordinator struct {
	balancer *tandem.Balancer
	mu       sync.RWMutex
	last     model.SpeedTargets
}

func NewTandemCoordinator(balancer *tandem.Balancer) *TandemCoordinator {
	return &TandemCoordinator{balancer: balancer}
}

func (c *TandemCoordinator) Apply(sample model.LoadSample) (model.SpeedTargets, error) {
	targets, err := c.balancer.Coordinate(sample)
	if err != nil {
		return model.SpeedTargets{}, err
	}
	c.mu.Lock()
	c.last = targets
	c.mu.Unlock()
	return targets, nil
}

func (c *TandemCoordinator) Last() model.SpeedTargets {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.last
}

func (c *TandemCoordinator) Balancer() *tandem.Balancer { return c.balancer }
