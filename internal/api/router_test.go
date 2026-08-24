package api

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/wyw14/cry-109/internal/control"
	"github.com/wyw14/cry-109/internal/timing"
)

func TestOperationalPagesAndHealth(t *testing.T) {
	runtime, err := control.NewRuntime(filepath.Join(t.TempDir(), "state"), timing.RealClock{})
	if err != nil {
		t.Fatal(err)
	}
	handler := NewServer(runtime).Handler()
	for _, path := range []string{"/healthz", "/lifts", "/spreader", "/motion", "/safety", "/api/snapshot"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s returned %d: %s", path, response.Code, response.Body.String())
		}
	}
}
