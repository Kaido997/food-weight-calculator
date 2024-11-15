package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
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

func NewTemplates() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("web/*.html")),
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			return
		}
		templ.Render(w, "index", translation)
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
        templ.Render(w, "calculation-result", response);
	})
    
	http.ListenAndServe(api.PORT, nil)

}
