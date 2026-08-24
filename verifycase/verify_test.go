package verifycase

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/wyw14/cry-109/internal/control"
	"github.com/wyw14/cry-109/internal/model"
)

type fixedClock struct{ current time.Time }

func (c *fixedClock) Now() time.Time { return c.current }
func (c *fixedClock) advance(duration time.Duration) { c.current = c.current.Add(duration) }

func TestWaveUnloadCannotProveContainerLanding(t *testing.T) {
	at := time.Date(2026, 8, 24, 10, 0, 0, 0, time.UTC)
	clock := &fixedClock{current: at}
	runtime, err := control.NewRuntime(filepath.Join(t.TempDir(), "state"), clock)
	if err != nil {
		t.Fatal(err)
	}
	session := "landing-7"
	release := runtime.PrimeLanding(session)
	runtime.Heave.State().Update(model.HeaveSample{CapturedAt: at, Meters: 0.5, Velocity: 0.8})
	first, err := runtime.ObserveLanding(release, control.LandingRequest{
		SessionID: session, LoadKN: 0.5, HookMeters: 8.4,
		ContactToleranceM: 0.08, LoadTransferComplete: true, ObservedAt: at,
	})
	if err != nil {
		t.Fatal(err)
	}
	clock.advance(400 * time.Millisecond)
	second, err := runtime.ObserveLanding(release, control.LandingRequest{
		SessionID: session, LoadKN: 0.4, HookMeters: 8.4,
		ContactToleranceM: 0.08, LoadTransferComplete: true, ObservedAt: clock.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if first.Landed || second.Landed || !second.TwistlocksHeld || second.TwistlocksOpen {
		t.Fatalf("wave unload released a suspended container: first=%+v second=%+v", first, second)
	}
}
