package utils

import (
	mathrand "math/rand"

	"github.com/google/uuid"
)

// GenerateID creates a proper UUID
func GenerateID() string {
	return uuid.New().String()
}

// GenerateRandomNickname generates a random nickname with 3-7 characters
func GenerateRandomNickname() string {
	// アルファベットと数字の組み合わせ
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// 3-7文字のランダムな長さ
	length := mathrand.Intn(5) + 3 // 3から7文字

	result := make([]byte, length)
	for i := range length {
		result[i] = chars[mathrand.Intn(len(chars))]
	}

	return string(result)
}
