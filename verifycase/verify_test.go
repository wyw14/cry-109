package verifycase

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/wyw14/cry-109/internal/control"
)

type fixedClock struct{ current time.Time }

func (c *fixedClock) Now() time.Time { return c.current }

func TestReevingChangeUpdatesHeightAndClearance(t *testing.T) {
	clock := &fixedClock{current: time.Date(2026, 8, 24, 10, 0, 0, 0, time.UTC)}
	runtime, err := control.NewRuntime(filepath.Join(t.TempDir(), "state"), clock)
	if err != nil {
		t.Fatal(err)
	}
	decision, err := runtime.ChangeReeving(8, 16, 47)
	if err != nil {
		t.Fatalf("current reeving did not reach clearance consumers: %v decision=%+v", err, decision)
	}
	height, config := runtime.Height.Estimate(16)
	if config.Factor != 8 || config.Revision != runtime.Reeving.Current().Revision || height < 47.9 || height > 48.1 {
		t.Fatalf("height estimator retained old reeving: height=%.2f config=%+v current=%+v", height, config, runtime.Reeving.Current())
	}
	if current := runtime.Clearance.Check(16, 47); !current.Allowed {
		t.Fatalf("clearance did not use updated height: %+v", current)
	}
}
