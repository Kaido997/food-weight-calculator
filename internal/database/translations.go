package database

import (
	"encoding/json"
	"os"
)

type Translation struct {
	UI                        map[string]string `json:"ui"`
	Title                     string            `json:"title"`
	RawWeightLabel            string            `json:"raw-weight-label"`
	RawWeightInputPlaceholder string            `json:"raw-weight-input-placeholder"`
	FoodTypeLabel             string            `json:"food-type-label"`
	CalcButtonLabel           string            `json:"calc-button-label"`
	ResultLabel               string            `json:"result-label"`
	MetaDescription           string            `json:"meta-description"`
	MetaKeywords              string            `json:"meta-keywords"`
	FoodTypes                 map[string]string `json:"food-types"`
}

func GetTranslation(locale string) (Translation, error) {
	var data Translation
	res, err := os.ReadFile("internal/database/translations/" + locale + ".json")
	if err != nil {
		return data, err
	}
	err = json.Unmarshal(res, &data)
	return data, err
}
