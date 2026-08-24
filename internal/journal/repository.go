package journal

import (
	"sync"

	"github.com/wyw14/cry-109/internal/model"
)

type Repository struct {
	log       *Log
	snapshots *SnapshotStore
	mu        sync.RWMutex
	lifts     map[string]model.Lift
}

func NewRepository(log *Log, snapshots *SnapshotStore) *Repository {
	return &Repository{log: log, snapshots: snapshots, lifts: make(map[string]model.Lift)}
}

func (r *Repository) Restore() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	restored := make(map[string]model.Lift)
	ok, err := r.snapshots.Load(&restored)
	if err != nil {
		return err
	}
	if ok {
		r.lifts = restored
	}
	return nil
}

func (r *Repository) PutLift(lift model.Lift, event model.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.log.Append(event); err != nil {
		return err
	}
	r.lifts[lift.ID] = lift
	return r.snapshots.Save(r.lifts)
}

func (r *Repository) Lift(id string) (model.Lift, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	lift, ok := r.lifts[id]
	return lift, ok
}

func (r *Repository) Lifts() []model.Lift {
	r.mu.RLock()
	defer r.mu.RUnlock()
	values := make([]model.Lift, 0, len(r.lifts))
	for _, lift := range r.lifts {
		values = append(values, lift)
	}
	return values
}

func (r *Repository) EventCount() (int, error) {
	events, err := r.log.ReadAll()
	return len(events), err
}
