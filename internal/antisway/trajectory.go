package antisway

import (
	"errors"
	"math"

	"github.com/google/uuid"
	"github.com/wyw14/cry-109/internal/model"
)

type Planner struct{}

func NewPlanner() *Planner { return &Planner{} }

func (p *Planner) Plan(liftID string, sway model.AntiswayModel, distance, maxSpeed float64) (model.Trajectory, error) {
	if distance == 0 || maxSpeed <= 0 {
		return model.Trajectory{}, errors.New("distance and max speed must define a movement")
	}
	direction := 1.0
	if distance < 0 {
		direction = -1
	}
	duration := math.Abs(distance) / maxSpeed
	points := make([]model.TrajectoryPoint, 0, 9)
	for step := 0; step <= 8; step++ {
		ratio := float64(step) / 8
		shape := math.Sin(math.Pi * ratio)
		velocity := direction * maxSpeed * shape * (1 - math.Min(sway.Gain, 0.3))
		points = append(points, model.TrajectoryPoint{
			AtSeconds:   duration * ratio,
			PositionM:   distance * ratio,
			VelocityMPS: velocity,
		})
	}
	return model.Trajectory{
		ID: uuid.NewString(), LiftID: liftID, ModelID: sway.ID,
		GeometryRevision: sway.GeometryRevision, Points: points, Valid: true,
	}, nil
}

func (p *Planner) Revalidate(trajectory model.Trajectory, current model.AntiswayModel) model.Trajectory {
	if trajectory.GeometryRevision != current.GeometryRevision || trajectory.ModelID != current.ID {
		trajectory.Valid = false
		trajectory.Reason = "anti-sway model no longer matches current spreader geometry"
	}
	return trajectory
}
