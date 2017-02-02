package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// Encrypt converts value to a brcypt hash
func encrypt(val string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(val), 10)
	return string(b), err
}

// ValidateCryptedValue checks that the saved hash and raw value hash match
func checkCrypt(hash, val string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(val))
}
