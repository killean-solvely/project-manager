package mcpserver

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/killeanjohnson/projectmanager/internal/models"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const projectURIPrefix = "pm://project/"

func (d *deps) registerResources(s *mcp.Server) {
	s.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "project",
		Title:       "Project",
		Description: "A project's summary and lifecycle state, as JSON.",
		MIMEType:    "application/json",
		URITemplate: "pm://project/{id}",
	}, d.readProject)

	s.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "project-doc",
		Title:       "Project document",
		Description: "A single project document's markdown content.",
		MIMEType:    "text/markdown",
		URITemplate: "pm://project/{id}/doc/{type}",
	}, d.readProjectDoc)

	s.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "project-board",
		Title:       "Project board",
		Description: "A project's board and columns, as JSON.",
		MIMEType:    "application/json",
		URITemplate: "pm://project/{id}/board",
	}, d.readProjectBoard)
}

func (d *deps) readProject(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	uri := req.Params.URI
	id, err := uuid.Parse(strings.TrimPrefix(uri, projectURIPrefix))
	if err != nil {
		return nil, mcp.ResourceNotFoundError(uri)
	}
	p, err := d.projects.GetProject(id)
	if err != nil {
		return nil, mcp.ResourceNotFoundError(uri)
	}
	return jsonResource(uri, newProjectView(p))
}

func (d *deps) readProjectDoc(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	uri := req.Params.URI
	parts := strings.Split(strings.TrimPrefix(uri, projectURIPrefix), "/")
	if len(parts) != 3 || parts[1] != "doc" {
		return nil, mcp.ResourceNotFoundError(uri)
	}
	id, err := uuid.Parse(parts[0])
	if err != nil {
		return nil, mcp.ResourceNotFoundError(uri)
	}
	docType := models.DocumentType(parts[2])
	list, err := d.docs.ListDocuments(id)
	if err != nil {
		return nil, mcp.ResourceNotFoundError(uri)
	}
	for _, doc := range list {
		if doc.Type == docType {
			return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{
				URI:      uri,
				MIMEType: "text/markdown",
				Text:     doc.Content,
			}}}, nil
		}
	}
	return nil, mcp.ResourceNotFoundError(uri)
}

func (d *deps) readProjectBoard(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	uri := req.Params.URI
	parts := strings.Split(strings.TrimPrefix(uri, projectURIPrefix), "/")
	if len(parts) != 2 || parts[1] != "board" {
		return nil, mcp.ResourceNotFoundError(uri)
	}
	id, err := uuid.Parse(parts[0])
	if err != nil {
		return nil, mcp.ResourceNotFoundError(uri)
	}
	board, err := d.boards.GetBoardForProject(id)
	if err != nil {
		return nil, mcp.ResourceNotFoundError(uri)
	}
	cols, err := d.boards.ListColumns(board.ID)
	if err != nil {
		return nil, mcp.ResourceNotFoundError(uri)
	}
	return jsonResource(uri, newBoardView(board, cols))
}

func jsonResource(uri string, v any) (*mcp.ReadResourceResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{
		URI:      uri,
		MIMEType: "application/json",
		Text:     string(data),
	}}}, nil
}
