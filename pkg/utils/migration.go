package utils

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

// RunMigrations executes SQL migration files from a directory
func RunMigrations(db *sql.DB, migrationsDir string) error {
	// Read the init schema file
	initSchemPath := filepath.Join(migrationsDir, "init_schema.sql")

	schemaContent, err := os.ReadFile(initSchemPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute the schema
	if _, err := db.Exec(string(schemaContent)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}
