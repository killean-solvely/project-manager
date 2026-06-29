package server

import (
	"time"

	"github.com/killeanjohnson/projectmanager/internal/docs"
	"github.com/killeanjohnson/projectmanager/internal/models"
)

// DTOs carry the JSON tags and form the API contract. They convert from domain
// models via FromModel. Request bodies live in request_types.go.

type projectDTO struct {
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

func (d *projectDTO) FromModel(m *models.Project) {
	d.ID = m.ID.String()
	d.Name = m.Name
	d.Summary = m.Summary
	d.Description = m.Description
	d.Status = string(m.Status)
	d.Mode = string(m.Mode)
	d.Tags = m.Tags
	if d.Tags == nil {
		d.Tags = []string{}
	}
	d.ArchivedReason = m.ArchivedReason
	d.PromotedAt = m.PromotedAt
	d.ArchivedAt = m.ArchivedAt
	d.CreatedAt = m.CreatedAt
	d.UpdatedAt = m.UpdatedAt
}

func projectDTOs(ms []*models.Project) []projectDTO {
	out := make([]projectDTO, len(ms))
	for i, m := range ms {
		out[i].FromModel(m)
	}
	return out
}

type documentDTO struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (d *documentDTO) FromModel(m *models.Document) {
	d.ID = m.ID.String()
	d.ProjectID = m.ProjectID.String()
	d.Type = string(m.Type)
	d.Title = m.Title
	d.Content = m.Content
	d.Status = string(m.Status)
	d.CreatedAt = m.CreatedAt
	d.UpdatedAt = m.UpdatedAt
}

func documentDTOs(ms []*models.Document) []documentDTO {
	out := make([]documentDTO, len(ms))
	for i, m := range ms {
		out[i].FromModel(m)
	}
	return out
}

type missingDocDTO struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

func missingDocDTOs(missing []docs.MissingDoc) []missingDocDTO {
	out := make([]missingDocDTO, len(missing))
	for i, m := range missing {
		out[i] = missingDocDTO{Type: string(m.Type), Reason: m.Reason}
	}
	return out
}

type columnDTO struct {
	ID       string `json:"id"`
	BoardID  string `json:"board_id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
}

func (d *columnDTO) FromModel(m *models.Column) {
	d.ID = m.ID.String()
	d.BoardID = m.BoardID.String()
	d.Name = m.Name
	d.Position = m.Position
}

type boardDTO struct {
	ID        string      `json:"id"`
	ProjectID string      `json:"project_id"`
	Name      string      `json:"name"`
	Columns   []columnDTO `json:"columns"`
}

func newBoardDTO(board *models.Board, columns []*models.Column) boardDTO {
	d := boardDTO{
		ID:        board.ID.String(),
		ProjectID: board.ProjectID.String(),
		Name:      board.Name,
		Columns:   make([]columnDTO, len(columns)),
	}
	for i, c := range columns {
		d.Columns[i].FromModel(c)
	}
	return d
}

type taskDTO struct {
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

func (d *taskDTO) FromModel(m *models.Task) {
	d.ID = m.ID.String()
	d.BoardID = m.BoardID.String()
	d.ColumnID = m.ColumnID.String()
	d.Title = m.Title
	d.Description = m.Description
	d.Priority = string(m.Priority)
	d.Labels = m.Labels
	if d.Labels == nil {
		d.Labels = []string{}
	}
	if m.DocumentID != nil {
		id := m.DocumentID.String()
		d.DocumentID = &id
	}
	d.Position = m.Position
	d.CompletedAt = m.CompletedAt
	d.CreatedAt = m.CreatedAt
	d.UpdatedAt = m.UpdatedAt
}

func taskDTOs(ms []*models.Task) []taskDTO {
	out := make([]taskDTO, len(ms))
	for i, m := range ms {
		out[i].FromModel(m)
	}
	return out
}
