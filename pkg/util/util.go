package util

import (
	"crypto/rand"
	"encoding/hex"
)

func RandomString(n int) string {
	bytes := make([]byte, n)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:n]
}
