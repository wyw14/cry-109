package model

import "time"

type AntiswayModel struct {
	ID               string  `json:"id"`
	GeometryRevision uint64  `json:"geometry_revision"`
	LengthFeet       int     `json:"length_feet"`
	RopeLength       float64 `json:"rope_length"`
	Gain             float64 `json:"gain"`
}

type TrajectoryPoint struct {
	AtSeconds   float64 `json:"at_seconds"`
	PositionM   float64 `json:"position_m"`
	VelocityMPS float64 `json:"velocity_mps"`
}

type Trajectory struct {
	ID               string            `json:"id"`
	LiftID           string            `json:"lift_id"`
	ModelID          string            `json:"model_id"`
	VesselRevision   uint64            `json:"vessel_revision"`
	GeometryRevision uint64            `json:"geometry_revision"`
	Points           []TrajectoryPoint `json:"points"`
	Valid            bool              `json:"valid"`
	Reason           string            `json:"reason,omitempty"`
}

type VesselProfile struct {
	ID          string    `json:"id"`
	Revision    uint64    `json:"revision"`
	DraftMeters float64   `json:"draft_meters"`
	DeckMeters  float64   `json:"deck_meters"`
	HatchMeters float64   `json:"hatch_meters"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RailFrame struct {
	ID       string  `json:"id"`
	Revision uint64  `json:"revision"`
	OriginM  float64 `json:"origin_m"`
}

type RailPose struct {
	Meter float64   `json:"meter"`
	Frame RailFrame `json:"frame"`
}

type ExclusionZone struct {
	ID     string    `json:"id"`
	StartM float64   `json:"start_m"`
	EndM   float64   `json:"end_m"`
	Frame  RailFrame `json:"frame"`
}

type TravelPlan struct {
	FromM   float64 `json:"from_m"`
	ToM     float64 `json:"to_m"`
	FrameID string  `json:"frame_id"`
	Allowed bool    `json:"allowed"`
	Reason  string  `json:"reason,omitempty"`
}
