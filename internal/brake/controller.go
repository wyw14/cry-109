package brake

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/wyw14/cry-109/internal/model"
)

type Controller struct {
	mu    sync.RWMutex
	state model.BrakeState
}

func NewController() *Controller {
	return &Controller{state: model.BrakeState{Mode: model.BrakeReleased, TransitionID: uuid.NewString()}}
}

func (c *Controller) SetMode(mode model.BrakeMode, torqueKNM float64, fault string) (model.BrakeState, error) {
	if torqueKNM < 0 {
		return model.BrakeState{}, errors.New("brake torque cannot be negative")
	}
	if (mode == model.BrakeFriction || mode == model.BrakeHeld) && torqueKNM == 0 {
		return model.BrakeState{}, errors.New("friction or held mode requires torque")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state.Mode = mode
	c.state.TorqueKNM = torqueKNM
	c.state.LastFault = fault
	c.state.TransitionID = uuid.NewString()
	if mode == model.BrakeReleased || mode == model.BrakeRegenerative {
		c.state.PressureBar = 0
		c.state.Holding = false
	}
	return c.state, nil
}

func (c *Controller) BuildPressure(pressureBar float64) model.BrakeState {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.state.Mode == model.BrakeFriction || c.state.Mode == model.BrakeHeld {
		c.state.PressureBar = pressureBar
		c.state.Holding = pressureBar >= 110 && c.state.TorqueKNM > 0
		if c.state.Holding {
			c.state.Mode = model.BrakeHeld
		}
	}
	return c.state
}

func (c *Controller) TransferToFriction(requiredTorque float64, cause error) (model.BrakeState, error) {
	fault := ""
	if cause != nil {
		fault = cause.Error()
	}
	return c.SetMode(model.BrakeFriction, requiredTorque, fault)
}

func (c *Controller) State() model.BrakeState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}
