package verifycase

import (
	"sync"
	"testing"
	"time"

	"github.com/wyw14/cry-109/internal/model"
	"github.com/wyw14/cry-109/internal/tandem"
)

func TestTandemLoadErrorProducesOneCoordinatedCorrection(t *testing.T) {
	balancer := tandem.NewBalancer(0.6, 0.8, 0.12)
	sample := model.LoadSample{Cycle: 91, LeftKN: 530, RightKN: 470, Captured: time.Now()}
	const workers = 32
	start := make(chan struct{})
	results := make(chan model.SpeedTargets, workers)
	errors := make(chan error, workers)
	var group sync.WaitGroup
	for index := 0; index < workers; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			value, err := balancer.Coordinate(sample)
			results <- value
			errors <- err
		}()
	}
	close(start)
	group.Wait()
	close(results)
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatalf("parallel coordinator returned an error: %v", err)
		}
	}
	var expected model.SpeedTargets
	for value := range results {
		if expected.Cycle == 0 {
			expected = value
		}
		if value != expected || !value.Coordinated {
			t.Fatalf("same cycle produced independent targets: first=%+v current=%+v", expected, value)
		}
	}
	cycles, commits := balancer.Stats()
	if cycles != 1 || commits != 1 {
		t.Fatalf("one sample cycle was committed more than once: cycles=%d commits=%d", cycles, commits)
	}
}
