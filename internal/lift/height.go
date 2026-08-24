package lift

import (
	"errors"
	"sync"

	"github.com/wyw14/cry-109/internal/model"
)

type HeightEstimator struct {
	mu        sync.RWMutex
	datumM    float64
	config    model.ReevingConfig
	laserBias float64
}

func NewHeightEstimator(datumM float64, config model.ReevingConfig) *HeightEstimator {
	return &HeightEstimator{datumM: datumM, config: config}
}

func (e *HeightEstimator) OnReevingChanged(_, current model.ReevingConfig) {
	e.mu.Lock()
	e.config = current
	e.mu.Unlock()
}

func (e *HeightEstimator) CalibrateLaser(measuredHeight, drumMeters float64) error {
	if measuredHeight < 0 {
		return errors.New("laser height must be non-negative")
	}
	e.mu.Lock()
	predicted := e.datumM - drumMeters/float64(e.config.Factor)
	e.laserBias = measuredHeight - predicted
	e.mu.Unlock()
	return nil
}

func (e *HeightEstimator) Estimate(drumMeters float64) (height float64, config model.ReevingConfig) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.datumM - drumMeters/float64(e.config.Factor) + e.laserBias, e.config
}
