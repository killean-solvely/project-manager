package mcpserver

import (
	"context"
	"fmt"

	"github.com/killeanjohnson/projectmanager/internal/models"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// --- tool input types (field schemas are generated from these tags) ---

type listProjectsIn struct {
	Status string `json:"status,omitempty" jsonschema:"optional lifecycle filter: idea, active, or archived"`
}

type projectIDIn struct {
	ID string `json:"id" jsonschema:"the project id"`
}

type createIdeaIn struct {
	Name        string   `json:"name" jsonschema:"the idea/project name"`
	Summary     string   `json:"summary,omitempty" jsonschema:"a one-line summary"`
	Description string   `json:"description,omitempty" jsonschema:"freeform markdown braindump"`
	Tags        []string `json:"tags,omitempty" jsonschema:"optional tags"`
}

type updateProjectIn struct {
	ID          string    `json:"id" jsonschema:"the project id"`
	Name        *string   `json:"name,omitempty"`
	Summary     *string   `json:"summary,omitempty"`
	Description *string   `json:"description,omitempty"`
	Tags        *[]string `json:"tags,omitempty"`
}

type promoteIn struct {
	ID   string `json:"id" jsonschema:"the project id"`
	Mode string `json:"mode,omitempty" jsonschema:"developing or maintaining; defaults to developing"`
}

type setModeIn struct {
	ID   string `json:"id" jsonschema:"the project id"`
	Mode string `json:"mode" jsonschema:"developing or maintaining"`
}

type archiveIn struct {
	ID     string `json:"id" jsonschema:"the project id"`
	Reason string `json:"reason,omitempty" jsonschema:"why it is being archived"`
}

type reviveIn struct {
	ID   string `json:"id" jsonschema:"the project id"`
	Mode string `json:"mode,omitempty" jsonschema:"developing or maintaining; defaults to maintaining"`
}

type upsertDocumentIn struct {
	ProjectID string `json:"project_id" jsonschema:"the project id"`
	Type      string `json:"type" jsonschema:"overview, technical, spec, api, runbook, or other"`
	Title     string `json:"title" jsonschema:"document title"`
	Content   string `json:"content" jsonschema:"the markdown body"`
	Status    string `json:"status,omitempty" jsonschema:"draft, in_review, or complete; defaults to draft"`
}

type createTaskIn struct {
	ProjectID   string   `json:"project_id" jsonschema:"the project whose board to add to"`
	ColumnID    string   `json:"column_id" jsonschema:"the target column id (see get_board)"`
	Title       string   `json:"title" jsonschema:"task title"`
	Description string   `json:"description,omitempty"`
	Priority    string   `json:"priority,omitempty" jsonschema:"none, low, medium, or high"`
	Labels      []string `json:"labels,omitempty"`
	DocumentID  *string  `json:"document_id,omitempty" jsonschema:"optional id of a document this task is about"`
}

type updateTaskIn struct {
	ID          string    `json:"id" jsonschema:"the task id"`
	Title       *string   `json:"title,omitempty"`
	Description *string   `json:"description,omitempty"`
	Priority    *string   `json:"priority,omitempty" jsonschema:"none, low, medium, or high"`
	Labels      *[]string `json:"labels,omitempty"`
}

type moveTaskIn struct {
	ID       string `json:"id" jsonschema:"the task id"`
	ColumnID string `json:"column_id" jsonschema:"the destination column id (see get_board)"`
	Position int    `json:"position" jsonschema:"position within the destination column (0-based)"`
}

type linkTaskDocumentIn struct {
	TaskID     string  `json:"task_id" jsonschema:"the task id"`
	DocumentID *string `json:"document_id,omitempty" jsonschema:"document id to link; omit or empty to unlink"`
}

type taskIDIn struct {
	ID string `json:"id" jsonschema:"the task id"`
}

func (d *deps) registerTools(s *mcp.Server) {
	// --- projects / lifecycle ---

	mcp.AddTool(s, &mcp.Tool{Name: "list_projects", Description: "List projects, optionally filtered by lifecycle status (idea, active, archived)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in listProjectsIn) (*mcp.CallToolResult, projectListView, error) {
			var filter *models.ProjectStatus
			if in.Status != "" {
				st := models.ProjectStatus(in.Status)
				filter = &st
			}
			list, err := d.projects.ListProjects(filter)
			if err != nil {
				return toolErr(err), projectListView{}, nil
			}
			return nil, projectListView{Projects: newProjectViews(list)}, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "get_project", Description: "Get a single project by id."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in projectIDIn) (*mcp.CallToolResult, projectView, error) {
			id, err := uuid.Parse(in.ID)
			if err != nil {
				return toolErr(fmt.Errorf("invalid id: %w", err)), projectView{}, nil
			}
			p, err := d.projects.GetProject(id)
			if err != nil {
				return toolErr(err), projectView{}, nil
			}
			return nil, newProjectView(p), nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "create_idea", Description: "Capture a new project in the idea stage."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in createIdeaIn) (*mcp.CallToolResult, projectView, error) {
			p, err := d.projects.CreateIdea(in.Name, in.Summary, in.Description, in.Tags)
			if err != nil {
				return toolErr(err), projectView{}, nil
			}
			return nil, newProjectView(p), nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "update_project", Description: "Update a project's name, summary, description, or tags. Only provided fields change."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in updateProjectIn) (*mcp.CallToolResult, projectView, error) {
			id, err := uuid.Parse(in.ID)
			if err != nil {
				return toolErr(fmt.Errorf("invalid id: %w", err)), projectView{}, nil
			}
			p, err := d.projects.UpdateDetails(id, in.Name, in.Summary, in.Description, in.Tags)
			if err != nil {
				return toolErr(err), projectView{}, nil
			}
			return nil, newProjectView(p), nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "promote_project", Description: "Promote an idea to active. Mode defaults to developing."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in promoteIn) (*mcp.CallToolResult, projectView, error) {
			id, err := uuid.Parse(in.ID)
			if err != nil {
				return toolErr(fmt.Errorf("invalid id: %w", err)), projectView{}, nil
			}
			p, err := d.projects.PromoteToActive(id, models.ProjectMode(in.Mode))
			if err != nil {
				return toolErr(err), projectView{}, nil
			}
			return nil, newProjectView(p), nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "set_project_mode", Description: "Switch an active project between developing and maintaining."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in setModeIn) (*mcp.CallToolResult, projectView, error) {
			id, err := uuid.Parse(in.ID)
			if err != nil {
				return toolErr(fmt.Errorf("invalid id: %w", err)), projectView{}, nil
			}
			p, err := d.projects.SetMode(id, models.ProjectMode(in.Mode))
			if err != nil {
				return toolErr(err), projectView{}, nil
			}
			return nil, newProjectView(p), nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "archive_project", Description: "Archive a project (deprecated/no longer used), recording an optional reason."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in archiveIn) (*mcp.CallToolResult, projectView, error) {
			id, err := uuid.Parse(in.ID)
			if err != nil {
				return toolErr(fmt.Errorf("invalid id: %w", err)), projectView{}, nil
			}
			p, err := d.projects.Archive(id, in.Reason)
			if err != nil {
				return toolErr(err), projectView{}, nil
			}
			return nil, newProjectView(p), nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "revive_project", Description: "Restore an archived project to active. Mode defaults to maintaining."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in reviveIn) (*mcp.CallToolResult, projectView, error) {
			id, err := uuid.Parse(in.ID)
			if err != nil {
				return toolErr(fmt.Errorf("invalid id: %w", err)), projectView{}, nil
			}
			p, err := d.projects.Revive(id, models.ProjectMode(in.Mode))
			if err != nil {
				return toolErr(err), projectView{}, nil
			}
			return nil, newProjectView(p), nil
		})

	// --- docs ---

	mcp.AddTool(s, &mcp.Tool{Name: "list_documents", Description: "List a project's documents, including their markdown content."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in projectIDIn) (*mcp.CallToolResult, documentListView, error) {
			id, err := uuid.Parse(in.ID)
			if err != nil {
				return toolErr(fmt.Errorf("invalid id: %w", err)), documentListView{}, nil
			}
			list, err := d.docs.ListDocuments(id)
			if err != nil {
				return toolErr(err), documentListView{}, nil
			}
			return nil, documentListView{Documents: newDocumentViews(list)}, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "upsert_document", Description: "Create or replace a project's document of a given type. There is one document per (project, type)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in upsertDocumentIn) (*mcp.CallToolResult, documentView, error) {
			id, err := uuid.Parse(in.ProjectID)
			if err != nil {
				return toolErr(fmt.Errorf("invalid project_id: %w", err)), documentView{}, nil
			}
			doc, err := d.docs.UpsertDocument(id, models.DocumentType(in.Type), in.Title, in.Content, models.DocumentStatus(in.Status))
			if err != nil {
				return toolErr(err), documentView{}, nil
			}
			return nil, newDocumentView(doc), nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "get_missing_docs", Description: "List the required documents a project has not yet completed. Use this to know what to fill in."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in projectIDIn) (*mcp.CallToolResult, missingDocsView, error) {
			id, err := uuid.Parse(in.ID)
			if err != nil {
				return toolErr(fmt.Errorf("invalid id: %w", err)), missingDocsView{}, nil
			}
			missing, err := d.docs.GetMissingDocs(id)
			if err != nil {
				return toolErr(err), missingDocsView{}, nil
			}
			out := make([]missingDocView, len(missing))
			for i, m := range missing {
				out[i] = missingDocView{Type: string(m.Type), Reason: m.Reason}
			}
			return nil, missingDocsView{Missing: out}, nil
		})

	// --- board / tasks ---

	mcp.AddTool(s, &mcp.Tool{Name: "get_board", Description: "Get a project's board and its columns (with ids needed to create/move tasks). Created on first access."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in projectIDIn) (*mcp.CallToolResult, boardView, error) {
			id, err := uuid.Parse(in.ID)
			if err != nil {
				return toolErr(fmt.Errorf("invalid id: %w", err)), boardView{}, nil
			}
			board, err := d.boards.GetBoardForProject(id)
			if err != nil {
				return toolErr(err), boardView{}, nil
			}
			cols, err := d.boards.ListColumns(board.ID)
			if err != nil {
				return toolErr(err), boardView{}, nil
			}
			return nil, newBoardView(board, cols), nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "list_tasks", Description: "List all task cards on a project's board."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in projectIDIn) (*mcp.CallToolResult, taskListView, error) {
			id, err := uuid.Parse(in.ID)
			if err != nil {
				return toolErr(fmt.Errorf("invalid id: %w", err)), taskListView{}, nil
			}
			board, err := d.boards.GetBoardForProject(id)
			if err != nil {
				return toolErr(err), taskListView{}, nil
			}
			tasks, err := d.boards.ListTasks(board.ID)
			if err != nil {
				return toolErr(err), taskListView{}, nil
			}
			return nil, taskListView{Tasks: newTaskViews(tasks)}, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "create_task", Description: "Create a task card in a column on a project's board."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in createTaskIn) (*mcp.CallToolResult, taskView, error) {
			pid, err := uuid.Parse(in.ProjectID)
			if err != nil {
				return toolErr(fmt.Errorf("invalid project_id: %w", err)), taskView{}, nil
			}
			colID, err := uuid.Parse(in.ColumnID)
			if err != nil {
				return toolErr(fmt.Errorf("invalid column_id: %w", err)), taskView{}, nil
			}
			docID, err := optionalUUID(in.DocumentID)
			if err != nil {
				return toolErr(fmt.Errorf("invalid document_id: %w", err)), taskView{}, nil
			}
			board, err := d.boards.GetBoardForProject(pid)
			if err != nil {
				return toolErr(err), taskView{}, nil
			}
			t, err := d.boards.CreateTask(board.ID, colID, in.Title, in.Description, models.TaskPriority(in.Priority), in.Labels, docID)
			if err != nil {
				return toolErr(err), taskView{}, nil
			}
			return nil, newTaskView(t), nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "update_task", Description: "Update a task's title, description, priority, or labels. Only provided fields change."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in updateTaskIn) (*mcp.CallToolResult, taskView, error) {
			id, err := uuid.Parse(in.ID)
			if err != nil {
				return toolErr(fmt.Errorf("invalid id: %w", err)), taskView{}, nil
			}
			var priority *models.TaskPriority
			if in.Priority != nil {
				p := models.TaskPriority(*in.Priority)
				priority = &p
			}
			t, err := d.boards.UpdateTask(id, in.Title, in.Description, priority, in.Labels)
			if err != nil {
				return toolErr(err), taskView{}, nil
			}
			return nil, newTaskView(t), nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "move_task", Description: "Move a task to a column and position. Moving into the Done column marks it completed."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in moveTaskIn) (*mcp.CallToolResult, taskView, error) {
			id, err := uuid.Parse(in.ID)
			if err != nil {
				return toolErr(fmt.Errorf("invalid id: %w", err)), taskView{}, nil
			}
			colID, err := uuid.Parse(in.ColumnID)
			if err != nil {
				return toolErr(fmt.Errorf("invalid column_id: %w", err)), taskView{}, nil
			}
			t, err := d.boards.MoveTask(id, colID, in.Position)
			if err != nil {
				return toolErr(err), taskView{}, nil
			}
			return nil, newTaskView(t), nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "link_task_document", Description: "Link a task to a document (or unlink, by omitting document_id)."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in linkTaskDocumentIn) (*mcp.CallToolResult, taskView, error) {
			id, err := uuid.Parse(in.TaskID)
			if err != nil {
				return toolErr(fmt.Errorf("invalid task_id: %w", err)), taskView{}, nil
			}
			docID, err := optionalUUID(in.DocumentID)
			if err != nil {
				return toolErr(fmt.Errorf("invalid document_id: %w", err)), taskView{}, nil
			}
			t, err := d.boards.LinkDocument(id, docID)
			if err != nil {
				return toolErr(err), taskView{}, nil
			}
			return nil, newTaskView(t), nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "delete_task", Description: "Delete a task card."},
		func(ctx context.Context, _ *mcp.CallToolRequest, in taskIDIn) (*mcp.CallToolResult, okView, error) {
			id, err := uuid.Parse(in.ID)
			if err != nil {
				return toolErr(fmt.Errorf("invalid id: %w", err)), okView{}, nil
			}
			if err := d.boards.DeleteTask(id); err != nil {
				return toolErr(err), okView{}, nil
			}
			return nil, okView{OK: true}, nil
		})
}

// optionalUUID parses an optional id string pointer. nil or "" yields (nil, nil).
func optionalUUID(s *string) (*uuid.UUID, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	id, err := uuid.Parse(*s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
