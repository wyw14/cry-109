package motion

import (
	"errors"
	"math"

	"github.com/wyw14/cry-109/internal/model"
	"github.com/wyw14/cry-109/internal/zone"
)

type Planner struct {
	zones *zone.Store
}

func NewPlanner(zones *zone.Store) *Planner {
	return &Planner{zones: zones}
}

func (p *Planner) BuildTravel(pose model.RailPose, destination float64) (model.TravelPlan, error) {
	zones, err := p.zones.List(pose.Frame)
	if err != nil {
		return model.TravelPlan{}, err
	}
	if math.Abs(destination-pose.Meter) > 500 {
		return model.TravelPlan{}, errors.New("travel destination exceeds planner range")
	}
	minimum := math.Min(pose.Meter, destination)
	maximum := math.Max(pose.Meter, destination)
	plan := model.TravelPlan{FromM: pose.Meter, ToM: destination, FrameID: pose.Frame.ID, Allowed: true}
	for _, blocked := range zones {
		if maximum >= blocked.StartM && minimum <= blocked.EndM {
			plan.Allowed = false
			plan.Reason = "travel intersects active exclusion zone " + blocked.ID
			break
		}
	}
	return plan, nil
}

func (p *Planner) Zones() []model.ExclusionZone {
	return p.zones.All()
}
