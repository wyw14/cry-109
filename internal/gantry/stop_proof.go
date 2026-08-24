package gantry

import (
	"math"
	"sync"
	"time"
)

type StopProof struct {
	mu         sync.RWMutex
	tolerance  float64
	required   time.Duration
	stillSince time.Time
	velocity   float64
}

func NewStopProof(tolerance float64, required time.Duration) *StopProof {
	return &StopProof{tolerance: tolerance, required: required}
}

func (p *StopProof) Observe(velocity float64, at time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.velocity = velocity
	if math.Abs(velocity) <= p.tolerance {
		if p.stillSince.IsZero() {
			p.stillSince = at
		}
		return
	}
	p.stillSince = time.Time{}
}

func (p *StopProof) Proven(at time.Time) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_ = at
	return math.Abs(p.velocity) <= p.tolerance
}

func (p *StopProof) Moving() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return math.Abs(p.velocity) > p.tolerance
}

func (p *StopProof) Snapshot() (velocity float64, stillSince time.Time, required time.Duration) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.velocity, p.stillSince, p.required
}
