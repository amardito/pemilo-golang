package domain

import (
	"errors"
	"regexp"
	"strings"
)

var tokenPattern = regexp.MustCompile(`^[A-Z0-9]{8}$`)

func NormalizeNIM(input string) string {
	cleaned := strings.ToUpper(strings.TrimSpace(input))
	return strings.Join(strings.Fields(cleaned), "")
}

func ValidateToken(token string) error {
	token = strings.TrimSpace(token)
	if !tokenPattern.MatchString(token) {
		return errors.New("token must match ^[A-Z0-9]{8}$")
	}
	return nil
}
