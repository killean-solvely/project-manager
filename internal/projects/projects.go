package projects

import (
	"errors"
	"fmt"
	"time"

	"github.com/killeanjohnson/projectmanager/internal/models"
	"github.com/killeanjohnson/projectmanager/internal/persistence"

	"github.com/google/uuid"
)

// Service holds the project lifecycle logic. It enforces the state machine and
// the invariants: Mode is set iff a project is active, ArchivedReason is set iff
// a project is archived.
type Service struct {
	projects persistence.ProjectsRepository
}

func NewService(projectsRepo persistence.ProjectsRepository) *Service {
	return &Service{projects: projectsRepo}
}

// CreateIdea captures a new project in the idea stage.
func (s *Service) CreateIdea(name, summary, description string, tags []string) (*models.Project, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	if tags == nil {
		tags = []string{}
	}
	now := time.Now()
	p := &models.Project{
		ID:          uuid.New(),
		Name:        name,
		Summary:     summary,
		Description: description,
		Status:      models.ProjectStatusIdea,
		Tags:        tags,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.projects.Save(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) GetProject(id uuid.UUID) (*models.Project, error) {
	return s.projects.FindByID(id)
}

// ListProjects returns all projects, optionally filtered by lifecycle status.
func (s *Service) ListProjects(status *models.ProjectStatus) ([]*models.Project, error) {
	all, err := s.projects.FindAll()
	if err != nil {
		return nil, err
	}
	if status == nil {
		return all, nil
	}
	out := make([]*models.Project, 0, len(all))
	for _, p := range all {
		if p.Status == *status {
			out = append(out, p)
		}
	}
	return out, nil
}

// UpdateDetails applies a partial update to the descriptive fields. Lifecycle
// changes go through the dedicated transition methods instead.
func (s *Service) UpdateDetails(id uuid.UUID, name, summary, description *string, tags *[]string) (*models.Project, error) {
	p, err := s.projects.FindByID(id)
	if err != nil {
		return nil, err
	}
	if name != nil {
		if *name == "" {
			return nil, errors.New("name cannot be empty")
		}
		p.Name = *name
	}
	if summary != nil {
		p.Summary = *summary
	}
	if description != nil {
		p.Description = *description
	}
	if tags != nil {
		p.Tags = *tags
	}
	p.UpdatedAt = time.Now()
	if err := s.projects.Save(p); err != nil {
		return nil, err
	}
	return p, nil
}

// PromoteToActive moves an idea into the active stage. Mode defaults to developing.
func (s *Service) PromoteToActive(id uuid.UUID, mode models.ProjectMode) (*models.Project, error) {
	p, err := s.projects.FindByID(id)
	if err != nil {
		return nil, err
	}
	if p.Status != models.ProjectStatusIdea {
		return nil, fmt.Errorf("cannot promote a project with status %q; only ideas can be promoted", p.Status)
	}
	if mode == "" {
		mode = models.ProjectModeDeveloping
	}
	if !validMode(mode) {
		return nil, fmt.Errorf("invalid mode %q", mode)
	}
	now := time.Now()
	p.Status = models.ProjectStatusActive
	p.Mode = mode
	p.PromotedAt = &now
	p.ArchivedReason = ""
	p.ArchivedAt = nil
	p.UpdatedAt = now
	if err := s.projects.Save(p); err != nil {
		return nil, err
	}
	return p, nil
}

// SetMode switches an active project between developing and maintaining.
func (s *Service) SetMode(id uuid.UUID, mode models.ProjectMode) (*models.Project, error) {
	p, err := s.projects.FindByID(id)
	if err != nil {
		return nil, err
	}
	if p.Status != models.ProjectStatusActive {
		return nil, fmt.Errorf("cannot set mode on a project with status %q; only active projects have a mode", p.Status)
	}
	if !validMode(mode) {
		return nil, fmt.Errorf("invalid mode %q", mode)
	}
	p.Mode = mode
	p.UpdatedAt = time.Now()
	if err := s.projects.Save(p); err != nil {
		return nil, err
	}
	return p, nil
}

// Archive retires an active project, recording why.
func (s *Service) Archive(id uuid.UUID, reason string) (*models.Project, error) {
	p, err := s.projects.FindByID(id)
	if err != nil {
		return nil, err
	}
	if p.Status == models.ProjectStatusArchived {
		return nil, errors.New("project is already archived")
	}
	return s.archive(p, reason)
}

// DropIdea archives an idea that won't be built.
func (s *Service) DropIdea(id uuid.UUID) (*models.Project, error) {
	p, err := s.projects.FindByID(id)
	if err != nil {
		return nil, err
	}
	if p.Status != models.ProjectStatusIdea {
		return nil, fmt.Errorf("cannot drop a project with status %q; only ideas can be dropped", p.Status)
	}
	return s.archive(p, "dropped")
}

func (s *Service) archive(p *models.Project, reason string) (*models.Project, error) {
	now := time.Now()
	p.Status = models.ProjectStatusArchived
	p.ArchivedReason = reason
	p.ArchivedAt = &now
	p.Mode = ""
	p.UpdatedAt = now
	if err := s.projects.Save(p); err != nil {
		return nil, err
	}
	return p, nil
}

// Revive restores an archived project to active. Mode defaults to maintaining.
func (s *Service) Revive(id uuid.UUID, mode models.ProjectMode) (*models.Project, error) {
	p, err := s.projects.FindByID(id)
	if err != nil {
		return nil, err
	}
	if p.Status != models.ProjectStatusArchived {
		return nil, fmt.Errorf("cannot revive a project with status %q; only archived projects can be revived", p.Status)
	}
	if mode == "" {
		mode = models.ProjectModeMaintaining
	}
	if !validMode(mode) {
		return nil, fmt.Errorf("invalid mode %q", mode)
	}
	now := time.Now()
	p.Status = models.ProjectStatusActive
	p.Mode = mode
	p.ArchivedReason = ""
	p.ArchivedAt = nil
	p.UpdatedAt = now
	if err := s.projects.Save(p); err != nil {
		return nil, err
	}
	return p, nil
}

func validMode(m models.ProjectMode) bool {
	return m == models.ProjectModeDeveloping || m == models.ProjectModeMaintaining
}
