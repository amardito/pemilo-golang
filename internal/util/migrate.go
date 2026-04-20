package util

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"

	"github.com/amard/pemilo-golang/migrations"
	"github.com/lib/pq"
)

// pgAlreadyExists reports whether err is a PostgreSQL "already exists" error.
// Covers: 42P07 duplicate_table, 42710 duplicate_object, 42701 duplicate_column,
// 42P16 invalid_table_definition (unique index already exists), 42P11 duplicate_cursor.
func pgAlreadyExists(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "42P07", // duplicate_table
			"42710", // duplicate_object  (extension, role, …)
			"42701", // duplicate_column
			"42P11", // duplicate_cursor
			"42P16": // duplicate_object (index)
			return true
		}
	}
	return false
}

// RunMigrations creates the schema_migrations tracking table (if absent) and
// applies any unapplied goose-format *.sql migrations in lexicographic order.
// If a migration fails because the objects already exist (e.g. the DB was
// previously set up by goose directly) it is recorded as applied and skipped.
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

		if _, execErr := tx.Exec(upSQL); execErr != nil {
			_ = tx.Rollback()
			if pgAlreadyExists(execErr) {
				// Objects already exist — record as applied so we never retry.
				if _, recErr := db.Exec(
					`INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING`, fname,
				); recErr != nil {
					return fmt.Errorf("record skipped migration %s: %w", fname, recErr)
				}
				log.Printf("[migrate] already exists, marked applied: %s", fname)
				continue
			}
			return fmt.Errorf("apply migration %s: %w", fname, execErr)
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
