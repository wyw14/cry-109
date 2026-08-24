package gantry

import (
	"errors"
	"sync"
	"time"
)

type Controller struct {
	mu       sync.RWMutex
	position float64
	velocity float64
	proof    *StopProof
}

func NewController(position float64, proof *StopProof) *Controller {
	return &Controller{position: position, proof: proof}
}

func (c *Controller) CommandVelocity(velocity float64, at time.Time) error {
	if velocity < -1.5 || velocity > 1.5 {
		return errors.New("gantry velocity exceeds travel limit")
	}
	c.mu.Lock()
	c.velocity = velocity
	c.mu.Unlock()
	c.proof.Observe(velocity, at)
	return nil
}

func (c *Controller) Advance(position float64) {
	c.mu.Lock()
	c.position = position
	c.mu.Unlock()
}

func (c *Controller) Snapshot() (position, velocity float64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.position, c.velocity
}

func (c *Controller) StopProof() *StopProof { return c.proof }
