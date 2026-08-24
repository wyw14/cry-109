package control

import (
	"errors"
	"time"

	"github.com/wyw14/cry-109/internal/model"
	"github.com/wyw14/cry-109/internal/twistlock"
)

type LandingRequest struct {
	SessionID            string    `json:"session_id"`
	LoadKN               float64   `json:"load_kn"`
	HookMeters           float64   `json:"hook_meters"`
	ContactToleranceM    float64   `json:"contact_tolerance_m"`
	LoadTransferComplete bool      `json:"load_transfer_complete"`
	ObservedAt           time.Time `json:"observed_at"`
}

type LandingResult struct {
	Landed           bool    `json:"landed"`
	TwistlocksHeld   bool    `json:"twistlocks_held"`
	TwistlocksOpen   bool    `json:"twistlocks_open"`
	DeckMeters       float64 `json:"deck_meters"`
	RelativeVelocity float64 `json:"relative_velocity_mps"`
	Reason           string  `json:"reason,omitempty"`
}

func (r *Runtime) ObserveLanding(release *twistlock.ReleaseService, request LandingRequest) (LandingResult, error) {
	if request.SessionID == "" {
		return LandingResult{}, errors.New("engage session is required")
	}
	if request.ObservedAt.IsZero() {
		request.ObservedAt = r.Clock.Now()
	}
	deck, velocity, fresh := r.Heave.State().RelativeDeck(request.ObservedAt, 250*time.Millisecond)
	if request.ContactToleranceM <= 0 {
		request.ContactToleranceM = 0.08
	}
	observation := model.LandingObservation{
		At: request.ObservedAt, LoadKN: request.LoadKN, HookMeters: request.HookMeters,
		DeckMeters: deck, RelativeVelocityMPS: velocity,
		ContactToleranceM:    request.ContactToleranceM,
		LoadTransferComplete: request.LoadTransferComplete && fresh,
	}
	landed := r.Landing.ApplyLoad(observation)
	if landed {
		if err := release.OnLanded(request.SessionID, true); err != nil {
			return LandingResult{}, err
		}
	}
	locked, released := release.State()
	_, reason, _ := r.Landing.Snapshot()
	return LandingResult{
		Landed: landed, TwistlocksHeld: locked, TwistlocksOpen: released,
		DeckMeters: deck, RelativeVelocity: velocity, Reason: reason,
	}, nil
}

func (r *Runtime) PrimeLanding(sessionID string) *twistlock.ReleaseService {
	release := twistlock.NewReleaseService(sessionID)
	proof := model.TwistlockProof{SessionID: sessionID, Complete: true}
	_ = release.ApplyProof(proof)
	return release
}
