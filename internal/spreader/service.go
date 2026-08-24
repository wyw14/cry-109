package spreader

import (
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-109/internal/model"
)

type Service struct {
	engage    *EngageController
	telescope *Telescope
}

func NewService() *Service {
	session := uuid.NewString()
	return &Service{
		engage:    NewEngageController(session),
		telescope: NewTelescope(model.Geometry{LengthFeet: 20, Revision: 1, MassTonnes: 11.5}),
	}
}

func (s *Service) StartEngage() string {
	session := uuid.NewString()
	s.engage.Begin(session)
	return session
}

func (s *Service) ConfirmAll(session string) model.TwistlockProof {
	var proof model.TwistlockProof
	for index, corner := range model.AllCorners() {
		proof = s.engage.Confirm(model.TwistlockAck{
			SessionID: session,
			Corner:    corner,
			Locked:    true,
			At:        time.Now().UTC().Add(time.Duration(index) * time.Millisecond),
		})
	}
	return proof
}

func (s *Service) EngageController() *EngageController { return s.engage }
func (s *Service) Telescope() *Telescope               { return s.telescope }
