package utils

import (
	"crypto/rand"
	"fmt"
)

// GenerateID creates a random hex string ID
func GenerateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}