package model

import "time"

type Corner string

const (
	FrontLeft  Corner = "front-left"
	FrontRight Corner = "front-right"
	RearLeft   Corner = "rear-left"
	RearRight  Corner = "rear-right"
)

func AllCorners() []Corner {
	return []Corner{FrontLeft, FrontRight, RearLeft, RearRight}
}

type TwistlockAck struct {
	SessionID string    `json:"session_id"`
	Corner    Corner    `json:"corner"`
	Locked    bool      `json:"locked"`
	Failed    bool      `json:"failed"`
	At        time.Time `json:"at"`
}

type TwistlockProof struct {
	SessionID string   `json:"session_id"`
	Complete  bool     `json:"complete"`
	Failed    bool     `json:"failed"`
	Missing   []Corner `json:"missing"`
	Reason    string   `json:"reason,omitempty"`
}

type Geometry struct {
	LengthFeet int     `json:"length_feet"`
	Revision   uint64  `json:"revision"`
	MassTonnes float64 `json:"mass_tonnes"`
}

type HeaveSample struct {
	CapturedAt time.Time `json:"captured_at"`
	Meters     float64   `json:"meters"`
	Velocity   float64   `json:"velocity"`
}

type HoistSample struct {
	CapturedAt time.Time `json:"captured_at"`
	HookMeters float64   `json:"hook_meters"`
	Velocity   float64   `json:"velocity"`
}

type Compensation struct {
	CapturedAt time.Time `json:"captured_at"`
	Velocity   float64   `json:"velocity"`
	Degraded   bool      `json:"degraded"`
	Reason     string    `json:"reason,omitempty"`
}

type LoadSample struct {
	Cycle    uint64    `json:"cycle"`
	LeftKN   float64   `json:"left_kn"`
	RightKN  float64   `json:"right_kn"`
	Captured time.Time `json:"captured_at"`
}

type SpeedTargets struct {
	Cycle       uint64  `json:"cycle"`
	LeftMPS     float64 `json:"left_mps"`
	RightMPS    float64 `json:"right_mps"`
	Coordinated bool    `json:"coordinated"`
}
