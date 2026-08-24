package hoist

import (
	"errors"
	"math"
	"sync"

	"github.com/wyw14/cry-109/internal/model"
)

type CompensationDrive struct {
	mu           sync.RWMutex
	command      model.Compensation
	maxVelocity  float64
	tensionFloor float64
}

func NewCompensationDrive(maxVelocity, tensionFloor float64) *CompensationDrive {
	return &CompensationDrive{maxVelocity: maxVelocity, tensionFloor: tensionFloor}
}

func (d *CompensationDrive) Apply(command model.Compensation, ropeTensionKN float64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if command.Degraded {
		d.command = command
		d.command.Velocity = 0
		return nil
	}
	if ropeTensionKN < d.tensionFloor {
		return errors.New("rope tension is below compensation floor")
	}
	if math.Abs(command.Velocity) > d.maxVelocity {
		return errors.New("compensation velocity exceeds drive limit")
	}
	d.command = command
	return nil
}

func (d *CompensationDrive) Command() model.Compensation {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.command
}
