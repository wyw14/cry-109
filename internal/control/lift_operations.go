package control

import (
	"errors"
	"time"

	"github.com/wyw14/cry-109/internal/lift"
	"github.com/wyw14/cry-109/internal/model"
)

type StartLiftRequest struct {
	ContainerID string `json:"container_id"`
	LengthFeet  int    `json:"length_feet"`
}

func (r *Runtime) StartLift(request StartLiftRequest) (model.Lift, error) {
	if request.LengthFeet == 0 {
		request.LengthFeet = 40
	}
	geometry, err := r.Spreader.Telescope().Extend(request.LengthFeet, 12.8)
	if err != nil {
		return model.Lift{}, err
	}
	session := r.Spreader.StartEngage()
	current, err := r.Lifts.Start(request.ContainerID, session, geometry.Revision, r.Clock.Now())
	if err != nil {
		return model.Lift{}, err
	}
	if current, err = r.Lifts.Advance(current.ID, model.LiftEngaging, r.Clock.Now()); err != nil {
		return model.Lift{}, err
	}
	proof := r.Spreader.ConfirmAll(session)
	if err := r.Spreader.EngageController().RequireLocked(); err != nil {
		return model.Lift{}, err
	}
	permit := lift.NewLoadPermit(session)
	if !permit.Evaluate(proof) {
		_, reason := permit.Status()
		return model.Lift{}, errors.New(reason)
	}
	if err := permit.Require(); err != nil {
		return model.Lift{}, err
	}
	return r.Lifts.Advance(current.ID, model.LiftLifting, r.Clock.Now().Add(time.Millisecond))
}

func (r *Runtime) ReplanTrolley(liftID string, distance, maxSpeed, ropeLength, payloadTonnes float64) (model.Trajectory, error) {
	current, ok := r.Lifts.Get(liftID)
	if !ok {
		return model.Trajectory{}, errors.New("lift not found")
	}
	geometry := r.Spreader.Telescope().Geometry()
	current.GeometryRevision = geometry.Revision
	sway := r.Models.GetOrBuild(current.ID, geometry, ropeLength, payloadTonnes)
	profile := r.Vessel.Current()
	replanned, err := r.Antisway.Plan(current.ID, sway, distance, maxSpeed)
	if err != nil {
		return model.Trajectory{}, err
	}
	replanned.VesselRevision = profile.Revision
	if err := r.Trajectories.Replace(current.ID, replanned); err != nil {
		return model.Trajectory{}, err
	}
	return replanned, nil
}

func (r *Runtime) PlanTrolley(liftID string, distance, maxSpeed, ropeLength, payloadTonnes float64) (model.Trajectory, error) {
	current, ok := r.Repository.Lift(liftID)
	if !ok {
		return model.Trajectory{}, errors.New("lift not found")
	}
	geometry := r.Spreader.Telescope().Geometry()
	if current.GeometryRevision != geometry.Revision {
		return model.Trajectory{}, errors.New("lift geometry is stale")
	}
	sway := r.Models.GetOrBuild(liftID, geometry, ropeLength, payloadTonnes)
	profile := r.Vessel.Current()
	clearancePlan := lift.PlanClearance(liftID, profile, 1.2, 14)
	if !clearancePlan.CanTraverse(profile) {
		return model.Trajectory{}, errors.New(clearancePlan.Reason)
	}
	return r.Trajectories.Plan(liftID, sway, profile.Revision, distance, maxSpeed)
}
