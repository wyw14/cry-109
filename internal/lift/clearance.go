package lift

import (
	"fmt"

	"github.com/wyw14/cry-109/internal/model"
)

type ClearancePlan struct {
	LiftID          string  `json:"lift_id"`
	VesselRevision  uint64  `json:"vessel_revision"`
	RequiredHeightM float64 `json:"required_height_m"`
	PlannedHeightM  float64 `json:"planned_height_m"`
	Valid           bool    `json:"valid"`
	Reason          string  `json:"reason,omitempty"`
}

func PlanClearance(liftID string, profile model.VesselProfile, margin, plannedHeight float64) ClearancePlan {
	required := profile.HatchMeters + margin
	valid := plannedHeight >= required
	reason := ""
	if !valid {
		reason = fmt.Sprintf("planned height %.2f is below required %.2f", plannedHeight, required)
	}
	return ClearancePlan{
		LiftID: liftID, VesselRevision: profile.Revision,
		RequiredHeightM: required, PlannedHeightM: plannedHeight,
		Valid: valid, Reason: reason,
	}
}

func (p ClearancePlan) Revalidate(profile model.VesselProfile, margin float64) ClearancePlan {
	if p.VesselRevision == profile.Revision {
		return p
	}
	return PlanClearance(p.LiftID, profile, margin, p.PlannedHeightM)
}

func (p ClearancePlan) CanTraverse(profile model.VesselProfile) bool {
	return p.Valid && p.VesselRevision == profile.Revision && p.PlannedHeightM >= p.RequiredHeightM
}
