package database

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Analytics struct {
	mu       sync.Mutex
	path     string
	counters map[string]uint
}

func OpenAnalytics(path string) (*Analytics, error) {
	a := &Analytics{path: path, counters: map[string]uint{"page-load": 0, "calculation": 0}}
	data, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(data, &a.counters); err != nil {
			return nil, err
		}
		if a.counters == nil {
			a.counters = make(map[string]uint)
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *Analytics) Increment(name string) error {
	// Serialize increments and disk writes so concurrent requests cannot lose counts.
	a.mu.Lock()
	defer a.mu.Unlock()
	a.counters[name]++
	data, err := json.Marshal(a.counters)
	if err != nil {
		return err
	}
	// Rename prevents an interrupted write from truncating the saved counters.
	if err := os.WriteFile(a.path+".tmp", data, 0644); err != nil {
		return err
	}
	return os.Rename(a.path+".tmp", a.path)
}

func (a *Analytics) Snapshot() map[string]uint {
	a.mu.Lock()
	defer a.mu.Unlock()
	snapshot := make(map[string]uint, len(a.counters))
	for key, value := range a.counters {
		snapshot[key] = value
	}
	return snapshot
}
