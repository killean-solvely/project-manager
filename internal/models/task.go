package models

import (
	"time"

	"github.com/google/uuid"
)

// TaskPriority is an enum-like priority for a task.
type TaskPriority string

const (
	TaskPriorityNone   TaskPriority = "none"
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
)

// Task is a card/ticket on a board. Its ColumnID determines where it sits (and
// thus its status); Position orders it within that column. A task may optionally
// link to a Document it is about (e.g. "write the API spec").
type Task struct {
	ID          uuid.UUID
	BoardID     uuid.UUID
	ColumnID    uuid.UUID
	Title       string
	Description string // markdown
	Priority    TaskPriority
	Labels      []string
	DocumentID  *uuid.UUID // optional link into the docs library
	Position    int
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
