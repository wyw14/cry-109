package verifycase

import (
	"testing"
	"time"

	"github.com/wyw14/cry-109/internal/brake"
	"github.com/wyw14/cry-109/internal/gantry"
	"github.com/wyw14/cry-109/internal/model"
	"github.com/wyw14/cry-109/internal/storm"
)

func TestStormAnchorWaitsForBrakeHeldDwell(t *testing.T) {
	at := time.Date(2026, 8, 24, 10, 0, 0, 0, time.UTC)
	stop := gantry.NewStopProof(0.01, 2*time.Second)
	hold := brake.NewHoldProof(2 * time.Second)
	sequence := storm.NewSequence(stop, hold)
	stop.Observe(0, at)
	hold.Observe(model.BrakeState{Mode: model.BrakeFriction, TorqueKNM: 90, PressureBar: 70}, at)
	if err := sequence.Request(at); err != nil {
		t.Fatal(err)
	}
	if state := sequence.Update(at); state != model.AnchorWaiting {
		t.Fatalf("transient zero speed started anchor descent: %s", state)
	}
	held := model.BrakeState{Mode: model.BrakeHeld, TorqueKNM: 90, PressureBar: 125, Holding: true}
	hold.Observe(held, at)
	stop.Observe(0, at)
	if state := sequence.Update(at.Add(2100 * time.Millisecond)); state != model.AnchorDescending {
		t.Fatalf("proved brake hold and standstill did not start descent: %s", state)
	}
	stop.Observe(0.08, at.Add(2200*time.Millisecond))
	if state := sequence.Update(at.Add(2200 * time.Millisecond)); state != model.AnchorAborted {
		t.Fatalf("movement during descent did not abort the anchor: %s", state)
	}
}
