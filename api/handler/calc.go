package api

import (
	"github.com/kaido997/weightcalc/internal/database"
	"log"
)

func CalculateCookedFood(foodType string, rawQantity float32) float32 {
	factor, err := database.Repository.GetFactorFor(foodType)
	if err != nil {
		log.Fatalf("An Error occured: %s", err)
	}
	return rawQantity * factor
}
