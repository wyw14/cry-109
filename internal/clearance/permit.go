package clearance

import (
	"fmt"
	"sync"

	"github.com/wyw14/cry-109/internal/model"
)

type Decision struct {
	Allowed         bool    `json:"allowed"`
	Reason          string  `json:"reason,omitempty"`
	HeightM         float64 `json:"height_m"`
	RequiredHeightM float64 `json:"required_height_m"`
	ReevingRevision uint64  `json:"reeving_revision"`
}

type Permit struct {
	mu             sync.RWMutex
	currentReeving model.ReevingConfig
}

func NewPermit(config model.ReevingConfig) *Permit {
	return &Permit{currentReeving: config}
}

func (p *Permit) OnReevingChanged(_, current model.ReevingConfig) {
	p.mu.Lock()
	p.currentReeving = current
	p.mu.Unlock()
}

func (p *Permit) Check(height float64, estimateConfig model.ReevingConfig, required float64) Decision {
	p.mu.RLock()
	current := p.currentReeving
	p.mu.RUnlock()
	decision := Decision{
		HeightM: height, RequiredHeightM: required,
		ReevingRevision: estimateConfig.Revision,
	}
	if estimateConfig.Revision != current.Revision || estimateConfig.Factor != current.Factor {
		decision.Reason = "height estimate uses stale reeving configuration"
		return decision
	}
	if height < required {
		decision.Reason = fmt.Sprintf("hook height %.2f is below clearance %.2f", height, required)
		return decision
	}
	decision.Allowed = true
	return decision
}
