package spreader

import (
	"strings"
	"testing"
	"time"

	"github.com/wyw14/cry-109/internal/model"
)

const engageSession = "engage-40ft"

func engageAck(corner model.Corner, locked, failed bool) model.TwistlockAck {
	return model.TwistlockAck{
		SessionID: engageSession, Corner: corner,
		Locked: locked, Failed: failed, At: time.Now().UTC(),
	}
}

// TestEngageControllerBlocksLiftUntilAllFourCornersLock reproduces the
// reported 40ft spreader fault and asserts the engage controller never
// reaches EngageLocked (the gate RequireLocked enforces before load
// transfer) until all four corners of the current session are locked.
func TestEngageControllerBlocksLiftUntilAllFourCornersLock(t *testing.T) {
	t.Parallel()
	controller := NewEngageController("placeholder")
	controller.Begin(engageSession)

	// First diagonal locks. The hoist must not be allowed to transfer load.
	for _, corner := range []model.Corner{model.FrontLeft, model.RearRight} {
		controller.Confirm(engageAck(corner, true, false))
		if state := controller.State(); state == EngageLocked {
			t.Fatalf("state reached locked after %s alone; all four corners are required", corner)
		}
		if err := controller.RequireLocked(); err == nil {
			t.Fatalf("RequireLocked must block load transfer with only two corners locked")
		}
	}

	// Third corner locks. Still incomplete.
	controller.Confirm(engageAck(model.FrontRight, true, false))
	if state := controller.State(); state == EngageLocked {
		t.Fatalf("state reached locked after three corners; all four corners are required")
	}
	if err := controller.RequireLocked(); err == nil {
		t.Fatalf("RequireLocked must block load transfer with three corners locked")
	}

	// Fourth corner locks. Only now may the engage complete.
	controller.Confirm(engageAck(model.RearLeft, true, false))
	if state := controller.State(); state != EngageLocked {
		t.Fatalf("state must be locked once all four corners lock, got %s", state)
	}
	if err := controller.RequireLocked(); err != nil {
		t.Fatalf("RequireLocked must pass once all four corners lock: %v", err)
	}
}

// TestEngageControllerStopsOnAnyCornerFailure verifies that any corner
// rejecting engagement drives the controller into EngageFailed and keeps
// RequireLocked blocking load transfer with a clear reason, even if the
// remaining corners subsequently lock.
func TestEngageControllerStopsOnAnyCornerFailure(t *testing.T) {
	t.Parallel()
	controller := NewEngageController("placeholder")
	controller.Begin(engageSession)

	controller.Confirm(engageAck(model.FrontLeft, true, false))
	controller.Confirm(engageAck(model.RearRight, false, true))

	if state := controller.State(); state != EngageFailed {
		t.Fatalf("state must be failed after a corner rejects engagement, got %s", state)
	}
	err := controller.RequireLocked()
	if err == nil {
		t.Fatalf("RequireLocked must block load transfer after a failed corner")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "failed") && !strings.Contains(strings.ToLower(err.Error()), "reject") {
		t.Fatalf("RequireLocked error must clearly state the failure, got: %v", err)
	}

	// Remaining corners locking must not recover the engage without a fresh
	// Begin; the hoist stays blocked.
	for _, corner := range []model.Corner{model.FrontRight, model.RearLeft} {
		controller.Confirm(engageAck(corner, true, false))
		if state := controller.State(); state != EngageFailed {
			t.Fatalf("engage must remain failed after %s lock; failure is sticky until reset", corner)
		}
		if err := controller.RequireLocked(); err == nil {
			t.Fatalf("RequireLocked must keep blocking load transfer after sticky failure")
		}
	}
}

// TestEngageControllerRequiresCurrentSession ensures a stale session can
// never satisfy the four-corner gate.
func TestEngageControllerRequiresCurrentSession(t *testing.T) {
	t.Parallel()
	controller := NewEngageController("placeholder")
	controller.Begin(engageSession)

	stale := engageAck(model.FrontLeft, true, false)
	stale.SessionID = "stale-session"
	controller.Confirm(stale)
	for _, corner := range []model.Corner{model.FrontRight, model.RearLeft, model.RearRight} {
		controller.Confirm(engageAck(corner, true, false))
	}
	if state := controller.State(); state == EngageLocked {
		t.Fatalf("a stale-session ack must not allow the engage to complete")
	}
	if err := controller.RequireLocked(); err == nil {
		t.Fatalf("RequireLocked must reject a proof missing the current session's front-left ack")
	}
}
