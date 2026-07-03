package mcpserver_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/killeanjohnson/projectmanager/internal/models"
	"github.com/killeanjohnson/projectmanager/internal/persistence/sqlite"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestStdioBinary spawns the real cmd/mcp binary and connects to it the way a
// real MCP client (e.g. Claude) does: over stdio, holding the pipe open.
func TestStdioBinary(t *testing.T) {
	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "smoke", Version: "0.0.0"}, nil)
	transport := &mcp.CommandTransport{
		Command: exec.Command("go", "run", "github.com/killeanjohnson/projectmanager/cmd/mcp"),
	}
	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("connect over stdio: %v", err)
	}
	defer cs.Close()

	tools, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools over stdio: %v", err)
	}
	if len(tools.Tools) < 18 {
		t.Fatalf("expected >= 18 tools over stdio, got %d", len(tools.Tools))
	}
}

// TestStdioBinarySharesDatabase proves the shared-store goal: a project written
// to the DB file by one process is visible to the MCP binary pointed at the same
// DB path - exactly how cmd/api and cmd/mcp will share data. It deliberately
// passes the legacy PM_DB_PATH alias (not DB_PATH) so the backward-compat name
// stays covered end to end.
func TestStdioBinarySharesDatabase(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "shared.db")

	// Simulate another process (e.g. the API) writing a project to the shared DB.
	db, err := sqlite.OpenAt(dbPath)
	if err != nil {
		t.Fatalf("seed open: %v", err)
	}
	now := time.Now()
	if err := sqlite.NewProjectsRepository(db).Save(&models.Project{
		ID: uuid.New(), Name: "Shared idea", Status: models.ProjectStatusIdea,
		Tags: []string{}, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("seed save: %v", err)
	}
	_ = db.Close()

	// Launch the MCP binary pointed at the same DB.
	cmd := exec.Command("go", "run", "github.com/killeanjohnson/projectmanager/cmd/mcp")
	// Filter config vars from the inherited env so a DB_PATH set in the
	// developer's or CI environment cannot outrank the alias under test.
	env := make([]string, 0, len(os.Environ())+1)
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "DB_PATH=") || strings.HasPrefix(kv, "PM_DB_PATH=") || strings.HasPrefix(kv, "PORT=") {
			continue
		}
		env = append(env, kv)
	}
	cmd.Env = append(env, "PM_DB_PATH="+dbPath)
	client := mcp.NewClient(&mcp.Implementation{Name: "smoke", Version: "0.0.0"}, nil)
	cs, err := client.Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer cs.Close()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "list_projects", Arguments: map[string]any{}})
	if err != nil {
		t.Fatalf("list_projects: %v", err)
	}
	var out struct {
		Projects []struct {
			Name string `json:"name"`
		} `json:"projects"`
	}
	if err := json.Unmarshal([]byte(contentText(res.Content)), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	found := false
	for _, p := range out.Projects {
		if p.Name == "Shared idea" {
			found = true
		}
	}
	if !found {
		t.Fatalf("MCP binary did not see the project written by the other process: %+v", out.Projects)
	}
}
