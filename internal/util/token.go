package util

import (
	"crypto/rand"
	"math/big"
)

const tokenCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const tokenLength = 8

// GenerateToken creates a cryptographically secure random 8-char uppercase alphanumeric token.
func GenerateToken() (string, error) {
	b := make([]byte, tokenLength)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(tokenCharset))))
		if err != nil {
			return "", err
		}
		b[i] = tokenCharset[n.Int64()]
	}
	return string(b), nil
}
