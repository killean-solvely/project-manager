package mcpserver_test

import "testing"

// TestListParamsAcceptStringOrArray verifies that list-shaped tool params (tags,
// labels) accept both a real JSON array and a comma-separated string — some MCP
// clients stringify arrays, and the tools should tolerate both.
func TestListParamsAcceptStringOrArray(t *testing.T) {
	cs, ctx := connect(t)

	// A comma-separated string is split into a list.
	var fromString struct {
		Tags []string `json:"tags"`
	}
	callTool(t, cs, ctx, "create_idea", map[string]any{"name": "A", "tags": "go, cli ,"}, &fromString)
	if len(fromString.Tags) != 2 || fromString.Tags[0] != "go" || fromString.Tags[1] != "cli" {
		t.Fatalf("string tags not coerced to [go cli]: %v", fromString.Tags)
	}

	// A real array is preserved.
	var fromArray struct {
		Tags []string `json:"tags"`
	}
	callTool(t, cs, ctx, "create_idea", map[string]any{"name": "B", "tags": []string{"x", "y"}}, &fromArray)
	if len(fromArray.Tags) != 2 || fromArray.Tags[0] != "x" || fromArray.Tags[1] != "y" {
		t.Fatalf("array tags not preserved: %v", fromArray.Tags)
	}
}
