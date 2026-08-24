package clearance

import (
	"github.com/wyw14/cry-109/internal/hoist"
	"github.com/wyw14/cry-109/internal/lift"
)

type Service struct {
	reeving   *hoist.ReevingService
	estimator *lift.HeightEstimator
	permit    *Permit
}

func NewService(reeving *hoist.ReevingService, estimator *lift.HeightEstimator, permit *Permit) *Service {
	reeving.Subscribe(estimator.OnReevingChanged)
	reeving.Subscribe(permit.OnReevingChanged)
	return &Service{reeving: reeving, estimator: estimator, permit: permit}
}

func (s *Service) Check(drumMeters, requiredHeight float64) Decision {
	height, config := s.estimator.Estimate(drumMeters)
	return s.permit.Check(height, config, requiredHeight)
}

func (s *Service) Reeving() *hoist.ReevingService   { return s.reeving }
func (s *Service) Estimator() *lift.HeightEstimator { return s.estimator }
