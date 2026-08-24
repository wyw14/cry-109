package lift

import (
	"errors"
	"sync"

	"github.com/wyw14/cry-109/internal/model"
)

type LoadPermit struct {
	mu        sync.Mutex
	sessionID string
	issued    bool
	reason    string
}

func NewLoadPermit(sessionID string) *LoadPermit {
	return &LoadPermit{sessionID: sessionID}
}

func (p *LoadPermit) Evaluate(proof model.TwistlockProof) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.issued = false
	switch {
	case proof.SessionID != p.sessionID:
		p.reason = "twistlock proof is from another engage session"
	case proof.Failed:
		p.reason = "twistlock engagement failed"
	case !proof.Complete || len(proof.Missing) != 0:
		p.reason = "four-corner lock proof is incomplete"
	default:
		p.issued = true
		p.reason = ""
	}
	return p.issued
}

func (p *LoadPermit) Require() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.issued {
		return errors.New(p.reason)
	}
	return nil
}

func (p *LoadPermit) Status() (bool, string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.issued, p.reason
}
