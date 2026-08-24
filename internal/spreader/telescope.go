package spreader

import (
	"errors"
	"sync"

	"github.com/wyw14/cry-109/internal/model"
)

type GeometryListener func(previous, current model.Geometry)

type Telescope struct {
	mu        sync.RWMutex
	geometry  model.Geometry
	listeners []GeometryListener
}

func NewTelescope(initial model.Geometry) *Telescope {
	return &Telescope{geometry: initial}
}

func (t *Telescope) Subscribe(listener GeometryListener) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.listeners = append(t.listeners, listener)
}

func (t *Telescope) Extend(lengthFeet int, massTonnes float64) (model.Geometry, error) {
	if lengthFeet != 20 && lengthFeet != 40 && lengthFeet != 45 {
		return model.Geometry{}, errors.New("unsupported spreader length")
	}
	if massTonnes <= 0 {
		return model.Geometry{}, errors.New("spreader mass must be positive")
	}
	t.mu.Lock()
	previous := t.geometry
	if previous.LengthFeet == lengthFeet && previous.MassTonnes == massTonnes {
		t.mu.Unlock()
		return previous, nil
	}
	t.geometry = model.Geometry{LengthFeet: lengthFeet, Revision: previous.Revision + 1, MassTonnes: massTonnes}
	current := t.geometry
	listeners := append([]GeometryListener(nil), t.listeners...)
	t.mu.Unlock()
	for _, listener := range listeners {
		listener(previous, current)
	}
	return current, nil
}

func (t *Telescope) Geometry() model.Geometry {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.geometry
}
