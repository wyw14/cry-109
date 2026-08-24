package heave

import (
	"sync"
	"time"

	"github.com/wyw14/cry-109/internal/model"
	"github.com/wyw14/cry-109/internal/timing"
)

type Fusion struct {
	mu          sync.Mutex
	window      *timing.Window[model.HeaveSample, model.HoistSample]
	maxAge      time.Duration
	lastOutput  model.Compensation
	pairedCount uint64
}

func NewFusion(maxSkew, maxAge time.Duration) *Fusion {
	return &Fusion{
		window: timing.NewWindow[model.HeaveSample, model.HoistSample](maxSkew),
		maxAge: maxAge,
	}
}

func (f *Fusion) AddHeave(sample model.HeaveSample, now time.Time) (model.Compensation, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.window.DropBefore(now.Add(-f.maxAge))
	if now.Sub(sample.CapturedAt) > f.maxAge {
		return f.degraded(now, "heave sample is stale"), false
	}
	f.window.AddLeft(timing.TimedValue[model.HeaveSample]{At: sample.CapturedAt, Value: sample})
	return f.match(now)
}

func (f *Fusion) AddHoist(sample model.HoistSample, now time.Time) (model.Compensation, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.window.DropBefore(now.Add(-f.maxAge))
	if now.Sub(sample.CapturedAt) > f.maxAge {
		return f.degraded(now, "hoist sample is stale"), false
	}
	f.window.AddRight(timing.TimedValue[model.HoistSample]{At: sample.CapturedAt, Value: sample})
	return f.match(now)
}

func (f *Fusion) match(now time.Time) (model.Compensation, bool) {
	pair, ok := f.window.MatchNewest()
	if !ok {
		return f.degraded(now, "waiting for time-aligned heave and hoist samples"), false
	}
	if now.Sub(pair.Left.At) > f.maxAge || now.Sub(pair.Right.At) > f.maxAge {
		return f.degraded(now, "aligned sample pair expired before use"), false
	}
	f.pairedCount++
	f.lastOutput = model.Compensation{
		CapturedAt: later(pair.Left.At, pair.Right.At),
		Velocity:   -(pair.Left.Value.Velocity - pair.Right.Value.Velocity),
		Degraded:   false,
	}
	return f.lastOutput, true
}

func (f *Fusion) degraded(now time.Time, reason string) model.Compensation {
	return model.Compensation{
		CapturedAt: now,
		Velocity:   0,
		Degraded:   true,
		Reason:     reason,
	}
}

func (f *Fusion) Last() (model.Compensation, uint64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.lastOutput, f.pairedCount
}

func later(left, right time.Time) time.Time {
	if left.After(right) {
		return left
	}
	return right
}
