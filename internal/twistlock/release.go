package twistlock

import (
	"errors"
	"sync"

	"github.com/wyw14/cry-109/internal/model"
)

type ReleaseService struct {
	mu       sync.Mutex
	locked   bool
	session  string
	released bool
}

func NewReleaseService(sessionID string) *ReleaseService {
	return &ReleaseService{locked: true, session: sessionID}
}

func (s *ReleaseService) OnLanded(sessionID string, landed bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sessionID != s.session {
		return errors.New("landing proof belongs to another session")
	}
	if !landed {
		return errors.New("container landing is not physically proven")
	}
	s.locked = false
	s.released = true
	return nil
}

func (s *ReleaseService) ApplyProof(proof model.TwistlockProof) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if proof.SessionID != s.session || !proof.Complete || proof.Failed {
		return errors.New("cannot arm release without current lock proof")
	}
	s.locked = true
	s.released = false
	return nil
}

func (s *ReleaseService) State() (locked bool, released bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.locked, s.released
}
