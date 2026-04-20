package migrations

import "embed"

// FS holds every *.sql migration file embedded at compile time.
//
//go:embed *.sql
var FS embed.FS
