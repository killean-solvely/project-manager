package models

import (
	"time"

	"github.com/google/uuid"
)

// Board is a project's kanban board. v1 has exactly one board per project; the
// schema permits more later.
type Board struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Column is an ordered lane on a board (e.g. Backlog, Todo, In progress, Done).
// A column is the source of truth for a task's status.
type Column struct {
	ID        uuid.UUID
	BoardID   uuid.UUID
	Name      string
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}
