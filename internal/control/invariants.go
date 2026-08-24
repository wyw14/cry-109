package control

import (
	"fmt"
	"sort"

	"github.com/wyw14/cry-109/internal/model"
)

type InvariantResult struct {
	Name    string `json:"name"`
	Holds   bool   `json:"holds"`
	Details string `json:"details"`
}

func (r *Runtime) CheckInvariants() []InvariantResult {
	checks := []InvariantResult{
		r.checkTwistlockInvariant(),
		r.checkBrakeInvariant(),
		r.checkRailFrameInvariant(),
		r.checkReevingInvariant(),
		r.checkPendingTrajectoryInvariant(),
	}
	sort.Slice(checks, func(i, j int) bool { return checks[i].Name < checks[j].Name })
	return checks
}

func (r *Runtime) checkTwistlockInvariant() InvariantResult {
	proof := r.Spreader.EngageController().Proof()
	state := r.Spreader.EngageController().State()
	holds := state != "locked" || (proof.Complete && !proof.Failed && len(proof.Missing) == 0)
	details := fmt.Sprintf("state=%s complete=%t failed=%t missing=%d", state, proof.Complete, proof.Failed, len(proof.Missing))
	return InvariantResult{Name: "four-corner-current-session-proof", Holds: holds, Details: details}
}

func (r *Runtime) checkBrakeInvariant() InvariantResult {
	state := r.Brakes.State()
	requiresTorque := state.Mode == model.BrakeFriction || state.Mode == model.BrakeHeld
	holds := !requiresTorque || state.TorqueKNM > 0
	if state.Holding {
		holds = holds && state.Mode == model.BrakeHeld && state.PressureBar >= 110
	}
	details := fmt.Sprintf("mode=%s torque=%.2f pressure=%.2f holding=%t", state.Mode, state.TorqueKNM, state.PressureBar, state.Holding)
	return InvariantResult{Name: "brake-mode-has-physical-torque", Holds: holds, Details: details}
}

func (r *Runtime) checkRailFrameInvariant() InvariantResult {
	pose := r.Calibration.Pose()
	zones := r.Zones.All()
	holds := true
	for _, active := range zones {
		if active.Frame.ID != pose.Frame.ID || active.Frame.Revision != pose.Frame.Revision {
			holds = false
			break
		}
	}
	details := fmt.Sprintf("pose-frame=%s/r%d zones=%d", pose.Frame.ID, pose.Frame.Revision, len(zones))
	return InvariantResult{Name: "rail-objects-share-frame", Holds: holds, Details: details}
}

func (r *Runtime) checkReevingInvariant() InvariantResult {
	current := r.Reeving.Current()
	_, estimated := r.Height.Estimate(0)
	decision := r.Clearance.Check(0, 0)
	holds := estimated.Factor == current.Factor && estimated.Revision == current.Revision && decision.ReevingRevision == current.Revision
	details := fmt.Sprintf("current=%d/r%d estimator=%d/r%d permit-r%d", current.Factor, current.Revision, estimated.Factor, estimated.Revision, decision.ReevingRevision)
	return InvariantResult{Name: "reeving-config-reaches-clearance", Holds: holds, Details: details}
}

func (r *Runtime) checkPendingTrajectoryInvariant() InvariantResult {
	geometry := r.Spreader.Telescope().Geometry()
	profile := r.Vessel.Current()
	holds := true
	checked := 0
	for _, current := range r.Lifts.List() {
		trajectory, exists := r.Trajectories.Pending(current.ID)
		if !exists {
			continue
		}
		checked++
		if trajectory.Valid && (trajectory.GeometryRevision != geometry.Revision || trajectory.VesselRevision != profile.Revision) {
			holds = false
			break
		}
	}
	details := fmt.Sprintf("pending=%d geometry-r%d vessel-r%d", checked, geometry.Revision, profile.Revision)
	return InvariantResult{Name: "pending-trajectories-use-current-geometry", Holds: holds, Details: details}
}

func invariantsHold(results []InvariantResult) bool {
	for _, result := range results {
		if !result.Holds {
			return false
		}
	}
	return true
}
