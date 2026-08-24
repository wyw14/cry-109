package energy

import (
	"errors"
	"sync"
	"time"
)

var ErrSinkUnavailable = errors.New("regen sink unavailable")

type Result struct {
	Accepted  bool      `json:"accepted"`
	PowerKW   float64   `json:"power_kw"`
	At        time.Time `json:"at"`
	ErrorText string    `json:"error,omitempty"`
}

func (r Result) Err() error {
	if r.Accepted {
		return nil
	}
	if r.ErrorText == "" {
		return ErrSinkUnavailable
	}
	return errors.New(r.ErrorText)
}

type Router struct {
	mu            sync.RWMutex
	sinkAvailable bool
	capacityKW    float64
	acceptedKWh   float64
}

func NewRouter(capacityKW float64) *Router {
	return &Router{sinkAvailable: true, capacityKW: capacityKW}
}

func (r *Router) SetSinkAvailable(available bool) {
	r.mu.Lock()
	r.sinkAvailable = available
	r.mu.Unlock()
}

func (r *Router) AcceptRegen(powerKW float64, at time.Time) Result {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.sinkAvailable || powerKW > r.capacityKW {
		return Result{Accepted: false, PowerKW: powerKW, At: at, ErrorText: ErrSinkUnavailable.Error()}
	}
	r.acceptedKWh += powerKW / 3600
	return Result{Accepted: true, PowerKW: powerKW, At: at}
}

func (r *Router) AcceptedKWh() float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.acceptedKWh
}
