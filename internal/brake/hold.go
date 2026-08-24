package brake

import (
	"sync"
	"time"

	"github.com/wyw14/cry-109/internal/model"
)

type HoldProof struct {
	mu        sync.RWMutex
	required  time.Duration
	heldSince time.Time
	state     model.BrakeState
}

func NewHoldProof(required time.Duration) *HoldProof {
	return &HoldProof{required: required}
}

func (p *HoldProof) Observe(state model.BrakeState, at time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = state
	if state.Holding && state.Mode == model.BrakeHeld {
		if p.heldSince.IsZero() {
			p.heldSince = at
		}
		return
	}
	p.heldSince = time.Time{}
}

func (p *HoldProof) Proven(at time.Time) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state.Holding && !p.heldSince.IsZero() && at.Sub(p.heldSince) >= p.required
}

func (p *HoldProof) Snapshot() (model.BrakeState, time.Time, time.Duration) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state, p.heldSince, p.required
}
