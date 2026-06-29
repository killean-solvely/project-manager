package mcpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/killeanjohnson/projectmanager/internal/models"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (d *deps) registerPrompts(s *mcp.Server) {
	projectArg := []*mcp.PromptArgument{{
		Name:        "project_id",
		Description: "the project to work on",
		Required:    true,
	}}

	s.AddPrompt(&mcp.Prompt{
		Name:        "flesh_out_idea",
		Title:       "Flesh out an idea",
		Description: "Expand a rough idea into a clear summary and initial documentation.",
		Arguments:   projectArg,
	}, d.promptFleshOutIdea)

	s.AddPrompt(&mcp.Prompt{
		Name:        "generate_missing_docs",
		Title:       "Generate missing docs",
		Description: "Write the required documents a project is still missing or hasn't completed.",
		Arguments:   projectArg,
	}, d.promptGenerateMissingDocs)

	s.AddPrompt(&mcp.Prompt{
		Name:        "draft_tasks_from_spec",
		Title:       "Draft tasks from the spec",
		Description: "Break a project's spec document into task cards on its board.",
		Arguments:   projectArg,
	}, d.promptDraftTasksFromSpec)
}

func (d *deps) promptFleshOutIdea(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	id, err := projectIDArg(req)
	if err != nil {
		return nil, err
	}
	p, err := d.projects.GetProject(id)
	if err != nil {
		return nil, err
	}
	text := fmt.Sprintf(`You are fleshing out a project idea in the project manager.

Project: %s (id: %s, status: %s)
Summary: %s

Current notes:
%s

Do the following, using the available tools:
1. Tighten the summary to one crisp sentence and save it with update_project.
2. Expand the description into a clear problem statement, goals, and rough scope (update_project).
3. Draft the required docs (overview, technical, spec) with upsert_document.
4. Call get_missing_docs to confirm nothing required is left incomplete.

Keep edits grounded in the existing notes; ask for specifics only if truly blocked.`,
		p.Name, p.ID, p.Status, valueOr(p.Summary, "(none)"), valueOr(p.Description, "(none yet)"))
	return userPrompt("Flesh out: "+p.Name, text)
}

func (d *deps) promptGenerateMissingDocs(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	id, err := projectIDArg(req)
	if err != nil {
		return nil, err
	}
	p, err := d.projects.GetProject(id)
	if err != nil {
		return nil, err
	}
	missing, err := d.docs.GetMissingDocs(id)
	if err != nil {
		return nil, err
	}
	if len(missing) == 0 {
		return userPrompt("Docs complete: "+p.Name,
			fmt.Sprintf("Project %q (id: %s) has all required documents complete. Nothing to do.", p.Name, p.ID))
	}

	var b strings.Builder
	for _, m := range missing {
		fmt.Fprintf(&b, "- %s (%s)\n", m.Type, m.Reason)
	}
	text := fmt.Sprintf(`Fill in the missing documentation for project %q (id: %s).

Required docs not yet complete:
%s
For each one, read any existing project context first (use list_documents and the pm://project/%s resource), then write the document with upsert_document, setting status to "complete" once it is solid. Re-check with get_missing_docs when done.`,
		p.Name, p.ID, b.String(), p.ID)
	return userPrompt("Generate missing docs: "+p.Name, text)
}

func (d *deps) promptDraftTasksFromSpec(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	id, err := projectIDArg(req)
	if err != nil {
		return nil, err
	}
	p, err := d.projects.GetProject(id)
	if err != nil {
		return nil, err
	}
	list, err := d.docs.ListDocuments(id)
	if err != nil {
		return nil, err
	}
	var spec *models.Document
	for _, doc := range list {
		if doc.Type == models.DocumentTypeSpec {
			spec = doc
			break
		}
	}

	if spec == nil {
		text := fmt.Sprintf(`Project %q (id: %s) has no spec document yet. Write one first with upsert_document (type "spec"), then break it into tasks with create_task. Call get_board to get the column ids before creating tasks.`,
			p.Name, p.ID)
		return userPrompt("Draft tasks (no spec yet): "+p.Name, text)
	}

	text := fmt.Sprintf(`Break the spec for project %q (id: %s) into concrete task cards on its board.

Spec (%q):
%s

Steps:
1. Call get_board to get the column ids (default columns: Backlog, Todo, In progress, Done).
2. For each unit of work in the spec, call create_task into the Backlog column with a clear title and a short description.
3. Set a sensible priority, and link a task to a document with link_task_document where relevant.

Aim for tasks that are independently completable; avoid one giant catch-all card.`,
		p.Name, p.ID, spec.Title, spec.Content)
	return userPrompt("Draft tasks from spec: "+p.Name, text)
}

func projectIDArg(req *mcp.GetPromptRequest) (uuid.UUID, error) {
	idStr := req.Params.Arguments["project_id"]
	if idStr == "" {
		return uuid.Nil, fmt.Errorf("project_id is required")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid project_id: %w", err)
	}
	return id, nil
}

func userPrompt(description, text string) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Description: description,
		Messages: []*mcp.PromptMessage{{
			Role:    mcp.Role("user"),
			Content: &mcp.TextContent{Text: text},
		}},
	}, nil
}
