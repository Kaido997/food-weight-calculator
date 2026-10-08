package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"

	api "github.com/kaido997/weightcalc/api/handler"
	"github.com/kaido997/weightcalc/internal/database"
	"github.com/kaido997/weightcalc/internal/food"
	authservice "github.com/kaido997/weightcalc/services/auth_service"
)

//go:embed web/*.html web/assets/* web/manifest.webmanifest web/sw.js
var resources embed.FS

type pageData struct {
	database.Translation
	Language string
	Foods    [][2]string
}

func newHandler(analytics *database.Analytics, foods *food.Service) (http.Handler, error) {
	templates, err := template.ParseFS(resources, "web/*.html")
	if err != nil {
		return nil, err
	}
	pages := make(map[string]pageData, 2)
	for _, language := range []string{"en", "it"} {
		translation, err := database.GetTranslation(language)
		if err != nil {
			return nil, err
		}
		page := pageData{Translation: translation, Language: language}

		pages[language] = page
	}
	pageFor := func(r *http.Request) pageData {
		if page, ok := pages[r.URL.Query().Get("lang")]; ok {
			return page
		}
		return pages["en"]
	}
	render := func(w http.ResponseWriter, status int, name string, data any) {
		var body bytes.Buffer
		if err := templates.ExecuteTemplate(&body, name, data); err != nil {
			log.Printf("render %s: %v", name, err)
			http.Error(w, "unable to render page", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(status)
		_, _ = body.WriteTo(w)
	}
	count := func(name string) {
		if err := analytics.Increment(name); err != nil {
			log.Printf("analytics: %v", err)
		}
	}
	mux := http.NewServeMux()
	api.Map(mux, foods)
	assets, err := fs.Sub(resources, "web/assets")
	if err != nil {
		return nil, err
	}
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServerFS(assets)))
	mux.HandleFunc("GET /manifest.webmanifest", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/manifest+json")
		http.ServeFileFS(w, r, resources, "web/manifest.webmanifest")
	})
	mux.HandleFunc("GET /sw.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFileFS(w, r, resources, "web/sw.js")
	})
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		ids, err := foods.GetAll()
		if err != nil {
			http.Error(w, "food data unavailable", http.StatusServiceUnavailable)
			return
		}
		page := pageFor(r)
		for _, id := range ids {
			label := page.FoodTypes[id]
			if label == "" {
				label = id
			}
			page.Foods = append(page.Foods, [2]string{id, label})
		}
		sort.Slice(page.Foods, func(i, j int) bool {
			if page.Foods[i][1] == page.Foods[j][1] {
				return page.Foods[i][0] < page.Foods[j][0]
			}
			return page.Foods[i][1] < page.Foods[j][1]
		})
		count("page-load")
		render(w, http.StatusOK, "index", page)
	})
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, resources, "web/assets/favicon.ico")
	})
	mux.HandleFunc("POST /calculate", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		quantity, err := strconv.ParseFloat(r.PostForm.Get("quantity"), 64)
		if err != nil {
			http.Error(w, "invalid quantity", http.StatusBadRequest)
			return
		}
		unit := r.PostForm.Get("unit")
		if unit == "" {
			unit = "g"
		}
		value, err := foods.Calculate(r.PostForm.Get("food-type"), quantity, unit)
		if err != nil {
			if errors.Is(err, database.ErrFoodData) {
				http.Error(w, "food data unavailable", http.StatusServiceUnavailable)
				return
			}
			if r.Header.Get("Accept") == "application/json" {
				code := "invalid-weight"
				if errors.Is(err, database.ErrUnknownFood) {
					code = "unknown-food"
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": code})
			} else {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			return
		}
		count("calculation")
		w.Header().Set("Cache-Control", "no-store")
		if r.Header.Get("Accept") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]float64{"cooked-weight": value})
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "%s %.2f (%s)", pageFor(r).ResultLabel, value, unit)
	})
	mux.HandleFunc("GET /admin/analytics", func(w http.ResponseWriter, r *http.Request) {
		if !authservice.CheckAuth(r.Header.Get("Authorization")) {
			render(w, http.StatusUnauthorized, "unauthorized", nil)
			return
		}
		counts := analytics.Snapshot()
		render(w, http.StatusOK, "analytics-counter", struct{ PageLoad, Calculation uint }{counts["page-load"], counts["calculation"]})
	})
	return mux, nil
}

func main() {
	store, err := database.OpenFoods("internal/database/foodtable.json")
	if err != nil {
		log.Fatal(err)
	}
	foods := food.New(store)
	analytics, err := database.OpenAnalytics("internal/database/analytics/analytics.json")
	if err != nil {
		log.Fatal(err)
	}
	handler, err := newHandler(analytics, foods)
	if err != nil {
		log.Fatal(err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
