package tandem

import (
	"errors"
	"sync"
	"time"

	"github.com/wyw14/cry-109/internal/model"
)

type CellPair struct {
	mu     sync.RWMutex
	latest model.LoadSample
	valid  bool
}

func NewCellPair() *CellPair { return &CellPair{} }

func (p *CellPair) Capture(cycle uint64, leftKN, rightKN float64, at time.Time) (model.LoadSample, error) {
	if leftKN <= 0 || rightKN <= 0 {
		return model.LoadSample{}, errors.New("load cell reading must be positive")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.valid && cycle <= p.latest.Cycle {
		return model.LoadSample{}, errors.New("load cycle must increase")
	}
	p.latest = model.LoadSample{Cycle: cycle, LeftKN: leftKN, RightKN: rightKN, Captured: at}
	p.valid = true
	return p.latest, nil
}

func (p *CellPair) Latest() (model.LoadSample, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.latest, p.valid
}
