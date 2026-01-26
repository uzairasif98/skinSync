package utils

import (
	"crypto/rand"
	"math/big"
)

const (
	lowercase = "abcdefghijklmnopqrstuvwxyz"
	uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits    = "0123456789"
	special   = "!@#$%&*"
	allChars  = lowercase + uppercase + digits + special
)

// GenerateSecurePassword generates a random secure password
// Length: 12 characters
// Contains: lowercase, uppercase, digits, special characters
func GenerateSecurePassword() (string, error) {
	length := 12
	password := make([]byte, length)

	// Ensure at least one of each type
	password[0] = randomChar(lowercase)
	password[1] = randomChar(uppercase)
	password[2] = randomChar(digits)
	password[3] = randomChar(special)

	// Fill the rest with random characters
	for i := 4; i < length; i++ {
		password[i] = randomChar(allChars)
	}

	// Shuffle the password
	shuffle(password)

	return string(password), nil
}

func randomChar(charset string) byte {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
	return charset[n.Int64()]
}

func shuffle(password []byte) {
	for i := len(password) - 1; i > 0; i-- {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		j := n.Int64()
		password[i], password[j] = password[j], password[i]
	}
}
