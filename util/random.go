package util

import (
	"fmt"
	"math/rand"
	"time"
)

const alphaNum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func init() {
	rand.NewSource(time.Now().UnixNano())
}

// RandomInt returns a random integer in [min, max]
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString returns a random string of length n
func RandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = alphaNum[rand.Intn(len(alphaNum))]
	}
	return string(b)
}

// RandomOwner returns a random owner name
func RandomOwner() string {
	return RandomString(6)
}

// RandomMoney returns a random money amount
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomCurrency returns a random currency
func RandomCurrency() string {
	currencies := []string{USD, EUR, CAD}
	return currencies[rand.Intn(len(currencies))]
}

// RandomEmail returns a random email
func RandomEmail() string {
	return fmt.Sprintf("%s@gmail.com", RandomString(6))
}
