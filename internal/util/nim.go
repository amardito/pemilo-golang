package util

import (
	"regexp"
	"strings"
)

var whitespaceRe = regexp.MustCompile(`\s+`)

// NormalizeNIM trims and removes all whitespace from a NIM string.
func NormalizeNIM(nim string) string {
	nim = strings.TrimSpace(nim)
	nim = whitespaceRe.ReplaceAllString(nim, "")
	return nim
}
