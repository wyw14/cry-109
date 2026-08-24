package control

import (
	"fmt"
	"sort"
	"time"

	"github.com/wyw14/cry-109/internal/model"
)

type ComponentHealth struct {
	Name      string    `json:"name"`
	Available bool      `json:"available"`
	State     string    `json:"state"`
	CheckedAt time.Time `json:"checked_at"`
}

type Diagnostics struct {
	Ready      bool              `json:"ready"`
	Components []ComponentHealth `json:"components"`
	Invariants []InvariantResult `json:"invariants"`
	Summary    string            `json:"summary"`
	CheckedAt  time.Time         `json:"checked_at"`
}

func (r *Runtime) Diagnostics() Diagnostics {
	now := r.Clock.Now()
	snapshot := r.Snapshot()
	anchor, anchorReason, _ := r.Storm.Sequence().Snapshot()
	_, paired := r.Heave.Fusion().Last()
	compensation := r.Compensation.Command()
	_, loadsValid := r.LoadCells.Latest()
	cycles, commits := r.Tandem.Stats()
	eventCount, journalErr := r.Repository.EventCount()
	components := []ComponentHealth{
		health("journal", journalErr == nil, fmt.Sprintf("lifts=%d events=%d", len(snapshot.Lifts), eventCount), now),
		health("spreader", r.Spreader != nil, fmt.Sprintf("geometry-r%d", snapshot.Geometry.Revision), now),
		health("twistlock", r.Spreader.EngageController() != nil, string(r.Spreader.EngageController().State()), now),
		health("heave", r.Heave != nil, fmt.Sprintf("paired=%d degraded=%t", paired, compensation.Degraded), now),
		health("tandem", r.Tandem != nil, fmt.Sprintf("load-sample=%t cycles=%d commits=%d", loadsValid, cycles, commits), now),
		health("antisway", r.Models != nil, fmt.Sprintf("models=%d", r.Models.Size()), now),
		health("trolley", r.Trolley != nil, trolleyState(r), now),
		health("vessel", r.Vessel != nil, fmt.Sprintf("profile-r%d", r.Vessel.Current().Revision), now),
		health("brake", r.Brakes != nil, string(snapshot.Brake.Mode), now),
		health("storm", r.Storm != nil, string(anchor)+reasonSuffix(anchorReason), now),
		health("gantry", r.Gantry != nil, gantryState(r), now),
		health("rail-calibration", r.Calibration != nil, fmt.Sprintf("frame-r%d", snapshot.RailPose.Frame.Revision), now),
		health("zones", r.Zones != nil, fmt.Sprintf("active=%d", len(r.Zones.All())), now),
		health("motion", r.Motion != nil, "planner-online", now),
		health("energy", r.Energy != nil, fmt.Sprintf("accepted-kwh=%.3f", r.Energy.AcceptedKWh()), now),
		health("clearance", r.Clearance != nil, reevingState(r), now),
		health("interlock", r.Incidents != nil, fmt.Sprintf("active=%d", len(snapshot.Incidents)), now),
	}
	sort.Slice(components, func(i, j int) bool { return components[i].Name < components[j].Name })
	ready := true
	for _, component := range components {
		ready = ready && component.Available
	}
	invariants := r.CheckInvariants()
	ready = ready && invariantsHold(invariants)
	return Diagnostics{
		Ready: ready, Components: components, Invariants: invariants,
		Summary:   fmt.Sprintf("%d controllers online, %d active incidents", len(components), len(snapshot.Incidents)),
		CheckedAt: now,
	}
}

func health(name string, available bool, state string, at time.Time) ComponentHealth {
	return ComponentHealth{Name: name, Available: available, State: state, CheckedAt: at}
}

func trolleyState(r *Runtime) string {
	position, velocity, trajectory := r.Trolley.Snapshot()
	if trajectory == "" {
		return fmt.Sprintf("idle@%.2f velocity=%.2f", position, velocity)
	}
	return fmt.Sprintf("trajectory=%s@%.2f velocity=%.2f", trajectory, position, velocity)
}

func gantryState(r *Runtime) string {
	position, velocity := r.Gantry.Snapshot()
	return fmt.Sprintf("rail=%.2f velocity=%.2f", position, velocity)
}

func reevingState(r *Runtime) string {
	config := r.Reeving.Current()
	decision := r.Clearance.Check(8, 46)
	hookTravel := r.Reeving.HookTravel(8)
	_, estimated := r.Clearance.Estimator().Estimate(8)
	return fmt.Sprintf("factor=%d revision=%d estimator-r%d hook-travel=%.2f clearance=%t", config.Factor, config.Revision, estimated.Revision, hookTravel, decision.Allowed)
}

func reasonSuffix(reason string) string {
	if reason == "" {
		return ""
	}
	return ":" + reason
}

func (r *Runtime) RecordIncident(kind, component, message string) model.SafetyIncident {
	return r.Incidents.Trip(kind, component, message, r.Clock.Now())
}
