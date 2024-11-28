package authservice

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"slices"
)

func getSSKey() (string, error) {
    key, validkey := os.LookupEnv("SECRET__SERVER_KEY")
    env, validenv := os.LookupEnv("ENV")

    if validenv {
        switch env {
        case "PROD":
        if validkey {
                return key, nil
            } else {
                return "", fmt.Errorf("Key not found")
            }
        case "DEV":
        return "test-key", nil
        default:
        return "", fmt.Errorf("Invalid enviroment")
        }
    } else {
        return "test-key", nil
    }
}

func CheckAuth(pass string) bool {

    sha := sha256.New()

    sha.Write([]byte(pass))

    adminPass, exist := os.LookupEnv("SECRET__ADMIN_PASSWORD")
    
    if !exist {
        return false
    }


    if slices.Compare([]byte(adminPass), []byte(hex.EncodeToString(sha.Sum(nil)))) == 0 {
        return true
    }
    
    return false
}
