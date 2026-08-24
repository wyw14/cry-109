package control

import (
	"github.com/wyw14/cry-109/internal/model"
)

func (r *Runtime) Snapshot() model.FleetSnapshot {
	anchor, _, _ := r.Storm.Sequence().Snapshot()
	return model.FleetSnapshot{
		Lifts: r.Lifts.List(), Geometry: r.Spreader.Telescope().Geometry(),
		Brake: r.Brakes.State(), Anchor: anchor, Incidents: r.Incidents.Active(),
		RailPose: r.Calibration.Pose(), CapturedAt: r.Clock.Now(),
	}
}
