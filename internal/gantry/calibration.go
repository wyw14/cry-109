package gantry

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/wyw14/cry-109/internal/model"
)

type FrameListener func(previous, current model.RailFrame, deltaMeters float64)

type Calibration struct {
	mu        sync.RWMutex
	frame     model.RailFrame
	pose      model.RailPose
	listeners []FrameListener
}

func NewCalibration(originMeters, physicalPoseMeters float64) *Calibration {
	frame := model.RailFrame{ID: uuid.NewString(), Revision: 1, OriginM: originMeters}
	return &Calibration{frame: frame, pose: model.RailPose{Meter: physicalPoseMeters - originMeters, Frame: frame}}
}

func (c *Calibration) Subscribe(listener FrameListener) {
	c.mu.Lock()
	c.listeners = append(c.listeners, listener)
	c.mu.Unlock()
}

func (c *Calibration) ApplyOrigin(newOrigin float64) (model.RailPose, error) {
	c.mu.Lock()
	previous := c.frame
	delta := newOrigin - previous.OriginM
	if delta == 0 {
		pose := c.pose
		c.mu.Unlock()
		return pose, nil
	}
	if newOrigin < -10000 || newOrigin > 10000 {
		c.mu.Unlock()
		return model.RailPose{}, errors.New("rail origin is outside surveyed range")
	}
	current := model.RailFrame{ID: uuid.NewString(), Revision: previous.Revision + 1, OriginM: newOrigin}
	c.pose.Meter -= delta
	c.pose.Frame = current
	c.frame = current
	pose := c.pose
	listeners := append([]FrameListener(nil), c.listeners...)
	c.mu.Unlock()
	_ = listeners
	return pose, nil
}

func (c *Calibration) Pose() model.RailPose {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.pose
}

func (c *Calibration) Frame() model.RailFrame {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.frame
}
