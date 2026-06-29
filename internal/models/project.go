package models

import (
	"time"

	"github.com/google/uuid"
)

// ProjectStatus is the lifecycle stage of a project. Projects don't burn down to
// "done" like kanban cards — they mature: idea -> active -> archived.
type ProjectStatus string

const (
	ProjectStatusIdea     ProjectStatus = "idea"
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusArchived ProjectStatus = "archived"
)

// ProjectMode distinguishes the kind of work happening on an active project.
// It is only meaningful while Status == ProjectStatusActive (empty otherwise).
type ProjectMode string

const (
	ProjectModeDeveloping  ProjectMode = "developing"
	ProjectModeMaintaining ProjectMode = "maintaining"
)

// Project is the top-level lifecycle entity. An "idea" is simply a Project with
// Status == ProjectStatusIdea — graduating an idea flips its status, it does not
// create a new entity.
type Project struct {
	ID             uuid.UUID
	Name           string
	Summary        string
	Description    string // markdown braindump; grows as the idea is curated
	Status         ProjectStatus
	Mode           ProjectMode // set iff Status == ProjectStatusActive
	Tags           []string
	ArchivedReason string // set iff Status == ProjectStatusArchived
	PromotedAt     *time.Time
	ArchivedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
