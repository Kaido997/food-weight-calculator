// Package food provides the shared application API for UI and HTTP callers.
package food

import (
	"fmt"
	"math"

	"github.com/kaido997/weightcalc/internal/database"
)

type Service struct{ store *database.FoodStore }

func New(store *database.FoodStore) *Service { return &Service{store: store} }

// Calculate returns cooked weight in the same unit as rawWeight.
// Cooking factors are ratios, so no intermediate unit conversion is necessary.
func (s *Service) Calculate(foodID string, rawWeight float64, unit string) (float64, error) {
	switch unit {
	case "g", "gr", "lbs":
	default:
		return 0, fmt.Errorf("unsupported unit: %q", unit)
	}
	if rawWeight < 0 || math.IsNaN(rawWeight) || math.IsInf(rawWeight, 0) {
		return 0, fmt.Errorf("quantity must be a finite, non-negative number")
	}
	factor, err := s.store.Factor(foodID)
	if err != nil {
		return 0, err
	}
	cooked := rawWeight * factor
	if math.IsInf(cooked, 0) {
		return 0, fmt.Errorf("quantity is too large")
	}
	return cooked, nil
}

// GetAll returns sorted food IDs from the current file, independent of translations.
func (s *Service) GetAll() ([]string, error) { return s.store.IDs() }
