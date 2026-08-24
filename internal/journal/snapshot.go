package journal

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type SnapshotStore struct {
	path string
	mu   sync.RWMutex
}

func NewSnapshotStore(path string) *SnapshotStore {
	return &SnapshotStore{path: path}
}

func (s *SnapshotStore) Save(value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(s.path), ".snapshot-*")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if _, err = temporary.Write(append(data, '\n')); err != nil {
		temporary.Close()
		return err
	}
	if err = temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err = temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryName, s.path)
}

func (s *SnapshotStore) Load(target any) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(data, target)
}
