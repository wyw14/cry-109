package spreader

import (
	"errors"
	"sync"

	"github.com/wyw14/cry-109/internal/model"
	"github.com/wyw14/cry-109/internal/twistlock"
)

type EngageState string

const (
	EngageOpen    EngageState = "open"
	EngageLocking EngageState = "locking"
	EngageLocked  EngageState = "locked"
	EngageFailed  EngageState = "failed"
)

type EngageController struct {
	mu      sync.Mutex
	session string
	state   EngageState
	proofs  *twistlock.ProofReducer
}

func NewEngageController(sessionID string) *EngageController {
	return &EngageController{session: sessionID, state: EngageOpen, proofs: twistlock.NewProofReducer(sessionID)}
}

func (c *EngageController) Begin(sessionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.session = sessionID
	c.state = EngageLocking
	c.proofs.Reset(sessionID)
}

func (c *EngageController) Confirm(ack model.TwistlockAck) model.TwistlockProof {
	c.mu.Lock()
	defer c.mu.Unlock()
	proof := c.proofs.Apply(ack)
	switch {
	case proof.Failed:
		c.state = EngageFailed
	case proof.Complete:
		c.state = EngageLocked
	default:
		c.state = EngageLocking
	}
	return proof
}

func (c *EngageController) Proof() model.TwistlockProof {
	return c.proofs.Snapshot()
}

func (c *EngageController) State() EngageState {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}

// RequireLocked gates load transfer on an explicit four-corner proof: the
// engage session must match, the proof must be complete and not failed, and
// no corner may be missing. The Complete flag is derived from all four
// corners locking in this session, but the missing-corner check is kept
// explicit so that a future weakening of that derivation can never, on its
// own, let the hoist transfer load before every twistlock is seated.
func (c *EngageController) RequireLocked() error {
	proof := c.Proof()
	switch {
	case proof.SessionID != c.session:
		return errors.New("spreader lock proof belongs to a different engage session")
	case proof.Failed:
		return errors.New("spreader does not have a complete current-session lock proof: a corner rejected engagement")
	case !proof.Complete:
		return errors.New("spreader does not have a complete current-session lock proof: not all corners locked")
	case len(proof.Missing) != 0:
		return errors.New("spreader does not have a complete current-session lock proof: corners still missing")
	}
	return nil
}
