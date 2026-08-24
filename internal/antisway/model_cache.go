package antisway

import (
	"sync"

	"github.com/google/uuid"
	"github.com/wyw14/cry-109/internal/model"
)

type modelKey struct {
	liftID           string
	geometryRevision uint64
	ropeMillimeters  int64
	payloadKilograms int64
}

type ModelCache struct {
	mu     sync.RWMutex
	models map[modelKey]model.AntiswayModel
}

func NewModelCache() *ModelCache {
	return &ModelCache{models: make(map[modelKey]model.AntiswayModel)}
}

func (c *ModelCache) GetOrBuild(liftID string, geometry model.Geometry, ropeLength, payloadTonnes float64) model.AntiswayModel {
	key := modelKey{
		liftID:           liftID,
		geometryRevision: geometry.Revision,
		ropeMillimeters:  int64(ropeLength * 1000),
		payloadKilograms: int64(payloadTonnes * 1000),
	}
	c.mu.RLock()
	current, ok := c.models[key]
	c.mu.RUnlock()
	if ok {
		return current
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if current, ok = c.models[key]; ok {
		return current
	}
	gain := (geometry.MassTonnes + payloadTonnes) / (ropeLength * float64(geometry.LengthFeet))
	current = model.AntiswayModel{
		ID: uuid.NewString(), GeometryRevision: geometry.Revision,
		LengthFeet: geometry.LengthFeet, RopeLength: ropeLength, Gain: gain,
	}
	c.models[key] = current
	return current
}

func (c *ModelCache) InvalidateGeometry(previous, current model.Geometry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if previous.Revision == current.Revision {
		return
	}
	for key, model := range c.models {
		if model.GeometryRevision != current.Revision {
			delete(c.models, key)
		}
	}
}

func (c *ModelCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.models)
}
