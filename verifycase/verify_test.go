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

func TestSpreaderTelescopeRebuildsAntiswayModel(t *testing.T) {
	runtime, err := control.NewRuntime(filepath.Join(t.TempDir(), "state"), timing.RealClock{})
	if err != nil {
		t.Fatal(err)
	}
	beforeGeometry := runtime.Spreader.Telescope().Geometry()
	before := runtime.Models.GetOrBuild("lift-4", beforeGeometry, 32, 28)
	handler := api.NewServer(runtime).Handler()
	request := httptest.NewRequest(http.MethodPost, "/api/spreader/telescope", bytes.NewBufferString(`{"length_feet":40,"mass_tonnes":12.8}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("telescope endpoint returned %d: %s", response.Code, response.Body.String())
	}
	afterGeometry := runtime.Spreader.Telescope().Geometry()
	after := runtime.Models.GetOrBuild("lift-4", afterGeometry, 32, 28)
	if after.ID == before.ID || after.GeometryRevision != afterGeometry.Revision || after.LengthFeet != 40 {
		t.Fatalf("anti-sway cache retained compact model: before=%+v after=%+v", before, after)
	}
}
