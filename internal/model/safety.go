package model

import "time"

type LandingObservation struct {
	At                   time.Time `json:"at"`
	LoadKN               float64   `json:"load_kn"`
	HookMeters           float64   `json:"hook_meters"`
	DeckMeters           float64   `json:"deck_meters"`
	RelativeVelocityMPS  float64   `json:"relative_velocity_mps"`
	ContactToleranceM    float64   `json:"contact_tolerance_m"`
	LoadTransferComplete bool      `json:"load_transfer_complete"`
}

type BrakeMode string

const (
	BrakeReleased     BrakeMode = "released"
	BrakeRegenerative BrakeMode = "regenerative"
	BrakeFriction     BrakeMode = "friction"
	BrakeHeld         BrakeMode = "held"
)

type BrakeState struct {
	Mode         BrakeMode `json:"mode"`
	TorqueKNM    float64   `json:"torque_knm"`
	PressureBar  float64   `json:"pressure_bar"`
	Holding      bool      `json:"holding"`
	LastFault    string    `json:"last_fault,omitempty"`
	TransitionID string    `json:"transition_id"`
}

type AnchorState string

const (
	AnchorRaised     AnchorState = "raised"
	AnchorWaiting    AnchorState = "waiting"
	AnchorDescending AnchorState = "descending"
	AnchorSecured    AnchorState = "secured"
	AnchorAborted    AnchorState = "aborted"
)

type SafetyIncident struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	Message   string    `json:"message"`
	RaisedAt  time.Time `json:"raised_at"`
	Resolved  bool      `json:"resolved"`
	Component string    `json:"component"`
}

type ReevingConfig struct {
	Factor    int       `json:"factor"`
	Revision  uint64    `json:"revision"`
	UpdatedAt time.Time `json:"updated_at"`
}
