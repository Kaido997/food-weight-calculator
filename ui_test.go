package main

import (
	"encoding/json"
	"image/png"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/kaido997/weightcalc/internal/database"
	"github.com/kaido997/weightcalc/internal/food"
)

func TestUIAndPWA(t *testing.T) {
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
	get := func(path string) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
		}
		return w
	}
	for _, lang := range []string{"en", "it"} {
		page := get("/?lang=" + lang).Body.String()
		for _, fragment := range []string{`lang="` + lang + `"`, `novalidate`, `aria-live="polite"`, `name="theme" value="system"`, `/manifest.webmanifest`, `name="unit" value="lbs"`, `id="food-search"`, `aria-controls="foodType"`} {
			if !strings.Contains(page, fragment) {
				t.Errorf("%s missing %s", lang, fragment)
			}
		}
		if strings.Contains(page, "Ready when you are") || strings.Contains(page, "Tutto pronto") {
			t.Fatal("obsolete readiness text")
		}
		match := regexp.MustCompile(`(?s)<script id="ui-messages" type="application/json">(.*?)</script>`).FindStringSubmatch(page)
		if len(match) != 2 {
			t.Fatal("missing UI messages")
		}
		var messages map[string]string
		if err := json.Unmarshal([]byte(match[1]), &messages); err != nil {
			t.Fatal(err)
		}
		translation, err := database.GetTranslation(lang)
		if err != nil {
			t.Fatal(err)
		}
		if messages["Calculate"] != translation.UI["Calculate"] {
			t.Fatal("messages do not follow locale")
		}
		for _, asset := range regexp.MustCompile(`(?:src|href)="(/assets/[^\"]+)"`).FindAllStringSubmatch(page, -1) {
			get(asset[1])
		}
	}
	var manifest struct {
		Display  string                        `json:"display"`
		StartURL string                        `json:"start_url"`
		Icons    []struct{ Src, Sizes string } `json:"icons"`
	}
	m := get("/manifest.webmanifest")
	if !strings.Contains(m.Header().Get("Content-Type"), "application/manifest+json") {
		t.Fatal("manifest MIME")
	}
	if err := json.Unmarshal(m.Body.Bytes(), &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Display != "standalone" || manifest.StartURL != "/" || len(manifest.Icons) != 2 {
		t.Fatalf("invalid manifest: %+v", manifest)
	}
	for _, icon := range manifest.Icons {
		config, err := png.DecodeConfig(get(icon.Src).Body)
		if err != nil {
			t.Fatal(err)
		}
		if config.Width != config.Height || (config.Width != 192 && config.Width != 512) {
			t.Fatal("invalid install icon dimensions")
		}
	}
	sw := get("/sw.js")
	if sw.Header().Get("Cache-Control") != "no-cache" {
		t.Fatal("service worker must revalidate")
	}
	for _, path := range []string{"/assets/app.css", "/assets/theme.js", "/assets/shell.js", "/assets/calculator.js", "/assets/food-prep.png"} {
		get(path)
	}

	for _, tc := range []struct {
		body   string
		status int
		want   string
	}{
		{"quantity=1.5&food-type=FOOD_RICE&unit=lbs", 200, `"cooked-weight":3.75`},
		{"quantity=1e308&food-type=FOOD_RICE&unit=g", 400, `"error":"invalid-weight"`},
		{"quantity=80&food-type=missing&unit=g", 400, `"error":"unknown-food"`},
	} {
		r := httptest.NewRequest("POST", "/calculate", strings.NewReader(tc.body))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != tc.status || !strings.Contains(w.Body.String(), tc.want) {
			t.Fatalf("JSON form: %d %s", w.Code, w.Body.String())
		}
		if tc.status == http.StatusOK && w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("calculation can be cached")
		}
	}
}
