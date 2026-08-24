package lift

import (
	"math"
	"sync"
	"time"

	"github.com/wyw14/cry-109/internal/model"
)

type LandingDetector struct {
	mu            sync.RWMutex
	loadFloorKN   float64
	requiredDwell time.Duration
	unloadedSince time.Time
	landed        bool
	reason        string
}

func NewLandingDetector(loadFloorKN float64, dwell time.Duration) *LandingDetector {
	return &LandingDetector{loadFloorKN: loadFloorKN, requiredDwell: dwell}
}

func (d *LandingDetector) ApplyLoad(observation model.LandingObservation) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if observation.LoadKN >= d.loadFloorKN {
		d.unloadedSince = time.Time{}
		d.landed = false
		d.reason = "load remains on the hoist"
		return false
	}
	contact := math.Abs(observation.HookMeters-observation.DeckMeters) <= observation.ContactToleranceM
	settled := math.Abs(observation.RelativeVelocityMPS) <= 0.05
	if !contact || !settled || !observation.LoadTransferComplete {
		d.unloadedSince = time.Time{}
		d.landed = false
		d.reason = "unload is not accompanied by physical deck contact"
		return false
	}
	if d.unloadedSince.IsZero() {
		d.unloadedSince = observation.At
	}
	d.landed = observation.At.Sub(d.unloadedSince) >= d.requiredDwell
	if d.landed {
		d.reason = ""
	} else {
		d.reason = "physical contact has not met landing dwell"
	}
	return d.landed
}

func (d *LandingDetector) Snapshot() (bool, string, time.Time) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.landed, d.reason, d.unloadedSince
}
