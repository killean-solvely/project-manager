package models

import (
	"time"

	"github.com/google/uuid"
)

// DocumentType is the kind of documentation a project carries. The set of types a
// "complete" project is expected to have is defined by the doc template (see the
// docs service) and drives the completeness checklist.
type DocumentType string

const (
	DocumentTypeOverview  DocumentType = "overview"
	DocumentTypeTechnical DocumentType = "technical"
	DocumentTypeSpec      DocumentType = "spec"
	DocumentTypeAPI       DocumentType = "api"
	DocumentTypeRunbook   DocumentType = "runbook"
	DocumentTypeOther     DocumentType = "other"
)

// DocumentStatus tracks how finished a document is. There is no "missing" status:
// a missing document is simply the absence of a Document row for an expected type.
type DocumentStatus string

const (
	DocumentStatusDraft    DocumentStatus = "draft"
	DocumentStatusInReview DocumentStatus = "in_review"
	DocumentStatusComplete DocumentStatus = "complete"
)

// Document is a first-class artifact in a project's docs library. Task cards may
// link to a Document, but documents live independently of the board.
type Document struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Type      DocumentType
	Title     string
	Content   string // markdown
	Status    DocumentStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}
