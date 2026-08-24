package control

import (
	"path/filepath"
	"time"

	"github.com/wyw14/cry-109/internal/antisway"
	"github.com/wyw14/cry-109/internal/brake"
	"github.com/wyw14/cry-109/internal/clearance"
	"github.com/wyw14/cry-109/internal/energy"
	"github.com/wyw14/cry-109/internal/gantry"
	"github.com/wyw14/cry-109/internal/heave"
	"github.com/wyw14/cry-109/internal/hoist"
	"github.com/wyw14/cry-109/internal/interlock"
	"github.com/wyw14/cry-109/internal/journal"
	"github.com/wyw14/cry-109/internal/lift"
	"github.com/wyw14/cry-109/internal/model"
	"github.com/wyw14/cry-109/internal/motion"
	"github.com/wyw14/cry-109/internal/spreader"
	"github.com/wyw14/cry-109/internal/storm"
	"github.com/wyw14/cry-109/internal/tandem"
	"github.com/wyw14/cry-109/internal/timing"
	"github.com/wyw14/cry-109/internal/trolley"
	"github.com/wyw14/cry-109/internal/vessel"
	"github.com/wyw14/cry-109/internal/zone"
)

type Runtime struct {
	Clock        timing.Clock
	Repository   *journal.Repository
	Lifts        *lift.Service
	Spreader     *spreader.Service
	Heave        *heave.Service
	Compensation *hoist.CompensationDrive
	LoadCells    *tandem.CellPair
	Tandem       *lift.TandemLift
	Models       *antisway.ModelCache
	Antisway     *antisway.Planner
	Trajectories *trolley.TrajectoryService
	Trolley      *trolley.Controller
	Vessel       *vessel.Store
	Brakes       *brake.Controller
	BrakeHold    *brake.HoldProof
	GantryStop   *gantry.StopProof
	Gantry       *gantry.Controller
	Storm        *storm.Service
	Landing      *lift.LandingDetector
	Energy       *energy.Router
	Lowering     *hoist.LoweringLoop
	Incidents    *interlock.Service
	Calibration  *gantry.Calibration
	Zones        *zone.Store
	Motion       *motion.Planner
	Reeving      *hoist.ReevingService
	Height       *lift.HeightEstimator
	Clearance    *clearance.Service
}

func NewRuntime(dataDir string, clock timing.Clock) (*Runtime, error) {
	if clock == nil {
		clock = timing.RealClock{}
	}
	repository := journal.NewRepository(
		journal.NewLog(filepath.Join(dataDir, "events.jsonl")),
		journal.NewSnapshotStore(filepath.Join(dataDir, "lifts.json")),
	)
	if err := repository.Restore(); err != nil {
		return nil, err
	}
	spreaderService := spreader.NewService()
	modelCache := antisway.NewModelCache()
	antiPlanner := antisway.NewPlanner()
	trajectoryService := trolley.NewTrajectoryService(antiPlanner)
	spreaderService.Telescope().Subscribe(modelCache.InvalidateGeometry)
	spreaderService.Telescope().Subscribe(trajectoryService.InvalidateGeometry)

	profile := model.VesselProfile{
		ID: "berth-7-vessel", Revision: 1, DraftMeters: 13.2,
		DeckMeters: 7.4, HatchMeters: 10.8, UpdatedAt: clock.Now(),
	}
	vesselStore := vessel.NewStore(profile)
	vesselStore.Subscribe(func(_, current model.VesselProfile) {
		trajectoryService.InvalidateVessel(current.Revision)
	})

	brakes := brake.NewController()
	brakeHold := brake.NewHoldProof(2 * time.Second)
	stopProof := gantry.NewStopProof(0.01, 2*time.Second)
	gantryController := gantry.NewController(280, stopProof)
	stormSequence := storm.NewSequence(stopProof, brakeHold)

	calibration := gantry.NewCalibration(0, 280)
	zones := zone.NewStore()
	calibration.Subscribe(zones.Transform)
	if _, err := zones.Add(312, 316, calibration.Frame()); err != nil {
		return nil, err
	}

	reeving := hoist.NewReevingService(4, clock.Now())
	height := lift.NewHeightEstimator(50, reeving.Current())
	clearancePermit := clearance.NewPermit(reeving.Current())
	clearanceService := clearance.NewService(reeving, height, clearancePermit)

	balancer := tandem.NewBalancer(0.6, 0.8, 0.12)
	coordinator := hoist.NewTandemCoordinator(balancer)
	runtime := &Runtime{
		Clock: clock, Repository: repository,
		Lifts: lift.NewService(repository), Spreader: spreaderService,
		Heave: heave.NewService(profile.DeckMeters), Compensation: hoist.NewCompensationDrive(1.2, 10),
		LoadCells: tandem.NewCellPair(), Tandem: lift.NewTandemLift(coordinator),
		Models: modelCache, Antisway: antiPlanner, Trajectories: trajectoryService,
		Trolley: trolley.NewController(0), Vessel: vesselStore,
		Brakes: brakes, BrakeHold: brakeHold, GantryStop: stopProof, Gantry: gantryController,
		Storm:   storm.NewService(stormSequence, gantryController, brakes),
		Landing: lift.NewLandingDetector(8, 300*time.Millisecond),
		Energy:  energy.NewRouter(850), Lowering: hoist.NewLoweringLoop(brakes, 140),
		Incidents: interlock.NewService(), Calibration: calibration,
		Zones: zones, Motion: motion.NewPlanner(zones),
		Reeving: reeving, Height: height, Clearance: clearanceService,
	}
	return runtime, nil
}
