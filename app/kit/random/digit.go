package random

import (
	"math/rand"
	"time"
)

// Digit : to generate the number string
func Digit(n int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	var letterRunes = []rune("1234567890")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	s := string(b)
	return s
}
