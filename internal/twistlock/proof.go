package twistlock

import (
	"fmt"
	"sync"

	"github.com/wyw14/cry-109/internal/model"
)

type ProofReducer struct {
	mu       sync.Mutex
	session  string
	acks     map[model.Corner]model.TwistlockAck
	complete bool
	failed   bool
}

func NewProofReducer(sessionID string) *ProofReducer {
	return &ProofReducer{session: sessionID, acks: make(map[model.Corner]model.TwistlockAck)}
}

func (r *ProofReducer) Reset(sessionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.session = sessionID
	r.acks = make(map[model.Corner]model.TwistlockAck)
	r.complete = false
	r.failed = false
}

func (r *ProofReducer) Apply(ack model.TwistlockAck) model.TwistlockProof {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ack.SessionID != r.session {
		return r.snapshot("ack belongs to a different engage session")
	}
	if _, known := validCorners[ack.Corner]; !known {
		return r.snapshot(fmt.Sprintf("unknown corner %q", ack.Corner))
	}
	r.acks[ack.Corner] = ack
	if ack.Failed || !ack.Locked {
		r.failed = true
	}
	r.complete = !r.failed && r.allLocked()
	return r.snapshot("")
}

func (r *ProofReducer) Snapshot() model.TwistlockProof {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.snapshot("")
}

func (r *ProofReducer) allLocked() bool {
	pairs := [][2]model.Corner{{model.FrontLeft, model.RearRight}, {model.FrontRight, model.RearLeft}}
	for _, pair := range pairs {
		left, leftOK := r.acks[pair[0]]
		right, rightOK := r.acks[pair[1]]
		if leftOK && rightOK && left.Locked && right.Locked && !left.Failed && !right.Failed {
			return left.SessionID == r.session && right.SessionID == r.session
		}
	}
	return false
}

func (r *ProofReducer) snapshot(reason string) model.TwistlockProof {
	missing := make([]model.Corner, 0, 4)
	for _, corner := range model.AllCorners() {
		ack, ok := r.acks[corner]
		if !ok || !ack.Locked || ack.Failed || ack.SessionID != r.session {
			missing = append(missing, corner)
		}
	}
	if r.failed && reason == "" {
		reason = "one or more twistlocks rejected engagement"
	}
	return model.TwistlockProof{
		SessionID: r.session,
		Complete:  r.complete,
		Failed:    r.failed,
		Missing:   missing,
		Reason:    reason,
	}
}

var validCorners = map[model.Corner]struct{}{
	model.FrontLeft: {}, model.FrontRight: {}, model.RearLeft: {}, model.RearRight: {},
}
