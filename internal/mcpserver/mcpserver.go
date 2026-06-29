// Package mcpserver is the MCP adapter over the service layer. Like the HTTP
// server, it holds no business logic — every tool, resource, and prompt calls
// straight into the projects/docs/boards services.
package mcpserver

import (
	"time"

	"github.com/killeanjohnson/projectmanager/internal/boards"
	"github.com/killeanjohnson/projectmanager/internal/docs"
	"github.com/killeanjohnson/projectmanager/internal/models"
	"github.com/killeanjohnson/projectmanager/internal/projects"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// deps bundles the services the MCP handlers call into.
type deps struct {
	projects *projects.Service
	docs     *docs.Service
	boards   *boards.Service
}

// New builds an MCP server exposing the project manager over tools, resources,
// and prompts. Run it with server.Run(ctx, &mcp.StdioTransport{}).
func New(projectsSvc *projects.Service, docsSvc *docs.Service, boardsSvc *boards.Service) *mcp.Server {
	d := &deps{projects: projectsSvc, docs: docsSvc, boards: boardsSvc}

	s := mcp.NewServer(&mcp.Implementation{
		Name:    "projectmanager",
		Title:   "Project Manager",
		Version: "0.1.0",
	}, nil)

	d.registerTools(s)
	d.registerResources(s)
	d.registerPrompts(s)
	return s
}

// toolErr reports a business error as a tool-level error result (IsError) rather
// than a transport error, so the model sees and can react to it.
func toolErr(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
	}
}

func valueOr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

// --- output views (the JSON the tools/resources return) ---

type projectView struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Summary        string     `json:"summary"`
	Description    string     `json:"description"`
	Status         string     `json:"status"`
	Mode           string     `json:"mode,omitempty"`
	Tags           []string   `json:"tags"`
	ArchivedReason string     `json:"archived_reason,omitempty"`
	PromotedAt     *time.Time `json:"promoted_at,omitempty"`
	ArchivedAt     *time.Time `json:"archived_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func newProjectView(m *models.Project) projectView {
	v := projectView{
		ID:             m.ID.String(),
		Name:           m.Name,
		Summary:        m.Summary,
		Description:    m.Description,
		Status:         string(m.Status),
		Mode:           string(m.Mode),
		Tags:           m.Tags,
		ArchivedReason: m.ArchivedReason,
		PromotedAt:     m.PromotedAt,
		ArchivedAt:     m.ArchivedAt,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
	if v.Tags == nil {
		v.Tags = []string{}
	}
	return v
}

func newProjectViews(ms []*models.Project) []projectView {
	out := make([]projectView, len(ms))
	for i, m := range ms {
		out[i] = newProjectView(m)
	}
	return out
}

type documentView struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func newDocumentView(m *models.Document) documentView {
	return documentView{
		ID:        m.ID.String(),
		ProjectID: m.ProjectID.String(),
		Type:      string(m.Type),
		Title:     m.Title,
		Content:   m.Content,
		Status:    string(m.Status),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func newDocumentViews(ms []*models.Document) []documentView {
	out := make([]documentView, len(ms))
	for i, m := range ms {
		out[i] = newDocumentView(m)
	}
	return out
}

type missingDocView struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

type columnView struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
}

type boardView struct {
	ID        string       `json:"id"`
	ProjectID string       `json:"project_id"`
	Name      string       `json:"name"`
	Columns   []columnView `json:"columns"`
}

func newBoardView(b *models.Board, columns []*models.Column) boardView {
	v := boardView{
		ID:        b.ID.String(),
		ProjectID: b.ProjectID.String(),
		Name:      b.Name,
		Columns:   make([]columnView, len(columns)),
	}
	for i, c := range columns {
		v.Columns[i] = columnView{ID: c.ID.String(), Name: c.Name, Position: c.Position}
	}
	return v
}

type taskView struct {
	ID          string     `json:"id"`
	BoardID     string     `json:"board_id"`
	ColumnID    string     `json:"column_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	Labels      []string   `json:"labels"`
	DocumentID  *string    `json:"document_id,omitempty"`
	Position    int        `json:"position"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func newTaskView(m *models.Task) taskView {
	v := taskView{
		ID:          m.ID.String(),
		BoardID:     m.BoardID.String(),
		ColumnID:    m.ColumnID.String(),
		Title:       m.Title,
		Description: m.Description,
		Priority:    string(m.Priority),
		Labels:      m.Labels,
		Position:    m.Position,
		CompletedAt: m.CompletedAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	if v.Labels == nil {
		v.Labels = []string{}
	}
	if m.DocumentID != nil {
		id := m.DocumentID.String()
		v.DocumentID = &id
	}
	return v
}

func newTaskViews(ms []*models.Task) []taskView {
	out := make([]taskView, len(ms))
	for i, m := range ms {
		out[i] = newTaskView(m)
	}
	return out
}

// list/result wrappers — structured tool output must marshal to a JSON object.
type projectListView struct {
	Projects []projectView `json:"projects"`
}

type documentListView struct {
	Documents []documentView `json:"documents"`
}

type missingDocsView struct {
	Missing []missingDocView `json:"missing"`
}

type taskListView struct {
	Tasks []taskView `json:"tasks"`
}

type okView struct {
	OK bool `json:"ok"`
}
