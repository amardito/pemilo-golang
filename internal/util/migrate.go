package util

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"

	"github.com/amard/pemilo-golang/migrations"
)

// RunMigrations creates the schema_migrations tracking table (if absent) and
// applies any unapplied goose-format *.sql migrations in lexicographic order.
func RunMigrations(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT        PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`)
	if err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, fname := range files {
		var count int
		if err := db.QueryRow(
			`SELECT COUNT(*) FROM schema_migrations WHERE version = $1`, fname,
		).Scan(&count); err != nil {
			return fmt.Errorf("check migration %s: %w", fname, err)
		}
		if count > 0 {
			continue
		}

		raw, err := fs.ReadFile(migrations.FS, fname)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", fname, err)
		}

		upSQL := extractGooseUp(string(raw))
		if strings.TrimSpace(upSQL) == "" {
			log.Printf("[migrate] %s: no Up section found, skipping", fname)
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", fname, err)
		}

		if _, err := tx.Exec(upSQL); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", fname, err)
		}

		if _, err := tx.Exec(
			`INSERT INTO schema_migrations (version) VALUES ($1)`, fname,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", fname, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", fname, err)
		}

		log.Printf("[migrate] applied: %s", fname)
	}

	return nil
}

// extractGooseUp returns the SQL between "-- +goose Up" and "-- +goose Down"
// markers. If there is no Up marker the whole content is returned as-is.
func extractGooseUp(content string) string {
	const upMarker = "-- +goose Up"
	const downMarker = "-- +goose Down"

	upIdx := strings.Index(content, upMarker)
	if upIdx == -1 {
		return strings.TrimSpace(content)
	}
	after := content[upIdx+len(upMarker):]
	if downIdx := strings.Index(after, downMarker); downIdx != -1 {
		after = after[:downIdx]
	}
	return strings.TrimSpace(after)
}
