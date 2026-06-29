package sqlite_test

import (
	"database/sql"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/killeanjohnson/projectmanager/internal/models"
	"github.com/killeanjohnson/projectmanager/internal/persistence/sqlite"

	"github.com/google/uuid"
)

func newDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sqlite.OpenAt(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestProjectRoundTripAndList(t *testing.T) {
	db := newDB(t)
	repo := sqlite.NewProjectsRepository(db)

	now := time.Now()
	promoted := now.Add(time.Minute)
	active := &models.Project{
		ID: uuid.New(), Name: "Active one", Summary: "s", Description: "d",
		Status: models.ProjectStatusActive, Mode: models.ProjectModeDeveloping,
		Tags: []string{"a", "b"}, PromotedAt: &promoted,
		CreatedAt: now, UpdatedAt: now,
	}
	idea := &models.Project{
		ID: uuid.New(), Name: "Idea one", Status: models.ProjectStatusIdea,
		Tags: []string{}, CreatedAt: now.Add(-time.Hour), UpdatedAt: now,
	}
	for _, p := range []*models.Project{active, idea} {
		if err := repo.Save(p); err != nil {
			t.Fatalf("save: %v", err)
		}
	}

	got, err := repo.FindByID(active.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.Name != "Active one" || got.Status != models.ProjectStatusActive || got.Mode != models.ProjectModeDeveloping {
		t.Fatalf("scalar fields wrong: %+v", got)
	}
	if !slices.Equal(got.Tags, []string{"a", "b"}) {
		t.Fatalf("tags round-trip wrong: %v", got.Tags)
	}
	if got.PromotedAt == nil || !got.PromotedAt.Equal(promoted) {
		t.Fatalf("promoted_at round-trip wrong: %v", got.PromotedAt)
	}
	if got.ArchivedAt != nil {
		t.Fatalf("archived_at should be nil, got %v", got.ArchivedAt)
	}

	// FindAll is ordered by created_at (idea is older, so first).
	all, err := repo.FindAll()
	if err != nil {
		t.Fatalf("find all: %v", err)
	}
	if len(all) != 2 || all[0].ID != idea.ID || all[1].ID != active.ID {
		t.Fatalf("unexpected order/count: %+v", all)
	}

	if err := repo.Delete(idea.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.FindByID(idea.ID); err == nil {
		t.Fatalf("expected not-found after delete")
	}
}

func TestDocumentByTypeAbsence(t *testing.T) {
	db := newDB(t)
	projects := sqlite.NewProjectsRepository(db)
	documents := sqlite.NewDocumentsRepository(db)

	pid := uuid.New()
	mustSaveProject(t, projects, pid)

	// Absent type yields (nil, nil), not an error.
	got, err := documents.FindByProjectIDAndType(pid, models.DocumentTypeSpec)
	if err != nil || got != nil {
		t.Fatalf("expected (nil,nil) for absent doc, got (%v,%v)", got, err)
	}

	now := time.Now()
	doc := &models.Document{
		ID: uuid.New(), ProjectID: pid, Type: models.DocumentTypeSpec,
		Title: "Spec", Content: "# spec", Status: models.DocumentStatusComplete,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := documents.Save(doc); err != nil {
		t.Fatalf("save doc: %v", err)
	}
	got, err = documents.FindByProjectIDAndType(pid, models.DocumentTypeSpec)
	if err != nil || got == nil || got.Content != "# spec" {
		t.Fatalf("expected the saved doc, got (%v,%v)", got, err)
	}
	list, err := documents.FindByProjectID(pid)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 doc for project, got %d (%v)", len(list), err)
	}
}

func TestBoardAbsenceAndTaskRoundTrip(t *testing.T) {
	db := newDB(t)
	projects := sqlite.NewProjectsRepository(db)
	documents := sqlite.NewDocumentsRepository(db)
	boards := sqlite.NewBoardsRepository(db)
	columns := sqlite.NewColumnsRepository(db)
	tasks := sqlite.NewTasksRepository(db)

	pid := uuid.New()
	mustSaveProject(t, projects, pid)

	if b, err := boards.FindByProjectID(pid); err != nil || b != nil {
		t.Fatalf("expected (nil,nil) for absent board, got (%v,%v)", b, err)
	}

	now := time.Now()
	board := &models.Board{ID: uuid.New(), ProjectID: pid, Name: "Board", CreatedAt: now, UpdatedAt: now}
	if err := boards.Save(board); err != nil {
		t.Fatalf("save board: %v", err)
	}

	// Columns come back ordered by position regardless of insert order.
	for _, pos := range []int{2, 0, 1} {
		col := &models.Column{ID: uuid.New(), BoardID: board.ID, Name: "c", Position: pos, CreatedAt: now, UpdatedAt: now}
		if err := columns.Save(col); err != nil {
			t.Fatalf("save column: %v", err)
		}
	}
	cols, err := columns.FindByBoardID(board.ID)
	if err != nil || len(cols) != 3 || cols[0].Position != 0 || cols[2].Position != 2 {
		t.Fatalf("columns not ordered by position: %+v (%v)", cols, err)
	}

	// Task with a document link and completion time round-trips. The linked
	// document must exist (the document_id foreign key enforces this).
	docID := uuid.New()
	if err := documents.Save(&models.Document{
		ID: docID, ProjectID: pid, Type: models.DocumentTypeSpec, Title: "Spec",
		Status: models.DocumentStatusDraft, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("save doc: %v", err)
	}
	completed := now.Add(time.Hour)
	task := &models.Task{
		ID: uuid.New(), BoardID: board.ID, ColumnID: cols[0].ID, Title: "T",
		Priority: models.TaskPriorityHigh, Labels: []string{"x"},
		DocumentID: &docID, Position: 0, CompletedAt: &completed,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := tasks.Save(task); err != nil {
		t.Fatalf("save task: %v", err)
	}
	got, err := tasks.FindByID(task.ID)
	if err != nil {
		t.Fatalf("find task: %v", err)
	}
	if got.DocumentID == nil || *got.DocumentID != docID {
		t.Fatalf("document_id round-trip wrong: %v", got.DocumentID)
	}
	if got.CompletedAt == nil || !got.CompletedAt.Equal(completed) {
		t.Fatalf("completed_at round-trip wrong: %v", got.CompletedAt)
	}
	if !slices.Equal(got.Labels, []string{"x"}) || got.Priority != models.TaskPriorityHigh {
		t.Fatalf("labels/priority round-trip wrong: %+v", got)
	}
}

// TestPersistenceAcrossReopen proves data survives closing and reopening the DB —
// i.e. what cmd/api and cmd/mcp (separate processes) rely on to share state.
func TestPersistenceAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "persist.db")

	db1, err := sqlite.OpenAt(path)
	if err != nil {
		t.Fatalf("open 1: %v", err)
	}
	pid := uuid.New()
	mustSaveProject(t, sqlite.NewProjectsRepository(db1), pid)
	if err := db1.Close(); err != nil {
		t.Fatalf("close 1: %v", err)
	}

	db2, err := sqlite.OpenAt(path)
	if err != nil {
		t.Fatalf("open 2: %v", err)
	}
	defer db2.Close()
	got, err := sqlite.NewProjectsRepository(db2).FindByID(pid)
	if err != nil || got == nil {
		t.Fatalf("project did not persist across reopen: (%v, %v)", got, err)
	}
}

func mustSaveProject(t *testing.T, repo *sqlite.ProjectsRepository, id uuid.UUID) {
	t.Helper()
	now := time.Now()
	p := &models.Project{
		ID: id, Name: "P", Status: models.ProjectStatusIdea,
		Tags: []string{}, CreatedAt: now, UpdatedAt: now,
	}
	if err := repo.Save(p); err != nil {
		t.Fatalf("save project: %v", err)
	}
}
