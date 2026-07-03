package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/killeanjohnson/projectmanager/internal/config"
)

// setup isolates a test from the host: it moves the working directory to an
// empty temp dir (so no stray .env is picked up) and clears every variable
// the loader reads. t.Setenv also registers restoration of the originals.
func setup(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	for _, name := range []string{"PORT", "DB_PATH", "PM_DB_PATH", "MCP_HTTP_ENABLED"} {
		t.Setenv(name, "")
		_ = os.Unsetenv(name)
	}
	return dir
}

func TestDefaults(t *testing.T) {
	setup(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != "4523" {
		t.Errorf("Port = %q, want %q", cfg.Port, "4523")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}
	want := filepath.Join(home, ".projectmanager", "projectmanager.db")
	if cfg.DBPath != want {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, want)
	}
	if !cfg.MCPHTTPEnabled {
		t.Errorf("MCPHTTPEnabled = false, want true (default)")
	}
}

func TestMCPHTTPEnabledDisable(t *testing.T) {
	setup(t)
	t.Setenv("MCP_HTTP_ENABLED", "false")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.MCPHTTPEnabled {
		t.Errorf("MCPHTTPEnabled = true, want false when MCP_HTTP_ENABLED=false")
	}
}

func TestEnvOverrides(t *testing.T) {
	setup(t)
	t.Setenv("PORT", "9000")
	t.Setenv("DB_PATH", "/tmp/custom.db")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != "9000" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9000")
	}
	if cfg.DBPath != "/tmp/custom.db" {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, "/tmp/custom.db")
	}
}

func TestLegacyPMDBPathAlias(t *testing.T) {
	setup(t)
	t.Setenv("PM_DB_PATH", "/tmp/legacy.db")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DBPath != "/tmp/legacy.db" {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, "/tmp/legacy.db")
	}
}

func TestDBPathWinsOverLegacyAlias(t *testing.T) {
	setup(t)
	t.Setenv("DB_PATH", "/tmp/new.db")
	t.Setenv("PM_DB_PATH", "/tmp/legacy.db")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DBPath != "/tmp/new.db" {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, "/tmp/new.db")
	}
}

func TestDotEnvFile(t *testing.T) {
	dir := setup(t)
	env := "PORT=8123\nDB_PATH=/tmp/dotenv.db\n"
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(env), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != "8123" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8123")
	}
	if cfg.DBPath != "/tmp/dotenv.db" {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, "/tmp/dotenv.db")
	}
}

func TestRealEnvBeatsDotEnv(t *testing.T) {
	dir := setup(t)
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("PORT=8123\n"), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	t.Setenv("PORT", "9000")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != "9000" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9000")
	}
}

func TestMalformedDotEnv(t *testing.T) {
	dir := setup(t)
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("not a valid line\n"), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	if _, err := config.Load(); err == nil {
		t.Error("Load succeeded, want error for malformed .env")
	}
}
