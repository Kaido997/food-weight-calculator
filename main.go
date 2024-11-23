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
	"strconv"
	"github.com/kaido997/weightcalc/api/handler"
	"github.com/kaido997/weightcalc/internal/database"
)

const (
    GRAMS = "gr"
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
    Text string
    Value float32
    Unit string
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
        if val != nil && val["lang"] !=nil && slices.Contains(in, val["lang"][0]) {
            translation = database.GetTranslation(val["lang"][0])
    
        }
        database.CounterIncr("page-load")
		templ.Render(w, "index", translation)
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
        
        response :=  ResultResponseDTO{
            Text: translation.ResultLabel, 
            Value: api.CalculateCookedFood(data.FoodType,data.Quantity),
            Unit: fmt.Sprintf("(%s.)", GRAMS),
            
        }
        database.CounterIncr("calculation")
        templ.Render(w, "calculation-result", response);
	})
    
    port := os.Getenv("PORT")
    if port == "" {
        log.Printf("PORT not found. found '%s' setting to default 8080", port)
        port = "8080"
    }
    log.Fatal(http.ListenAndServe(":"+port, nil));

}
