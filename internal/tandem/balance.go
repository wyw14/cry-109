package tandem

import (
	"errors"
	"math"
	"sync"

	"github.com/wyw14/cry-109/internal/model"
)

type Balancer struct {
	mu        sync.Mutex
	baseSpeed float64
	gain      float64
	limit     float64
	cycles    map[uint64]model.SpeedTargets
	commits   uint64
}

func NewBalancer(baseSpeed, gain, limit float64) *Balancer {
	return &Balancer{baseSpeed: baseSpeed, gain: gain, limit: limit, cycles: make(map[uint64]model.SpeedTargets)}
}

func (b *Balancer) Coordinate(sample model.LoadSample) (model.SpeedTargets, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if sample.LeftKN <= 0 || sample.RightKN <= 0 {
		return model.SpeedTargets{}, errors.New("both tandem load cells must be healthy")
	}
	total := sample.LeftKN + sample.RightKN
	errorRatio := (sample.LeftKN - sample.RightKN) / total
	correction := clamp(errorRatio*b.gain, -b.limit, b.limit)
	targets := model.SpeedTargets{
		Cycle:       sample.Cycle,
		LeftMPS:     b.baseSpeed - correction,
		RightMPS:    b.baseSpeed + correction,
		Coordinated: true,
	}
	if math.Abs((targets.LeftMPS+targets.RightMPS)-2*b.baseSpeed) > 1e-9 {
		return model.SpeedTargets{}, errors.New("coordinated targets violate speed conservation")
	}
	b.cycles[sample.Cycle] = targets
	b.commits++
	return targets, nil
}

func (b *Balancer) Stats() (cycles int, commits uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.cycles), b.commits
}

func (b *Balancer) Prune(before uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for cycle := range b.cycles {
		if cycle < before {
			delete(b.cycles, cycle)
		}
	}
}

func clamp(value, minimum, maximum float64) float64 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
