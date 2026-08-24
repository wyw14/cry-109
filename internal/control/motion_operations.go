package control

import (
	"errors"

	"github.com/wyw14/cry-109/internal/model"
)

type MotionStatus struct {
	TrolleyPosition float64               `json:"trolley_position"`
	TrolleyVelocity float64               `json:"trolley_velocity"`
	TrajectoryID    string                `json:"trajectory_id,omitempty"`
	RailPose        model.RailPose        `json:"rail_pose"`
	Zones           []model.ExclusionZone `json:"zones"`
	Travel          model.TravelPlan      `json:"travel"`
}

func (r *Runtime) MotionStatus(destination float64) (MotionStatus, error) {
	position, velocity, trajectoryID := r.Trolley.Snapshot()
	pose := r.Calibration.Pose()
	travel, err := r.Motion.BuildTravel(pose, destination)
	if err != nil {
		return MotionStatus{}, err
	}
	return MotionStatus{
		TrolleyPosition: position, TrolleyVelocity: velocity, TrajectoryID: trajectoryID,
		RailPose: pose, Zones: r.Motion.Zones(), Travel: travel,
	}, nil
}

func (r *Runtime) RecalibrateRail(origin, destination float64) (MotionStatus, error) {
	if _, err := r.Calibration.ApplyOrigin(origin); err != nil {
		return MotionStatus{}, err
	}
	return r.MotionStatus(destination)
}

func (r *Runtime) UpdateVesselDraft(draft, deck, hatch float64) (model.VesselProfile, error) {
	return r.Vessel.UpdateDraft(draft, deck, hatch, r.Clock.Now())
}

func (r *Runtime) BalanceTandem(cycle uint64, leftKN, rightKN float64) (model.SpeedTargets, error) {
	if cycle == 0 {
		return model.SpeedTargets{}, errors.New("cycle must be positive")
	}
	sample, err := r.LoadCells.Capture(cycle, leftKN, rightKN, r.Clock.Now())
	if err != nil {
		return model.SpeedTargets{}, err
	}
	targets, err := r.Tandem.Balance(sample)
	if cycle > 128 {
		r.Tandem.Prune(cycle - 128)
	}
	return targets, err
}

func (r *Runtime) ApplyHeave(heave model.HeaveSample, hoist model.HoistSample) model.Compensation {
	_, _ = r.Heave.IngestHeave(heave, r.Clock.Now())
	command, ok := r.Heave.IngestHoist(hoist, r.Clock.Now())
	if !ok {
		return command
	}
	if err := r.Compensation.Apply(command, 32); err != nil {
		command.Degraded = true
		command.Reason = err.Error()
		command.Velocity = 0
	}
	return command
}
