package food

import (
	"math"
	"sort"
	"testing"

	"github.com/kaido997/weightcalc/internal/database"
)

func TestCalculateCookedFood(t *testing.T) {
	t.Chdir("../..")
	table, err := database.OpenFoods("internal/database/foodtable.json")
	if err != nil {
		t.Fatal(err)
	}
	translation, err := database.GetTranslation("en")
	if err != nil {
		t.Fatal(err)
	}
	for food := range translation.FoodTypes {
		factor, err := table.Factor(food)
		if err != nil {
			t.Fatal(err)
		}
		got, err := New(table).Calculate(food, 80, "g")
		if err != nil || got != 80*factor {
			t.Errorf("%s: %v %v", food, got, err)
		}
	}
	for _, quantity := range []float64{-1, math.NaN(), math.Inf(1), math.MaxFloat64} {
		if _, err := New(table).Calculate("FOOD_RICE", quantity, "g"); err == nil {
			t.Errorf("accepted %v", quantity)
		}
	}
	if _, err := New(table).Calculate("unknown", 80, "g"); err == nil {
		t.Fatal("accepted unknown food")
	}
}

func TestUnitsAndGetAll(t *testing.T) {
	t.Chdir("../..")
	store, err := database.OpenFoods("internal/database/foodtable.json")
	if err != nil {
		t.Fatal(err)
	}
	service := New(store)
	for _, unit := range []string{"g", "gr", "lbs"} {
		got, err := service.Calculate("FOOD_RICE", 1.5, unit)
		if err != nil || got != 3.75 {
			t.Errorf("%s: %v %v", unit, got, err)
		}
	}
	for _, unit := range []string{"", "oz", "invalid"} {
		if _, err := service.Calculate("FOOD_RICE", 1, unit); err == nil {
			t.Errorf("accepted %q", unit)
		}
	}
	ids, err := service.GetAll()
	if err != nil || len(ids) == 0 || !sort.StringsAreSorted(ids) {
		t.Fatalf("IDs: %v %v", ids, err)
	}
	got, err := service.Calculate("FOOD_RICE", 0, "g")
	if err != nil || got != 0 {
		t.Fatalf("zero: %v %v", got, err)
	}
}
