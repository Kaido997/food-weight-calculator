package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/kaido997/weightcalc/internal/database"
	"github.com/kaido997/weightcalc/internal/food"
)

func TestRoutes(t *testing.T) {
	store, err := database.OpenFoods("internal/database/foodtable.json")
	if err != nil {
		t.Fatal(err)
	}
	analytics, err := database.OpenAnalytics(filepath.Join(t.TempDir(), "analytics.json"))
	if err != nil {
		t.Fatal(err)
	}
	handler, err := newHandler(analytics, food.New(store))
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte("test-password"))
	t.Setenv("SECRET__ADMIN_PASSWORD", hex.EncodeToString(digest[:]))
	cases := []struct {
		name, method, path, body, auth string
		status                         int
		want                           string
	}{
		{"Italian", "GET", "/?lang=it", "", "", 200, `lang="it"`},
		{"default remains English", "GET", "/", "", "", 200, "Cooked Weight Calculator"},
		{"unknown locale", "GET", "/?lang=fr", "", "", 200, `lang="en"`},
		{"English rice", "POST", "/calculate", "quantity=80&food-type=FOOD_RICE", "", 200, "The cooked weight is 200.00"},
		{"Italian rice", "POST", "/calculate?lang=it", "quantity=80&food-type=FOOD_RICE", "", 200, "Il peso cotto"},
		{"decimal", "POST", "/calculate", "quantity=1.5&food-type=FOOD_RICE", "", 200, "3.75"},
		{"zero", "POST", "/calculate", "quantity=0&food-type=FOOD_RICE", "", 200, "0.00"},
		{"missing", "POST", "/calculate", "", "", 400, "invalid quantity"},
		{"bad encoding", "POST", "/calculate", "quantity=%zz", "", 400, "invalid form"},
		{"unknown food", "POST", "/calculate", "quantity=80&food-type=missing", "", 400, "unknown food"},
		{"negative", "POST", "/calculate", "quantity=-1&food-type=FOOD_RICE", "", 400, "non-negative"},
		{"NaN", "POST", "/calculate", "quantity=NaN&food-type=FOOD_RICE", "", 400, "finite"},
		{"overflow", "POST", "/calculate", "quantity=1e308&food-type=FOOD_RICE", "", 400, "too large"},
		{"API", "POST", "/api/v1/calculate-cooked", `{"food-type":"FOOD_RICE","quantity":80}`, "", 200, `{"cooked-weight": 200.0}`},
		{"bad JSON", "POST", "/api/v1/calculate-cooked", `{`, "", 400, "invalid"},
		{"missing quantity", "POST", "/api/v1/calculate-cooked", `{"food-type":"FOOD_RICE"}`, "", 400, "invalid"},
		{"null", "POST", "/api/v1/calculate-cooked", `null`, "", 400, "invalid"},
		{"trailing JSON", "POST", "/api/v1/calculate-cooked", `{"food-type":"FOOD_RICE","quantity":80}{}`, "", 400, "one JSON"},
		{"API missing food", "POST", "/api/v1/calculate-cooked", `{"quantity":80}`, "", 400, "unknown food"},
		{"oversized", "POST", "/api/v1/calculate-cooked", strings.Repeat(" ", 4097) + `{}`, "", 400, "invalid"},
		{"method", "GET", "/calculate", "", "", 405, "Method Not Allowed"},
		{"unknown route", "GET", "/missing", "", "", 404, "404"},
		{"favicon", "GET", "/favicon.ico", "", "", 200, ""},
		{"unauthorized", "GET", "/admin/analytics", "", "", 401, "Unauthorized"},
		{"wrong password", "GET", "/admin/analytics", "", "wrong", 401, "Unauthorized"},
		{"analytics", "GET", "/admin/analytics", "", "test-password", 200, "Page load: 3"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if strings.HasPrefix(tc.path, "/api/") {
				r.Header.Set("Content-Type", "application/json")
			} else {
				r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
			r.Header.Set("Authorization", tc.auth)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if w.Code != tc.status || !strings.Contains(w.Body.String(), tc.want) {
				t.Fatalf("got %d %s", w.Code, w.Body.String())
			}
		})
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Go(func() {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/?lang=it", nil))
			if w.Code != 200 || !strings.Contains(w.Body.String(), `lang="it"`) {
				t.Errorf("concurrent page failed")
			}
		})
	}
	wg.Wait()
	if got := analytics.Snapshot()["page-load"]; got != 23 {
		t.Fatalf("page count: %d", got)
	}
}

