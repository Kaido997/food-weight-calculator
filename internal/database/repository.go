package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type foodTable struct {
	Name   string  `json:"name"`
	Factor float32 `json:"factor"`
}

type table struct {
	t    []foodTable
	Keys []string
}

type notFound struct {
	arg string
}

var Repository *table

func (e *notFound) Error() string {
	return fmt.Sprintf("%s - not found", e.arg)
}

func loadFile(path string) ([]byte, error) {
	curr, _ := os.Getwd()
	res, err := os.ReadFile(curr + path)
	if err != nil {
		log.Fatalf("Load File Error: %s", err)
		return nil, err
	}
	return res, nil
}

func LoadTable() (*table, error) {
	res, err := loadFile("/internal/database/foodtable.json")
	if err != nil {
		log.Fatalf("Load table Error: %s", err)
		return nil, err
	}
	data := []foodTable{}

	if err := json.Unmarshal(res, &data); err != nil {
		log.Fatalf("Json parse Error: %s", err)
		return nil, err
	}
	var keys []string
	for val := range data {
		keys = append(keys, data[val].Name)
	}
	Repository = &table{t: data, Keys: keys}
	log.Println("Database loaded")
	return Repository, nil
}

func (repo *table) GetFactorFor(food string) (float32, error) {
	if Repository == nil {
		return 0, fmt.Errorf("Repository not initialized")
	}
	for i := 0; i < len(repo.t); i++ {
		if repo.t[i].Name == food {
			return repo.t[i].Factor, nil
		}
	}
	return 0, &notFound{arg: food}
}

type Translation struct {
	Title                     string            `json:"title"`
	RawWeightLabel            string            `json:"raw-weight-label"`
	RawWeightInputPlaceholder string            `json:"raw-weight-input-placeholder"`
	FoodTypeLabel             string            `json:"food-type-label"`
	CalcButtonLabel           string            `json:"calc-button-label"`
	ResultLabel               string            `json:"result-label"`
	FoodTypes                 map[string]string `json:"food-types"`
}

func GetTranslation(locale string) Translation {
	res, err := loadFile("/internal/database/translations/" + locale + ".json")
	if err != nil {
		log.Fatalf("Load table Error: %s", err)
	}
	var data Translation

	if err := json.Unmarshal(res, &data); err != nil {
		log.Fatalf("Json parse Error: %s", err)
	}

	return data

}

func GetFaviconPath() string {
    curr, _ := os.Getwd()
    return fmt.Sprintf("%s/web/assets/favicon.ico", curr)
}
