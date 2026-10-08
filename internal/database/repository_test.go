package database

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestFoodData(t *testing.T) {
	t.Chdir("../..")
	table, err := OpenFoods("internal/database/foodtable.json")
	if err != nil {
		t.Fatal(err)
	}
	ids, err := table.IDs()
	if err != nil {
		t.Fatal(err)
	}
	for _, locale := range []string{"en", "it"} {
		translation, err := GetTranslation(locale)
		if err != nil {
			t.Fatal(err)
		}
		if len(translation.FoodTypes) != len(ids) {
			t.Fatalf("%s food count differs", locale)
		}
		for _, food := range ids {
			factor, _ := table.Factor(food)
			if factor <= 0 || translation.FoodTypes[food] == "" {
				t.Errorf("invalid food %s in %s", food, locale)
			}
		}
	}
	if factor, err := table.Factor("FOOD_RICE"); err != nil || factor != 2.5 {
		t.Fatalf("rice: %v %v", factor, err)
	}
	if _, err := table.Factor("missing"); err == nil {
		t.Fatal("missing food accepted")
	}
}

func TestAnalytics(t *testing.T) {
	path := filepath.Join(t.TempDir(), "new", "analytics.json")
	a, err := OpenAnalytics(path)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Go(func() {
			if err := a.Increment("calculation"); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	snapshot := a.Snapshot()
	snapshot["calculation"] = 0
	loaded, err := OpenAnalytics(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Snapshot()["calculation"] != 30 || a.Snapshot()["calculation"] != 30 {
		t.Fatal("lost increments or exposed mutable state")
	}
	if err := os.WriteFile(path, []byte("broken"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenAnalytics(path); err == nil {
		t.Fatal("corrupt analytics accepted")
	}
}

func TestFoodReplacement(t *testing.T) {
	path := filepath.Join(t.TempDir(), "foods.json")
	replace := func(data string) {
		t.Helper()
		if err := os.WriteFile(path+".new", []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(path+".new", path); err != nil {
			t.Fatal(err)
		}
	}
	replace(`[{"name":"rice","factor":2.5}]`)
	store, err := OpenFoods(path)
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	replace(`[{"name":"rice","factor":3.5}]`)
	// A same-size replacement with the same timestamp must still be detected.
	if err := os.Chtimes(path, before.ModTime(), before.ModTime()); err != nil {
		t.Fatal(err)
	}
	if got, err := store.Factor("rice"); err != nil || got != 3.5 {
		t.Fatalf("replacement: %v %v", got, err)
	}
	replace(`[{"name":"beans","factor":2}]`)
	ids, err := store.IDs()
	if err != nil || len(ids) != 1 || ids[0] != "beans" {
		t.Fatalf("IDs: %v %v", ids, err)
	}
	ids[0] = "changed"
	if _, err := store.Factor("rice"); !errors.Is(err, ErrUnknownFood) {
		t.Fatalf("removed food: %v", err)
	}
	for _, data := range []string{``, `{`, `null`, `[]`, `[{"name":"x"}]`, `[{"name":"","factor":2}]`, `[{"name":"x","factor":-1}]`, `[{"name":"x","factor":2},{"name":"x","factor":3}]`} {
		replace(data)
		if _, err := store.IDs(); !errors.Is(err, ErrFoodData) {
			t.Errorf("accepted invalid table %s: %v", data, err)
		}
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Factor("beans"); !errors.Is(err, ErrFoodData) {
		t.Fatalf("missing file: %v", err)
	}
	replace(`[{"name":"recovered","factor":2}]`)
	if got, err := store.Factor("recovered"); err != nil || got != 2 {
		t.Fatalf("recovery: %v %v", got, err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Go(func() {
			for j := 0; j < 20; j++ {
				got, err := store.Factor("recovered")
				if err != nil || (got != 2 && got != 3) {
					t.Errorf("concurrent reload: %v %v", got, err)
				}
				if _, err := store.IDs(); err != nil {
					t.Error(err)
				}
			}
		})
	}
	for i := 0; i < 20; i++ {
		replace(`[{"name":"recovered","factor":3}]`)
		replace(`[{"name":"recovered","factor":2}]`)
	}
	wg.Wait()
}

func TestOpenFoodsRejectsEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "foods.json")
	if err := os.WriteFile(path, nil, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenFoods(path); !errors.Is(err, ErrFoodData) {
		t.Fatalf("empty file accepted: %v", err)
	}
}
