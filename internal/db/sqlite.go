package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

type DB struct {
	*sql.DB
}

func New(path string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL mode: %w", err)
	}

	return &DB{DB: db}, nil
}

func (db *DB) Migrate() error {
	version := db.getSchemaVersion()

	if version == 0 {
		if _, err := db.Exec(schemaSQL); err != nil {
			return fmt.Errorf("apply schema: %w", err)
		}
		if err := db.setSchemaVersion(1); err != nil {
			return fmt.Errorf("set schema version: %w", err)
		}
	}

	// Add future migrations here
	// if version < 2 { ... }

	return nil
}

func (db *DB) getSchemaVersion() int {
	var version int
	row := db.QueryRow("SELECT version FROM schema_version ORDER BY version DESC LIMIT 1")
	if err := row.Scan(&version); err != nil {
		return 0
	}
	return version
}

func (db *DB) setSchemaVersion(version int) error {
	_, err := db.Exec("INSERT OR REPLACE INTO schema_version (version) VALUES (?)", version)
	return err
}

// GetDBPath returns the default database path following XDG spec
func GetDBPath() string {
	if dataDir := os.Getenv("XDG_DATA_HOME"); dataDir != "" {
		return filepath.Join(dataDir, "tsk", "tsk.db")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "tsk", "tsk.db")
}