func TestSharedFoodServiceReload(t *testing.T) {
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
	replace(`[{"name":"FOOD_RICE","factor":2.5}]`)
	store, err := database.OpenFoods(path)
	if err != nil {
		t.Fatal(err)
	}
	service := food.New(store)
	analytics, err := database.OpenAnalytics(filepath.Join(t.TempDir(), "analytics.json"))
	if err != nil {
		t.Fatal(err)
	}
	handler, err := newHandler(analytics, service)
	if err != nil {
		t.Fatal(err)
	}
	request := func(method, path, body string, status int) string {
		t.Helper()
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if strings.HasPrefix(path, "/api/") {
			r.Header.Set("Content-Type", "application/json")
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != status {
			t.Fatalf("%s: got %d: %s", path, w.Code, w.Body.String())
		}
		return w.Body.String()
	}
	for _, locale := range []string{"en", "it"} {
		page := request("GET", "/?lang="+locale, "", 200)
		if !strings.Contains(page, `value="FOOD_RICE"`) || strings.Contains(page, `value="FOOD_BEANS"`) {
			t.Fatal("UI options do not reflect database")
		}
	}
	request("POST", "/api/v1/calculate-cooked", `{"food-type":"FOOD_RICE","quantity":1,"unit":"oz"}`, 400)
	replace(`[{"name":"NEW_FOOD","factor":3}]`)
	for _, locale := range []string{"en", "it"} {
		page := request("GET", "/?lang="+locale, "", 200)
		if !strings.Contains(page, `value="NEW_FOOD">NEW_FOOD</option>`) || strings.Contains(page, `value="FOOD_RICE"`) {
			t.Fatal("UI did not reload IDs or apply label fallback")
		}
	}
	ids := request("GET", "/api/v1/foods", "", 200)
	if strings.TrimSpace(ids) != `["NEW_FOOD"]` {
		t.Fatalf("food list: %s", ids)
	}
	got, err := service.Calculate("NEW_FOOD", 2, "lbs")
	if err != nil || got != 6 {
		t.Fatalf("internal calculation: %v %v", got, err)
	}
	result := request("POST", "/api/v1/calculate-cooked", `{"food-type":"NEW_FOOD","quantity":2,"unit":"lbs"}`, 200)
	var response map[string]float64
	if err := json.Unmarshal([]byte(result), &response); err != nil || response["cooked-weight"] != got {
		t.Fatalf("public calculation differs: %s", result)
	}
	result = request("POST", "/calculate", "food-type=NEW_FOOD&quantity=2", 200)
	if !strings.Contains(result, "6.00") {
		t.Fatalf("UI calculation differs: %s", result)
	}
	request("POST", "/api/v1/calculate-cooked", `{"food-type":"FOOD_RICE","quantity":2}`, 400)
	replace(`broken`)
	request("GET", "/", "", 503)
	request("GET", "/api/v1/foods", "", 503)
	request("POST", "/calculate", "food-type=NEW_FOOD&quantity=2", 503)
	request("POST", "/api/v1/calculate-cooked", `{"food-type":"NEW_FOOD","quantity":2}`, 503)
	replace(`[{"name":"NEW_FOOD","factor":4}]`)
	if result := request("POST", "/calculate", "food-type=NEW_FOOD&quantity=2", 200); !strings.Contains(result, "8.00") {
		t.Fatalf("recovery: %s", result)
	}
}
