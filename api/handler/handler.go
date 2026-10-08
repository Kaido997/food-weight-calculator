package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/kaido997/weightcalc/internal/database"
	"github.com/kaido997/weightcalc/internal/food"
)

func (h handler) calculateWeight(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	decoder := json.NewDecoder(r.Body)
	var data struct {
		Unit     string   `json:"unit"`
		FoodType string   `json:"food-type"`
		Quantity *float64 `json:"quantity"`
	}
	if err := decoder.Decode(&data); err != nil || data.Quantity == nil {
		http.Error(w, "invalid calculation request", http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		http.Error(w, "expected one JSON object", http.StatusBadRequest)
		return
	}
	// Keep existing v1 clients working when unit is omitted.
	if data.Unit == "" {
		data.Unit = "g"
	}
	value, err := h.foods.Calculate(data.FoodType, *data.Quantity, data.Unit)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	// Preserve the endpoint's existing one-decimal precision.
	fmt.Fprintf(w, `{"cooked-weight": %.1f}`, value)
}

type handler struct{ foods *food.Service }

func Map(mux *http.ServeMux, foods *food.Service) {
	h := handler{foods: foods}
	mux.HandleFunc("POST /api/v1/calculate-cooked", h.calculateWeight)
	mux.HandleFunc("GET /api/v1/foods", func(w http.ResponseWriter, r *http.Request) {
		ids, err := foods.GetAll()
		if err != nil {
			writeError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ids)
	})
}

func writeError(w http.ResponseWriter, err error) {
	if errors.Is(err, database.ErrFoodData) {
		http.Error(w, "food data unavailable", http.StatusServiceUnavailable)
		return
	}
	http.Error(w, err.Error(), http.StatusBadRequest)
}
