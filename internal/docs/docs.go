package docs

import (
	"fmt"
	"time"

	"github.com/killeanjohnson/projectmanager/internal/models"
	"github.com/killeanjohnson/projectmanager/internal/persistence"

	"github.com/google/uuid"
)

// DocTemplateEntry describes one doc type a project is expected to carry and
// whether it is required for the project to count as "complete".
type DocTemplateEntry struct {
	Type     models.DocumentType
	Required bool
}

// defaultTemplate is the v1 global template that drives the completeness checklist.
// Per-project overrides are a later addition.
var defaultTemplate = []DocTemplateEntry{
	{Type: models.DocumentTypeOverview, Required: true},
	{Type: models.DocumentTypeTechnical, Required: true},
	{Type: models.DocumentTypeSpec, Required: true},
	{Type: models.DocumentTypeAPI, Required: false},
	{Type: models.DocumentTypeRunbook, Required: false},
}

// MissingDoc is a required doc type a project has not yet completed.
type MissingDoc struct {
	Type   models.DocumentType
	Reason string // "missing" (no document) or "incomplete" (exists but not complete)
}

// Service manages a project's docs library and the completeness checklist.
type Service struct {
	documents persistence.DocumentsRepository
}

func NewService(documentsRepo persistence.DocumentsRepository) *Service {
	return &Service{documents: documentsRepo}
}

// DocTemplate returns the expected set of doc types.
func (s *Service) DocTemplate() []DocTemplateEntry {
	out := make([]DocTemplateEntry, len(defaultTemplate))
	copy(out, defaultTemplate)
	return out
}

func (s *Service) ListDocuments(projectID uuid.UUID) ([]*models.Document, error) {
	return s.documents.FindByProjectID(projectID)
}

func (s *Service) GetDocument(id uuid.UUID) (*models.Document, error) {
	return s.documents.FindByID(id)
}

// UpsertDocument creates or replaces the document of a given type for a project.
// There is at most one document per (project, type).
func (s *Service) UpsertDocument(projectID uuid.UUID, docType models.DocumentType, title, content string, status models.DocumentStatus) (*models.Document, error) {
	if !validType(docType) {
		return nil, fmt.Errorf("invalid document type %q", docType)
	}
	if status == "" {
		status = models.DocumentStatusDraft
	}
	if !validStatus(status) {
		return nil, fmt.Errorf("invalid document status %q", status)
	}

	existing, err := s.documents.FindByProjectIDAndType(projectID, docType)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if existing != nil {
		existing.Title = title
		existing.Content = content
		existing.Status = status
		existing.UpdatedAt = now
		if err := s.documents.Save(existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	doc := &models.Document{
		ID:        uuid.New(),
		ProjectID: projectID,
		Type:      docType,
		Title:     title,
		Content:   content,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.documents.Save(doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *Service) DeleteDocument(id uuid.UUID) error {
	if _, err := s.documents.FindByID(id); err != nil {
		return err
	}
	return s.documents.Delete(id)
}

// GetMissingDocs reports which required doc types are absent or not yet complete.
// This is the surface an AI uses to "fill in the missing details".
func (s *Service) GetMissingDocs(projectID uuid.UUID) ([]MissingDoc, error) {
	existing, err := s.documents.FindByProjectID(projectID)
	if err != nil {
		return nil, err
	}
	byType := make(map[models.DocumentType]*models.Document, len(existing))
	for _, d := range existing {
		byType[d.Type] = d
	}

	missing := make([]MissingDoc, 0)
	for _, entry := range defaultTemplate {
		if !entry.Required {
			continue
		}
		doc, ok := byType[entry.Type]
		switch {
		case !ok:
			missing = append(missing, MissingDoc{Type: entry.Type, Reason: "missing"})
		case doc.Status != models.DocumentStatusComplete:
			missing = append(missing, MissingDoc{Type: entry.Type, Reason: "incomplete"})
		}
	}
	return missing, nil
}

func validType(t models.DocumentType) bool {
	switch t {
	case models.DocumentTypeOverview, models.DocumentTypeTechnical, models.DocumentTypeSpec,
		models.DocumentTypeAPI, models.DocumentTypeRunbook, models.DocumentTypeOther:
		return true
	}
	return false
}

func validStatus(s models.DocumentStatus) bool {
	switch s {
	case models.DocumentStatusDraft, models.DocumentStatusInReview, models.DocumentStatusComplete:
		return true
	}
	return false
}
