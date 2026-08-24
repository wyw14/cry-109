package verifycase

import (
	"path/filepath"
	"testing"

	"github.com/wyw14/cry-109/internal/control"
	"github.com/wyw14/cry-109/internal/model"
	"github.com/wyw14/cry-109/internal/timing"
)

func TestRegenRejectionTransfersToFrictionBrake(t *testing.T) {
	runtime, err := control.NewRuntime(filepath.Join(t.TempDir(), "state"), timing.RealClock{})
	if err != nil {
		t.Fatal(err)
	}
	state, err := runtime.ApplyRegen(420, false)
	if err != nil {
		t.Fatal(err)
	}
	if state.Mode != model.BrakeFriction || state.TorqueKNM < 100 || state.LastFault != "regen sink unavailable" {
		t.Fatalf("regen rejection did not arm friction braking: %+v", state)
	}
	mode, result := runtime.Lowering.Snapshot()
	if mode != model.BrakeFriction || result.Accepted {
		t.Fatalf("lowering loop retained regenerative mode: mode=%s result=%+v", mode, result)
	}
}
