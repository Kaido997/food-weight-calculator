package authservice

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"os"
)

func CheckAuth(pass string) bool {
	expected, err := hex.DecodeString(os.Getenv("SECRET__ADMIN_PASSWORD"))
	if err != nil || len(expected) != sha256.Size || pass == "" {
		return false
	}
	actual := sha256.Sum256([]byte(pass))
	return subtle.ConstantTimeCompare(expected, actual[:]) == 1
}
