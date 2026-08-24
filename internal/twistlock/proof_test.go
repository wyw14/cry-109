package twistlock

import (
	"testing"
	"time"

	"github.com/wyw14/cry-109/internal/model"
)

const testSession = "session-40ft"

func ack(corner model.Corner, locked, failed bool) model.TwistlockAck {
	return model.TwistlockAck{
		SessionID: testSession, Corner: corner,
		Locked: locked, Failed: failed, At: time.Now().UTC(),
	}
}

// TestProofReducerRejectsPartialDiagonalLock reproduces the 40ft spreader
// fault: front-left and rear-right lock first, and the system must NOT treat
// the engage as complete while the other diagonal is still turning. Before
// the fix a single diagonal pair set Complete=true and the hoist transferred
// load, lifting one side of the container and dropping it back onto the
// hatch. The proof must require all four corners of this session locked.
func TestProofReducerRejectsPartialDiagonalLock(t *testing.T) {
	t.Parallel()
	reducer := NewProofReducer(testSession)

	// First diagonal reports locked. This is exactly the unsafe trigger.
	for _, corner := range []model.Corner{model.FrontLeft, model.RearRight} {
		proof := reducer.Apply(ack(corner, true, false))
		if proof.Complete {
			t.Fatalf("%s locked: proof must not be complete with only two corners", corner)
		}
		if proof.Failed {
			t.Fatalf("%s locked: proof must not be failed before any corner rejects", corner)
		}
		if !contains(proof.Missing, model.FrontRight) || !contains(proof.Missing, model.RearLeft) {
			t.Fatalf("%s locked: remaining corners must be reported missing, got %v", corner, proof.Missing)
		}
	}

	// Remaining corners report locked. Only now may the proof complete.
	for _, corner := range []model.Corner{model.FrontRight, model.RearLeft} {
		proof := reducer.Apply(ack(corner, true, false))
		if proof.Failed {
			t.Fatalf("%s locked: a successful lock must never set failed", corner)
		}
		if corner == model.FrontRight && proof.Complete {
			t.Fatalf("proof must not complete until all four corners lock, not after three")
		}
	}
	proof := reducer.Snapshot()
	if !proof.Complete {
		t.Fatalf("proof must be complete once all four corners lock: %+v", proof)
	}
	if len(proof.Missing) != 0 {
		t.Fatalf("complete proof must report no missing corners, got %v", proof.Missing)
	}
}

// TestProofReducerRequiresEveryCorner guards against any future relaxation
// back toward a subset of corners: no single corner, diagonal, or three of
// four corners may complete the proof.
func TestProofReducerRequiresEveryCorner(t *testing.T) {
	t.Parallel()
	for _, drop := range model.AllCorners() {
		reducer := NewProofReducer(testSession)
		for _, corner := range model.AllCorners() {
			if corner == drop {
				continue
			}
			reducer.Apply(ack(corner, true, false))
		}
		if proof := reducer.Snapshot(); proof.Complete {
			t.Fatalf("proof completed after locking all corners except %s", drop)
		}
	}
}

// TestProofReducerFailsOnAnyCornerFailure ensures that a corner rejecting
// engagement stops the engage and retains a clear reason so the hoist never
// lifts. The failed flag is sticky: a later lock on the same corner does not
// clear it without a fresh engage session.
func TestProofReducerFailsOnAnyCornerFailure(t *testing.T) {
	t.Parallel()
	reducer := NewProofReducer(testSession)

	failed := reducer.Apply(ack(model.RearLeft, false, true))
	if !failed.Failed {
		t.Fatalf("a corner that rejects engagement must mark the proof failed")
	}
	if failed.Complete {
		t.Fatalf("a failed proof must not be complete")
	}
	if failed.Reason == "" {
		t.Fatalf("a failed proof must retain a clear reason for the operator")
	}
	if !contains(failed.Missing, model.RearLeft) {
		t.Fatalf("the failed corner must remain in the missing list")
	}

	// Even if the remaining three corners lock afterward, the engage stays
	// failed: load transfer must stay blocked.
	for _, corner := range []model.Corner{model.FrontLeft, model.FrontRight, model.RearRight} {
		proof := reducer.Apply(ack(corner, true, false))
		if !proof.Failed {
			t.Fatalf("engagement must remain failed after %s lock; failure must be sticky", corner)
		}
		if proof.Complete {
			t.Fatalf("a failed engagement must never report complete")
		}
	}
}

// TestProofReducerIgnoresOtherSessionAcks guarantees acks from a different
// engage session cannot contribute to the current proof, so a stale ack
// cannot satisfy the four-corner requirement.
func TestProofReducerIgnoresOtherSessionAcks(t *testing.T) {
	t.Parallel()
	reducer := NewProofReducer(testSession)

	stale := ack(model.FrontLeft, true, false)
	stale.SessionID = "stale-session"
	proof := reducer.Apply(stale)
	if proof.Complete {
		t.Fatalf("ack from another session must not complete the proof")
	}
	if !contains(proof.Missing, model.FrontLeft) {
		t.Fatalf("ack from another session must leave its corner missing")
	}

	for _, corner := range []model.Corner{model.FrontRight, model.RearLeft, model.RearRight} {
		reducer.Apply(ack(corner, true, false))
	}
	if proof = reducer.Snapshot(); proof.Complete {
		t.Fatalf("three current-session locks plus one stale ack must not complete the proof")
	}
}

func contains(corners []model.Corner, want model.Corner) bool {
	for _, corner := range corners {
		if corner == want {
			return true
		}
	}
	return false
}
