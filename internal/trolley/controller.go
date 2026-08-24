package trolley

import (
	"errors"
	"sync"

	"github.com/wyw14/cry-109/internal/model"
)

type Controller struct {
	mu       sync.RWMutex
	position float64
	velocity float64
	active   string
}

func NewController(position float64) *Controller {
	return &Controller{position: position}
}

func (c *Controller) Start(trajectory model.Trajectory) error {
	if !trajectory.Valid || len(trajectory.Points) < 2 {
		return errors.New("trolley trajectory is not executable")
	}
	c.mu.Lock()
	c.active = trajectory.ID
	c.velocity = trajectory.Points[1].VelocityMPS
	c.mu.Unlock()
	return nil
}

func (c *Controller) Stop() {
	c.mu.Lock()
	c.velocity = 0
	c.active = ""
	c.mu.Unlock()
}

func (c *Controller) Snapshot() (position, velocity float64, trajectoryID string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.position, c.velocity, c.active
}

func (c *Controller) Advance(position float64) {
	c.mu.Lock()
	c.position = position
	c.mu.Unlock()
}
