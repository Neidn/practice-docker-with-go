package util

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword returns the hash of a password.
func HashPassword(password string) (string, error) {
	fromPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(fromPassword), nil
}

// CheckPassword checks if the provided password is correct or not.
func CheckPassword(password string, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
