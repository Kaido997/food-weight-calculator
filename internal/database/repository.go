package database

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
)

var ErrUnknownFood = errors.New("unknown food")
var ErrFoodData = errors.New("food data unavailable")

// FoodStore reloads the file on demand; it owns the cached, validated snapshot.
// Replace the file by atomic rename so readers see a complete old or new version.
type FoodStore struct {
	path    string
	mu      sync.Mutex
	source  []byte
	factors map[string]float64
}

func OpenFoods(path string) (*FoodStore, error) {
	store := &FoodStore{path: path}
	_, err := store.snapshot()
	return store, err
}

func (s *FoodStore) snapshot() (map[string]float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFoodData, err)
	}
	// Compare content, not timestamps: a replacement can preserve size and mtime.
	// The small JSON file is read each time, but parsed only when it changes.
	if s.factors != nil && bytes.Equal(data, s.source) {
		return s.factors, nil
	}
	var rows []struct {
		Name   string  `json:"name"`
		Factor float64 `json:"factor"`
	}
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFoodData, err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("%w: empty food table", ErrFoodData)
	}
	factors := make(map[string]float64, len(rows))
	for _, row := range rows {
		if strings.TrimSpace(row.Name) == "" || row.Factor <= 0 || math.IsInf(row.Factor, 0) || math.IsNaN(row.Factor) {
			return nil, fmt.Errorf("%w: invalid food entry %q", ErrFoodData, row.Name)
		}
		if _, exists := factors[row.Name]; exists {
			return nil, fmt.Errorf("%w: duplicate food %q", ErrFoodData, row.Name)
		}
		factors[row.Name] = row.Factor
	}
	// Publish only complete, valid snapshots. Returned maps are never mutated.
	s.source, s.factors = data, factors
	return factors, nil
}

func (s *FoodStore) Factor(id string) (float64, error) {
	factors, err := s.snapshot()
	if err != nil {
		return 0, err
	}
	factor, ok := factors[id]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrUnknownFood, id)
	}
	return factor, nil
}

func (s *FoodStore) IDs() ([]string, error) {
	factors, err := s.snapshot()
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(factors))
	for id := range factors {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids, nil
}
