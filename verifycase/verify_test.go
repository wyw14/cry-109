package verifycase

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/wyw14/cry-109/internal/api"
	"github.com/wyw14/cry-109/internal/control"
	"github.com/wyw14/cry-109/internal/timing"
)

func TestRailOriginCorrectionTransformsActiveExclusionZones(t *testing.T) {
	runtime, err := control.NewRuntime(filepath.Join(t.TempDir(), "state"), timing.RealClock{})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/motion/rail-origin", bytes.NewBufferString(`{"origin":4,"destination":312}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	api.NewServer(runtime).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("rail calibration endpoint returned %d: %s", response.Code, response.Body.String())
	}
	var status control.MotionStatus
	if err := json.Unmarshal(response.Body.Bytes(), &status); err != nil {
		t.Fatal(err)
	}
	if status.Travel.Allowed || len(status.Zones) != 1 {
		t.Fatalf("planner did not retain the physical exclusion zone: %+v", status)
	}
	if status.Zones[0].StartM != 308 || status.Zones[0].Frame.ID != status.RailPose.Frame.ID {
		t.Fatalf("zone was not transformed into the new frame: pose=%+v zone=%+v", status.RailPose, status.Zones[0])
	}
}
