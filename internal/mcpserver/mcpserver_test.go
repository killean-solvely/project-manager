package mcpserver_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/killeanjohnson/projectmanager/internal/boards"
	"github.com/killeanjohnson/projectmanager/internal/docs"
	"github.com/killeanjohnson/projectmanager/internal/mcpserver"
	"github.com/killeanjohnson/projectmanager/internal/persistence/memory"
	"github.com/killeanjohnson/projectmanager/internal/projects"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// connect wires the services, builds the MCP server, and connects an in-process
// client to it over an in-memory transport.
func connect(t *testing.T) (*mcp.ClientSession, context.Context) {
	t.Helper()
	ctx := context.Background()

	projectsSvc := projects.NewService(memory.NewProjectsRepository())
	docsSvc := docs.NewService(memory.NewDocumentsRepository())
	boardsSvc := boards.NewService(
		memory.NewBoardsRepository(),
		memory.NewColumnsRepository(),
		memory.NewTasksRepository(),
	)
	srv := mcpserver.New(projectsSvc, docsSvc, boardsSvc)

	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	if _, err := srv.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0.0.0"}, nil)
	cs, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs, ctx
}

func callTool(t *testing.T, cs *mcp.ClientSession, ctx context.Context, name string, args map[string]any, out any) {
	t.Helper()
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("call %s: transport error: %v", name, err)
	}
	text := contentText(res.Content)
	if res.IsError {
		t.Fatalf("call %s: tool error: %s", name, text)
	}
	if out != nil {
		if err := json.Unmarshal([]byte(text), out); err != nil {
			t.Fatalf("call %s: unmarshal result %q: %v", name, text, err)
		}
	}
}

func contentText(content []mcp.Content) string {
	var b strings.Builder
	for _, c := range content {
		if tc, ok := c.(*mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}

func TestServerSurface(t *testing.T) {
	cs, ctx := connect(t)

	// Tools are registered.
	tools, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if len(tools.Tools) < 18 {
		t.Fatalf("expected >= 18 tools, got %d", len(tools.Tools))
	}
	if !hasTool(tools.Tools, "create_idea") || !hasTool(tools.Tools, "get_missing_docs") || !hasTool(tools.Tools, "move_task") {
		t.Fatalf("missing expected tools; got %v", toolNames(tools.Tools))
	}

	// Prompts are registered.
	prompts, err := cs.ListPrompts(ctx, nil)
	if err != nil {
		t.Fatalf("list prompts: %v", err)
	}
	if len(prompts.Prompts) != 3 {
		t.Fatalf("expected 3 prompts, got %d", len(prompts.Prompts))
	}
}

func TestEndToEndFlow(t *testing.T) {
	cs, ctx := connect(t)

	// Create an idea.
	var created struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	callTool(t, cs, ctx, "create_idea", map[string]any{
		"name":    "Demo",
		"summary": "smoke test",
		"tags":    []string{"demo"},
	}, &created)
	if created.ID == "" || created.Status != "idea" {
		t.Fatalf("unexpected created project: %+v", created)
	}
	id := created.ID

	// A fresh idea is missing all three required docs.
	var missing struct {
		Missing []struct {
			Type   string `json:"type"`
			Reason string `json:"reason"`
		} `json:"missing"`
	}
	callTool(t, cs, ctx, "get_missing_docs", map[string]any{"id": id}, &missing)
	if len(missing.Missing) != 3 {
		t.Fatalf("expected 3 missing docs, got %d (%+v)", len(missing.Missing), missing.Missing)
	}

	// Promote it.
	var promoted struct {
		Status string `json:"status"`
		Mode   string `json:"mode"`
	}
	callTool(t, cs, ctx, "promote_project", map[string]any{"id": id}, &promoted)
	if promoted.Status != "active" || promoted.Mode != "developing" {
		t.Fatalf("unexpected promote result: %+v", promoted)
	}

	// Board exists with default columns; grab the Todo column.
	var board struct {
		Columns []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"columns"`
	}
	callTool(t, cs, ctx, "get_board", map[string]any{"id": id}, &board)
	var todoCol string
	for _, c := range board.Columns {
		if c.Name == "Todo" {
			todoCol = c.ID
		}
	}
	if todoCol == "" {
		t.Fatalf("no Todo column on board: %+v", board.Columns)
	}

	// Create a task in Todo.
	var task struct {
		ID       string `json:"id"`
		ColumnID string `json:"column_id"`
	}
	callTool(t, cs, ctx, "create_task", map[string]any{
		"project_id": id,
		"column_id":  todoCol,
		"title":      "Write the spec",
		"priority":   "high",
	}, &task)
	if task.ID == "" || task.ColumnID != todoCol {
		t.Fatalf("unexpected task: %+v", task)
	}

	// It shows up in the task list.
	var tasks struct {
		Tasks []json.RawMessage `json:"tasks"`
	}
	callTool(t, cs, ctx, "list_tasks", map[string]any{"id": id}, &tasks)
	if len(tasks.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks.Tasks))
	}

	// Read the project resource.
	res, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "pm://project/" + id})
	if err != nil {
		t.Fatalf("read resource: %v", err)
	}
	if len(res.Contents) == 0 || !strings.Contains(res.Contents[0].Text, "Demo") {
		t.Fatalf("project resource did not contain the project name: %+v", res.Contents)
	}

	// Get a prompt, templated with live project state.
	pr, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{
		Name:      "flesh_out_idea",
		Arguments: map[string]string{"project_id": id},
	})
	if err != nil {
		t.Fatalf("get prompt: %v", err)
	}
	if len(pr.Messages) == 0 {
		t.Fatalf("expected prompt messages")
	}
}

func hasTool(tools []*mcp.Tool, name string) bool {
	for _, tool := range tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}

func toolNames(tools []*mcp.Tool) []string {
	names := make([]string, len(tools))
	for i, tool := range tools {
		names[i] = tool.Name
	}
	return names
}
