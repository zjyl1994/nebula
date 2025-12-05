package util

import (
	"math/rand/v2"
)

func COALESCE[T comparable](v ...T) T {
	var defaultValue T
	for _, val := range v {
		if val != defaultValue {
			return val
		}
	}
	return defaultValue
}

func RandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.IntN(len(letterBytes))]
	}
	return string(b)
}
