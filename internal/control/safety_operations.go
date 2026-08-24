package control

import (
	"errors"
	"time"

	"github.com/wyw14/cry-109/internal/model"
)

type SafetyStatus struct {
	Brake     model.BrakeState       `json:"brake"`
	Anchor    model.AnchorState      `json:"anchor"`
	WindMPS   float64                `json:"wind_mps"`
	Incidents []model.SafetyIncident `json:"incidents"`
	Clearance any                    `json:"clearance"`
}

func (r *Runtime) TriggerStorm(windMPS float64) (SafetyStatus, error) {
	now := r.Clock.Now()
	if _, err := r.Storm.ObserveWind(windMPS, now); err != nil {
		return SafetyStatus{}, err
	}
	state := r.Brakes.BuildPressure(125)
	r.BrakeHold.Observe(state, now)
	r.GantryStop.Observe(0, now)
	return r.SafetyStatus(now), nil
}

func (r *Runtime) AdvanceStorm(at time.Time) model.AnchorState {
	return r.Storm.Sequence().Update(at)
}

func (r *Runtime) ApplyRegen(powerKW float64, sinkAvailable bool) (model.BrakeState, error) {
	r.Energy.SetSinkAvailable(sinkAvailable)
	if err := r.Lowering.Begin(); err != nil {
		return model.BrakeState{}, err
	}
	result := r.Energy.AcceptRegen(powerKW, r.Clock.Now())
	return r.Lowering.ApplyEnergyResult(result)
}

func (r *Runtime) ChangeReeving(factor int, drumMeters, requiredHeight float64) (any, error) {
	if _, err := r.Reeving.Apply(factor, r.Clock.Now()); err != nil {
		return nil, err
	}
	decision := r.Clearance.Check(drumMeters, requiredHeight)
	if !decision.Allowed {
		return decision, errors.New(decision.Reason)
	}
	return decision, nil
}

func (r *Runtime) SafetyStatus(at time.Time) SafetyStatus {
	anchor, _, _ := r.Storm.Sequence().Snapshot()
	return SafetyStatus{
		Brake: r.Brakes.State(), Anchor: anchor, WindMPS: r.Storm.Wind(),
		Incidents: r.Incidents.Active(), Clearance: r.Clearance.Check(8, 46),
	}
}
