package authservice

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestCheckAuth(t *testing.T) {
	hash := sha256.Sum256([]byte("secret"))
	for _, tc := range []struct {
		digest, password string
		want             bool
	}{
		{hex.EncodeToString(hash[:]), "secret", true},
		{hex.EncodeToString(hash[:]), "wrong", false},
		{hex.EncodeToString(hash[:]), "", false},
		{"", "secret", false},
		{"not-hex", "secret", false},
		{"ab", "secret", false},
	} {
		t.Setenv("SECRET__ADMIN_PASSWORD", tc.digest)
		if got := CheckAuth(tc.password); got != tc.want {
			t.Errorf("got %v, want %v", got, tc.want)
		}
	}
}
