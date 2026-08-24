package verifycase

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/wyw14/cry-109/internal/api"
	"github.com/wyw14/cry-109/internal/control"
	"github.com/wyw14/cry-109/internal/timing"
)

func TestDraftChangeRevalidatesPendingClearanceTrajectory(t *testing.T) {
	runtime, err := control.NewRuntime(filepath.Join(t.TempDir(), "state"), timing.RealClock{})
	if err != nil {
		t.Fatal(err)
	}
	geometry := runtime.Spreader.Telescope().Geometry()
	sway := runtime.Models.GetOrBuild("lift-5", geometry, 30, 26)
	profile := runtime.Vessel.Current()
	if _, err := runtime.Trajectories.Plan("lift-5", sway, profile.Revision, 24, 1.8); err != nil {
		t.Fatal(err)
	}
	handler := api.NewServer(runtime).Handler()
	request := httptest.NewRequest(http.MethodPost, "/api/motion/vessel-draft", bytes.NewBufferString(`{"draft_meters":12.5,"deck_meters":8.1,"hatch_meters":11.5}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("draft endpoint returned %d: %s", response.Code, response.Body.String())
	}
	pending, ok := runtime.Trajectories.Pending("lift-5")
	if !ok || pending.Valid || pending.VesselRevision == runtime.Vessel.Current().Revision {
		t.Fatalf("pending trajectory continued using the old clearance envelope: %+v", pending)
	}
}
