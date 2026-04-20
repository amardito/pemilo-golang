package util

import (
	"regexp"
	"strings"
)

var tokenRegex = regexp.MustCompile(`^[A-Z0-9]{8}$`)

// NormalizeToken uppercases and trims a token string.
func NormalizeToken(token string) string {
	return strings.ToUpper(strings.TrimSpace(token))
}

// ValidateToken checks if a token matches the expected format.
func ValidateToken(token string) bool {
	return tokenRegex.MatchString(token)
}
