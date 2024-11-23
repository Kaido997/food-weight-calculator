package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var Analytics map[string]uint

func save(data map[string]uint) error {

	encoded, err := json.Marshal(data)

	if err != nil {
	    return fmt.Errorf("Error while dumping json: %s", err)
	}

    wd, _ := os.Getwd()
	if err := os.WriteFile(filepath.Join(wd, "internal/database/analytics/analytics.json"), encoded, 0644); err != nil {
	    return fmt.Errorf("Error while wrinting file: %s", err)
	}
    return nil

}

func GetAnalytics() error {
	file, err := loadFile("internal/database/analytics/analytics.json")

	if err != nil {
		return fmt.Errorf("File loading error: %s", err)
	}

	data := make(map[string]uint)

	if err := json.Unmarshal(file, &data); err != nil {
		log.Fatalf("Unable to parse json: %s", err)
	}

	Analytics = data
	return nil

}

func NewCounterAnalytics(name string) {
	if Analytics == nil {
        Analytics = make(map[string]uint)
        Analytics[name] = 0
        if error := save(Analytics); error != nil {
            log.Fatalf("Save gone wrong %s", error)
        }
        return;

	}

	if len(Analytics) > 0 {
		keys := make([]string, 0, len(Analytics))
		for k := range Analytics {
			keys = append(keys, k)
		}
		for k := range keys {
			if keys[k] == name {
				log.Print("Analytics already exist")
                return
			}
		}
		Analytics[name] = 0
	} else {
		Analytics[name] = 0
	}
}

func CounterIncr(path string) error {

	if Analytics == nil {
		return fmt.Errorf("Undefined analytics")
	}

	if len(Analytics) > 0 {
		keys := make([]string, 0, len(Analytics))
		for k := range Analytics {
			keys = append(keys, k)
		}
		for k := range keys {
			if keys[k] == path {
				Analytics[path]++
                if err := save(Analytics); err != nil {
                    log.Fatalf("Cannot increment '%s' because %s occured", path, err)
                }

				break
			}
		}
	} else {
		return fmt.Errorf("Empty map")
	}

	return fmt.Errorf("Path '%s not found", path)
}
