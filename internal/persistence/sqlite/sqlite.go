// Package sqlite provides SQLite-backed implementations of the persistence
// interfaces, using the pure-Go modernc.org/sqlite driver (no cgo). The repos
// here are drop-in replacements for the memory repos behind the same interfaces.
package sqlite

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// Open opens the database at the resolved path (PM_DB_PATH, or
// ~/.projectmanager/projectmanager.db by default) and applies the schema. Using
// the same path from cmd/api and cmd/mcp is what lets them share one store.
func Open() (*sql.DB, string, error) {
	path, err := resolveDBPath()
	if err != nil {
		return nil, "", err
	}
	db, err := OpenAt(path)
	if err != nil {
		return nil, "", err
	}
	return db, path, nil
}

// OpenAt opens (creating if needed) the database at an explicit path.
func OpenAt(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// One connection serializes access — simplest correct choice for a local
	// single-user tool, and keeps per-connection PRAGMAs reliable.
	db.SetMaxOpenConns(1)

	for _, pragma := range []string{
		"PRAGMA busy_timeout = 5000",
		"PRAGMA journal_mode = WAL",
		"PRAGMA foreign_keys = ON",
	} {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("apply %q: %w", pragma, err)
		}
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return db, nil
}

func resolveDBPath() (string, error) {
	if p := os.Getenv("PM_DB_PATH"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".projectmanager")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "projectmanager.db"), nil
}

// --- scan/serialize helpers ---

// rowScanner is satisfied by both *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, s)
}

func nullableTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return formatTime(*t)
}

func parseNullableTime(ns sql.NullString) (*time.Time, error) {
	if !ns.Valid || ns.String == "" {
		return nil, nil
	}
	t, err := parseTime(ns.String)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func nullableUUID(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return id.String()
}

func parseNullableUUID(ns sql.NullString) (*uuid.UUID, error) {
	if !ns.Valid || ns.String == "" {
		return nil, nil
	}
	id, err := uuid.Parse(ns.String)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func encodeStrings(ss []string) string {
	if ss == nil {
		ss = []string{}
	}
	b, _ := json.Marshal(ss)
	return string(b)
}

func decodeStrings(s string) []string {
	if s == "" {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal([]byte(s), &out); err != nil || out == nil {
		return []string{}
	}
	return out
}
