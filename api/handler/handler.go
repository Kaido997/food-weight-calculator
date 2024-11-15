package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)
const ( 
    PORT = ":3000"
    VERSION = "1"
    API_BASE_URL = "/api/v" + VERSION + "/" 
)

type CalcCookedDTO struct {
	FoodType string  `json:"food-type"`
	Quantity float32 `json:"quantity"`
}


func calculateWeight(w http.ResponseWriter, req *http.Request) {
    if req.Method != "POST" {
        return;
    }
	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Fprintf(w, "Error")
		return
	}
	var data CalcCookedDTO
	if err := json.Unmarshal(body, &data); err != nil {
        fmt.Fprintf(w, "BadRequest: %s", err)
		return
	}
	fmt.Fprintf(w, "{cooked-weight: %.1f}", CalculateCookedFood(data.FoodType, data.Quantity))
}

func Map() {
    var routesMap map[string]func(w http.ResponseWriter, req *http.Request) 
    routesMap = make(map[string]func(w http.ResponseWriter, req *http.Request)) 
    
    routesMap[API_BASE_URL + "calculate-cooked"] = calculateWeight

    log.Printf("Mapped %d routes", len(routesMap))
    for k, v := range routesMap {
        log.Printf("Mapped %s", k);
        http.HandleFunc(k, v);

    }
}
