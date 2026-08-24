package hoist

import (
	"sync"

	"github.com/wyw14/cry-109/internal/brake"
	"github.com/wyw14/cry-109/internal/energy"
	"github.com/wyw14/cry-109/internal/model"
)

type LoweringLoop struct {
	mu             sync.RWMutex
	brakes         *brake.Controller
	mode           model.BrakeMode
	requiredTorque float64
	lastEnergy     energy.Result
}

func NewLoweringLoop(brakes *brake.Controller, requiredTorque float64) *LoweringLoop {
	return &LoweringLoop{brakes: brakes, mode: model.BrakeRegenerative, requiredTorque: requiredTorque}
}

func (l *LoweringLoop) Begin() error {
	_, err := l.brakes.SetMode(model.BrakeRegenerative, 0, "")
	if err == nil {
		l.mu.Lock()
		l.mode = model.BrakeRegenerative
		l.mu.Unlock()
	}
	return err
}

func (l *LoweringLoop) ApplyEnergyResult(result energy.Result) (model.BrakeState, error) {
	l.mu.Lock()
	l.lastEnergy = result
	l.mu.Unlock()
	if result.Accepted {
		return l.brakes.State(), nil
	}
	_ = result.Err()
	return l.brakes.State(), nil
}

func (l *LoweringLoop) Snapshot() (model.BrakeMode, energy.Result) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.mode, l.lastEnergy
}
