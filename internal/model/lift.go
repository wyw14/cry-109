package model

import (
	"errors"
	"time"
)

type LiftPhase string

const (
	LiftApproaching LiftPhase = "approaching"
	LiftEngaging    LiftPhase = "engaging"
	LiftLifting     LiftPhase = "lifting"
	LiftTraversing  LiftPhase = "traversing"
	LiftLowering    LiftPhase = "lowering"
	LiftLanding     LiftPhase = "landing"
	LiftReleasing   LiftPhase = "releasing"
	LiftComplete    LiftPhase = "complete"
	LiftAborted     LiftPhase = "aborted"
)

type Lift struct {
	ID               string    `json:"id"`
	ContainerID      string    `json:"container_id"`
	Phase            LiftPhase `json:"phase"`
	SessionID        string    `json:"session_id"`
	GeometryRevision uint64    `json:"geometry_revision"`
	StartedAt        time.Time `json:"started_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Reason           string    `json:"reason,omitempty"`
}

var phaseOrder = map[LiftPhase]int{
	LiftApproaching: 0,
	LiftEngaging:    1,
	LiftLifting:     2,
	LiftTraversing:  3,
	LiftLowering:    4,
	LiftLanding:     5,
	LiftReleasing:   6,
	LiftComplete:    7,
	LiftAborted:     8,
}

func (l Lift) Advance(next LiftPhase, at time.Time) (Lift, error) {
	if next == LiftAborted {
		l.Phase = next
		l.UpdatedAt = at
		return l, nil
	}
	current, okCurrent := phaseOrder[l.Phase]
	wanted, okWanted := phaseOrder[next]
	if !okCurrent || !okWanted || wanted != current+1 {
		return l, errors.New("invalid lift phase transition")
	}
	l.Phase = next
	l.UpdatedAt = at
	return l, nil
}

type FleetSnapshot struct {
	Lifts      []Lift           `json:"lifts"`
	Geometry   Geometry         `json:"geometry"`
	Brake      BrakeState       `json:"brake"`
	Anchor     AnchorState      `json:"anchor"`
	Incidents  []SafetyIncident `json:"incidents"`
	RailPose   RailPose         `json:"rail_pose"`
	CapturedAt time.Time        `json:"captured_at"`
}
