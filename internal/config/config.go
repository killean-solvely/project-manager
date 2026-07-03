// Package config owns all configuration loading for the backend binaries.
// Settings come from, in order of precedence: real environment variables,
// a .env file in the working directory (loaded via godotenv, which never
// overrides variables already set), and built-in defaults (via viper).
//
// Load is called once at each composition root (cmd/api, cmd/mcp) and the
// resulting Config is passed down explicitly; no other package reads the
// environment.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds every runtime setting for the backend.
type Config struct {
	// Port is the HTTP listen port for cmd/api. Env: PORT.
	Port string
	// DBPath is the SQLite database file path shared by cmd/api and
	// cmd/mcp. Env: DB_PATH, with PM_DB_PATH kept as a backward-compat
	// alias; DB_PATH wins when both are set.
	DBPath string
	// MCPHTTPEnabled gates whether cmd/api mounts the MCP server as an
	// HTTP handler at /mcp. Env: MCP_HTTP_ENABLED; defaults to true. The
	// stdio binary (cmd/mcp) is unaffected by this setting.
	MCPHTTPEnabled bool
}

// Load reads the optional .env file and the environment, and returns the
// resolved configuration. A missing .env is not an error; a malformed or
// unreadable one is.
func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	v := viper.New()
	v.SetDefault("port", "4523")
	v.SetDefault("mcp_http_enabled", true)

	// Later names are fallbacks: DB_PATH wins over the legacy PM_DB_PATH.
	_ = v.BindEnv("port", "PORT")
	_ = v.BindEnv("db_path", "DB_PATH", "PM_DB_PATH")
	_ = v.BindEnv("mcp_http_enabled", "MCP_HTTP_ENABLED")

	// Resolve the default lazily so an explicit DB_PATH/PM_DB_PATH still
	// works in environments without a home directory.
	dbPath := v.GetString("db_path")
	if dbPath == "" {
		var err error
		dbPath, err = defaultDBPath()
		if err != nil {
			return Config{}, fmt.Errorf("resolve default db path: %w", err)
		}
	}

	return Config{
		Port:           v.GetString("port"),
		DBPath:         dbPath,
		MCPHTTPEnabled: v.GetBool("mcp_http_enabled"),
	}, nil
}

// defaultDBPath computes (but does not create) the default database location,
// ~/.projectmanager/projectmanager.db. Directory creation is an open-time
// concern handled by sqlite.OpenAt.
func defaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".projectmanager", "projectmanager.db"), nil
}
