package projects_test

import (
	"testing"

	"github.com/killeanjohnson/projectmanager/internal/models"
	"github.com/killeanjohnson/projectmanager/internal/persistence/memory"
	"github.com/killeanjohnson/projectmanager/internal/projects"
)

func newService() *projects.Service {
	return projects.NewService(memory.NewProjectsRepository())
}

func TestLifecycleTransitions(t *testing.T) {
	s := newService()

	p, err := s.CreateIdea("Test", "", "", nil)
	if err != nil {
		t.Fatalf("create idea: %v", err)
	}
	if p.Status != models.ProjectStatusIdea {
		t.Fatalf("want status idea, got %q", p.Status)
	}
	if p.Mode != "" {
		t.Fatalf("an idea should have no mode, got %q", p.Mode)
	}

	p, err = s.PromoteToActive(p.ID, "")
	if err != nil {
		t.Fatalf("promote: %v", err)
	}
	if p.Status != models.ProjectStatusActive {
		t.Fatalf("want status active, got %q", p.Status)
	}
	if p.Mode != models.ProjectModeDeveloping {
		t.Fatalf("promote should default to developing, got %q", p.Mode)
	}
	if p.PromotedAt == nil {
		t.Fatalf("promotedAt should be set")
	}

	if _, err := s.PromoteToActive(p.ID, ""); err == nil {
		t.Fatalf("promoting an already-active project should fail")
	}

	p, err = s.SetMode(p.ID, models.ProjectModeMaintaining)
	if err != nil {
		t.Fatalf("set mode: %v", err)
	}
	if p.Mode != models.ProjectModeMaintaining {
		t.Fatalf("want mode maintaining, got %q", p.Mode)
	}

	p, err = s.Archive(p.ID, "shipped")
	if err != nil {
		t.Fatalf("archive: %v", err)
	}
	if p.Status != models.ProjectStatusArchived {
		t.Fatalf("want status archived, got %q", p.Status)
	}
	if p.Mode != "" {
		t.Fatalf("archiving should clear mode, got %q", p.Mode)
	}
	if p.ArchivedReason != "shipped" {
		t.Fatalf("archive reason not recorded, got %q", p.ArchivedReason)
	}

	if _, err := s.SetMode(p.ID, models.ProjectModeDeveloping); err == nil {
		t.Fatalf("setting mode on an archived project should fail")
	}

	p, err = s.Revive(p.ID, "")
	if err != nil {
		t.Fatalf("revive: %v", err)
	}
	if p.Status != models.ProjectStatusActive {
		t.Fatalf("want status active after revive, got %q", p.Status)
	}
	if p.Mode != models.ProjectModeMaintaining {
		t.Fatalf("revive should default to maintaining, got %q", p.Mode)
	}
	if p.ArchivedReason != "" {
		t.Fatalf("revive should clear archived reason, got %q", p.ArchivedReason)
	}
}

func TestCreateIdeaRequiresName(t *testing.T) {
	s := newService()
	if _, err := s.CreateIdea("", "", "", nil); err == nil {
		t.Fatalf("expected an error for an empty name")
	}
}

func TestListProjectsByStatus(t *testing.T) {
	s := newService()
	a, _ := s.CreateIdea("a", "", "", nil)
	_, _ = s.CreateIdea("b", "", "", nil)
	if _, err := s.PromoteToActive(a.ID, ""); err != nil {
		t.Fatalf("promote: %v", err)
	}

	idea := models.ProjectStatusIdea
	ideas, _ := s.ListProjects(&idea)
	if len(ideas) != 1 {
		t.Fatalf("want 1 idea, got %d", len(ideas))
	}

	active := models.ProjectStatusActive
	actives, _ := s.ListProjects(&active)
	if len(actives) != 1 {
		t.Fatalf("want 1 active, got %d", len(actives))
	}

	all, _ := s.ListProjects(nil)
	if len(all) != 2 {
		t.Fatalf("want 2 projects total, got %d", len(all))
	}
}
