package main

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"sort"
	"strconv"

	"github.com/kaido997/weightcalc/api/handler"
	"github.com/kaido997/weightcalc/internal/database"
	authservice "github.com/kaido997/weightcalc/services/auth_service"
)

const (
	GRAMS  = "gr"
	POUNDS = "lbs"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

//go:embed web/*.html
var resources embed.FS

func NewTemplates() *Templates {

	return &Templates{
		templates: template.Must(template.ParseFS(resources, "web/*.html")),
	}
}

type ResultResponseDTO struct {
	Text  string
	Value float32
	Unit  string
}
type TranslationDTO struct {
	Title                     string      `json:"title"`
	RawWeightLabel            string      `json:"raw-weight-label"`
	RawWeightInputPlaceholder string      `json:"raw-weight-input-placeholder"`
	FoodTypeLabel             string      `json:"food-type-label"`
	CalcButtonLabel           string      `json:"calc-button-label"`
	ResultLabel               string      `json:"result-label"`
	MetaDescription           string      `json:"meta-description"`
	MetaKeywords              string      `json:"meta-keywords"`
	FoodTypes                 [][2]string `json:"food-types"`
}

func buildTrasnlationResponseDTO(data database.Translation) TranslationDTO {
	keys := make([]string, 0, len(data.FoodTypes))
	for k := range data.FoodTypes {
		keys = append(keys, k)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return data.FoodTypes[keys[i]] < data.FoodTypes[keys[j]]
	})

	sorted := make([][2]string, len(keys))

	for i, k := range keys {
		sorted[i] = [2]string{k, data.FoodTypes[k]}
	}
    return TranslationDTO{
        Title: data.Title,
        RawWeightLabel: data.RawWeightLabel,
        RawWeightInputPlaceholder: data.RawWeightInputPlaceholder,
        FoodTypeLabel: data.FoodTypeLabel,
        CalcButtonLabel: data.CalcButtonLabel,
        ResultLabel: data.ResultLabel,
        MetaDescription: data.MetaDescription,
        MetaKeywords: data.MetaKeywords,
        FoodTypes: sorted,
    }

}

func main() {
	log.Println("Starting application")
	database.LoadTable()
	api.Map()
	translation := database.GetTranslation("en")
	templ := NewTemplates()

	database.GetAnalytics()

	database.NewCounterAnalytics("page-load")
	database.NewCounterAnalytics("calculation")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			return
		}
		in := []string{"en", "it"}
		val := r.URL.Query()
		if val != nil && val["lang"] != nil && slices.Contains(in, val["lang"][0]) {
			translation = database.GetTranslation(val["lang"][0])

		}
		database.CounterIncr("page-load")
		templ.Render(w, "index", buildTrasnlationResponseDTO(translation))
	})
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("get icon")
		if r.Method != "GET" {
			return
		}
		http.ServeFile(w, r, database.GetFaviconPath())
	})

	http.HandleFunc("/calculate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}

		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "BadRequest: %s", err)
		}
		float, err := strconv.ParseFloat(r.PostForm["quantity"][0], 32)
		if err != nil {
			fmt.Fprintf(w, "BadRequest: %s", err)
		}
		var data api.CalcCookedDTO = api.CalcCookedDTO{FoodType: r.PostForm["food-type"][0], Quantity: float32(float)}

		response := ResultResponseDTO{
			Text:  translation.ResultLabel,
			Value: api.CalculateCookedFood(data.FoodType, data.Quantity),
			Unit:  fmt.Sprintf("(%s.)", GRAMS),
		}
		database.CounterIncr("calculation")
		templ.Render(w, "calculation-result", response)
	})

	http.HandleFunc("/admin/analytics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			return
		}
		type AnalyticsDTO struct {
			PageLoad    uint
			Calculation uint
		}

		val := r.Header.Get("authorization")
        log.Println(val)
		if val != ""  {
			if authservice.CheckAuth(val) {
				database.GetAnalytics()
				templ.Render(w, "analytics-counter", AnalyticsDTO{PageLoad: database.Analytics["page-load"], Calculation: database.Analytics["calculation"]})
			} else {
				templ.Render(w, "unauthorized", nil)

			}

		} else {
			templ.Render(w, "unauthorized", nil)
		}

	})

	port := os.Getenv("PORT")
	if port == "" {
        log.Printf("PORT not found. found '%s' setting default port: 8080", port)
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))

}
